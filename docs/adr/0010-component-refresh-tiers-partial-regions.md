---
status: accepted
date: 2026-07-27
---

# Component refresh tiers and spatial partial updates

HUD components declare a **refresh tier** and, when eligible, a **fixed layout slot** (`image.Rectangle` in canvas space). After each state update, a **refresh orchestrator** compares stable **change keys** per component (not SVG string diffs), unions dirty rectangles, and chooses: **no draw**, **spatial partial** (Waveshare only), **full bitmap + partial waveform**, or **full bitmap + full waveform** (ghosting cadence per ADR 0004). Semantic gating (ADR 0005) remains the first filter; tiers refine *what* to recomposite and *how much* of the panel to push over SPI.

Today Waveshare “partial” is only a **faster full-frame update mode** (`setWindow` covers the whole panel). This ADR adds **optional spatial windows** when the dirty union is small enough and only **partial-capable** components changed.

## Component metadata

Each primitive or screen region exposes (compile-time or layout-time constants where possible):

| Field | Purpose |
|-------|---------|
| `Tier` | `Static` · `Slow` · `Fast` · `PartialOK` |
| `Slot` | Fixed `Rect` after layout; required for `PartialOK` |
| `ChangeKey(props)` | Hashable value(s) that mean “pixels in this slot may differ” |

**Tier semantics (intent):**

- **Static** — chrome, rules, footer hints on a given screen; redraw on screen/mode change or forced full refresh only.
- **Slow** — road name, media title, maneuver glyph; redraw when matching protocol fields change (or screen change).
- **Fast** — distance bucket, ETA, link heartbeat; may change under ADR 0005 without other nav fields changing.
- **PartialOK** — same as Fast/Slow for change detection, but **allowed** to participate in a spatial partial upload if its slot is stable and dirty area is below policy threshold.

Components that **reflow siblings** (flex grow, variable-width text outside a fixed slot) must **not** be `PartialOK`; they force at least a **full recomposite** of the affected column (often the whole main area).

## Orchestrator decision flow

```text
BLE / button → RefreshGate (ADR 0005) → if !ShouldRedraw → stop
                ↓
         Build props per component → ChangeKey vs last frame
                ↓
         Dirty set = components whose keys changed
                ↓
    All dirty ⊆ PartialOK AND union(Slot) area ≤ maxPartialPixels?
         yes → raster dirty slots (or layers) → display.ShowRegion(rects, bitmap)
         no  → full Render() → display.Show(full)  [partial or full waveform per ADR 0004]
                ↓
    partialsSinceFull++; at N → force full waveform + driver re-init (existing Waveshare logic)
```

**Policy knobs** (defaults tuned on hardware):

- `maxPartialPixels` — e.g. 30–40% of canvas; above → full frame recomposite even if components are PartialOK.
- `maxPartialCount` — already `waveshareFullEveryN` (5); spatial partials count toward the same counter.
- **Alignment** — dirty rects expanded to **8px horizontal** boundaries before `setWindow` (EPD byte columns).

The orchestrator **does not** infer dirtiness from SVG or raster diffs in v1; **explicit ChangeKey** keeps Pi cost predictable and matches protocol-shaped updates.

## Rendering engine responsibilities

1. **Compose** (`hudui/compose` + templ) assigns stable slots, change keys, tiers, and optional `Patch` closures per layer in `plan.ScreenPlan`. The `hud` engine and orchestrator consume that plan only — they do not embed screen-specific geometry.
2. **Retain** last frame grayscale buffer on hosts that support spatial partial.
3. **On partial path** — call each dirty layer’s `Patch()`, blit at `Layer.Slot`, crop to aligned union, pass to display.
4. **On full path** — `BodySVG` from the plan plus chrome shell → `BuildPixelSVGFromVars` → rasterize entire 250×122.

`templ` components stay pure; tier/slot/key/patch live on compose layout nodes (`plan.Layer`). The BLE link glyph uses a fixed chrome slot (`NodeLink`, `TierPartialOK`) and is composited after the shell SVG (full frame and spatial patch).

## Display host capabilities

| Host | Spatial partial | Waveform partial | Notes |
|------|-----------------|------------------|-------|
| `waveshare` | Optional (`ShowRegion`) | Yes (ADR 0004) | Implement sub-window `setWindow` + subset SPI write |
| `inky` | No | No | Always full frame, ~20s |
| `lcd`, `png`, `emu` | N/A | N/A | Full buffer each draw; emulator should visualize dirty rects for debug |

`Display` interface grows a capability-based API, e.g. `Show(img)` unchanged as default full frame; `RegionalDisplay` or optional `ShowPartial(img, []image.Rectangle)` for Waveshare.

## Considered options

- **Chosen** — declarative tiers + change keys + union rect policy; orchestrator above compositor; extend Waveshare driver for real windows when profitable.
- **Rejected: framebuffer auto-diff** — too CPU-heavy on Pi Zero; fragile with dither/anti-alias thresholds.
- **Rejected: SVG tree diff** — expensive; layout shift breaks bbox assumptions.
- **Rejected: spatial partial without fixed slots** — flex reflow makes correct dirty regions impractical.

## Consequences

- ADR 0009 layout layer must support **fixed slots** for PartialOK components; distance readout is the first candidate.
- `RefreshGate` may shrink to coarse “any nav tick worth considering” while orchestrator handles Fast vs Static within an allowed redraw.
- Tests: golden full frames plus unit tests on orchestrator decisions (dirty set → full vs partial).
- Emulator gains optional “highlight dirty rect” mode for tuning `maxPartialPixels`.

## Related

- [0004](0004-waveshare-partial-refresh.md) — partial vs full **waveform** cadence
- [0005](0005-eink-refresh-gating.md) — when updates are considered at all
- [0009](0009-hud-ui-templ-component-layout.md) — templ components and layout
