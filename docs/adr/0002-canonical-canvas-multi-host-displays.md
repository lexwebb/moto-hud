---
status: accepted
date: 2026-07-26
---

# Canonical 250×122 canvas with swappable display hosts

Compositing always targets a **250×122 grayscale** canvas. Hardware differences are isolated behind `-host` adapters (`inky`, `waveshare`, `lcd`, `png`, `emu`, `test`) so layout, WASM emulator, and protocol stay shared.

## Considered options

- **Chosen** — one logical resolution; letterbox or dither at the adapter (LCD 320×240 letterboxes the HUD).
- **Rejected: native resolution per panel** — would fork layouts and emulator profiles per SKU.

## Consequences

- LCD Action button remaps to BCM 16 because Display HAT Mini backlight uses 13 (see `docs/BUTTONS.md`).
- Emulator device picker mirrors the same host profiles (timing + chrome).
