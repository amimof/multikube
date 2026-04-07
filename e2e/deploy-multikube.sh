#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "$0")/lib.sh"

ensure_dirs

log "deploying multikube into management cluster"
kubectl_mgmt create namespace "${E2E_NAMESPACE}" --dry-run=client -o yaml | kubectl_mgmt apply -f - >/dev/null
ensure_backend_secret
kubectl_mgmt apply -f "${E2E_ROOT}/manifests/multikube.yaml"
wait_for_deployment "${E2E_NAMESPACE}" multikube

start_port_forward
wait_for_proxy "https://127.0.0.1:${E2E_PROXY_PORT}/"
init_multikubectl_config

log "multikube deployment is ready"
