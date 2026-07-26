---
status: accepted
date: 2026-07-26
---

# Gate e-ink redraws on maneuver and coarse distance

Because e-ink updates are slow and visually noisy, the hub/emulator only redraw on **maneuver / road / screen changes**, **≈50 m distance buckets**, or **force** — not every GPS tick. LCD hosts skip the e-ink gate.

## Consequences

- Distance shown on e-ink is stepped (nearest ~50 m), not metre-accurate live countdown.
- Busy panel queues at most one pending force refresh.
