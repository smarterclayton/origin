package graph

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/glog"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	boltlog "github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	"k8s.io/client-go/pkg/util/sets"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/healthz"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	apigraph "github.com/openshift/origin/pkg/api/graph"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploygraph "github.com/openshift/origin/pkg/deploy/graph"
	deploynodes "github.com/openshift/origin/pkg/deploy/graph/nodes"
	"github.com/openshift/origin/pkg/util/proc"
)

const (
	graphLong    = ""
	graphExample = ""
)

// NewCmdGraph creates the graph command.
func NewCmdGraph(fullName string, f *clientcmd.Factory, out, errOut io.Writer) *cobra.Command {
	options := &GraphOptions{
		baseCommandName: fullName,
		include:         []string{"deploymentconfigs"},
		defaultExcludes: []unversioned.GroupResource{
			{Resource: "appliedclusterresourcequotas"},
			{Resource: "bindings"},
			{Resource: "deploymentconfigrollbacks"},
			{Resource: "events"},
			{Resource: "imagestreamimages"}, {Resource: "imagestreamtags"}, {Resource: "imagestreammappings"}, {Resource: "imagestreamimports"},
			{Resource: "projectrequests"}, {Resource: "projects"},
			{Resource: "componentstatuses"},
			{Resource: "clusterrolebindings"}, {Resource: "rolebindings"},
			{Resource: "clusterroles"}, {Resource: "roles"},
			{Resource: "resourceaccessreviews"}, {Resource: "localresourceaccessreviews"}, {Resource: "subjectaccessreviews"},
			{Resource: "selfsubjectrulesreviews"}, {Resource: "localsubjectaccessreviews"},
			{Resource: "replicationcontrollerdummies.extensions"},
			{Resource: "podtemplates"},
			{Resource: "useridentitymappings"},
		},
		// Resources known to share the same storage
		overlappingResources: []sets.String{
			sets.NewString("horizontalpodautoscalers.autoscaling", "horizontalpodautoscalers.extensions"),
			sets.NewString("jobs.batch", "jobs.extensions"),
		},
	}

	cmd := &cobra.Command{
		Use:     "graph RESOURCE [-- COMMAND ...]",
		Short:   "Maintain a representation of the state of a server in a graph database",
		Long:    fmt.Sprintf(graphLong, fullName),
		Example: fmt.Sprintf(graphExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args, out, errOut); err != nil {
				cmdutil.CheckErr(err)
			}
			if err := options.Validate(args); err != nil {
				cmdutil.CheckErr(cmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.Run(); err != nil {
				cmdutil.CheckErr(err)
			}
		},
	}

	// flags controlling what to select
	cmd.Flags().BoolVar(&options.allNamespaces, "all-namespaces", true, "If true, list the requested object(s) across all projects. Project in current context is ignored.")

	// control graph program behavior
	cmd.Flags().StringVar(&options.listenAddr, "listen-addr", options.listenAddr, "The name of an interface to listen on to expose metrics and health checking.")
	cmd.Flags().DurationVar(&options.resyncPeriod, "resync-period", 0, "When non-zero, periodically reprocess every item from the server as a Sync event. Use to ensure external systems are kept up to date. Requires --names")

	return cmd
}

type GraphOptions struct {
	out, errOut io.Writer
	debugOut    io.Writer

	include              []string
	overlappingResources []sets.String
	defaultExcludes      []unversioned.GroupResource

	client           resource.RESTClient
	mapping          *meta.RESTMapping
	includeNamespace bool

	// which resources to select
	namespace     string
	allNamespaces bool

	// additional debugging information
	listenAddr string

	// when to exit or reprocess the list of items
	resyncPeriod time.Duration

	baseCommandName string
}

func (o *GraphOptions) Complete(f *clientcmd.Factory, cmd *cobra.Command, args []string, out, errOut io.Writer) error {
	var err error

	if len(args) > 0 {
		o.include = args
	}

	switch len(o.include) {
	case 0:
		return fmt.Errorf("you must specify at least one argument containing the resource to graph")
	}

	oclient, _, err := f.Clients()
	if err != nil {
		return err
	}
	mapper, _ := f.Object()

	resourceNames := sets.NewString()
	for i, s := range o.include {
		if resourceNames.Has(s) {
			continue
		}
		if s != "*" {
			resourceNames.Insert(s)
			break
		}

		all, err := clientcmd.FindAllCanonicalResources(oclient.Discovery(), mapper)
		if err != nil {
			return fmt.Errorf("could not calculate the list of available resources: %v", err)
		}
		exclude := sets.NewString()
		for _, gr := range o.defaultExcludes {
			exclude.Insert(gr.String())
		}
		candidate := sets.NewString()
		for _, gr := range all {
			// if the user specifies a resource that matches resource or resource+group, skip it
			if resourceNames.Has(gr.Resource) || resourceNames.Has(gr.String()) || exclude.Has(gr.String()) {
				continue
			}
			candidate.Insert(gr.String())
		}
		candidate.Delete(exclude.List()...)
		include := candidate
		if len(o.overlappingResources) > 0 {
			include = sets.NewString()
			for _, k := range candidate.List() {
				reduce := k
				for _, others := range o.overlappingResources {
					if !others.Has(k) {
						continue
					}
					reduce = others.List()[0]
					break
				}
				include.Insert(reduce)
			}
		}
		glog.V(4).Infof("Found the following resources from the server: %v", include.List())
		last := o.include[i+1:]
		o.include = append([]string{}, o.include[:i]...)
		o.include = append(o.include, include.List()...)
		o.include = append(o.include, last...)
		break
	}

	m, _ := f.Object()
	_, gr := unversioned.ParseResourceArg(o.include[0])
	gvk, err := m.KindFor(gr.WithVersion(""))
	if err != nil {
		return err
	}
	o.mapping, err = m.RESTMapping(gvk.GroupKind())
	if err != nil {
		return err
	}
	o.client, err = f.ClientForMapping(o.mapping)
	if err != nil {
		return err
	}

	return nil
}

func (o *GraphOptions) Validate(args []string) error {
	return nil
}

func (o *GraphOptions) Run() error {
	// watch the given resource for changes
	store := &graphStore{nodeType: o.mapping.Resource}
	if err := store.Open("bolt://test:test@10.1.2.2:7687"); err != nil {
		return err
	}
	lw := restListWatcher{Helper: resource.NewHelper(o.client, o.mapping)}
	if !o.allNamespaces {
		lw.namespace = o.namespace
	}

	// ensure any child processes are reaped if we are running as PID 1
	proc.StartReaper()

	// listen on the provided address for metrics
	if len(o.listenAddr) > 0 {
		errWaitingForSync := fmt.Errorf("waiting for initial sync")
		healthz.InstallHandler(http.DefaultServeMux, healthz.NamedCheck("ready", func(r *http.Request) error {
			if !store.HasSynced() {
				return errWaitingForSync
			}
			return nil
		}))
		http.Handle("/metrics", prometheus.Handler())
		go func() {
			glog.Fatalf("Unable to listen on %q: %v", o.listenAddr, http.ListenAndServe(o.listenAddr, nil))
		}()
		glog.V(2).Infof("Listening on %s at /metrics and /healthz", o.listenAddr)
	}

	// start the reflector
	reflector := cache.NewNamedReflector("graph", lw, nil, store, o.resyncPeriod)
	reflector.Run()

	// wait forever
	select {}
}

type restListWatcher struct {
	*resource.Helper
	namespace string
}

func (lw restListWatcher) List(opt api.ListOptions) (runtime.Object, error) {
	return lw.Helper.List(lw.namespace, "", opt.LabelSelector, false)
}

func (lw restListWatcher) Watch(opt api.ListOptions) (watch.Interface, error) {
	return lw.Helper.Watch(lw.namespace, opt.ResourceVersion, "", opt.LabelSelector)
}

type graphStore struct {
	nodeType string
	conn     bolt.Conn
}

func (s *graphStore) HasSynced() bool              { return true }
func (s *graphStore) Add(obj interface{}) error    { return nil }
func (s *graphStore) Update(obj interface{}) error { return nil }
func (s *graphStore) Delete(obj interface{}) error { return nil }
func (s *graphStore) List() []interface{}          { return nil }
func (s *graphStore) ListKeys() []string           { return nil }
func (s *graphStore) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, nil
}
func (s *graphStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

func (s *graphStore) Replace(objs []interface{}, rv string) error {
	g := apigraph.New()
	for _, obj := range objs {
		switch t := obj.(type) {
		case *deployapi.DeploymentConfig:
			deploynodes.EnsureDeploymentConfigNode(g, t)
		}
	}
	deploygraph.AddAllTriggerEdges(g)
	deploygraph.AddAllVolumeClaimEdges(g)
	if err := applyGraph(g, s.conn); err != nil {
		return err
	}
	return markNamespacedNodesDeleted(s.conn, deploynodes.DeploymentConfigNodeKind, g)
}

func markNamespacedNodesDeleted(conn bolt.Conn, nodeKind string, existing apigraph.NodeFinder) error {
	qs, err := conn.PrepareNeo(fmt.Sprintf("MATCH (n:%s) RETURN n.namespace, n.name, ID(n)", nodeKind))
	if err != nil {
		return err
	}
	defer qs.Close()

	var removed []interface{}
	rows, err := qs.QueryNeo(nil)
	if err != nil {
		return err
	}
	for {
		r, _, err := rows.NextNeo()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if existing.Find(apigraph.UniqueName(fmt.Sprintf("%s|%s/%s", nodeKind, r[0].(string), r[1].(string)))) == nil {
			removed = append(removed, r[2].(int64))
		}
	}
	qs.Close()

	return markNodesDeletedByID(conn, nodeKind, removed...)
}

func markNodesDeletedByID(conn bolt.Conn, nodeKind string, ids ...interface{}) error {
	_, err := conn.ExecNeo(fmt.Sprintf(`
		MATCH (n:%s)
		WHERE ID(n) IN {ids} AND NOT exists(n.deletedAt)
		SET n.deletedAt = timestamp()
		`, nodeKind),
		map[string]interface{}{"ids": ids},
	)
	return err
}

func (s *graphStore) Resync() error { return nil }

func (s *graphStore) Open(connStr string) error {
	if glog.V(4) {
		boltlog.SetLevel("trace")
	}
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(connStr)
	if err != nil {
		return err
	}
	//defer conn.Close()
	s.conn = conn
	return nil
}

func simple(conn bolt.Conn, objs []interface{}, nodeType string) error {
	params := map[string]interface{}{"ns": nil, "name": nil, "rv": nil}
	ms, err := conn.PrepareNeo(fmt.Sprintf(`
		MERGE (n:%[1]s { namespace: $ns, name: $name })
		ON CREATE SET n.rv = $rv
		ON MATCH SET n.rv = $rv, n.deletedAt = NULL
	`, nodeType))
	if err != nil {
		return err
	}

	existing := make(map[string]struct{})
	for _, obj := range objs {
		switch t := obj.(type) {
		case *deployapi.DeploymentConfig:
			key := t.Namespace + "/" + t.Name
			existing[key] = struct{}{}
			params["ns"] = t.Namespace
			params["name"] = t.Name
			params["rv"] = t.ResourceVersion
			if _, err := ms.ExecNeo(params); err != nil {
				glog.Infof("unable to replace deployment %s in %s: %v", t.Name, t.Namespace, err)
				continue
			}

			glog.Infof("Replace updated %s", key)
		}
	}
	ms.Close()
	return nil
}
