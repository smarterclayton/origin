package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/golang/glog"

	kapierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kexternalinformers "k8s.io/kubernetes/pkg/client/informers/informers_generated/externalversions"
	kinternalinformers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"

	osapi "github.com/openshift/origin/pkg/sdn/api"
)

const (
	// randomly-selected conntrack zone ID; FIXME, make this configurable
	networkPolicyConntrackZone = 13877
)

type networkPolicyPlugin struct {
	node  *OsdnNode
	vnids *nodeVNIDMap

	lock        sync.Mutex
	namespaces  map[uint32]*npNamespace
	kNamespaces map[string]*v1.Namespace
	pods        map[ktypes.UID]kapi.Pod

	kubeInformers         kinternalinformers.SharedInformerFactory
	kubeExternalInformers kexternalinformers.SharedInformerFactory
}

// npNamespace tracks NetworkPolicy-related data for a Namespace
type npNamespace struct {
	name     string
	vnid     uint32
	isolated bool
	refs     int
	inUse    bool

	policies map[ktypes.UID]*npPolicy
}

// npPolicy is a parsed version of a single NetworkPolicy object
type npPolicy struct {
	policy            extensions.NetworkPolicy
	watchesNamespaces bool
	watchesPods       bool

	flows []string
}

func NewNetworkPolicyPlugin() osdnPolicy {
	return &networkPolicyPlugin{
		namespaces:  make(map[uint32]*npNamespace),
		kNamespaces: make(map[string]*v1.Namespace),
		pods:        make(map[ktypes.UID]kapi.Pod),
	}
}

func (np *networkPolicyPlugin) Name() string {
	return osapi.NetworkPolicyPluginName
}

func (np *networkPolicyPlugin) Start(node *OsdnNode) error {
	np.node = node
	np.kubeInformers = node.kubeInformers
	np.kubeExternalInformers = node.kubeExternalInformers
	np.vnids = newNodeVNIDMap(np, node.osClient)
	if err := np.vnids.Start(); err != nil {
		return err
	}

	otx := node.oc.NewTransaction()
	otx.AddFlow("table=21, priority=200, ip, nw_dst=%s, actions=goto_table:30", np.node.networkInfo.ServiceNetwork.String())
	otx.AddFlow("table=21, priority=100, ip, actions=ct(zone=%d,commit,table=30)", networkPolicyConntrackZone)
	otx.AddFlow("table=80, priority=50, ip, actions=ct(zone=%d,commit,table=81)", networkPolicyConntrackZone)
	otx.AddFlow("table=81, priority=100, ip, ct_state=+trk+est, actions=output:NXM_NX_REG2[]")
	otx.AddFlow("table=81, priority=0, actions=drop")
	if err := otx.EndTransaction(); err != nil {
		return err
	}

	if err := np.initNamespaces(); err != nil {
		return err
	}
	if err := np.node.SetupEgressNetworkPolicy(); err != nil {
		return err
	}

	np.watchNamespaces()
	np.watchPods()
	go utilwait.Forever(np.watchNetworkPolicies, 0)
	return nil
}

func (np *networkPolicyPlugin) initNamespaces() error {
	np.lock.Lock()
	defer np.lock.Unlock()

	namespaces, err := np.node.kClient.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range namespaces.Items {
		np.kNamespaces[ns.Name] = &ns

		if vnid, err := np.vnids.WaitAndGetVNID(ns.Name); err == nil {
			np.namespaces[vnid] = &npNamespace{
				name:     ns.Name,
				vnid:     vnid,
				isolated: namespaceIsIsolated(&ns),
				refs:     0,
				policies: make(map[ktypes.UID]*npPolicy),
			}
		}
	}

	policies, err := np.node.kInternalClient.Extensions().NetworkPolicies(kapi.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		if kapierrs.IsForbidden(err) {
			glog.Errorf("Unable to query NetworkPolicies (%v) - please ensure your nodes have access to view NetworkPolicy (eg, 'oadm policy reconcile-cluster-roles')", err)
		}
		return err
	}
	for _, policy := range policies.Items {
		vnid, err := np.vnids.WaitAndGetVNID(policy.Namespace)
		if err != nil {
			continue
		}
		npns := np.namespaces[vnid]
		np.updateNetworkPolicy(npns, &policy)
	}

	return nil
}

