#!/usr/bin/env bash
# Flash Raspberry Pi OS Lite (32-bit / armhf) to an SD card, then customise bootfs.
#
# Default target: Pi Zero / Zero W — thin image, no desktop.
#
# Usage:
#   ./scripts/flash-lite.sh --disk disk12 \
#     --user motohud --password motohud \
#     --ssid 'WebbWideWeb' --psk 'iliketrains1'
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DISK=""
USER_NAME="motohud"
PASSWORD="motohud"
HOSTNAME="motohud"
SSID=""
PSK=""
COUNTRY="GB"
BWR=0
IMAGE_XZ="${ROOT}/out/images/2026-06-18-raspios-trixie-armhf-lite.img.xz"
IMAGE_URL="https://downloads.raspberrypi.com/raspios_lite_armhf/images/raspios_lite_armhf-2026-06-19/2026-06-18-raspios-trixie-armhf-lite.img.xz"

usage() {
  cat <<'EOF'
Usage: flash-lite.sh --disk diskN [options]

Options:
  --disk ID         diskutil identifier, e.g. disk12 (required; NOT rdisk)
  --image PATH      path to .img.xz (downloaded if missing)
  --user NAME       (default: motohud)
  --password PASS   (default: motohud)
  --hostname NAME   (default: motohud)
  --ssid SSID       Wi-Fi SSID (2.4 GHz on Zero W)
  --psk PASS        Wi-Fi password
  --country CODE    (default: GB)
  --bwr             Waveshare HAT (B) red/black/white panel
  -h, --help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --disk) DISK="${2:-}"; shift 2 ;;
    --image) IMAGE_XZ="${2:-}"; shift 2 ;;
    --user) USER_NAME="${2:-}"; shift 2 ;;
    --password) PASSWORD="${2:-}"; shift 2 ;;
    --hostname) HOSTNAME="${2:-}"; shift 2 ;;
    --ssid) SSID="${2:-}"; shift 2 ;;
    --psk) PSK="${2:-}"; shift 2 ;;
    --country) COUNTRY="${2:-}"; shift 2 ;;
    --bwr) BWR=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown arg: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$DISK" ]]; then
  echo "--disk is required (see: diskutil list)" >&2
  exit 2
fi
if [[ "$DISK" == */* || "$DISK" == rdisk* ]]; then
  echo "pass bare id like disk12 (script uses /dev/rdiskN for speed)" >&2
  exit 2
fi
if [[ ! -e "/dev/$DISK" ]]; then
  echo "/dev/$DISK not found" >&2
  exit 1
fi

# Refuse obvious system disks
case "$DISK" in
  disk0|disk1|disk2|disk3) echo "refusing to flash $DISK (looks like internal)" >&2; exit 1 ;;
esac

mkdir -p "$(dirname "$IMAGE_XZ")"
if [[ ! -f "$IMAGE_XZ" ]]; then
  echo "==> downloading Lite image"
  curl -L --fail --progress-bar -o "${IMAGE_XZ}.partial" "$IMAGE_URL"
  mv "${IMAGE_XZ}.partial" "$IMAGE_XZ"
fi
ls -lh "$IMAGE_XZ"

echo "==> unmounting /dev/$DISK"
diskutil unmountDisk "/dev/$DISK"

RDISK="/dev/r${DISK}"
echo "==> writing image to $RDISK (this wipes the card)"
# shellcheck disable=SC2024
xz -dc "$IMAGE_XZ" | sudo dd of="$RDISK" bs=4m status=progress
sync

echo "==> remounting bootfs"
sleep 2
diskutil mountDisk "/dev/$DISK" || true
# Wait for boot volume
BOOT=""
for _ in $(seq 1 30); do
  if [[ -f /Volumes/bootfs/config.txt ]]; then
    BOOT=/Volumes/bootfs
    break
  fi
  # Some images name it "boot"
  if [[ -f /Volumes/boot/config.txt ]]; then
    BOOT=/Volumes/boot
    break
  fi
  sleep 1
done
if [[ -z "$BOOT" ]]; then
  echo "boot partition did not appear under /Volumes; mount it and run prepare-bootfs.sh" >&2
  exit 1
fi

ARGS=(--boot "$BOOT" --user "$USER_NAME" --password "$PASSWORD" --hostname "$HOSTNAME")
if [[ -n "$SSID" ]]; then
  ARGS+=(--ssid "$SSID" --psk "$PSK" --country "$COUNTRY")
fi
if [[ "$BWR" -eq 1 ]]; then
  ARGS+=(--bwr)
fi
"$ROOT/scripts/prepare-bootfs.sh" "${ARGS[@]}"

diskutil eject "/dev/$DISK" || diskutil eject "$BOOT" || true
cat <<EOF

==> flash complete.

Hardware:
  • Pi Zero / Zero W: use the micro-USB DATA port (closest to HDMI), not PWR
  • Insert SD, plug into Mac USB — first boot runs firstrun then reboots (~2–4 min)

macOS USB (do NOT use Internet Sharing — broken for this gadget on Sequoia/Tahoe):
  ./scripts/macos-usb-gadget.sh
  ssh ${USER_NAME}@10.12.194.1

Wi-Fi / mDNS (after associate):
  ssh ${USER_NAME}@${HOSTNAME}.local

HTTP: http://10.12.194.1:8787/  or  http://${HOSTNAME}.local:8787/
EOF
