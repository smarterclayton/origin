// Code generated by lister-gen. DO NOT EDIT.

package internalversion

import (
	authorization "github.com/openshift/origin/pkg/authorization/apis/authorization"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterPolicyBindingLister helps list ClusterPolicyBindings.
type ClusterPolicyBindingLister interface {
	// List lists all ClusterPolicyBindings in the indexer.
	List(selector labels.Selector) (ret []*authorization.ClusterPolicyBinding, err error)
	// Get retrieves the ClusterPolicyBinding from the index for a given name.
	Get(name string) (*authorization.ClusterPolicyBinding, error)
	ClusterPolicyBindingListerExpansion
}

// clusterPolicyBindingLister implements the ClusterPolicyBindingLister interface.
type clusterPolicyBindingLister struct {
	indexer cache.Indexer
}

// NewClusterPolicyBindingLister returns a new ClusterPolicyBindingLister.
func NewClusterPolicyBindingLister(indexer cache.Indexer) ClusterPolicyBindingLister {
	return &clusterPolicyBindingLister{indexer: indexer}
}

// List lists all ClusterPolicyBindings in the indexer.
func (s *clusterPolicyBindingLister) List(selector labels.Selector) (ret []*authorization.ClusterPolicyBinding, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*authorization.ClusterPolicyBinding))
	})
	return ret, err
}

// Get retrieves the ClusterPolicyBinding from the index for a given name.
func (s *clusterPolicyBindingLister) Get(name string) (*authorization.ClusterPolicyBinding, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(authorization.Resource("clusterpolicybinding"), name)
	}
	return obj.(*authorization.ClusterPolicyBinding), nil
}
