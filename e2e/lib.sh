#!/usr/bin/env bash

set -euo pipefail

E2E_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "${E2E_ROOT}/.." && pwd)

E2E_TMP_DIR=${E2E_TMP_DIR:-"${E2E_ROOT}/tmp"}
E2E_BIN_DIR=${E2E_BIN_DIR:-"${E2E_TMP_DIR}/bin"}
E2E_KUBECONFIG_DIR=${E2E_KUBECONFIG_DIR:-"${E2E_TMP_DIR}/kubeconfigs"}
E2E_ARTIFACT_DIR=${E2E_ARTIFACT_DIR:-"${E2E_TMP_DIR}/artifacts"}
E2E_CLIENT_CONFIG=${E2E_CLIENT_CONFIG:-"${E2E_TMP_DIR}/multikube.yaml"}

E2E_KIND_MANAGEMENT_CLUSTER=${E2E_KIND_MANAGEMENT_CLUSTER:-mk-mgmt}
E2E_KIND_BACKEND_ONE=${E2E_KIND_BACKEND_ONE:-mk-be-1}
E2E_KIND_BACKEND_TWO=${E2E_KIND_BACKEND_TWO:-mk-be-2}

E2E_NAMESPACE=${E2E_NAMESPACE:-multikube-system}
E2E_IMAGE=${E2E_IMAGE:-multikube-e2e:dev}

E2E_API_PORT=${E2E_API_PORT:-5743}
E2E_PROXY_PORT=${E2E_PROXY_PORT:-8443}

E2E_MANAGEMENT_KUBECONFIG=${E2E_MANAGEMENT_KUBECONFIG:-"${E2E_KUBECONFIG_DIR}/management.yaml"}
E2E_BACKEND_ONE_KUBECONFIG=${E2E_BACKEND_ONE_KUBECONFIG:-"${E2E_KUBECONFIG_DIR}/backend-one.yaml"}
E2E_BACKEND_TWO_KUBECONFIG=${E2E_BACKEND_TWO_KUBECONFIG:-"${E2E_KUBECONFIG_DIR}/backend-two.yaml"}
E2E_BACKENDS_KUBECONFIG=${E2E_BACKENDS_KUBECONFIG:-"${E2E_KUBECONFIG_DIR}/backends.yaml"}

E2E_PORT_FORWARD_PID_FILE=${E2E_PORT_FORWARD_PID_FILE:-"${E2E_TMP_DIR}/port-forward.pid"}
E2E_PORT_FORWARD_LOG=${E2E_PORT_FORWARD_LOG:-"${E2E_ARTIFACT_DIR}/port-forward.log"}

E2E_MULTIKUBECTL=${E2E_MULTIKUBECTL:-"${E2E_BIN_DIR}/multikubectl"}

log() {
  printf '==> %s\n' "$*"
}

fail() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

ensure_dirs() {
  mkdir -p "${E2E_TMP_DIR}" "${E2E_BIN_DIR}" "${E2E_KUBECONFIG_DIR}" "${E2E_ARTIFACT_DIR}"
}

kind_context() {
  printf 'kind-%s\n' "$1"
}

backend_name_for_context() {
  printf '%s-backend\n' "$1"
}

management_context() {
  kind_context "${E2E_KIND_MANAGEMENT_CLUSTER}"
}

backend_one_context() {
  kind_context "${E2E_KIND_BACKEND_ONE}"
}

backend_two_context() {
  kind_context "${E2E_KIND_BACKEND_TWO}"
}

kind_cluster_exists() {
  kind get clusters | grep -qx "$1"
}

create_kind_cluster() {
  local name=$1
  local config=$2

  if kind_cluster_exists "${name}"; then
    log "kind cluster ${name} already exists"
    return 0
  fi

  log "creating kind cluster ${name}"
  kind create cluster --name "${name}" --config "${config}"
}

export_kind_kubeconfig() {
  local cluster=$1
  local outfile=$2

  log "exporting kubeconfig for ${cluster}"
  kind get kubeconfig --name "${cluster}" >"${outfile}"
}

export_kind_internal_kubeconfig() {
  local cluster=$1
  local outfile=$2

  log "exporting internal kubeconfig for ${cluster}"
  kind get kubeconfig --internal --name "${cluster}" >"${outfile}"
}

