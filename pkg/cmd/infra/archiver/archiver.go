package archiver

import (
	"fmt"
	"os"
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

	// locate the oldest snapshot
	snapshotSize := uint64(1000)
	snapshotWindow := snapshotSize
	resp, err := client.Get("/", false, false)
	if err != nil {
		return err
	}
	recentIndex := uint64(1)
	if resp.EtcdIndex > snapshotSize {
		recentIndex = resp.EtcdIndex - snapshotWindow + 1
	}

	// initialize the handlers
	archiver, err := OpenBoltArchiver("openshift-archive.boltdb", 0640)
	if err != nil {
		return err
	}
	defer archiver.Close()
	handler := &kubernetesResourceHandler{
		archiver: archiver,
	}

	watches := make(chan chan *etcd.Response)

	go util.Forever(func() {
		ch := make(chan *etcd.Response)
		watches <- ch
		if _, err := client.Watch("/", recentIndex, true, ch, nil); err != nil {
			snapshotWindow = snapshotWindow * 9 / 10
			if etcdError, ok := err.(*etcd.EtcdError); ok {
				recentIndex = etcdError.Index - snapshotWindow
			}
			glog.Errorf("Unable to watch: %v", err)
			return
		}
		snapshotWindow = snapshotSize
	}, 1*time.Second)

	lowestIndex := uint64(0)
	go util.Forever(func() {
		glog.Infof("Ready to archive changes from etcd ...")

		for ch := range watches {
			glog.Infof("Watching ...")
			for resp := range ch {
				index := uint64(0)
				path := ""
				creation := false
				deletion := resp.Action == "delete"
				if resp.Node != nil {
					path = resp.Node.Key
					index = resp.Node.ModifiedIndex
					creation = !deletion && resp.Node.ModifiedIndex == resp.Node.CreatedIndex
				} else if resp.PrevNode != nil {
					path = resp.PrevNode.Key
					index = resp.Node.ModifiedIndex
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

	for _ = range time.NewTicker(5 * time.Second).C {
		archiver.Dump(os.Stdout)
	}
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

type Archiver interface {
	Create(resource, namespace, name string, index uint64, current *etcd.Node) error
	Update(resource, namespace, name string, index uint64, current, previous *etcd.Node) error
	Delete(resource, namespace, name string, index uint64, previous *etcd.Node) error
}

type kubernetesResourceHandler struct {
	archiver Archiver
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
		glog.Infof("created %s %q %q", resource, namespace, name)
		if err := h.archiver.Create(resource, namespace, name, index, current); err != nil {
			glog.Errorf("Unable to record creation: %v", err)
		}
	case deletion:
		glog.Infof("deleted %s %q %q", resource, namespace, name)
		if err := h.archiver.Delete(resource, namespace, name, index, previous); err != nil {
			glog.Errorf("Unable to record deletion: %v", err)
		}
	default:
		glog.Infof("updated %s %q %q", resource, namespace, name)
		if err := h.archiver.Update(resource, namespace, name, index, current, previous); err != nil {
			glog.Errorf("Unable to record update: %v", err)
		}
	}
}

type archiver struct {
}

func (a *archiver) Create(resource, namespace, name string, index uint64, current *etcd.Node) error {
	glog.Infof("created %s %q %q %d", resource, namespace, name, index)
	return nil
}

func (a *archiver) Update(resource, namespace, name string, index uint64, current *etcd.Node, previous *etcd.Node) error {
	glog.Infof("updated %s %q %q %d", resource, namespace, name, index)
	return nil
}

func (a *archiver) Delete(resource, namespace, name string, index uint64, previous *etcd.Node) error {
	glog.Infof("deleted %s %q %q %d", resource, namespace, name, index)
	return nil
}
