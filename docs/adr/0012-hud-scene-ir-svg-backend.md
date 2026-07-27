---
status: accepted
date: 2026-07-27
---

# HUD scene IR; SVG as render backend only

Compose, refresh plans, and patches must not build SVG tags or depend on `<text>` / `<g>` strings. They produce a small **integer scene graph** (1-bit black/white semantics on the 250×122 canvas). **SVG is a backend**: one adapter lowers scene → SVG → existing `pixelfont` + rasterize path (Pi, WASM, design preview). **templ** authors screen trees in `hudui/screens/`; **`scenetempl`** turns templ components into scene nodes at compose time. `scene.RawSVG` remains for ribbon, minimap, lanes, and maneuver paths until native vector nodes exist.

## Considered options

- **Chosen** — `hudui/scene` display list + `hudui/render/svg` serializer; `plan.Layer.Patch` returns `scene.Document`; full-frame body migrates incrementally (`BodySVG` → scene or templ→scene at the adapter).
- **Rejected: SVG everywhere in compose** — couples layout/refresh to tag soup and duplicates patch vs full-frame paths.
- **Rejected: parse SVG back to scene** — fragile, costly on Pi Zero.
- **Rejected: immediate direct scene→Gray rasterizer** — optional later for hot patches; reuse SVG rasterizer until profiling says otherwise.

## Consequences

- Rename `NavSVGDeps` toward **`DrawDeps`** (fit/wrap/maneuver/ribbon/minimap/lane helpers); patches and nav chrome use `scene.Builder` / `DrawDeps.*Nodes`.
- Engine `PatchLayer` serializes `scene.Document` only inside `render/svg`.
- Compose no longer uses **RawSVG** for nav chrome or link; keep the type only for legacy/tests until removed.
- Related: [0010](0010-hud-ui-templ-component-layout.md), [0011](0011-component-refresh-tiers-partial-regions.md).
