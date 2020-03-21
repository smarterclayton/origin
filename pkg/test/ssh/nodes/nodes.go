package nodes

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/objx"
	"golang.org/x/crypto/ssh"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func New(nodeSelector string, interval time.Duration) (*SSH, error) {
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	clusterConfig, err := cfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load client configuration: %v", err)
	}
	client, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}

	s := &SSH{
		port:         22,
		nodeSelector: nodeSelector,
		interval:     interval,
		client:       client,
		configClient: dynamicClient.Resource(schema.GroupVersionResource{
			Group:    "machineconfiguration.openshift.io",
			Version:  "v1",
			Resource: "machineconfigs",
		}),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

type SSH struct {
	nodeSelector string
	interval     time.Duration
	client       kubernetes.Interface
	configClient dynamic.ResourceInterface
	port         int

	lock           sync.Mutex
	lastRefresh    time.Time
	addresses      map[string]string
	authorizedKeys []ssh.PublicKey
}

func (s *SSH) AddressFor(address string) (string, int, bool, error) {
	err := s.load()

	s.lock.Lock()
	defer s.lock.Unlock()
	val, ok := s.addresses[address]
	return val, s.port, ok, err
}

func (s *SSH) AuthorizedKeys() ([]ssh.PublicKey, error) {
	err := s.load()

	s.lock.Lock()
	defer s.lock.Unlock()
	return s.authorizedKeys, err
}

func (s *SSH) load() error {
	var last time.Time
	s.lock.Lock()
	last = s.lastRefresh
	s.lock.Unlock()

	if time.Now().Sub(last) < s.interval {
		return nil
	}

	var desiredMachineConfig string
	addresses := make(map[string]string)
	nodes, err := s.client.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: s.nodeSelector})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		if v := node.Annotations["machineconfiguration.openshift.io/desiredConfig"]; len(v) > 0 {
			desiredMachineConfig = v
		}
		for _, address := range node.Status.Addresses {
			if len(address.Address) == 0 {
				continue
			}
			addresses[address.Address] = address.Address
		}
		if address := preferredNodeAddress(&node, v1.NodeInternalIP, v1.NodeExternalIP); len(address) > 0 {
			addresses[node.Name] = address
		}
	}

	var authorizedKeys []ssh.PublicKey
	if len(desiredMachineConfig) > 0 {
		config, err := s.configClient.Get(desiredMachineConfig, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj := objx.Map(config.UnstructuredContent())
		for _, user := range objects(obj.Get("spec.config.passwd.users")) {
			for i, val := range asStrings(user.Get("sshAuthorizedKeys")) {
				key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(val))
				if err != nil {
					klog.Warningf("Authorized key %d for user %q is not valid", i, user.Get("name").String())
					continue
				}
				authorizedKeys = append(authorizedKeys, key)
			}
		}
	}

	targets := sets.NewString()
	for k := range addresses {
		targets.Insert(k)
	}
	klog.V(2).Infof("Accepting SSH traffic from %d keys to nodes: %s", len(authorizedKeys), strings.Join(targets.List(), ", "))

	now := time.Now()

	s.lock.Lock()
	defer s.lock.Unlock()
	s.lastRefresh = now
	s.addresses = addresses
	s.authorizedKeys = authorizedKeys
	return nil
}

func preferredNodeAddress(node *v1.Node, addressTypes ...v1.NodeAddressType) string {
	for _, addressType := range addressTypes {
		for _, address := range node.Status.Addresses {
			if address.Type == addressType {
				return address.Address
			}
		}
	}
	return ""
}

func objects(from *objx.Value) []objx.Map {
	var values []objx.Map
	switch {
	case from.IsObjxMapSlice():
		return from.ObjxMapSlice()
	case from.IsInterSlice():
		for _, i := range from.InterSlice() {
			if msi, ok := i.(map[string]interface{}); ok {
				values = append(values, objx.Map(msi))
			}
		}
	}
	return values
}

func asStrings(from *objx.Value) []string {
	var values []string
	switch {
	case from.IsStrSlice():
		return from.StrSlice()
	case from.IsInterSlice():
		for _, i := range from.InterSlice() {
			if s, ok := i.(string); ok {
				values = append(values, s)
			}
		}
	}
	return values
}
