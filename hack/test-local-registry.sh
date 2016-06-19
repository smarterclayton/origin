#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

OS_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${OS_ROOT}/hack/lib/init.sh"

# Go to the top of the tree.
cd "${OS_ROOT}"

os::build::setup_env

hack/build-go.sh cmd/dockerregistry
dockerregistry="$( os::build::find-binary dockerregistry )"

# find the first builder service account token
token="$(oc get $(oc get secrets -o name | grep builder-token | head -n 1) --template '{{ .data.token }}' | base64 -D)"
echo
echo Login with:
echo   docker login -p "${token}" -u user IP:5000
echo

REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY="${REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY:-/tmp/registry}" \
  DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL:-localhost:5000}" \
	KUBECONFIG=openshift.local.config/master/openshift-registry.kubeconfig \
	${dockerregistry} images/dockerregistry/config.yml
