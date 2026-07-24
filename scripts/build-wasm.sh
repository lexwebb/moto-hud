#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="/opt/homebrew/bin:$PATH"
cd "$ROOT/pi"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$ROOT/web/emulator/wasm_exec.js"
GOOS=js GOARCH=wasm go build -o "$ROOT/web/emulator/motohud.wasm" ./cmd/motohud-wasm
ls -lh "$ROOT/web/emulator/motohud.wasm"
echo "wrote web/emulator/motohud.wasm"
