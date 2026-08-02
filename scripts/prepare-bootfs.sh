#!/usr/bin/env bash
# Cross-compile motohud for Pi Zero / Zero W and customise a mounted Raspberry Pi
# OS Lite (Trixie) boot partition for Waveshare e-ink + Wi-Fi + USB Ethernet gadget.
#
# Learnings baked in:
#   - Trixie uses NetworkManager + cloud-init (network-config), not wpa_supplicant.conf alone
#   - macOS Sequoia/Tahoe cannot Internet-Share to CHANNEL_IO USB gadgets (no bridge100);
#     use Pi SHARED mode at 10.12.194.1/28 and a Mac static IP without a default gateway
#   - Official enable: rpi.enable_usb_gadget + dtoverlay=dwc2,dr_mode=peripheral
#
# Usage:
#   ./scripts/prepare-bootfs.sh --boot /Volumes/bootfs \
#     --user motohud --password motohud \
#     --ssid 'YourWifi' --psk '…'
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BOOT=""
USER_NAME="motohud"
PASSWORD=""
HOSTNAME="motohud"
HOST_KIND="waveshare"
SSID=""
PSK=""
COUNTRY="GB"
DRY_RUN=0
USB_ADDR="10.12.194.1/28"
BWR=0
DEMO=1
BLE=1

usage() {
  cat <<'EOF'
Usage: prepare-bootfs.sh --boot /Volumes/bootfs --password '…' [options]

Options:
  --boot PATH       Mounted Raspberry Pi boot partition (required)
  --user NAME       Linux username (default: motohud)
  --password PASS   Login password (required)
  --hostname NAME   Hostname (default: motohud)
  --host KIND       motohud -host value (default: waveshare)
  --ssid SSID       Wi-Fi network name (2.4 GHz; Zero W has no 5 GHz)
  --psk PASS        Wi-Fi password
  --country CODE    Wi-Fi country code (default: GB)
  --bwr             Waveshare 2.13" HAT (B) black/white/red (sets MOTOHUD_WAVESHARE_BWR=1)
  --no-demo         Omit -demo on first boot (default includes a static nav frame)
  --no-ble          Stub BLE only (no Docker / BlueZ CGO); default builds real BLE via Docker
  --dry-run         Build + stage only; do not write bootfs
  -h, --help        Show this help
EOF
}

yaml_quote() {
  # Double-quoted YAML scalar with minimal escapes.
  local s=$1
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  printf '"%s"' "$s"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --boot) BOOT="${2:-}"; shift 2 ;;
    --user) USER_NAME="${2:-}"; shift 2 ;;
    --password) PASSWORD="${2:-}"; shift 2 ;;
    --hostname) HOSTNAME="${2:-}"; shift 2 ;;
    --host) HOST_KIND="${2:-}"; shift 2 ;;
    --ssid) SSID="${2:-}"; shift 2 ;;
    --psk) PSK="${2:-}"; shift 2 ;;
    --country) COUNTRY="${2:-}"; shift 2 ;;
    --bwr) BWR=1; shift ;;
    --no-demo) DEMO=0; shift ;;
    --no-ble) BLE=0; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown arg: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$BOOT" ]]; then
  echo "--boot is required" >&2
  usage >&2
  exit 2
fi
if [[ -z "$PASSWORD" ]]; then
  echo "--password is required" >&2
  usage >&2
  exit 2
fi
if [[ ! -f "$BOOT/config.txt" || ! -f "$BOOT/cmdline.txt" ]]; then
  echo "bootfs looks wrong: need config.txt + cmdline.txt under $BOOT" >&2
  exit 1
fi
if [[ -n "$SSID" && -z "$PSK" ]]; then
  echo "--psk is required when --ssid is set" >&2
  exit 2
fi

export PATH="/opt/homebrew/bin:$PATH"
BIN_OUT="$ROOT/bin/motohud-linux-armv6"
STAGE="$ROOT/out/bootfs-stage"
PAYLOAD="$STAGE/moto-hud-payload"
INSTANCE_ID="motohud-$(date -u +%Y%m%d%H%M%S)-$$"

mkdir -p "$ROOT/bin" "$PAYLOAD/bin" "$PAYLOAD/assets/hud"
if [[ "$BLE" -eq 1 ]]; then
  echo "==> cross-compiling motohud (linux/armv6 + BLE via Docker)"
  "$ROOT/scripts/build-armv6.sh" --ble