func (np *networkPolicyPlugin) AddNetNamespace(netns *osapi.NetNamespace) {
	np.lock.Lock()
	defer np.lock.Unlock()

	if _, exists := np.namespaces[netns.NetID]; exists {
		glog.Warning("Got AddNetNamespace for already-existing namespace %s (%d)", netns.NetName, netns.NetID)
		return
	}

	isolated := false
	if kns, exists := np.kNamespaces[netns.NetName]; exists {
		isolated = namespaceIsIsolated(kns)
	}

	np.namespaces[netns.NetID] = &npNamespace{
		name:     netns.NetName,
		vnid:     netns.NetID,
		isolated: isolated,
		refs:     0,
		policies: make(map[ktypes.UID]*npPolicy),
	}
}

func (np *networkPolicyPlugin) UpdateNetNamespace(netns *osapi.NetNamespace, oldNetID uint32) {
	if netns.NetID != oldNetID {
		glog.Warning("Got VNID change for namespace %s while using %s plugin", netns.NetName, osapi.NetworkPolicyPluginName)
	}

	np.node.podManager.UpdateLocalMulticastRules(netns.NetID)
}

func (np *networkPolicyPlugin) DeleteNetNamespace(netns *osapi.NetNamespace) {
	np.lock.Lock()
	defer np.lock.Unlock()

	delete(np.namespaces, netns.NetID)
}

func (np *networkPolicyPlugin) GetVNID(namespace string) (uint32, error) {
	return np.vnids.WaitAndGetVNID(namespace)
}

func (np *networkPolicyPlugin) GetNamespaces(vnid uint32) []string {
	return np.vnids.GetNamespaces(vnid)
}

func (np *networkPolicyPlugin) GetMulticastEnabled(vnid uint32) bool {
	return np.vnids.GetMulticastEnabled(vnid)
}

func (np *networkPolicyPlugin) syncNamespace(npns *npNamespace) {
	inUse := npns.refs > 0
	if !inUse && !npns.inUse {
		return
	}

	glog.V(5).Infof("syncNamespace %d", npns.vnid)
	otx := np.node.oc.NewTransaction()
	otx.DeleteFlows("table=80, reg1=%d", npns.vnid)
	if inUse {
		if npns.isolated {
			for _, npp := range npns.policies {
				for _, flow := range npp.flows {
					otx.AddFlow("table=80, priority=100, reg1=%d, %s actions=output:NXM_NX_REG2[]", npns.vnid, flow)
				}
			}
		} else {
			otx.AddFlow("table=80, priority=100, reg1=%d, actions=output:NXM_NX_REG2[]", npns.vnid)
		}
	}
	if err := otx.EndTransaction(); err != nil {
		glog.Errorf("Error syncing OVS flows for VNID: %v", err)
	}
	npns.inUse = inUse
}

func (np *networkPolicyPlugin) RefVNID(vnid uint32) {
	np.lock.Lock()
	defer np.lock.Unlock()

	npns, exists := np.namespaces[vnid]
	if !exists {
		return
	}

	npns.refs += 1
	np.syncNamespace(npns)
}

func (np *networkPolicyPlugin) UnrefVNID(vnid uint32) {
	np.lock.Lock()
	defer np.lock.Unlock()

	npns, exists := np.namespaces[vnid]
	if !exists {
		return
	}
	if npns.refs == 0 {
		glog.Warningf("refcounting error on vnid %d", vnid)
		return
	}

	npns.refs -= 1
	np.syncNamespace(npns)
}

