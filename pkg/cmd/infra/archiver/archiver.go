package archiver

import (
	//"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/cmd/util/etcdcmd"
)

const longCommandDesc = `
Start an OpenShift archiver

This command launches an archiver connected to your etcd store. The archiver copies changes to
etcd into a backing store that can provide historical state over OpenShift and Kubernetes
resources
`

type config struct {
	Config         *etcdcmd.Config
	DeploymentName string
}

func NewCommandArchiver(name string) *cobra.Command {
	cfg := &config{
		Config: etcdcmd.NewConfig(),
	}

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s%s", name, etcdcmd.ConfigSyntax),
		Short: "Start an OpenShift archiver",
		Long:  longCommandDesc,
		Run: func(c *cobra.Command, args []string) {
			if err := start(cfg); err != nil {
				glog.Fatal(err)
			}
		},
	}

	flag := cmd.Flags()
	cfg.Config.Bind(flag)

	return cmd
}

// start begins archiving
func start(cfg *config) error {
	client, err := cfg.Config.Client(true)
	if err != nil {
		return err
	}

	handler := &kubernetesResourceHandler{
		archiver: &archiver{},
	}

	watches := make(chan chan *etcd.Response)

	recentIndex := uint64(1)
	go util.Forever(func() {
		ch := make(chan *etcd.Response)
		watches <- ch
		if _, err := client.Watch("/", recentIndex, true, ch, nil); err != nil {
			if etcdError, ok := err.(*etcd.EtcdError); ok {
				recentIndex = etcdError.Index
			}
			glog.Errorf("Unable to watch: %v", err)
			return
		}
		glog.Infof("next loop")
	}, 1*time.Second)

	lowestIndex := uint64(0)
	go util.Forever(func() {
		glog.Infof("Ready to archive changes from etcd ...")

		for ch := range watches {
			glog.Infof("Watching ...")
			for resp := range ch {
				index := uint64(0)
				path := ""
				deletion := false
				creation := false
				if resp.Node != nil {
					path = resp.Node.Key
					index = resp.Node.ModifiedIndex
					creation = resp.PrevNode == nil
				} else if resp.PrevNode != nil {
					path = resp.PrevNode.Key
					index = resp.Node.ModifiedIndex
					deletion = true
				}

				// ignore results we've already seen
				if index <= lowestIndex {
					glog.V(4).Infof("Already seen %d", index)
					continue
				}
				lowestIndex = index

				handler.Change(index, path, creation, deletion, resp.Node, resp.PrevNode)
			}
		}
	}, 10*time.Millisecond)

	select {}
	return nil
}

const kubernetesCommonPrefix = "/registry/"

var kubernetesTypes = map[string]struct {
	Namespaced bool
}{
	"pods": {
		Namespaced: true,
	},
}

type kubernetesResourceHandler struct {
	archiver *archiver
}

func (h *kubernetesResourceHandler) Change(index uint64, path string, creation, deletion bool, current, previous *etcd.Node) {
	if !strings.HasPrefix(path, kubernetesCommonPrefix) {
		glog.V(4).Infof("Ignored key %q(%d)", path, index)
		return
	}
	path = path[len(kubernetesCommonPrefix):]

	segments := strings.SplitN(path, "/", 4)
	t, ok := kubernetesTypes[segments[0]]
	if !ok {
		//glog.V(4).Infof("Ignored key %q(%d)", path, index)
		return
	}

	resource, name, namespace := segments[0], "", ""
	if t.Namespaced {
		if len(segments) != 3 {
			glog.V(4).Infof("Ignored key %q(%d)", path, index)
			return
		}
		namespace, name = segments[1], segments[2]
	} else {
		if len(segments) != 2 {
			glog.V(4).Infof("Ignored key %q(%d)", path, index)
			return
		}
		name = segments[1]
	}

	switch {
	case creation:
		if err := h.archiver.Create(resource, namespace, name, current); err != nil {
			glog.Errorf("Unable to record creation: %v", err)
		}
	case deletion:
		if err := h.archiver.Delete(resource, namespace, name, previous); err != nil {
			glog.Errorf("Unable to record deletion: %v", err)
		}
	default:
		if err := h.archiver.Update(resource, namespace, name, current, previous); err != nil {
			glog.Errorf("Unable to record update: %v", err)
		}
	}
}

type metadata struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type common struct {
	Metadata metadata `json:"metadata"`

	UID       string `json:"uid"`
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
}

func (c *common) Reset() {
	c.ID, c.Namespace, c.UID = "", "", ""
	c.Metadata.Name, c.Metadata.Namespace, c.UID = "", "", ""
}

type archiver struct {
}

func (a *archiver) Create(resource, namespace, name string, current *etcd.Node) error {
	glog.Infof("created %s %q %q", resource, namespace, name)
	return nil
}

func (a *archiver) Update(resource, namespace, name string, current *etcd.Node, previous *etcd.Node) error {
	glog.Infof("updated %s %q %q", resource, namespace, name)
	return nil
}

func (a *archiver) Delete(resource, namespace, name string, previous *etcd.Node) error {
	glog.Infof("deleted %s %q %q", resource, namespace, name)
	return nil
}