else
  echo "==> cross-compiling motohud (linux/arm GOARM=6, stub BLE, no CGO)"
  (
    cd "$ROOT/pi"
    CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 \
      go build -ldflags='-s -w' -o "$BIN_OUT" ./cmd/motohud
  )
fi
cp "$BIN_OUT" "$PAYLOAD/bin/motohud"
cp "$ROOT/assets/hud/"*.svg "$PAYLOAD/assets/hud/" 2>/dev/null || true
mkdir -p "$PAYLOAD/assets/fonts/terminus"
cp "$ROOT/assets/fonts/terminus/"*.bdf "$PAYLOAD/assets/fonts/terminus/" 2>/dev/null || true
cp "$ROOT/assets/fonts/terminus/OFL.TXT" "$PAYLOAD/assets/fonts/terminus/" 2>/dev/null || true

DEMO_FLAG=""
if [[ "$DEMO" -eq 1 ]]; then
  DEMO_FLAG=" -demo"
fi
cat >"$PAYLOAD/motohud.service" <<EOF
[Unit]
Description=Moto HUD e-ink service
After=bluetooth.target network.target
Wants=bluetooth.target

[Service]
Type=simple
User=${USER_NAME}
WorkingDirectory=/home/${USER_NAME}/moto-hud
# 1 = Waveshare 2.13" HAT (B) BWR; 0 = pure B/W HAT (preferred for nav speed)
Environment=MOTOHUD_WAVESHARE_BWR=${BWR}
ExecStart=/home/${USER_NAME}/moto-hud/bin/motohud -host ${HOST_KIND}${DEMO_FLAG} -out /tmp/hud.png -http :8787 -assets /home/${USER_NAME}/moto-hud/assets/hud
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

cat >"$PAYLOAD/install.sh" <<EOF
#!/bin/bash
# Runs on the Pi from firstrun.sh / cloud-init.
set -euo pipefail
USER_NAME="${USER_NAME}"
HOSTNAME="${HOSTNAME}"
HOME_DIR="/home/\${USER_NAME}"
DEST="\${HOME_DIR}/moto-hud"
SRC=""
for candidate in /boot/firmware/moto-hud-payload /boot/moto-hud-payload; do
  if [[ -d "\$candidate" ]]; then
    SRC="\$candidate"
    break
  fi
done
if [[ -z "\$SRC" ]]; then
  echo "moto-hud payload not found under /boot" >&2
  exit 1
fi

install -d -o "\$USER_NAME" -g "\$USER_NAME" "\$DEST/bin" "\$DEST/assets"
cp -a "\$SRC/bin/." "\$DEST/bin/"
cp -a "\$SRC/assets/." "\$DEST/assets/"
chown -R "\$USER_NAME:\$USER_NAME" "\$DEST"
chmod 755 "\$DEST/bin/motohud"

usermod -aG spi,gpio,i2c,bluetooth "\$USER_NAME" 2>/dev/null || true
raspi-config nonint do_spi 0 || true
hostnamectl set-hostname "\$HOSTNAME" 2>/dev/null || echo "\$HOSTNAME" >/etc/hostname
if [[ -f /etc/hosts ]]; then
  grep -qE "[[:space:]]\$HOSTNAME\$" /etc/hosts 2>/dev/null || echo "10.0.0.2\t\$HOSTNAME" >>/etc/hosts
fi

# Passwordless sudo for headless setup scripts
printf '%s ALL=(ALL) NOPASSWD:ALL\n' "\$USER_NAME" >/etc/sudoers.d/010-\${USER_NAME}-nopasswd
chmod 440 /etc/sudoers.d/010-\${USER_NAME}-nopasswd

install -m 644 "\$SRC/motohud.service" /etc/systemd/system/motohud.service
systemctl daemon-reload
systemctl enable --now motohud.service
EOF
chmod +x "$PAYLOAD/install.sh"

HASH="$(openssl passwd -6 "$PASSWORD")"
FIRSTRUN="$STAGE/firstrun.sh"

