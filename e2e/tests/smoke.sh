#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "$0")/../lib.sh"

ensure_dirs

backend_one_ctx=$(backend_one_context)
backend_two_ctx=$(backend_two_context)
backend_one_name=$(backend_name_for_context "${backend_one_ctx}")
backend_two_name=$(backend_name_for_context "${backend_two_ctx}")

log "creating marker namespaces in backend clusters"
ensure_namespace_present "${backend_one_ctx}" e2e-backend-1
ensure_namespace_present "${backend_two_ctx}" e2e-backend-2

log "importing backend clusters into multikube"
mkctl import "${backend_one_ctx}" --kubeconfig "${E2E_BACKENDS_KUBECONFIG}" --force
mkctl import "${backend_two_ctx}" --kubeconfig "${E2E_BACKENDS_KUBECONFIG}" --force

log "verifying imported backends are visible"
backends_output=$(mkctl get backends)
printf '%s\n' "${backends_output}" >"${E2E_ARTIFACT_DIR}/backends.txt"
printf '%s\n' "${backends_output}" | grep -q "${backend_one_name}" || fail "backend ${backend_one_name} not found in output"
printf '%s\n' "${backends_output}" | grep -q "${backend_two_name}" || fail "backend ${backend_two_name} not found in output"

log "creating header-based routes"
if ! mkctl get routes e2e-route-one >/dev/null 2>&1; then
  mkctl create route e2e-route-one --backend-ref "${backend_one_name}" --header-name X-Cluster --header-value one
fi
if ! mkctl get routes e2e-route-two >/dev/null 2>&1; then
  mkctl create route e2e-route-two --backend-ref "${backend_two_name}" --header-name X-Cluster --header-value two
fi

bearer_token=$(mkctl create token --subject "e2e-tester")

log "verifying proxy can reach backend one"
curl -sk -H "Authorization: Bearer ${bearer_token}" -H 'X-Cluster: one' "https://127.0.0.1:${E2E_PROXY_PORT}/api/v1/namespaces/e2e-backend-1" >"${E2E_ARTIFACT_DIR}/backend-one.json"
grep -q '"name": "e2e-backend-1"' "${E2E_ARTIFACT_DIR}/backend-one.json" || fail "backend one response did not include expected namespace"

log "verifying proxy can reach backend two"
curl -sk -H "Authorization: Bearer ${bearer_token}" -H 'X-Cluster: two' "https://127.0.0.1:${E2E_PROXY_PORT}/api/v1/namespaces/e2e-backend-2" >"${E2E_ARTIFACT_DIR}/backend-two.json"
grep -q '"name": "e2e-backend-2"' "${E2E_ARTIFACT_DIR}/backend-two.json" || fail "backend two response did not include expected namespace"

log "smoke test passed"