func (np *networkPolicyPlugin) selectNamespaces(lsel *metav1.LabelSelector) []uint32 {
	vnids := []uint32{}
	sel, err := metav1.LabelSelectorAsSelector(lsel)
	if err != nil {
		// Shouldn't happen
		glog.Errorf("ValidateNetworkPolicy() failure! Invalid NamespaceSelector: %v", err)
		return vnids
	}
	for vnid, ns := range np.namespaces {
		if kns, exists := np.kNamespaces[ns.name]; exists {
			if sel.Matches(labels.Set(kns.Labels)) {
				vnids = append(vnids, vnid)
			}
		}
	}
	return vnids
}

func (np *networkPolicyPlugin) selectPods(npns *npNamespace, lsel *metav1.LabelSelector) []string {
	ips := []string{}
	sel, err := metav1.LabelSelectorAsSelector(lsel)
	if err != nil {
		// Shouldn't happen
		glog.Errorf("ValidateNetworkPolicy() failure! Invalid PodSelector: %v", err)
		return ips
	}
	for _, pod := range np.pods {
		if (npns.name == pod.Namespace) && sel.Matches(labels.Set(pod.Labels)) {
			ips = append(ips, pod.Status.PodIP)
		}
	}
	return ips
}

func (np *networkPolicyPlugin) parseNetworkPolicy(npns *npNamespace, policy *extensions.NetworkPolicy) (*npPolicy, error) {
	npp := &npPolicy{policy: *policy}

	var destFlows []string
	if len(policy.Spec.PodSelector.MatchLabels) > 0 || len(policy.Spec.PodSelector.MatchExpressions) > 0 {
		npp.watchesPods = true
		for _, ip := range np.selectPods(npns, &policy.Spec.PodSelector) {
			destFlows = append(destFlows, fmt.Sprintf("ip, nw_dst=%s, ", ip))
		}
	} else {
		destFlows = []string{""}
	}

	for _, rule := range policy.Spec.Ingress {
		var portFlows, peerFlows []string
		if len(rule.Ports) == 0 {
			portFlows = []string{""}
		}
		for _, port := range rule.Ports {
			var protocol string
			if port.Protocol == nil {
				protocol = "tcp"
			} else if *port.Protocol == kapi.ProtocolTCP || *port.Protocol == kapi.ProtocolUDP {
				protocol = strings.ToLower(string(*port.Protocol))
			} else {
				// FIXME: validation should catch this
				return nil, fmt.Errorf("policy specifies unrecognized protocol %q", *port.Protocol)
			}
			var portNum int
			if port.Port.Type == intstr.Int {
				portNum = int(port.Port.IntVal)
				if portNum < 0 || portNum > 0xFFFF {
					// FIXME: validation should catch this
					return nil, fmt.Errorf("port value out of bounds %q", port.Port.IntVal)
				}
			} else {
				// FIXME: implement this
				return nil, fmt.Errorf("named port values (%q) are not yet implemented", port.Port.StrVal)
			}
			portFlows = append(portFlows, fmt.Sprintf("%s, tp_dst=%d, ", protocol, portNum))
		}

		if len(rule.From) == 0 {
			peerFlows = []string{""}
		}
		for _, peer := range rule.From {
			if peer.PodSelector != nil {
				if len(peer.PodSelector.MatchLabels) == 0 && len(peer.PodSelector.MatchExpressions) == 0 {
					// The PodSelector is empty, meaning it selects all pods in this namespace
					peerFlows = append(peerFlows, fmt.Sprintf("reg0=%d, ", npns.vnid))
				} else {
					npp.watchesPods = true
					for _, ip := range np.selectPods(npns, peer.PodSelector) {
						peerFlows = append(peerFlows, fmt.Sprintf("reg0=%d, ip, nw_src=%s, ", npns.vnid, ip))
					}
				}
			} else {
				if len(peer.NamespaceSelector.MatchLabels) == 0 && len(peer.NamespaceSelector.MatchExpressions) == 0 {
					// The NamespaceSelector is empty, meaning it selects all namespaces
					peerFlows = append(peerFlows, "")
				} else {
					npp.watchesNamespaces = true
					for _, otherVNID := range np.selectNamespaces(peer.NamespaceSelector) {
						peerFlows = append(peerFlows, fmt.Sprintf("reg0=%d, ", otherVNID))
					}
				}
			}
		}

		for _, destFlow := range destFlows {
			for _, peerFlow := range peerFlows {
				for _, portFlow := range portFlows {
					npp.flows = append(npp.flows, fmt.Sprintf("%s%s%s", destFlow, peerFlow, portFlow))
				}
			}
		}
	}

	sort.Strings(npp.flows)
	glog.V(5).Infof("Parsed NetworkPolicy: %#v", npp)
	return npp, nil
}