# --- firstrun: runs once via systemd.run, then reboots and strips itself ---
{
  cat <<'HDR'
#!/bin/bash
set +e
exec >/tmp/motohud-firstrun.log 2>&1
echo "motohud firstrun $(date -u +%Y-%m-%dT%H:%M:%SZ)"

systemctl enable --now ssh 2>/dev/null || systemctl enable --now sshd 2>/dev/null || true
install -d /etc/ssh/sshd_config.d
printf 'PasswordAuthentication yes\n' >/etc/ssh/sshd_config.d/99-motohud-pw.conf
systemctl restart ssh 2>/dev/null || systemctl restart sshd 2>/dev/null || true

# Official USB Ethernet gadget (Trixie+)
command -v rpi-usb-gadget >/dev/null && rpi-usb-gadget on || true
modprobe dwc2 2>/dev/null || true
modprobe g_ether 2>/dev/null || true

# Wait for usb0, prefer SHARED profile, always pin static for macOS hosts
for _ in $(seq 1 45); do
  ip link show usb0 >/dev/null 2>&1 && break
  sleep 1
done
nmcli con down "USB Gadget (client)" 2>/dev/null || true
nmcli con up "USB Gadget (shared)" 2>/dev/null || true
ip link set usb0 up 2>/dev/null || true
ip addr replace 10.12.194.1/28 dev usb0 2>/dev/null || true
if ! nmcli -t -f NAME con show | grep -qx 'motohud-usb0'; then
  nmcli con add type ethernet ifname usb0 con-name motohud-usb0 \
    ipv4.method manual ipv4.addresses 10.12.194.1/28 ipv6.method ignore \
    connection.autoconnect yes 2>/dev/null || true
fi
nmcli con up motohud-usb0 2>/dev/null || true
ip addr replace 10.12.194.1/28 dev usb0 2>/dev/null || true

rfkill unblock wifi 2>/dev/null || true
nmcli radio wifi on 2>/dev/null || true
nmcli device set wlan0 managed yes 2>/dev/null || true

# Wait for Wi-Fi radio / firmware
for _ in $(seq 1 60); do
  ip link show wlan0 >/dev/null 2>&1 && break
  sleep 1
done
ip link set wlan0 up 2>/dev/null || true
HDR

  if [[ -n "$SSID" && -n "$PSK" ]]; then
    cat <<EOF
# Persist Wi-Fi via NetworkManager (Trixie ignores boot wpa_supplicant.conf)
SSID=$(printf '%q' "$SSID")
PSK=$(printf '%q' "$PSK")
for attempt in 1 2 3 4 5 6 7 8; do
  echo "wifi connect attempt \$attempt"
  nmcli device wifi rescan ifname wlan0 2>/dev/null || true
  sleep 2
  if nmcli -t -f NAME con show | grep -qx 'moto-hud-wifi'; then
    nmcli con up moto-hud-wifi ifname wlan0 && break
  fi
  nmcli device wifi connect "\$SSID" password "\$PSK" ifname wlan0 name moto-hud-wifi \\
    && break
  nmcli device wifi connect "\$SSID" password "\$PSK" name moto-hud-wifi \\
    && break
  sleep 5
done
nmcli con modify moto-hud-wifi connection.autoconnect yes 2>/dev/null || true
nmcli con modify moto-hud-wifi connection.autoconnect-priority 10 2>/dev/null || true
nmcli -f DEVICE,STATE,CONNECTION device status || true
ip -br a || true
EOF
  fi

  cat <<'EOF'
if [[ -x /boot/firmware/moto-hud-payload/install.sh ]]; then
  /boot/firmware/moto-hud-payload/install.sh
elif [[ -x /boot/moto-hud-payload/install.sh ]]; then
  /boot/moto-hud-payload/install.sh
fi

systemctl enable --now avahi-daemon 2>/dev/null || true

rm -f /boot/firstrun.sh /boot/firmware/firstrun.sh
for cmdfile in /boot/firmware/cmdline.txt /boot/cmdline.txt; do
  if [[ -f "$cmdfile" ]]; then
    sed -i 's| systemd.run=[^ ]*||g; s| systemd.run_success_action=[^ ]*||g; s| systemd.unit=[^ ]*||g' "$cmdfile"
  fi
done
sync
echo "firstrun done"
exit 0
EOF
} >"$FIRSTRUN"
chmod +x "$FIRSTRUN"

