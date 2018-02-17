package secretref

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/kubernetes/fake"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	rbaclisters "k8s.io/client-go/listers/rbac/v1"
	clientgotesting "k8s.io/client-go/testing"
)

func Test_roleHasInvalidElements(t *testing.T) {
	type args struct {
		role *rbacv1.Role
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "too short", args: args{role: &rbacv1.Role{}}, want: true},
		{name: "too long", args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{}, {}}}}, want: true},
		{name: "empty", args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{}}}}, want: true},
		{
			name: "valid",
			args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"test"},
			}}}},
			want: false,
		},
		{
			name: "wrong verb",
			args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"post"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"test"},
			}}}},
			want: true,
		},
		{
			name: "wrong group",
			args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{"core"},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"test"},
			}}}},
			want: true,
		},
		{
			name: "wrong resource",
			args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"newsecrets"},
				ResourceNames: []string{"test"},
			}}}},
			want: true,
		},
		{
			name: "no secrets",
			args: args{role: &rbacv1.Role{Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{},
			}}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roleHasInvalidElements(tt.args.role); got != tt.want {
				t.Errorf("roleHasInvalidElements() = %v, want %v", got, tt.want)
			}
		})
	}
}

type roleLister struct {
	Err   error
	Items []*rbacv1.Role
}

func (r *roleLister) List(selector labels.Selector) (ret []*rbacv1.Role, err error) {
	return r.Items, r.Err
}
func (r *roleLister) Roles(namespace string) rbaclisters.RoleNamespaceLister {
	return &nsRoleLister{r: r, ns: namespace}
}

type nsRoleLister struct {
	r  *roleLister
	ns string
}

func (r *nsRoleLister) List(selector labels.Selector) (ret []*rbacv1.Role, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsRoleLister) Get(name string) (*rbacv1.Role, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{}, name)
}

type ingressLister struct {
	Err   error
	Items []*extensionsv1beta1.Ingress
}

func (r *ingressLister) List(selector labels.Selector) (ret []*extensionsv1beta1.Ingress, err error) {
	return r.Items, r.Err
}
func (r *ingressLister) Ingresses(namespace string) extensionslisters.IngressNamespaceLister {
	return &nsIngressLister{r: r, ns: namespace}
}

type nsIngressLister struct {
	r  *ingressLister
	ns string
}

func (r *nsIngressLister) List(selector labels.Selector) (ret []*extensionsv1beta1.Ingress, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsIngressLister) Get(name string) (*extensionsv1beta1.Ingress, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{}, name)
}

type secretLister struct {
	Err   error
	Items []*v1.Secret
}

func (r *secretLister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	return r.Items, r.Err
}
func (r *secretLister) Secrets(namespace string) corelisters.SecretNamespaceLister {
	return &nsSecretLister{r: r, ns: namespace}
}

type nsSecretLister struct {
	r  *secretLister
	ns string
}

func (r *nsSecretLister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	return r.r.Items, r.r.Err
}
func (r *nsSecretLister) Get(name string) (*v1.Secret, error) {
	for _, s := range r.r.Items {
		if s.Name == name && r.ns == s.Namespace {
			return s, nil
		}
	}
	return nil, errors.NewNotFound(schema.GroupResource{}, name)
}

