---
status: accepted
date: 2026-07-28
---

# Semantics-first junction IR replaces meter minimap

Live nav left-column geometry is a **kind-discriminated junction IR** (`nav.junction`), not geographic polylines. The Pi draws idealized templates by `kind`; the phone fills rich topology when it can (OsmAnd Full Library), otherwise the Pi synthesizes `{ kind, outbound }` from `maneuver`. Immediate decision only — no interchange spaghetti on 70×80 e-ink. `dual_carriageway` is v1; dual detection is ours (OBF has no tag).

## Considered options

- **Chosen** — semantic IR + template draw; replace `nav.minimap`; phone optional / Pi fallback.
- **Rejected: keep meter polylines** — RDP/octilinear still looks geographic and fights the panel size.
- **Rejected: keep minimap beside junction** — two geometry paths; HUD must pick one.
- **Rejected: wait for OsmAnd `dual_carriageway` tag** — not in OBF routing encoding; infer from oneway + bearings.

## Consequences

- Wire: optional `junction` on `nav`; omit `minimap`. Prefer junction over ribbon; ribbon when neither junction nor synthesizable maneuver exists.
- Schema: [`protocol/junction.ts`](../../protocol/junction.ts); kinds and mapping in [`protocol/README.md`](../../protocol/README.md).
- Producers: Full Library only for rich IR; AIDL stays thin; emulator/classifier is lab.
- Unknown `kind` → render as `simple` using `outbound` / `sides`.
