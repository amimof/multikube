#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "$0")/lib.sh"

stop_port_forward

for cluster in "${E2E_KIND_MANAGEMENT_CLUSTER}" "${E2E_KIND_BACKEND_ONE}" "${E2E_KIND_BACKEND_TWO}"; do
  if kind_cluster_exists "${cluster}"; then
    log "deleting kind cluster ${cluster}"
    kind delete cluster --name "${cluster}"
  fi
done

if [[ "${E2E_KEEP:-0}" != "1" ]]; then
  log "removing e2e temporary artifacts"
  rm -rf "${E2E_TMP_DIR}"
fi

log "e2e teardown complete"
