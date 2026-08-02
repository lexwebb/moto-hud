#!/usr/bin/env bash
# Configure macOS for a Raspberry Pi USB Ethernet gadget (rpi-usb-gadget SHARED).
#
# Why this exists: on macOS Sequoia/Tahoe, USB gadgets often have CHANNEL_IO and
# cannot join Internet Sharing's bridge100. Official "share Wi-Fi → gadget" fails
# silently (no 192.168.2.1). Use the Pi's SHARED address instead:
#   Pi:  10.12.194.1/28
#   Mac: 10.12.194.2/28 with NO router (so Wi-Fi stays your default route)
#
# Usage (Pi plugged in, "Raspberry Pi USB Gadget" visible in Network):
#   ./scripts/macos-usb-gadget.sh
#   ssh motohud@10.12.194.1
set -euo pipefail

SERVICE="Raspberry Pi USB Gadget"
IP="10.12.194.2"
MASK="255.255.255.240"
# 0.0.0.0 = no usable default gateway via the gadget (keeps Mac internet on Wi-Fi)
ROUTER="0.0.0.0"

if ! networksetup -listallnetworkservices | grep -Fxq "$SERVICE"; then
  echo "Network service '$SERVICE' not found." >&2
  echo "Plug the Pi into the OTG/data USB port and wait until macOS shows the gadget." >&2
  exit 1
fi

echo "==> static ${IP}/${MASK} on '${SERVICE}' (router ${ROUTER})"
networksetup -setmanual "$SERVICE" "$IP" "$MASK" "$ROUTER"

# Put Wi-Fi above the gadget so a mistaken router never steals the default route.
# Must include every enabled service (+ disabled Hotspot Shield if present).
mapfile_services() {
  networksetup -listallnetworkservices | awk '
    NR==1 && /asterisk/ { next }
    /^\*/ { gsub(/^\*/,""); disabled[++d]=$0; next }
    { enabled[++e]=$0 }
    END {
      # Wi-Fi first among enabled, gadget last, then disabled
      for (i=1;i<=e;i++) if (enabled[i]=="Wi-Fi") print enabled[i]
      for (i=1;i<=e;i++) if (enabled[i]!="Wi-Fi" && enabled[i]!="Raspberry Pi USB Gadget") print enabled[i]
      for (i=1;i<=e;i++) if (enabled[i]=="Raspberry Pi USB Gadget") print enabled[i]
      for (i=1;i<=d;i++) print disabled[i]
    }'
}

# bash 3.2 (macOS) — no mapfile; build argv
args=()
while IFS= read -r line; do
  [[ -n "$line" ]] && args+=("$line")
done < <(mapfile_services)

if [[ ${#args[@]} -gt 0 ]]; then
  echo "==> network service order (Wi-Fi first, gadget last)"
  networksetup -ordernetworkservices "${args[@]}" || echo "(service order unchanged; set Wi-Fi above the gadget in System Settings → Network)" >&2
fi

echo "==> current config"
networksetup -getinfo "$SERVICE"
echo ""
echo "Default route should stay on Wi-Fi:"
route -n get default 2>/dev/null | awk '/gateway:|interface:/{print}'
echo ""
echo "When the Pi finishes first-boot (usb0 up):"
echo "  ping -c2 10.12.194.1"
echo "  ssh motohud@10.12.194.1"
