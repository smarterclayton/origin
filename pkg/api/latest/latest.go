package latest

import (
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	_ "github.com/openshift/origin/pkg/api"
	"github.com/openshift/origin/pkg/api/v1"
	"github.com/openshift/origin/pkg/api/v1beta3"
)

// Version is the string that represents the current external default version.
var Version = v1.SchemeGroupVersion

// OldestVersion is the string that represents the oldest server version supported,
// for client code that wants to hardcode the lowest common denominator.
var OldestVersion = v1beta3.SchemeGroupVersion

// Versions is the list of versions that are recognized in code. The order provided
// may be assumed to be most preferred to least preferred, and clients may
// choose to prefer the earlier items in the list over the latter items when presented
// with a set of versions to choose.
var Versions = []unversioned.GroupVersion{v1.SchemeGroupVersion, v1beta3.SchemeGroupVersion}

// originTypes are the hardcoded types defined by the OpenShift API.
var originTypes = make(map[unversioned.GroupVersionKind]bool)

// UserResources are the resource names that apply to the primary, user facing resources used by
// client tools. They are in deletion-first order - dependent resources should be last.
var UserResources = []string{
	"buildConfigs", "builds",
	"imageStreams",
	"deploymentConfigs", "replicationControllers",
	"routes", "services",
	"pods",
}

// OriginKind returns true if OpenShift owns the GroupVersionKind.
func OriginKind(gvk unversioned.GroupVersionKind) bool {
	return originTypes[gvk]
}

// IsKindInAnyOriginGroup returns true if OpenShift owns the kind described in any apiVersion.
// TODO: this may not work once we divide builds/deployments/images into their own API groups
func IsKindInAnyOriginGroup(kind string) bool {
	for _, version := range Versions {
		if OriginKind(version.WithKind(kind)) {
			return true
		}
	}

	return false
}

func init() {

	// enumerate all supported versions, get the kinds, and register with the mapper how to address our resources
	for _, version := range Versions {
		for kind, t := range api.Scheme.KnownTypes(version) {
			if !strings.Contains(t.PkgPath(), "openshift/origin") {
				continue
			}
			gvk := version.WithKind(kind)
			originTypes[gvk] = true
		}
	}
}
