#!/usr/bin/env bash
# Cross-compile motohud for Pi Zero / Zero W (linux/arm GOARM=6) via Docker.
#
# Works on macOS/Linux/Windows (Docker required). No on-Pi Go toolchain.
#
# Usage:
#   ./scripts/build-armv6.sh           # stub BLE (pure Go, fast)
#   ./scripts/build-armv6.sh --ble     # real BlueZ peripheral (CGO)
#   ./scripts/build-armv6.sh --ble --deploy 10.12.194.1
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BLE=0
DEPLOY_HOST=""
USER_NAME="motohud"
PASSWORD="motohud"
PLATFORM="linux/arm/v6"
OUT="$ROOT/bin/motohud-linux-armv6"

usage() {
  cat <<'EOF'
Usage: build-armv6.sh [--ble] [--deploy HOST] [--user NAME] [--password PASS]

  --ble           build with -tags ble (BlueZ / tinygo bluetooth, CGO)
  --deploy HOST   scp binary to HOST and restart motohud.service
  --user NAME     SSH user for --deploy (default: motohud)
  --password PASS sshpass password if sshpass is installed (default: motohud)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ble) BLE=1; shift ;;
    --deploy) DEPLOY_HOST="${2:-}"; shift 2 ;;
    --user) USER_NAME="${2:-}"; shift 2 ;;
    --password) PASSWORD="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown arg: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "docker daemon not running — start Docker Desktop / dockerd" >&2
  exit 1
fi

# Ensure buildx + qemu can target arm/v6 (idempotent).
docker buildx version >/dev/null
docker buildx inspect motohud-cross >/dev/null 2>&1 \
  || docker buildx create --name motohud-cross --driver docker-container --use >/dev/null
docker buildx use motohud-cross >/dev/null
docker run --privileged --rm tonistiigi/binfmt --install arm >/dev/null 2>&1 || true

mkdir -p "$ROOT/bin"
TMP_OUT="$ROOT/bin/.armv6-out"
rm -rf "$TMP_OUT"
mkdir -p "$TMP_OUT"

echo "==> docker buildx (${PLATFORM}) BLE=${BLE}"
docker buildx build \
  --builder motohud-cross \
  --platform "$PLATFORM" \
  --build-arg "BLE=${BLE}" \
  -f "$ROOT/pi/Dockerfile.cross" \
  --target export \
  -o "type=local,dest=${TMP_OUT}" \
  "$ROOT/pi"

cp "${TMP_OUT}/motohud-linux-armv6" "$OUT"
rm -rf "$TMP_OUT"
chmod +x "$OUT"
file "$OUT" || true
ls -lh "$OUT"
echo "==> wrote $OUT"

if [[ -n "$DEPLOY_HOST" ]]; then
  DEST="/home/${USER_NAME}/moto-hud/bin/motohud"
  echo "==> deploy ${USER_NAME}@${DEPLOY_HOST}:${DEST}"
  ssh_cmd=(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ServerAliveInterval=30)
  scp_cmd=(scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null)
  if command -v sshpass >/dev/null 2>&1; then
    ssh_cmd=(sshpass -p "$PASSWORD" "${ssh_cmd[@]}"
      -o PreferredAuthentications=password -o PubkeyAuthentication=no)
    scp_cmd=(sshpass -p "$PASSWORD" "${scp_cmd[@]}"
      -o PreferredAuthentications=password -o PubkeyAuthentication=no)
  fi
  "${scp_cmd[@]}" "$OUT" "${USER_NAME}@${DEPLOY_HOST}:/tmp/motohud.new"
  "${ssh_cmd[@]}" "${USER_NAME}@${DEPLOY_HOST}" bash -s <<REMOTE
set -euo pipefail
SUDO=sudo
sudo -n true 2>/dev/null || SUDO="echo ${PASSWORD} | sudo -S"
eval "\$SUDO install -m 755 /tmp/motohud.new ${DEST}"
eval "\$SUDO rfkill unblock bluetooth" || true
eval "\$SUDO bluetoothctl power on" || true
eval "\$SUDO systemctl restart motohud"
sleep 4
journalctl -u motohud -n 25 --no-pager || true
REMOTE
  echo "==> deployed"
fi
