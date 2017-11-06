#!/bin/bash

# This script sets up a go workspace locally and builds all go components.
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

OS_BUILD_RELEASE_ARCHIVES=n \
  OS_ONLY_BUILD_PLATFORMS="$(os::build::host_platform)" \
  "${OS_ROOT}/hack/build-cross.sh" "$@"