func (np *networkPolicyPlugin) updateNetworkPolicy(npns *npNamespace, policy *extensions.NetworkPolicy) bool {
	npp, err := np.parseNetworkPolicy(npns, policy)
	if err != nil {
		glog.Infof("Unsupported NetworkPolicy %s/%s (%v); treating as deny-all", policy.Namespace, policy.Name, err)
		npp = &npPolicy{policy: *policy}
	}

	oldNPP, existed := npns.policies[policy.UID]
	npns.policies[policy.UID] = npp

	changed := !existed || !reflect.DeepEqual(oldNPP.flows, npp.flows)
	if !changed {
		glog.V(5).Infof("NetworkPolicy %s/%s is unchanged", policy.Namespace, policy.Name)
	}
	return changed
}

func (np *networkPolicyPlugin) watchNetworkPolicies() {
	RunEventQueue(np.node.kInternalClient.Extensions().RESTClient(), NetworkPolicies, func(delta cache.Delta) error {
		policy := delta.Object.(*extensions.NetworkPolicy)

		glog.V(5).Infof("Watch %s event for NetworkPolicy %s/%s", delta.Type, policy.Namespace, policy.Name)

		vnid, err := np.vnids.WaitAndGetVNID(policy.Namespace)
		if err != nil {
			return err
		}

		np.lock.Lock()
		defer np.lock.Unlock()
		npns, exists := np.namespaces[vnid]
		if !exists {
			// NetNamespace was deleted after WaitAndGetVNID() returned!
			return nil
		}

		switch delta.Type {
		case cache.Sync, cache.Added, cache.Updated:
			if changed := np.updateNetworkPolicy(npns, policy); changed {
				np.syncNamespace(npns)
			}
		case cache.Deleted:
			delete(npns.policies, policy.UID)
			np.syncNamespace(npns)
		}

		return nil
	})
}

const (
	NetworkPolicyAnnotation = "net.beta.kubernetes.io/network-policy"
)

type IngressIsolationPolicy string

const (
	DefaultDeny IngressIsolationPolicy = "DefaultDeny"
)

type NamespaceNetworkPolicy struct {
	Ingress *NamespaceIngressPolicy `json:"ingress,omitempty"`
}

type NamespaceIngressPolicy struct {
	Isolation *IngressIsolationPolicy `json:"isolation,omitempty"`
}

func namespaceIsIsolated(ns *v1.Namespace) bool {
	annotation, exists := ns.Annotations[NetworkPolicyAnnotation]
	if !exists {
		return false
	}
	var policy NamespaceNetworkPolicy
	if err := json.Unmarshal([]byte(annotation), &policy); err != nil {
		glog.Warningf("Namespace %q has unparsable %q annotation %q", ns.Name, NetworkPolicyAnnotation, annotation)
		return false
	} else if policy.Ingress != nil && *policy.Ingress.Isolation == DefaultDeny {
		return true
	} else {
		return false
	}
}

func (np *networkPolicyPlugin) watchPods() {
	RegisterSharedInformerEventHandlers(
		np.kubeInformers,
		np.kubeExternalInformers,
		np.handleAddOrUpdatePod, np.handleDeletePod, Pods)
}

