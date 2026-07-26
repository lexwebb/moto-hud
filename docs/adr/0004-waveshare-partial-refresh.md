---
status: accepted
date: 2026-07-26
---

# Waveshare: partial refresh with full clean every 5 frames

The Waveshare 2.13″ B/W driver uses **partial** updates (~0.3s, no flicker) for ordinary gated redraws, and a **full** refresh (~2s, multi-flash) for the base image and every 5th update thereafter, per vendor guidance against ghosting. Emulator Waveshare profile matches this cadence.

## Consequences

- Re-init before leaving partial mode for a global refresh (Waveshare FAQ).
- Inky remains full-frame only; LCD is instant (no e-ink gate timing).
