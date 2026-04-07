#!/usr/bin/env bash

set -euo pipefail

source "$(dirname "$0")/lib.sh"

"${E2E_ROOT}/setup.sh"
"${E2E_ROOT}/deploy-multikube.sh"
"${E2E_ROOT}/tests/smoke.sh"

if [[ "${E2E_KEEP:-0}" != "1" ]]; then
  "${E2E_ROOT}/teardown.sh"
else
  log "keeping e2e environment because E2E_KEEP=1"
fi
