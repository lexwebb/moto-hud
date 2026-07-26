---
status: accepted
date: 2026-07-01
---

# BLE GATT with small JSON messages (nav / media / cmd / heartbeat)

Phone↔Pi link uses a custom GATT service (**MotoHUD**) with JSON payloads for `nav`, `media`, `cmd`, and `heartbeat`. Same JSON shapes are reused by the Pi’s HTTP injector for bring-up without BLE. Payloads stay notification-sized for typical ATT MTUs.

## Considered options

- **Chosen** — JSON over BLE characteristics (simple to debug, shared with HTTP/WASM tools).
- **Rejected: protobuf/binary-only** — harder to poke from emulator and curl; size win not worth it at HUD rates.
- **Rejected: Classic Bluetooth SPP** — BLE fits phone background + low power better for this form factor.

## Consequences

- Optional rich fields (`ribbon_points`, `minimap`, future `lanes`) must stay compact or the link needs MTU negotiation / chunking later.
- Schema documented in `protocol/README.md` and mirrored in Go + Kotlin.
