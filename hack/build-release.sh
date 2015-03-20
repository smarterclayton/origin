#!/bin/bash

# This script generates release zips into _output/releases. It requires the openshift/origin-release
# image to be built prior to executing this command via hack/build-base-images.sh.

set -o errexit
set -o nounset
set -o pipefail

OS_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${OS_ROOT}/hack/common.sh"

# Go to the top of the tree.
cd "${OS_ROOT}"

context="${OS_ROOT}/_output/buildenv-context"

# Clean existing output.
rm -rf "${OS_ROOT}/_output/local/releases"
rm -rf "${OS_ROOT}/_output/local/go/bin"
rm -rf "${context}"
mkdir -p "${context}"
mkdir -p "${OS_ROOT}/_output/local"

# Generate version definitions.
os::build::get_version_vars
os::build::save_version_vars "${context}/os-version-defs"

# Create the input archive.
git archive --format=tar -o "${context}/archive.tar" HEAD
tar -rf "${context}/archive.tar" -C "${context}" os-version-defs
gzip -f "${context}/archive.tar"

# Perform the build and release in Docker.
cat "${context}/archive.tar.gz" | docker run -i --cidfile="${context}/cid" openshift/origin-release
docker cp $(cat ${context}/cid):/go/src/github.com/openshift/origin/_output/local/releases "${OS_ROOT}/_output/local"
echo "${OS_GIT_COMMIT}" > "${OS_ROOT}/_output/local/releases/.commit"

# Copy the linux release archives release back to the local _output/local/go/bin directory.
os::build::detect_local_release_tars "linux"

mkdir -p "${OS_LOCAL_BINPATH}"
tar mxzf "${OS_PRIMARY_RELEASE_TAR}" -C "${OS_LOCAL_BINPATH}"
tar mxzf "${OS_IMAGE_RELEASE_TAR}" -C "${OS_LOCAL_BINPATH}"

os::build::make_openshift_binary_symlinks
