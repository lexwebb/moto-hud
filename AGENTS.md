# AGENTS.md

## Cursor Cloud specific instructions

Moto HUD is a monorepo for one product (a motorcycle e-ink turn-by-turn HUD). The software-testable services here — no hardware required — are:

- `pi/` — the **core** Go HUD service (`go 1.25`). Renders frames, applies nav/media, and serves an HTTP injector + browser tools. This is the heart of the product.
- `site/` — Astro project site + WASM ride emulator (npm).
- `android/` — Kotlin companion app. Optional and heavy: needs the Android SDK + a system `gradle` 8.11.1 + JDK 17 (no `gradlew` wrapper is committed). Not set up here; skip unless specifically working on Android.

Standard build/run/test commands live in `README.md` and `.github/workflows/ci.yml`. CI is `go test ./...` in `pi/` (Go 1.25.x) and `gradle :app:testDebugUnitTest` in `android/`.

### Environment notes (non-obvious)

- **Go 1.25 is required** (`pi/go.mod`). Cloud agents get it from `.cursor/Dockerfile` (referenced by `.cursor/environment.json`); the update/`install` script only refreshes modules (`go mod download` + `npm install`), it does not install the toolchain. If `go version` reports < 1.25, rebuild/repair the environment from that Dockerfile rather than patching the running VM.
- The WASM emulator artifact (`web/emulator/motohud.wasm`, ~15 MB) is gitignored and must be built before the browser emulator will work. Build it with `cd site && npm run build:wasm` (or `./scripts/build-wasm.sh`); it requires Go. It is not part of `npm install`, so rebuild after Go/core changes.
- `scripts/build-wasm.sh` and `scripts/emu.sh` prepend `/opt/homebrew/bin` to `PATH` (macOS-oriented) but still work on Linux since the real `go` is found later on `PATH`.

### Running the core service (software-only, no hardware)

Run the hub, then feed it nav/media over HTTP — this exercises the full product loop and needs no Pi/display:

```bash
cd pi && go build -o ../bin/motohud ./cmd/motohud && go build -o ../bin/mock-nav ./cmd/mock-nav
./bin/motohud -host png -out out/hud.png -http :8787      # add -demo for a static start frame
go run ./cmd/mock-nav -scenario approach                  # or -scenario arrive|idle|media
```

- Health: `GET http://127.0.0.1:8787/health`. Rendered 1-bit frame: `GET /frame.png` (250×122). Browser preview: `/preview/`. Full Leaflet ride emulator: `/emulator/` (needs the WASM built first).
- `GPIO not found` warnings on startup are expected off-hardware and harmless.
- Inject nav via `POST /nav` with the `NavMessage` JSON shape from `pi/internal/protocol/protocol.go` (e.g. `maneuver` must be a valid enum like `left`, `eta_min` is an int). Prefer `cmd/mock-nav` to avoid schema mistakes.
- Hardware hosts (`-host inky|waveshare|lcd`) and real BLE (`-tags ble`, needs BlueZ) only work on a Pi; use `-host png` or `-host emu` in the cloud VM.
