---
status: accepted
date: 2026-07-27
---

# HUD UI: templ components, layout props, pure render tree

The Pi compositor will move from hand-built SVG strings in `layout.go` toward **compile-time templ components** plus a small **integer flex layout** layer (`Row` / `Col` / `Box`) that emits SVG groups. Components are **pure functions of props and `RenderContext`** (faces, tokens, canvas size); side effects (BLE, refresh gating, rasterize) stay in the hub. The browser design kit (`design/`) remains the visual reference; Go components mirror the same primitive names and tokens.

## Considered options

- **Chosen** — templ for markup/ergonomics; in-repo layout subset (gap, pad, justify, align, baseline rows with `pixelfont` measure); no embedded JS/React on device.
- **Rejected: runtime React in Go** — too heavy for Pi Zero; SVG needs a custom host anyway.
- **Rejected: gomponents-only** — viable but weaker “screen file” DX vs templ for nested HUD trees.
- **Rejected: full CSS/grid engine** — 250×122 instrument panel does not justify it.

## Consequences

- New package(s) under `pi/internal/hudui/` (or similar): layout pass → SVG serialize; existing `pixelfont` + `RasterizeSVG` path unchanged.
- Screen migration is incremental (one kit screen at a time); `layout.go` shrinks as screens move.
- **Chrome** (header, legend column, link glyph slot) lives in `hudui/screens/chrome.templ`; compose wraps each screen `BodySVG` via `FrameVars`.
- Shared tokens should converge with `design/tokens/` (codegen or checked-in JSON) to limit drift.
- Interactive “hooks” are not a runtime: state lives on `hud.State`; preview interactivity uses the same props from tests or emulator, not `useEffect` in components.
- Refresh tiers, fixed slots, and spatial partial orchestration are specified in [ADR 0010](0010-component-refresh-tiers-partial-regions.md).