func TestController_sync(t *testing.T) {
	tlsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-1",
			Namespace: "test",
		},
		Type: v1.SecretTypeTLS,
	}
	opaqueSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-0",
			Namespace: "test",
		},
		Type: v1.SecretTypeOpaque,
	}
	secrets := &secretLister{Items: []*v1.Secret{opaqueSecret, tlsSecret}}

	type fields struct {
		i extensionslisters.IngressLister
		r rbaclisters.RoleLister
		s corelisters.SecretLister
	}
	type args struct {
		ns string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantRole   *rbacv1.Role
		wantPatch  string
		wantDelete bool
	}{
		{name: "no changes", fields: fields{i: &ingressLister{}, r: &roleLister{}}, args: args{ns: "test"}},
		{
			name: "create role",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-1"},
							},
						},
					},
				}},
				r: &roleLister{},
			},
			args: args{ns: "test"},
			wantRole: &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ingress-secretref",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"secrets"},
						ResourceNames: []string{"secret-1"},
					},
				},
			},
		},
		{
			name: "update role",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-1"},
							},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-0"},
						},
					},
				}}},
			},
			args:      args{ns: "test"},
			wantPatch: `{"rules":[{"verbs":["get"],"apiGroups":[""],"resources":["secrets"],"resourceNames":["secret-1"]}]}`,
		},
		{
			name: "no-op changes",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-1"},
							},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-1"},
						},
					},
				}}},
			},
			args: args{ns: "test"},
		},
		{
			name: "update role if rules have other things in them",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-1"},
							},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-1"},
						},
					},
				}}},
			},
			args:      args{ns: "test"},
			wantPatch: `{"rules":[{"verbs":["get"],"apiGroups":[""],"resources":["secrets"],"resourceNames":["secret-1"]}]}`,
		},
		{
			name: "delete role when no secrets referenced",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-1"},
						},
					},
				}}},
			},
			args:       args{ns: "test"},
			wantDelete: true,
		},
		{
			name: "delete role when secret is not TLS",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-0"},
							},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-0"},
						},
					},
				}}},
			},
			args:       args{ns: "test"},
			wantDelete: true,
		},
		{
			name: "delete role when secret does not exist",
			fields: fields{
				s: secrets,
				i: &ingressLister{Items: []*extensionsv1beta1.Ingress{
					&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "1",
							Namespace: "test",
						},
						Spec: extensionsv1beta1.IngressSpec{
							TLS: []extensionsv1beta1.IngressTLS{
								{SecretName: "secret-2"},
							},
						},
					},
				}},
				r: &roleLister{Items: []*rbacv1.Role{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ingress-secretref",
						Namespace: "test",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{"secret-2"},
						},
					},
				}}},
			},
			args:       args{ns: "test"},
			wantDelete: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &fake.Clientset{}
			kc.AddReactor("*", "roles", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			c := &Controller{
				roleName:      "ingress-secretref",
				client:        kc.Rbac(),
				ingressLister: tt.fields.i,
				roleLister:    tt.fields.r,
				secretLister:  tt.fields.s,
			}
			if err := c.sync(tt.args.ns); (err != nil) != tt.wantErr {
				t.Errorf("Controller.sync() error = %v, wantErr %v", err, tt.wantErr)
			}

			actions := kc.Actions()

			// delete is always the first action
			if tt.wantDelete {
				if len(actions) < 1 || actions[0].GetVerb() != "delete" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", kc.Actions())
				}
				action := actions[0].(clientgotesting.DeleteAction)
				if action.GetName() != c.roleName || action.GetNamespace() != tt.args.ns {
					t.Fatalf("unexpected action: %s", action)
				}
				actions = actions[1:]
			}

			switch {
			case tt.wantRole != nil:
				if len(actions) != 1 || actions[0].GetVerb() != "create" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", actions)
				}
				action := actions[0].(clientgotesting.CreateAction)
				r := action.GetObject().(*rbacv1.Role)
				if !reflect.DeepEqual(tt.wantRole, r) {
					t.Fatalf("unexpected create: %s", diff.ObjectReflectDiff(tt.wantRole, r))
				}
				if action.GetNamespace() != tt.args.ns {
					t.Fatalf("unexpected action: %s", action)
				}
			case len(tt.wantPatch) > 0:
				if len(actions) != 1 || actions[0].GetVerb() != "patch" {
					t.Fatalf("Controller.sync() unexpected actions: %#v", actions)
				}
				action := actions[0].(clientgotesting.PatchAction)
				patch := action.GetPatch()
				if !reflect.DeepEqual([]byte(tt.wantPatch), patch) {
					t.Fatalf("unexpected patch: %s", patch)
				}
				if action.GetName() != c.roleName || action.GetNamespace() != tt.args.ns {
					t.Fatalf("unexpected action: %s", action)
				}
			default:
				if len(actions) != 0 {
					t.Fatalf("Controller.sync() unexpected actions: %#v", actions)
				}
			}
		})
	}
}
