---
status: accepted
date: 2026-07-01
---

# Terminus Bold pixel fonts, four sizes, 1-bit output

HUD text is rendered from **Terminus Bold** BDF faces at fixed `data-pixel` sizes (16×32, 12×24, 8×16, 6×12) into hard-thresholded 1-bit frames. No runtime font scaling. Design-kit browser monospace is a stand-in only.

## Considered options

- **Chosen** — Terminus Bold (SIL OFL) for crisp e-ink glyphs.
- **Rejected: scalable outline fonts on device** — soft edges and size sprawl fight 250×122 1-bit.

## Consequences

- SVG templates use placeholder text + `data-pixel`; production path is rect glyphs → PNG/panel.
- Preview and Pi must share the same raster path (`/frame.png`).