write_boot() {
  local dest="$1"
  echo "==> writing payload → $dest"
  rm -rf "$dest/moto-hud-payload"
  cp -R "$PAYLOAD" "$dest/moto-hud-payload"
  cp "$FIRSTRUN" "$dest/firstrun.sh"
  chmod +x "$dest/firstrun.sh" "$dest/moto-hud-payload/install.sh"
  find "$dest/moto-hud-payload" "$dest/firstrun.sh" -name '._*' -delete 2>/dev/null || true

  printf '%s:%s\n' "$USER_NAME" "$HASH" >"$dest/userconf.txt"
  : >"$dest/ssh"

  # Fresh cloud-init instance every prepare (re-apply on already-booted cards)
  cat >"$dest/meta-data" <<EOF
#cloud-init meta-data
dsmode: local
instance_id: ${INSTANCE_ID}
EOF

  # NetworkManager via cloud-init: static usb0 (macOS path) + optional Wi-Fi
  {
    cat <<EOF
network:
  version: 2
  ethernets:
    usb0:
      dhcp4: false
      addresses:
        - ${USB_ADDR}
      optional: true
EOF
    if [[ -n "$SSID" && -n "$PSK" ]]; then
      cat <<EOF
  wifis:
    wlan0:
      dhcp4: true
      optional: true
      regulatory-domain: ${COUNTRY}
      access-points:
        $(yaml_quote "$SSID"):
          password: $(yaml_quote "$PSK")
EOF
    fi
  } >"$dest/network-config"

  # Full user-data (replace stock) — gadget + SSH user + install hooks
  {
    cat <<EOF
#cloud-config
hostname: ${HOSTNAME}
manage_etc_hosts: true
ssh_pwauth: true
enable_ssh: true

rpi:
  enable_usb_gadget: true

users:
  - name: ${USER_NAME}
    gecos: Moto HUD
    primary_group: users
    groups: [sudo, spi, gpio, i2c, bluetooth, adm]
    shell: /bin/bash
    lock_passwd: false
    passwd: ${HASH}
    sudo: ALL=(ALL) NOPASSWD:ALL

packages:
  - avahi-daemon
EOF
    if [[ -n "$SSID" && -n "$PSK" ]]; then
      # Belt-and-suspenders NM keyfile (literal SSID/PSK; no shell expansion in content)
      {
        printf '\nwrite_files:\n'
        printf '  - path: /etc/NetworkManager/system-connections/moto-hud-wifi.nmconnection\n'
        printf '    permissions: "0600"\n'
        printf '    content: |\n'
        printf '      [connection]\n'
        printf '      id=moto-hud-wifi\n'
        printf '      type=wifi\n'
        printf '      autoconnect=true\n'
        printf '      autoconnect-priority=10\n'
        printf '\n'
        printf '      [wifi]\n'
        printf '      mode=infrastructure\n'
        printf '      ssid=%s\n' "$SSID"
        printf '\n'
        printf '      [wifi-security]\n'
        printf '      key-mgmt=wpa-psk\n'
        printf '      psk=%s\n' "$PSK"
        printf '\n'
        printf '      [ipv4]\n'
        printf '      method=auto\n'
        printf '\n'
        printf '      [ipv6]\n'
        printf '      method=auto\n'
      }
    fi
    cat <<'EOF'

runcmd:
  - [ bash, -lc, "command -v rpi-usb-gadget >/dev/null && rpi-usb-gadget on || true" ]
  - [ bash, -lc, "nmcli con down 'USB Gadget (client)' 2>/dev/null || true; nmcli con up 'USB Gadget (shared)' 2>/dev/null || true" ]
  - [ bash, -lc, "ip link set usb0 up 2>/dev/null; ip addr replace 10.12.194.1/28 dev usb0 2>/dev/null || true" ]
  - [ bash, -lc, "nmcli con up moto-hud-wifi 2>/dev/null || true" ]
  - [ bash, -lc, "test -x /boot/firmware/moto-hud-payload/install.sh && /boot/firmware/moto-hud-payload/install.sh || true" ]
  - [ systemctl, enable, --now, avahi-daemon ]
  - [ systemctl, enable, --now, ssh ]
EOF
  } >"$dest/user-data"

  # Legacy marker (ignored by Trixie NM, kept for older images / debugging)
  if [[ -n "$SSID" && -n "$PSK" ]]; then
    cat >"$dest/wpa_supplicant.conf" <<EOF
country=${COUNTRY}
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1

network={
	ssid="${SSID}"
	psk="${PSK}"
}
EOF
  fi

  # Waveshare HAT uses hardware SPI0 CE0 (GPIO8). Do NOT use dtoverlay=spi0-0cs —
  # that frees CS as a GPIO input (pulled high) and the panel ignores all SPI data
  # while BUSY stays idle-low, so the driver looks "healthy" with a blank screen.
  # (spi0-0cs is for Inky-style soft-CS; not for Waveshare.)
  if grep -qE '^dtoverlay=spi0-0cs' "$dest/config.txt"; then
    sed -i '/^dtoverlay=spi0-0cs/d' "$dest/config.txt"
  fi
  if ! grep -q 'Moto HUD' "$dest/config.txt"; then
    cat >>"$dest/config.txt" <<'CFG'

# Moto HUD — Waveshare SPI0 CE0 (hardware CS)
[all]
dtparam=spi=on
CFG
  fi
  # Strip mistaken host-mode dwc2 if present (peripheral must win for Zero gadget)
  sed -i '/^dtoverlay=dwc2,dr_mode=host$/d' "$dest/config.txt" 2>/dev/null || true
  # Official USB Ethernet gadget overlay
  if grep -qE '^dtoverlay=dwc2' "$dest/config.txt"; then
    perl -i -pe 's/^dtoverlay=dwc2(?!,dr_mode=host)$/dtoverlay=dwc2,dr_mode=peripheral/; s/^dtoverlay=dwc2,dr_mode=host$/dtoverlay=dwc2,dr_mode=peripheral/' "$dest/config.txt" 2>/dev/null || true
  fi
  if ! grep -q 'dtoverlay=dwc2,dr_mode=peripheral' "$dest/config.txt"; then
    printf '\n# USB Ethernet gadget\n[all]\ndtoverlay=dwc2,dr_mode=peripheral\n' >>"$dest/config.txt"
  fi

  local cmd
  cmd="$(tr -d '\n' <"$dest/cmdline.txt")"
  cmd="${cmd// modules-load=dwc2,g_ether/}"
  cmd=$(printf '%s' "$cmd" | sed -E 's/ systemd\.run=[^ ]+//g; s/ systemd\.run_success_action=[^ ]+//g; s/ systemd\.unit=[^ ]+//g')
  if [[ "$cmd" != *systemd.run=* && -f "$FIRSTRUN" ]]; then
    cmd="$cmd systemd.run=/boot/firmware/firstrun.sh systemd.run_success_action=reboot systemd.unit=kernel-command-line.target"
  fi
  if [[ "$cmd" != *raspberrypi-sys-mods/firstboot* && "$cmd" != *resize* ]]; then
    cmd="${cmd/ rootwait/ rootwait init=\/usr\/lib\/raspberrypi-sys-mods\/firstboot}"
  fi
  printf '%s\n' "$cmd" >"$dest/cmdline.txt"
}

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "==> dry-run; staged at $STAGE"
  ls -lh "$BIN_OUT"
  file "$BIN_OUT"
  exit 0
fi

write_boot "$BOOT"
echo "==> done (instance_id=${INSTANCE_ID})."
echo ""
echo "First boot: wait for TWO reboots (~2–4 min), then SSH:"
echo "  ssh ${USER_NAME}@10.12.194.1          # USB gadget (always)"
echo "  ssh ${USER_NAME}@${HOSTNAME}.local    # mDNS when Wi-Fi/USB up"
if [[ -n "$SSID" ]]; then
  echo "  Wi-Fi SSID: ${SSID} (2.4 GHz required on Zero W)"
fi
echo ""
if [[ "$BWR" -eq 1 ]]; then
  echo "  Display: Waveshare HAT (B) BWR (MOTOHUD_WAVESHARE_BWR=1)"
else
  echo "  Display: Waveshare B/W (add --bwr for HAT (B) red/black/white)"
fi
echo ""
echo "macOS USB tip (Internet Sharing is broken for this gadget):"
echo "  ./scripts/macos-usb-gadget.sh"
echo "  → sets 10.12.194.2/28 with NO default gateway; keeps Wi-Fi as default route"
echo ""
echo "HTTP: http://${HOSTNAME}.local:8787/  or  http://10.12.194.1:8787/"
