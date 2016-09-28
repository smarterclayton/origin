#!/bin/bash

# Script to create latest swagger spec.
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

function cleanup() {
    out=$?
    cleanup_openshift

    if [ $out -ne 0 ]; then
        echo "[FAIL] !!!!! Generate Failed !!!!"
        echo
        tail -100 "${LOG_DIR}/openshift.log"
        echo
        echo -------------------------------------
        echo
    fi
    exit $out
}

trap "exit" INT TERM
trap "cleanup" EXIT

export ALL_IP_ADDRESSES=127.0.0.1
export SERVER_HOSTNAME_LIST=127.0.0.1
export API_BIND_HOST=127.0.0.1
export API_PORT=38443
export ETCD_PORT=34001
export ETCD_PEER_PORT=37001
os::util::environment::setup_all_server_vars "generate-swagger-spec/"
reset_tmp_dir
configure_os_server


SWAGGER_SPEC_REL_DIR="${1:-}"
SWAGGER_SPEC_OUT_DIR="${OS_ROOT}/${SWAGGER_SPEC_REL_DIR}/api/swagger-spec"
mkdir -p "${SWAGGER_SPEC_OUT_DIR}"
SWAGGER_API_PATH="${MASTER_ADDR}/swaggerapi/"

# Start openshift
start_os_master

os::log::info "Updating ${SWAGGER_SPEC_OUT_DIR}:"

endpoint_types=("oapi" "api")
for type in "${endpoint_types[@]}"; do
    endpoints=("v1")
    for endpoint in "${endpoints[@]}"; do
        os::log::info "Updating ${SWAGGER_SPEC_OUT_DIR}/${type}-${endpoint}.json from ${SWAGGER_API_PATH}${type}/${endpoint}..."
        curl -w "\n" "${SWAGGER_API_PATH}${type}/${endpoint}" > "${SWAGGER_SPEC_OUT_DIR}/${type}-${endpoint}.json"

        os::util::sed 's|https://127.0.0.1:38443|https://127.0.0.1:8443|g' "${SWAGGER_SPEC_OUT_DIR}/${type}-${endpoint}.json"
    done
done

# Swagger 2.0 / OpenAPI docs
curl -w "\n" "${MASTER_ADDR}/swagger.json" > "${SWAGGER_SPEC_OUT_DIR}/openshift-openapi-spec.json"
os::util::sed 's|https://127.0.0.1:38443|https://127.0.0.1:8443|g' "${SWAGGER_SPEC_OUT_DIR}/openshift-openapi-spec.json"
os::util::sed -r 's|"version": "[^\"]+"|"version: "latest"|g' "${SWAGGER_SPEC_OUT_DIR}/openshift-openapi-spec.json"

# Copy all protobuf generated specs into the api/protobuf-spec directory
PROTO_SPEC_OUT_DIR="${OS_ROOT}/${SWAGGER_SPEC_REL_DIR}/api/protobuf-spec"
mkdir -p "${PROTO_SPEC_OUT_DIR}"
find "${OS_ROOT}/pkg" "${OS_ROOT}/vendor/k8s.io/kubernetes/pkg" -name generated.proto | \
  xargs grep -E '^package' | \
  sed -rn 's/(.+)\:package (.+);/\1\n\2/p' | \
  xargs -n 2 bash -c 'cp "$1" "$0/$( echo $2 | sed -rn "s/\./_/pg" ).proto"' "${PROTO_SPEC_OUT_DIR}"

os::log::info "SUCCESS"
