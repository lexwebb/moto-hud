#!/usr/bin/env bash
# Start motohud (HTTP fallback) + open the browser emulator.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="/opt/homebrew/bin:$PATH"

"$ROOT/scripts/build-wasm.sh"

cd "$ROOT/pi"
go build -o "$ROOT/bin/motohud" ./cmd/motohud

pkill -f 'bin/motohud' 2>/dev/null || true
"$ROOT/bin/motohud" -host emu -demo -out "$ROOT/out/hud.png" -http :8787 \
  -assets "$ROOT/assets/hud" >/tmp/motohud-emu.log 2>&1 &
echo $! > /tmp/motohud-emu.pid
sleep 1
echo "Emulator: http://127.0.0.1:8787/emulator/"
echo "HUD still also at: http://127.0.0.1:8787/preview/"
echo "log: /tmp/motohud-emu.log"
