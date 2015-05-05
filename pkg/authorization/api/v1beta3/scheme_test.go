package v1beta3

import (
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/latest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/conversion"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

func TestEmbedded(t *testing.T) {
	// gob.Register(&authorizationapi.IsPersonalSubjectAccessReview{})

	rule := authorizationapi.PolicyRule{}
	rule.AttributeRestrictions = runtime.EmbeddedObject{&authorizationapi.IsPersonalSubjectAccessReview{}}

	role := &authorizationapi.Role{}
	role.Name = "role.name"
	role.Rules = append(role.Rules, rule)

	originalData, err := latest.Codec.Encode(role)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t.Logf("originalRole = %v\n", string(originalData))

	copyOfRole, err := conversion.DeepCopy(role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copiedData, err := latest.Codec.Encode(copyOfRole.(runtime.Object))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t.Logf("copyOfRole   = %v\n", string(copiedData))
}
