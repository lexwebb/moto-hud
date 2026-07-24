# Moto HUD

Motorcycle e-ink turn-by-turn HUD: Android companion reads Google Maps notifications + media, sends them over BLE to a Go service on a Raspberry Pi Zero that renders a sparse UI on an [Inky pHAT](https://shop.pimoroni.com/products/inky-phat).

```
Google Maps ──► NotificationListener ──┐
Music apps  ──► MediaController      ──┼─► Android app ──BLE──► Go (Pi) ──► Inky pHAT
Buttons (GPIO) ────────────────────────┘         ▲
                                                 │ cmd notify (play/pause, skip)
```

## Repo layout

| Path | Purpose |
|------|---------|
| [`protocol/`](protocol/) | BLE UUIDs + JSON message schema |
| [`pi/`](pi/) | Go HUD service, mock injector, systemd unit |
| [`android/`](android/) | Kotlin companion app |
| [`enclosure/`](enclosure/) | OpenSCAD bench case (CAD + mesh exports) |
| [`site/`](site/) | Astro project site (GitHub Pages) |


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
- **E-ink emulation** toggle: flash all pixels on→off, then ~1s fade-in; distance as **≈ nearest 50 m**; redraws about every **50 m** (plus maneuver / road / screen changes)
- Build WASM only: `./scripts/build-wasm.sh` → `web/emulator/motohud.wasm` (gitignored, ~15MB)

## Hardware hosts

`motohud` talks to the panel/buttons/phone through `pi/internal/platform` ports so adapters can be swapped:

| `-host` | Screen | Controls | Phone link |
|---------|--------|----------|------------|
| `auto` (default) | PNG, or Inky if `-inky` | keyboard / GPIO | BLE stub (or `-tags ble`) |
| `png` | PNG file | keyboard / GPIO | BLE stub |
| `inky` | Inky pHAT | GPIO | BLE |
| `emu` | memory (+ optional PNG) | channel (injectable) | loopback linked |
| `test` | memory | channel | loopback |

```bash
./bin/motohud -demo -host png -out out/hud.png -http :8787
./bin/motohud -host inky -inky   # on the Pi
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

HTTP injector (also used by `mock-nav`):

- `POST /nav` — JSON nav message
- `POST /media` — JSON media message
- `POST /button` — body `prev` \| `next` \| `prev_long` \| `next_long` \| `action` \| `action_long` \| `skip_prev` \| `skip_next`

## Pi Zero + Inky

1. Enable SPI: `sudo raspi-config` → Interface → SPI. Add to `/boot/firmware/config.txt`:

   ```
   dtoverlay=spi0-0cs
   ```

2. Prefer a **black/white** Inky pHAT (~1–2s refresh). Colour panels take ~15–20s; there is no supported partial update.

3. Build on the Pi (or cross-compile):

   ```bash
   cd pi
   go build -o ../bin/motohud ./cmd/motohud
   # optional real BLE peripheral (BlueZ):
   go build -tags ble -o ../bin/motohud ./cmd/motohud
   ```

4. Run with hardware display:

   ```bash
   ./bin/motohud -inky -out /tmp/hud.png -http :8787
   ```

5. Install systemd unit from [`pi/systemd/motohud.service`](pi/systemd/motohud.service) (adjust paths/user).

### Button wiring (BCM)

Active-low with internal pull-ups. Avoid Inky pins **17** (BUSY), **27** (RESET), **22** (DC).

| Button | BCM GPIO | Role |
|--------|----------|------|
| Prev | 5 | Previous screen / prev track on Media |
| Next | 6 | Next screen / next track on Media |
| Action | 13 | Context action; **long-press** → Nav home |

Screens: **Nav** → **Media** → **Status**.

- Prev/Next (not on Media) → change screen  
- Media + short Prev/Next → skip track (`prev_track` / `next_track`)  
- Media + long Prev/Next → change screen  
- Media + Action → `play_pause`  
- Status + Action → force full redraw  
- Long-press Action → Nav home

## Android companion

Open [`android/`](android/) in Android Studio, sync Gradle, side-load on a phone (or emulator).

1. Grant **notification access** (required for Maps + media sessions).
2. Start Google Maps navigation; if fields are empty, disable Maps **Live Updates** / **Live info** notification categories.
3. Tap **Start HUD link** — scans for BLE device **MotoHUD**.
4. Keep the foreground notification running while riding.

**Dev HTTP (no Pi BLE):** enable **Also POST nav/media over HTTP** and set the base URL (emulator → host is `http://10.0.2.2:8787`). Run `motohud -host png -http :8787` on the PC. BLE scan still runs; HTTP posts happen whenever nav/media update. See [`protocol/README.md`](protocol/README.md).

Unit tests: `ManeuverParser` — from Android Studio or CI (`gradle :app:testDebugUnitTest`).

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs `go test ./...` in `pi/` and Android unit tests on push/PR.

## Notes

- Inky refresh is always full-frame; the Go compositor only redraws on maneuver/road changes, coarse distance thresholds, screen changes, or force.
- Audio stays on the phone → helmet Bluetooth; the Pi is display + button bridge only.
