# Moto HUD Design System

## What this is
A visual system for a **motorcycle heads-up display** on a 2.13" e-ink panel — 250×122px, 1-bit black-on-white, no touch, three physical buttons (Prev / Next / Action, long-press Action = home/Nav). The phone runs Google Maps + music; this tiny Raspberry-Pi-class panel shows only distilled state pushed over BLE. There is no existing brand, codebase, or Figma file behind this — the system was designed from scratch against the hardware brief supplied in the project's initial prompt (reproduced in full in the original task; see "Data we have" below for the operative constraints).

This is not a phone UI, not CarPlay, not a moving map. Treat it as an **instrument face**: a glance-readable panel like a speedometer, not a screen you interact with.

## Data we actually have (design boundary)
Phone → HUD packets:
- **Navigation**: active y/n; maneuver type (left/right/slight-left/slight-right/straight/u-turn/roundabout/arrive/depart/unknown); distance_m + short distance_text; road name/instruction string; optional eta_min
- **Media**: playing y/n; title; artist
- **Link**: BLE connected y/n; heartbeat

No live map tiles, no continuous GPS track, no full turn list, no album art, no bike telemetry (v1), no touch/voice. See `guidelines/data-scope.card.html` for the excluded-vs-reserved breakdown.

## Components (10)
- **glyphs/** — `ManeuverGlyph` (10 maneuver types, pixel-grid arrows), `ConnectionMark` (BLE link + heartbeat blink)
- **readouts/** — `DistanceReadout` (hero number), `ETAReadout`, `MediaLine` (now-playing title/artist)
- **navigation/** — `RoadRibbon` (schematic corridor band, reserved for the road-view stretch feature), `ProgressTicks` (coarse distance ticks)
- **chrome/** — `ModeHeader` (mode label + link mark), `FooterHints` (≤11px button legend)
- **core/** — `PixelDivider` (solid/dashed/dither rule)

### Intentional additions
This system has no prior component inventory to enumerate (no attached codebase/Figma), so the primitive set above was authored from scratch, sized to what an e-ink instrument panel actually needs — not a generic web kit. There is deliberately no Button/Input/Card/Dialog: the device has no touch surface, so standard interactive web primitives don't apply.

## UI kit
`ui_kits/hud/` — all six delivered screens, interactive (click Prev/Next/Action, hold for long-press):
1. Nav active — no ribbon
2. Nav active — with road-view ribbon
3. Nav idle / waiting for route
4. Media focus
5. Status / link diagnostics
6. Nav + media hybrid

## Index
- `styles.css` — root stylesheet, `@import`s only
- `tokens/colors.css`, `typography.css`, `spacing.css`, `patterns.css` — design tokens
- `components/glyphs/`, `components/readouts/`, `components/navigation/`, `components/chrome/`, `components/core/` — the 10 primitives, each with `.jsx` + `.d.ts` + `.prompt.md` + a `.card.html` demo
- `guidelines/` — foundation specimen cards: colors, type scale/weights, spacing scale, panel grid, stroke weights, hierarchy/eye-path, data scope
- `ui_kits/hud/` — the six screen compositions + interactive `index.html`
- `thumbnail.html` — project tile
- `SKILL.md` — portable skill for Claude Code / other agents

## Content fundamentals
Copy on the panel is functional, not conversational — it's an instrument, not an assistant. No "you" address, no first person, no punctuation flourish. Road names and instructions are passed through verbatim from the phone (e.g. "Ridge Rd", "onto Harbor Blvd") — never rewritten or softened. Units render as short suffixes (m, km, min) directly beside numbers, never spelled out. No emoji, ever — a dashboard doesn't smile. The only invented copy is chrome: mode labels are one word, all-caps, 11px (NAV / MEDIA / STATUS); footer hints are `BUTTON + one-word verb` (PREV Mode, NEXT Skip, HOLD Home, ACTION Redraw). Idle/empty states state the fact plainly ("Waiting for route…") rather than being cute about it.

## Visual foundations
- **Color**: exactly two values, `--ink` (#000) and `--paper` (#fff) — this is a 1-bit display, there is no gray. Where a "quiet"/disabled/ghost state is needed, use a **dither pattern** (checkerboard, see `--dither-25`/`--dither-50`/`--hatch` in `tokens/patterns.css`) instead of opacity or a lighter color — opacity doesn't exist on e-ink.
- **Type**: a single pixel-grid family (Terminus Bold, bold BDF faces) at four fixed sizes — 12 (meta/footer), 16 (road/ETA), 24 (titles), 32 (hero distance). No sizes in between; no other families. Bold = signal (numbers, labels), Regular = context (road names, secondary meta).
- **Spacing**: small-integer scale (2/4/6/8/12/16px) — the panel is 250×122, an 8px web grid is too coarse.
- **Backgrounds**: flat paper white, full-bleed. No images, no photography, no gradients — a 1-bit panel can't render either.
- **Icons/glyphs**: hand-built on a 40×40 pixel grid, 3px square-capped strokes, miter joins — no rounded caps (rounding anti-aliases on rasterization). See Iconography below.
- **Animation**: none, except a single discrete on/off blink (`steps(1)`, ~1.6s) on the BLE heartbeat dot. No fades, no easing curves, no smooth transitions anywhere — the panel's refresh is a full 1–2s redraw, so anything implying continuous motion (a smoothly filling progress bar, a scrolling ribbon) is a lie the hardware can't back up. `ProgressTicks` and `RoadRibbon` both step in discrete jumps at coarse thresholds instead.
- **Hover/press states**: not applicable — no touch, no cursor. The only "state" feedback is the physical button legend in `FooterHints`.
- **Borders**: solid black rules only, 1px (dividers) / 2px (bold rules, ticks) / 3px (glyph strokes) — never a hairline that could vanish on threshold-to-1-bit conversion.
- **Corners**: square, everywhere. No border-radius on anything — an instrument face, not a phone app.
- **Cards/shadows/blur/transparency**: none of these exist in this system. No drop shadows, no blur, no translucency — all of that requires gray or alpha, which 1-bit e-ink doesn't have.
- **Layout**: fixed, not responsive — every screen is composed against the literal 250×122 canvas (see `guidelines/panel-grid.card.html`), with a reserved 34–40px horizontal band for the road-view ribbon so it can be added without restructuring the nav screens.
- **Imagery**: none — no photos, no illustration, no album art (the phone doesn't send any).

## Iconography
No icon library, no icon font, no CDN glyph set, no emoji, no Unicode symbol substitution — every glyph on the panel (maneuver arrows, play/pause, link/heartbeat marks) is a **hand-built vector on a pixel grid**, authored directly in `components/glyphs/ManeuverGlyph.jsx` and `ConnectionMark.jsx` using plain SVG `line`/`polygon`/`path`/`circle`. This is a deliberate exception to "never draw your own icons": there is no existing brand or icon system to source from, and the maneuver/status glyphs are core functional content the brief explicitly asks this system to define (a "component sheet" of maneuver glyphs). Rules for any new glyph: 40×40 (or 12×12 for marks) viewBox, 3px (or 1.6px for marks) square-capped strokes, miter joins, no curves except the u-turn/roundabout arcs, `shape-rendering:crispEdges` for preview. Production art gets thresholded to 1-bit by a separate raster pipeline — these SVGs are the vector source, not the shipped bitmap.

## Logo / brand mark
No company or product brand was provided — this is a from-scratch hardware concept, not an existing company. No logo exists and none was invented. Where a mark might normally sit (the Nav-idle empty state), the product is referred to in plain type as "MOTO HUD," set in the same pixel font as everything else — a descriptive label, not a designed wordmark. If this becomes a real product with a name/mark, replace that one string.

## Font substitution note
No existing typeface was specified (from-scratch build). **Terminus Bold** (Google Fonts, bold BDF faces) was chosen as the system's only typeface because it's designed natively as a pixel/bitmap font — it renders as clean whole-pixel strokes at small sizes without anti-aliasing, which is the actual hardware requirement here. Production rasters Terminus Bold BDFs (`ter-u12b` / `16b` / `24b` / `32b`, SIL OFL). This design kit uses a monospace stand-in in the browser; sizes match the BDF pixel sizes.