combine_backend_kubeconfigs() {
  log "combining backend kubeconfigs"
  KUBECONFIG="${E2E_BACKEND_ONE_KUBECONFIG}:${E2E_BACKEND_TWO_KUBECONFIG}" \
    kubectl config view --flatten >"${E2E_BACKENDS_KUBECONFIG}"
}

build_multikubectl() {
  log "building multikubectl helper binary"
  go build -o "${E2E_MULTIKUBECTL}" ./cmd/multikubectl
}

build_e2e_image() {
  log "building local multikube image ${E2E_IMAGE}"
  docker build -t "${E2E_IMAGE}" .
}

load_image_into_management_cluster() {
  log "loading ${E2E_IMAGE} into ${E2E_KIND_MANAGEMENT_CLUSTER}"
  kind load docker-image "${E2E_IMAGE}" --name "${E2E_KIND_MANAGEMENT_CLUSTER}"
}

kubectl_mgmt() {
  kubectl --context "$(management_context)" "$@"
}

kubectl_backend_one() {
  kubectl --context "$(backend_one_context)" "$@"
}

kubectl_backend_two() {
  kubectl --context "$(backend_two_context)" "$@"
}

wait_for_deployment() {
  local namespace=$1
  local deployment=$2
  log "waiting for deployment/${deployment} in ${namespace}"
  kubectl_mgmt -n "${namespace}" rollout status deployment/"${deployment}" --timeout=180s
}

wait_for_proxy() {
  local url=$1
  local attempts=${2:-60}
  local sleep_seconds=${3:-2}

  log "waiting for proxy endpoint ${url}"
  for _ in $(seq 1 "${attempts}"); do
    if curl -sk --max-time 5 --output /dev/null "${url}"; then
      return 0
    fi
    sleep "${sleep_seconds}"
  done

  fail "proxy endpoint did not become reachable: ${url}"
}

stop_port_forward() {
  if [[ -f "${E2E_PORT_FORWARD_PID_FILE}" ]]; then
    local pid
    pid=$(cat "${E2E_PORT_FORWARD_PID_FILE}")
    if kill -0 "${pid}" >/dev/null 2>&1; then
      log "stopping existing port-forward ${pid}"
      kill "${pid}" >/dev/null 2>&1 || true
      wait "${pid}" 2>/dev/null || true
    fi
    rm -f "${E2E_PORT_FORWARD_PID_FILE}"
  fi
}

start_port_forward() {
  stop_port_forward
  ensure_dirs

  log "starting port-forward to multikube service"
  kubectl_mgmt -n "${E2E_NAMESPACE}" port-forward svc/multikube "${E2E_API_PORT}:${E2E_API_PORT}" "${E2E_PROXY_PORT}:${E2E_PROXY_PORT}" \
    >"${E2E_PORT_FORWARD_LOG}" 2>&1 &
  local pid=$!
  echo "${pid}" >"${E2E_PORT_FORWARD_PID_FILE}"
}

ensure_backend_secret() {
  log "refreshing backends kubeconfig secret in management cluster"
  kubectl_mgmt -n "${E2E_NAMESPACE}" create secret generic multikube-kubeconfig \
    --from-file=kubeconfig="${E2E_BACKENDS_KUBECONFIG}" \
    --dry-run=client -o yaml | kubectl_mgmt apply -f - >/dev/null
}

init_multikubectl_config() {
  log "initializing isolated multikubectl config"
  rm -f "${E2E_CLIENT_CONFIG}"
  "${E2E_MULTIKUBECTL}" --config "${E2E_CLIENT_CONFIG}" config init >/dev/null
  "${E2E_MULTIKUBECTL}" --config "${E2E_CLIENT_CONFIG}" config create-server e2e \
    --address "127.0.0.1:${E2E_API_PORT}" \
    --tls \
    --insecure \
    --current >/dev/null
}

mkctl() {
  "${E2E_MULTIKUBECTL}" --config "${E2E_CLIENT_CONFIG}" "$@"
}

ensure_namespace_present() {
  local context=$1
  local namespace=$2

  kubectl --context "${context}" create namespace "${namespace}" --dry-run=client -o yaml | kubectl --context "${context}" apply -f - >/dev/null
}
