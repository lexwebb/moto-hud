---
status: accepted
date: 2026-07-26
---

# Prefer Waveshare 2.13″ B/W over colour e-ink for navigation

Four-colour Inky pHAT (BWRY) full refresh is ~15–20s even when the framebuffer is only black/white — too slow for turn-by-turn. **Waveshare 2.13″ black/white** (~2s full, ~0.3s partial) is the preferred new e-ink SKU. Existing Inky boards remain supported via `-host inky` for prototyping.

## Considered options

- **Chosen** — Waveshare B/W as preferred e-ink; keep Inky path.
- **Rejected: colour e-ink for nav** — flash duration dominates usefulness.
- **Rejected: LCD-only** — optional via `-host lcd`, not the primary product look.

## Consequences

- Pinouts differ (Waveshare DC=25 RST=17 BUSY=24 vs Inky); docs and platform wiring must stay per-host.
- Do not buy Waveshare colour (B/G) variants for this project.
