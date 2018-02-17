package secretref

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	extensionsinformers "k8s.io/client-go/informers/extensions/v1beta1"
	rbacinformers "k8s.io/client-go/informers/rbac/v1"
	kv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacclient "k8s.io/client-go/kubernetes/typed/rbac/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	rbaclisters "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

const (
	// maxRetries is the number of times an ingress will be retried before it is dropped out of the queue.
	maxRetries          = 5
	maxResourceInterval = 30 * time.Second
)

// defaultResourceFailureDelay will retry failures forever, but implements an exponential
// capped backoff after a certain limit.
func defaultResourceFailureDelay(requeue int) (time.Duration, bool) {
	if requeue > 5 {
		return maxResourceInterval, true
	}
	t := time.Duration(math.Pow(2.0, float64(requeue)) * float64(time.Second))
	if t > maxResourceInterval {
		t = maxResourceInterval
	}
	return t, true
}

// Controller ensures that a role and role-binding exist in each namespace that grant
// access to view secrets referenced by ingresses.
//
// Invariants:
//
// 1. For every ingress that references a TLS secret, a role should exist in the ingress
//    namespace that grants get access to the secret by name
// 2. TODO: binding
//
type Controller struct {
	eventRecorder record.EventRecorder

	// Allows injection for testing, controls requeues on errors
	resourceFailureDelayFn func(requeue int) (time.Duration, bool)

	roleName string

	client        rbacclient.RolesGetter
	secretLister  corelisters.SecretLister
	ingressLister extensionslisters.IngressLister
	roleLister    rbaclisters.RoleLister

	// queue is the list of namespace keys that must be synced.
	queue workqueue.RateLimitingInterface

	// syncs are the items that must return true before the queue can be processed
	syncs []cache.InformerSynced
}

func NewEventBroadcaster(client kv1core.CoreV1Interface) record.EventBroadcaster {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	// TODO: remove the wrapper when every client has moved to use the clientset.
	eventBroadcaster.StartRecordingToSink(&kv1core.EventSinkImpl{Interface: client.Events("")})
	return eventBroadcaster
}

// NewController instantiates a Controller
func NewController(eventBroadcaster record.EventBroadcaster, client rbacclient.RolesGetter, ingresses extensionsinformers.IngressInformer, roles rbacinformers.RoleInformer, secrets coreinformers.SecretInformer) *Controller {
	c := &Controller{
		eventRecorder: eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: "ingress-secretref-controller"}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-secretref"),

		roleName:      "ingress-secretref",
		client:        client,
		ingressLister: ingresses.Lister(),
		roleLister:    roles.Lister(),
		secretLister:  secrets.Lister(),

		resourceFailureDelayFn: defaultResourceFailureDelay,
	}

	// process the whole namespace
	processKeyFn := func(obj interface{}) {
		switch t := obj.(type) {
		case metav1.Object:
			ns := t.GetNamespace()
			if len(ns) == 0 {
				utilruntime.HandleError(fmt.Errorf("object %T has no namespace", obj))
				return
			}
			c.queue.Add(ns)
		default:
			utilruntime.HandleError(fmt.Errorf("couldn't get key for object %T", obj))
		}
	}

	// any change to a secret of type TLS in the namespace
	secrets.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *v1.Secret:
				return t.Type == v1.SecretTypeTLS
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    processKeyFn,
			DeleteFunc: processKeyFn,
			UpdateFunc: func(oldObj, newObj interface{}) {
				processKeyFn(newObj)
			},
		},
	})

	// any change to a role with the expected name triggers the namespace
	roles.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *rbacv1.Role:
				return t.Name == c.roleName
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    processKeyFn,
			DeleteFunc: processKeyFn,
			UpdateFunc: func(oldObj, newObj interface{}) {
				processKeyFn(newObj)
			},
		},
	})

	// changes to ingresses that have TLS rules
	ingresses.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *extensionsv1beta1.Ingress:
				for _, rule := range t.Spec.TLS {
					if len(rule.SecretName) > 0 {
						return true
					}
				}
				return false
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    processKeyFn,
			DeleteFunc: processKeyFn,
			UpdateFunc: func(oldObj, newObj interface{}) {
				switch t := oldObj.(type) {
				case *extensionsv1beta1.Ingress:
					switch t2 := newObj.(type) {
					case *extensionsv1beta1.Ingress:
						if reflect.DeepEqual(t.Spec.TLS, t2.Spec.TLS) {
							// filter out updates that don't alter secret names
							return
						}
					}
				}
				processKeyFn(newObj)
			},
		},
	})

	c.syncs = []cache.InformerSynced{ingresses.Informer().HasSynced, roles.Informer().HasSynced}

	return c
}