func (np *networkPolicyPlugin) handleAddOrUpdatePod(obj, _ interface{}, eventType watch.EventType) {
	pod := obj.(*kapi.Pod)
	glog.V(5).Infof("Watch %s event for Pod %q", eventType, getPodFullName(pod))

	// Ignore pods with HostNetwork=true, SDN is not involved in this case
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.HostNetwork {
		return
	}
	if pod.Status.PodIP == "" {
		glog.Warningf("PodIP is not set for pod %q", getPodFullName(pod))
		return
	}

	// We don't want to grab np.Lock for every Pod.Status change...
	// But it's safe to look up oldPod without locking here because no other
	// threads modify this map.
	oldPod, podExisted := np.pods[pod.UID]
	if podExisted && oldPod.Status.PodIP == pod.Status.PodIP && reflect.DeepEqual(oldPod.Labels, pod.Labels) {
		return
	}

	np.lock.Lock()
	defer np.lock.Unlock()

	np.pods[pod.UID] = *pod
	np.refreshNetworkPolicies(Pods)
}

func (np *networkPolicyPlugin) handleDeletePod(obj interface{}) {
	pod := obj.(*kapi.Pod)
	glog.V(5).Infof("Watch %s event for Pod %q", watch.Deleted, getPodFullName(pod))

	_, podExisted := np.pods[pod.UID]
	if !podExisted {
		return
	}

	np.lock.Lock()
	defer np.lock.Unlock()

	delete(np.pods, pod.UID)
	np.refreshNetworkPolicies(Pods)
}

func (np *networkPolicyPlugin) watchNamespaces() {
	RegisterSharedInformerEventHandlers(
		np.kubeInformers,
		np.kubeExternalInformers,
		np.handleAddOrUpdateNamespace, np.handleDeleteNamespace, Namespaces)
}

func (np *networkPolicyPlugin) handleAddOrUpdateNamespace(obj, _ interface{}, eventType watch.EventType) {
	ns := obj.(*v1.Namespace)
	glog.V(5).Infof("Watch %s event for Namespace %q", eventType, ns.Name)
	// Don't grab the lock yet since this may block
	vnid, err := np.vnids.WaitAndGetVNID(ns.Name)
	if err != nil {
		glog.Error(err)
		return
	}

	np.lock.Lock()
	defer np.lock.Unlock()
	np.kNamespaces[ns.Name] = ns
	if npns, exists := np.namespaces[vnid]; exists {
		npns.isolated = namespaceIsIsolated(ns)
		np.syncNamespace(npns)
	}
	// else the NetNamespace doesn't exist yet, but we will initialize
	// npns.isolated from the v1.Namespace when it's created

	np.refreshNetworkPolicies(Namespaces)
}

func (np *networkPolicyPlugin) handleDeleteNamespace(obj interface{}) {
	ns := obj.(*v1.Namespace)
	glog.V(5).Infof("Watch %s event for Namespace %q", watch.Deleted, ns.Name)
	np.lock.Lock()
	defer np.lock.Unlock()
	delete(np.kNamespaces, ns.Name)

	// We don't need to np.syncNamespace() because if the NetNamespace
	// still existed, it will be deleted as part of deleting the Namespace.

	np.refreshNetworkPolicies(Namespaces)
}

func (np *networkPolicyPlugin) refreshNetworkPolicies(watchResourceName ResourceName) {
	for _, npns := range np.namespaces {
		changed := false
		for _, npp := range npns.policies {
			if ((watchResourceName == Namespaces) && npp.watchesNamespaces) ||
				((watchResourceName == Pods) && npp.watchesPods) {
				if np.updateNetworkPolicy(npns, &npp.policy) {
					changed = true
					break
				}
			}
		}
		if changed {
			np.syncNamespace(npns)
		}
	}
}

func getPodFullName(pod *kapi.Pod) string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}
