// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/openshift/api/authorization/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeSubjectRulesReviews implements SubjectRulesReviewInterface
type FakeSubjectRulesReviews struct {
	Fake *FakeAuthorizationV1
	ns   string
}

var subjectrulesreviewsResource = schema.GroupVersionResource{Group: "authorization.openshift.io", Version: "v1", Resource: "subjectrulesreviews"}

var subjectrulesreviewsKind = schema.GroupVersionKind{Group: "authorization.openshift.io", Version: "v1", Kind: "SubjectRulesReview"}

// Create takes the representation of a subjectRulesReview and creates it.  Returns the server's representation of the subjectRulesReview, and an error, if there is any.
func (c *FakeSubjectRulesReviews) Create(subjectRulesReview *v1.SubjectRulesReview) (result *v1.SubjectRulesReview, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(subjectrulesreviewsResource, c.ns, subjectRulesReview), &v1.SubjectRulesReview{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.SubjectRulesReview), err
}