// Run begins watching and syncing.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting controller")

	if !cache.WaitForCacheSync(stopCh, c.syncs...) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	glog.Infof("Shutting down controller")
}

func (c *Controller) worker() {
	for c.processNextNamespace() {
	}
	glog.V(4).Infof("Worker stopped")
}

func (c *Controller) processNextNamespace() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(key.(string))
	c.handleNamespaceErr(err, key)

	return true
}

func (c *Controller) handleNamespaceErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		glog.V(4).Infof("Error syncing %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	glog.V(4).Infof("Dropping %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *Controller) sync(ns string) error {
	// identify the set of secrets that must match
	ingresses, err := c.ingressLister.Ingresses(ns).List(labels.Everything())
	if err != nil {
		return err
	}
	var secrets sets.String
	for _, ingress := range ingresses {
		for _, tls := range ingress.Spec.TLS {
			if len(tls.SecretName) > 0 {
				s, err := c.secretLister.Secrets(ns).Get(tls.SecretName)
				if err != nil || s.Type != v1.SecretTypeTLS {
					continue
				}
				if secrets == nil {
					secrets = sets.NewString()
				}
				secrets.Insert(tls.SecretName)
			}
		}
	}

	// check whether we need to delete the existing role
	role, err := c.roleLister.Roles(ns).Get(c.roleName)
	switch {
	case err != nil && !errors.IsNotFound(err):
		return err
	case role != nil && len(secrets) == 0:
		// role should be deleted
		err = c.client.Roles(ns).Delete(c.roleName, nil)
		if errors.IsNotFound(err) {
			err = nil
		}
		return err
	case len(secrets) == 0:
		return nil
	}

	// create a new role
	if role == nil {
		role = newRoleForSecrets(secrets)
		role.Name = c.roleName
		_, err = c.client.Roles(ns).Create(role)
		return err
	}

	switch {
	case roleHasInvalidElements(role):
		// if the role is invalid, always reset the rules
		role = newRoleForSecrets(secrets)
	case secrets.Equal(sets.NewString(role.Rules[0].ResourceNames...)):
		return nil
	default:
		role = newRoleForSecrets(secrets)
	}

	data, err := json.Marshal(role.Rules)
	if err != nil {
		return err
	}
	data = []byte(fmt.Sprintf(`{"rules":%s}`, data))
	_, err = c.client.Roles(ns).Patch(c.roleName, types.MergePatchType, data)
	return err
}

func newRoleForSecrets(secrets sets.String) *rbacv1.Role {
	return &rbacv1.Role{
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: secrets.List(),
			},
		},
	}
}

func roleHasInvalidElements(role *rbacv1.Role) bool {
	if len(role.Rules) != 1 {
		return true
	}
	for _, rule := range role.Rules {
		switch {
		case len(rule.NonResourceURLs) != 0,
			len(rule.Resources) != 1 || rule.Resources[0] != "secrets",
			len(rule.APIGroups) != 1 || rule.APIGroups[0] != "",
			len(rule.Verbs) != 1 || rule.Verbs[0] != "get",
			len(rule.ResourceNames) == 0:
			// the rule has unrecognized elements
			return true
		}
	}
	return false
}
