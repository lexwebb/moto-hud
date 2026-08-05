# Moto HUD

Motorcycle e-ink turn-by-turn HUD: Android companion reads turn-by-turn from **OsmAnd** (AIDL or embedded Full Library) or Google Maps notifications + media, sends them over BLE to a Go service on a Raspberry Pi Zero that renders a sparse UI on a 250×122 panel ([Inky pHAT](https://shop.pimoroni.com/products/inky-phat) or [Waveshare 2.13″ B/W e-Paper HAT](https://www.waveshare.com/2.13inch-e-paper-hat.htm); optional LCD).

```
OsmAnd app  ──► AIDL (default) ───────────────┐
:osmand DF  ──► RoutingHelper (on-demand) ────┼─► Android app ──BLE──► Go (Pi) ──► display HAT
Google Maps ──► NotificationListener ─────────┤         ▲
Music apps  ──► MediaController ──────────────┘         │
Buttons (GPIO) ─────────────────────────────────────────┘  cmd notify
```

## Repo layout

| Path | Purpose |
|------|---------|
| [`protocol/`](protocol/) | BLE UUIDs + JSON message schema |
| [`pi/`](pi/) | Go HUD service, mock injector, systemd unit |
| [`android/`](android/) | Kotlin companion app |
| [`enclosure/`](enclosure/) | OpenSCAD bench case (CAD + mesh exports) |
| [`site/`](site/) | Astro project site (GitHub Pages) |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records |
| [`CONTEXT.md`](CONTEXT.md) | Domain language / glossary |


## Project site (GitHub Pages)

Public demos: enclosure 3D viewer, design HUD kit, WASM ride emulator.

**https://lexwebb.github.io/moto-hud/**

```bash
cd site
npm install
npm run build:wasm   # optional locally; needs Go — CI builds it for Pages
npm run dev          # http://localhost:4321/moto-hud/
```

Deploy: push to `main` (workflow `.github/workflows/pages.yml`). One-time: repo **Settings → Pages → Source: GitHub Actions**.

Go-backed live `/preview` (frame.png / fonts) is still local-only via `./scripts/emu.sh`.

Full ride simulation with Leaflet (OSM) + HUD driven by the same Go core (WASM preferred, HTTP fallback):

```bash
./scripts/emu.sh
# open http://127.0.0.1:8787/emulator/
```

- Map animates a real OSRM street route Whitehall → Farringdon (`web/emulator/routes/whitehall-farringdon.json`)
- Each tick posts nav distance/maneuver into the HUD
- Virtual Prev/Next/Action buttons
- **Device** picker: Inky 4-colour (~20s BWRY wipe), Waveshare B/W (partial ~0.3s, full ~2s every 5), or Display HAT Mini LCD (instant letterbox); e-ink modes use **≈ nearest 50 m** and redraw ~every **50 m**
- Build WASM only: `./scripts/build-wasm.sh` → `web/emulator/motohud.wasm` (gitignored, ~15MB)

## Hardware hosts

`motohud` talks to the panel/buttons/phone through `pi/internal/platform` ports so adapters can be swapped. Canonical compose is always **250×122 gray**; backends map to hardware.

| `-host` | Screen | Controls | Phone link |
|---------|--------|----------|------------|
| `auto` (default) | PNG, or Inky if `-inky` | keyboard / GPIO | BLE stub (or `-tags ble`) |
| `png` | PNG file | keyboard / GPIO | BLE stub |
| `inky` | Inky pHAT (kept for existing boards) | GPIO | BLE |
| `waveshare` | Waveshare 2.13″ B/W e-Paper | GPIO | BLE |
| `lcd` | Display HAT Mini 320×240 (letterboxed) | GPIO (Action→16) | BLE |
| `emu` | memory (+ optional PNG) | channel (injectable) | loopback linked |
| `test` | memory | channel | loopback |

```bash
./bin/motohud -demo -host png -out out/hud.png -http :8787
./bin/motohud -host inky -inky          # existing Inky on the Pi
./bin/motohud -host waveshare           # Waveshare 2.13 B/W
./bin/motohud -host lcd                 # Display HAT Mini (optional)
```

The browser emulator uses Go WASM (`pi/cmd/motohud-wasm`) for the HUD core, with HTTP → `-host emu` as fallback.


Reference pack from Claude Design lives in [`design/`](design/) (tokens, React kit, guidelines). **Production type is Terminus Bold**, not the kit’s browser monospace stand-in. Interactive mock: open `design/ui_kits/hud/index.html` via a static server.


Layouts live in [`assets/hud/`](assets/hud/) as 250×122 SVG with `{{placeholders}}`.

```bash
cd /Users/lex/src/moto-hud
./bin/motohud -demo -out out/hud.png
# open http://127.0.0.1:8787/preview/   ← 1-bit /frame.png @ 2×
# SVG source (debug): http://127.0.0.1:8787/frame.svg
```

**Typography (SVG + Terminus Bold):** Templates in `assets/hud/*.svg` use `<text data-pixel="12x24" font-size="24" …>`. At render those become **1×1 pixel `<rect>`s** from Terminus Bold BDFs; arrows are stroked then hard-thresholded to 1-bit. The web preview shows **`/frame.png`** (same raster as the Pi). Compare faces at `/preview/fonts.html`.

| Role | `data-pixel` | Face |
|------|--------------|------|
| Distance | `16x32` | `ter-u32b` |
| Road / titles | `12x24` | `ter-u24b` |
| Body / ETA | `8x16` | `ter-u16b` |
| Meta | `6x12` | `ter-u12b` |

Only those four sizes — no runtime scaling. License: SIL OFL 1.1 (`assets/fonts/terminus/OFL.TXT`).

```bash
cd pi
go run ./cmd/motohud -demo -out ../out/hud.png -http :8787
# another terminal:
go run ./cmd/mock-nav -scenario approach
# open out/hud.png — updates when distance crosses 500/200/100/50/20 m
```

Keyboard buttons while `motohud` runs: `p`/`n`/`a` short, `P`/`N`/`A` long (+ Enter).

**HUD UI layer** (`pi/internal/hudui`, ADR 0010–0012): [templ](https://templ.guide) screen components, integer flex layout helpers, and `hud.RenderEngine` for tier-aware draws (distance-only patches when the refresh orchestrator allows). Regenerate after editing `.templ` files: `./scripts/generate-hudui.sh`.

HTTP injector (also used by `mock-nav`):

- `POST /nav` — JSON nav message
- `POST /media` — JSON media message
- `POST /button` — body `prev` \| `next` \| `prev_long` \| `next_long` \| `action` \| `action_long` \| `skip_prev` \| `skip_next`

## Pi Zero + display HATs

### Recommended boards

| Role | Board | Notes |
|------|-------|-------|
| **Preferred new e-ink** | [Waveshare 2.13″ e-Paper HAT (black/white)](https://www.waveshare.com/2.13inch-e-paper-hat.htm) or [HAT+](https://www.waveshare.com/2.13inch-e-paper-hat-plus.htm) | Same 250×122 canvas; ~2s full / ~0.3s partial. **Do not** buy colour (B/G) variants — ~15s+ refresh. |
| Existing / regression | Inky pHAT **black/white** | Still supported via `-host inky`. B/W Inky is discontinued; 4-colour Inky is too slow for nav (~15–20s). |
| Optional LCD | [Pimoroni Display HAT Mini](https://shop.pimoroni.com/products/display-hat-mini) | 320×240 IPS; HUD letterboxed. Instant refresh; backlight + sun glare trade-offs. |

1. Enable SPI: `sudo raspi-config` → Interface → SPI (`dtparam=spi=on`).

   - **Waveshare 2.13″ HAT:** use default SPI0 **with hardware CE0**. Do **not** add `dtoverlay=spi0-0cs` — that leaves CS floating high and the panel stays blank while the driver still reports “ready”.
   - **Inky pHAT:** may need `dtoverlay=spi0-0cs` (soft CS); see Pimoroni docs.
   - Display HAT Mini uses **SPI0 CE1**; `dtparam=spi=on` is enough.

2. Cross-compile for the Pi (Docker; any host OS):

   ```bash
   ./scripts/build-armv6.sh           # stub BLE (pure Go)
   ./scripts/build-armv6.sh --ble     # real BlueZ peripheral (CGO)
   ./scripts/build-armv6.sh --ble --deploy 10.12.194.1   # scp + restart
   ```

   On-device `go build -tags ble` still works if you have Go + `libbluetooth-dev` on the Pi, but Docker is the supported path.

3. Run with hardware display:

   ```bash
   ./bin/motohud -host waveshare -out /tmp/hud.png -http :8787
   # or: -host inky    /  -host lcd
   ```

4. Install systemd unit from [`pi/systemd/motohud.service`](pi/systemd/motohud.service) (adjust paths/user).

**SD card from a Mac (Pi Zero W + Waveshare):** use **Raspberry Pi OS Lite** (32-bit / armhf) — not desktop. Flash + customise:

```bash
./scripts/flash-lite.sh --disk disk12 \
  --user motohud --password motohud \
  --ssid 'YourWifi' --psk '…' \   # 2.4 GHz only on Zero W
  --bwr                           # only if you have HAT (B) red/black/white
```

Or, with `bootfs` already mounted on an existing Lite card:

```bash
./scripts/prepare-bootfs.sh --boot /Volumes/bootfs \
  --user motohud --password motohud \
  --ssid 'YourWifi' --psk '…' \
  --bwr   # HAT (B) only
```

That writes an ARMv6 `motohud` binary, Waveshare SPI (hardware CE0 — never `spi0-0cs`), official USB Ethernet gadget (`rpi-usb-gadget`), cloud-init NetworkManager Wi‑Fi, and a firstrun installer. Use the Zero’s **data** micro‑USB port (next to HDMI), not PWR. Wait for **two** boots (~2–4 min).

**SSH over USB (recommended first):** macOS Sequoia/Tahoe cannot Internet‑Share to this gadget (`CHANNEL_IO` → no `bridge100`). Skip Sharing; run:

```bash
./scripts/macos-usb-gadget.sh   # 10.12.194.2/28, no gateway; Wi‑Fi stays default route
ssh motohud@10.12.194.1         # Pi SHARED address
```

Then mDNS / Wi‑Fi: `ssh motohud@motohud.local`. HTTP: `http://10.12.194.1:8787/`.

**BLE (phone link):** `flash-lite` / `prepare-bootfs` build a BlueZ binary via Docker by default. To refresh a running Pi:

```bash
./scripts/build-armv6.sh --ble --deploy 10.12.194.1
```

Android companion scans for BLE name **MotoHUD**.

### Display pin maps (BCM)

Avoid these for buttons:

| Backend | DC | RST | BUSY / other | SPI |
|---------|----|-----|--------------|-----|
| Inky | 22 | 27 | BUSY 17 | SPI0.0 |
| Waveshare 2.13 | 25 | 17 | BUSY 24 | SPI0.0 |
| Display HAT Mini | 9 | — | BL **13**, CE1 | SPI0.1 |

### Button wiring (BCM)

Active-low with internal pull-ups. Full soldering dummies guide, enclosure BOM, and Waveshare 8-pin note: [`docs/BUTTONS.md`](docs/BUTTONS.md). Caps + tactile wells: [`enclosure/README.md`](enclosure/README.md).

| Button | BCM (inky / waveshare) | BCM (`-host lcd`) | Role |
|--------|------------------------|-------------------|------|
| Prev | 5 | 5 (HAT A) | Previous screen / prev track on Media |
| Next | 6 | 6 (HAT B) | Next screen / next track on Media |
| Action | 13 | **16** (HAT X) | Context action; **long-press** → Nav home |

On Display HAT Mini, BCM 13 is the backlight, so Action automatically remaps to button X (16).

Screens: **Nav** → **Media** → **Status**.

- Prev/Next (not on Media) → change screen  
- Media + short Prev/Next → skip track (`prev_track` / `next_track`)  
- Media + long Prev/Next → change screen  
- Media + Action → `play_pause`  
- Status + Action → force full redraw  
- Long-press Action → Nav home

## Android companion

Open [`android/`](android/) in Android Studio, sync Gradle, side-load on a phone (or emulator).

**Delivery model** ([ADR 0006](docs/adr/0006-engine-agnostic-nav-android-osmand-ios-mapkit.md)):

| Piece | Size (approx) | Nav |
|-------|----------------|-----|
| Base APK | ~6.5 MB | External [OsmAnd](https://play.google.com/store/apps/details?id=net.osmand) via AIDL + Maps scrape fallback |
| On-demand `:osmand` module | ~215 MB uncompressed (arm64-only, mini basemap stripped); Play compresses further | OsmAnd Full Library — lanes, then-next, ETA |
| Full debug `.aab` (base+module) | ~100 MB | both modules packaged; Play still delivers base first |
| Region maps | downloaded in OsmAnd UI | not bundled |

```bash
# base APK (AIDL)
gradle :app:installDebug

# Play / local module testing — build an App Bundle, then:
gradle :app:bundleDebug
# bundletool build-apks --bundle=app/build/outputs/bundle/debug/app-debug.aab \
#   --output=/tmp/motohud.apks --local-testing
# bundletool install-apks --apks=/tmp/motohud.apks
```

1. Install **OsmAnd** for the default AIDL path (or skip if you only use Maps scrape).
2. Optional: tap **Download rich OsmAnd nav** → restart when prompted → **Open OsmAnd map** → download your region → navigate (lanes on the HUD).
3. Grant **notification access** (media + Maps/OsmAnd enrichment).
4. Tap **Start HUD link** — starts the nav engine, scans for BLE **MotoHUD**.

Size knobs: arm64-only natives, Play Feature Delivery on-demand module, `World_basemap_mini.obf` stripped (maps stay dynamic), AAB ABI/language/density splits.

If using Maps and fields are empty, disable Maps **Live Updates** / **Live info** notification categories. iOS will use MapKit later (no self-hosted Valhalla).

**Dev HTTP (no Pi BLE):** enable **Also POST nav/media over HTTP** and set the base URL (emulator → host is `http://10.0.2.2:8787`). Run `motohud -host png -http :8787` on the PC. BLE scan still runs; HTTP posts happen whenever nav/media update. See [`protocol/README.md`](protocol/README.md).

Unit tests: `gradle :app:testDebugUnitTest`.

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs `go test ./...` in `pi/` and Android unit tests on push/PR.

## Notes

- Panel refresh is full-frame on Inky (~20s). Waveshare uses partial updates (~0.3s, no flicker) with a full refresh (~2s) every 5 frames to clear ghosting. The Go compositor only redraws on maneuver/road changes, coarse distance thresholds, screen changes, or force.
- Audio stays on the phone → helmet Bluetooth; the Pi is display + button bridge only.
