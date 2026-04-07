#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "$0")/lib.sh"

log "verifying e2e dependencies"
require_cmd kind
require_cmd kubectl
require_cmd docker
require_cmd go
require_cmd curl

ensure_dirs
build_multikubectl
build_e2e_image

create_kind_cluster "${E2E_KIND_MANAGEMENT_CLUSTER}" "${E2E_ROOT}/manifests/kind-management.yaml"
create_kind_cluster "${E2E_KIND_BACKEND_ONE}" "${E2E_ROOT}/manifests/kind-workload.yaml"
create_kind_cluster "${E2E_KIND_BACKEND_TWO}" "${E2E_ROOT}/manifests/kind-workload.yaml"

export_kind_kubeconfig "${E2E_KIND_MANAGEMENT_CLUSTER}" "${E2E_MANAGEMENT_KUBECONFIG}"
export_kind_internal_kubeconfig "${E2E_KIND_BACKEND_ONE}" "${E2E_BACKEND_ONE_KUBECONFIG}"
export_kind_internal_kubeconfig "${E2E_KIND_BACKEND_TWO}" "${E2E_BACKEND_TWO_KUBECONFIG}"
combine_backend_kubeconfigs

load_image_into_management_cluster

log "e2e setup complete"
