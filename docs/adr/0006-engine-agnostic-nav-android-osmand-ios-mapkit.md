---
status: accepted
date: 2026-07-27
---

# Engine-agnostic nav protocol; OsmAnd on Android, MapKit on iOS

The BLE/`nav` JSON schema is **engine-agnostic**. Platform companions fill it from different nav engines. **Android** targets **OsmAnd** (AIDL today; Full Library / richer fields including lanes as we own the flow — GPL and APK size are acceptable for this open-source project). **iOS** will use **Apple MapKit** `MKDirections` plus a local guidance loop so we avoid self-hosting Valhalla or paying a directions SaaS. Optional fields such as `lanes[]` are omitted when an engine cannot provide them.

## Considered options

- **Chosen** — shared protocol; OsmAnd-class richness on Android; MapKit for free iOS routing hosted by Apple.
- **Rejected: self-hosted Valhalla / GraphHopper for iOS** — ongoing infra cost unfit for this OSS project.
- **Rejected: OsmAnd as the cross-platform engine** — no usable iOS embed SDK; AIDL/Full Library are Android-only.
- **Rejected: Magic Lane / HERE / Mapbox as required path** — not free for production.
- **Rejected: Google Maps scrape as the long-term source** — brittle; kept as Android fallback only.

## Consequences

- iOS will lag Android on lanes, motorcycle profiles, and offline routing; the Pi must tolerate missing optional fields.
- Do not put OsmAnd `TurnType` ints or MapKit types on the wire — map to protocol `maneuver` / `lanes` in the companion.
- Near-term Android work can deepen AIDL then move to Full Library without changing the Pi contract.
- Full Library ships as an on-demand Play Feature Delivery module (see ADR 0009) so the base install stays AIDL-sized.
