package fake

import (
	authorization "github.com/openshift/origin/pkg/authorization/apis/authorization"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterRoles implements ClusterRoleInterface
type FakeClusterRoles struct {
	Fake *FakeAuthorization
}

var clusterrolesResource = schema.GroupVersionResource{Group: "authorization.openshift.io", Version: "", Resource: "clusterroles"}

var clusterrolesKind = schema.GroupVersionKind{Group: "authorization.openshift.io", Version: "", Kind: "ClusterRole"}

func (c *FakeClusterRoles) Create(clusterRole *authorization.ClusterRole) (result *authorization.ClusterRole, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(clusterrolesResource, clusterRole), &authorization.ClusterRole{})
	if obj == nil {
		return nil, err
	}
	return obj.(*authorization.ClusterRole), err
}

func (c *FakeClusterRoles) Update(clusterRole *authorization.ClusterRole) (result *authorization.ClusterRole, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clusterrolesResource, clusterRole), &authorization.ClusterRole{})
	if obj == nil {
		return nil, err
	}
	return obj.(*authorization.ClusterRole), err
}

func (c *FakeClusterRoles) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(clusterrolesResource, name), &authorization.ClusterRole{})
	return err
}

func (c *FakeClusterRoles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(clusterrolesResource, listOptions)

	_, err := c.Fake.Invokes(action, &authorization.ClusterRoleList{})
	return err
}

func (c *FakeClusterRoles) Get(name string, options v1.GetOptions) (result *authorization.ClusterRole, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(clusterrolesResource, name), &authorization.ClusterRole{})
	if obj == nil {
		return nil, err
	}
	return obj.(*authorization.ClusterRole), err
}

func (c *FakeClusterRoles) List(opts v1.ListOptions) (result *authorization.ClusterRoleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(clusterrolesResource, clusterrolesKind, opts), &authorization.ClusterRoleList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &authorization.ClusterRoleList{}
	for _, item := range obj.(*authorization.ClusterRoleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterRoles.
func (c *FakeClusterRoles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(clusterrolesResource, opts))
}

// Patch applies the patch and returns the patched clusterRole.
func (c *FakeClusterRoles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *authorization.ClusterRole, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(clusterrolesResource, name, data, subresources...), &authorization.ClusterRole{})
	if obj == nil {
		return nil, err
	}
	return obj.(*authorization.ClusterRole), err
}
