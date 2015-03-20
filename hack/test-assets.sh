#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

OS_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${OS_ROOT}/hack/common.sh"

# Go to the top of the tree.
cd "${OS_ROOT}"

# If we are running inside of Travis then do not run the rest of this
# script unless we want to TEST_ASSETS
if [[ "${TRAVIS-}" == "true" && "${TEST_ASSETS-}" == "false" ]]; then
  exit
fi

pushd "${OS_ROOT}/assets" > /dev/null
  grunt test
  grunt build
popd > /dev/null

pushd "${OS_ROOT}" > /dev/null
  Godeps/_workspace/bin/go-bindata -nocompress -prefix "assets/dist" -pkg "assets" -o "_output/test/assets/bindata.go" -ignore "\\.gitignore" assets/dist/...
  echo "Validating checked in bindata.go is up to date..."
  if ! diff _output/test/assets/bindata.go pkg/assets/bindata.go ; then

    pushd "${OS_ROOT}/assets" > /dev/null

      if [[ "${TRAVIS-}" == "true" ]]; then
        echo ""
        echo "Bower versions..."
        bower list -o

        echo ""
        echo "NPM versions..."
        npm list
      fi

    popd > /dev/null  

    exit 1
  fi
popd > /dev/null