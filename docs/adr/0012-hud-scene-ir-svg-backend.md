---
status: accepted
date: 2026-07-27
---

# HUD scene IR; bitmap is the only rasterizer

Compose, refresh plans, and patches produce a small **integer scene graph** (1-bit black/white on 250×122). **`hudui/render/bitmap`** is the sole path from scene → `*image.Gray` — full frames (`compose.FrameDocument`), patch layers, minimap/junction panes, Pi, and WASM. **`hudui/render/svg`** serializes scene → SVG markup for `/frame.svg`, designer export, and lab debug strings only — it does not rasterize.

## Considered options

- **Chosen** — scene IR + `render/bitmap` everywhere pixels are produced.
- **Rejected: SVG/`canvas` on the hot path** — megabyte allocs; blocked MCU ports; WASM paid the same cost.
- **Rejected: HTML/CSS runtime** — ADR 0010; fights partial-refresh slots.

## Consequences

- `tdewolff/canvas` is not a runtime dependency of `motohud` / WASM.
- `RawSVG` nodes fail bitmap rasterize; production compose must emit typed nodes.
- Related: [0010](0010-hud-ui-templ-component-layout.md), [0011](0011-component-refresh-tiers-partial-regions.md).
