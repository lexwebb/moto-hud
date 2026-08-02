#!/usr/bin/env bash
# Deprecated wrapper: BLE builds happen via Docker cross-compile now.
# Prefer: ./scripts/build-armv6.sh --ble --deploy 10.12.194.1
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HOST="10.12.194.1"
EXTRA=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) HOST="${2:-}"; shift 2 ;;
    *) EXTRA+=("$1"); shift ;;
  esac
done
echo "note: deploy-ble-pi.sh → build-armv6.sh --ble --deploy (Docker cross-compile)" >&2
exec "$ROOT/scripts/build-armv6.sh" --ble --deploy "$HOST" "${EXTRA[@]}"
