---
status: accepted
date: 2026-07-27
---

# Android: small base APK + on-demand OsmAnd Full Library module

Ship the companion as a **~6.5 MB base** (AIDL + Maps scrape) and put OsmAnd Full Library in a Play Feature Delivery **on-demand** module (`:osmand`). Shrink that module with **arm64-only** natives, **stripped `World_basemap_mini.obf`**, and **AAB ABI/language/density splits**. Region maps stay runtime downloads inside OsmAnd.

## Considered options

- **Chosen** — base AIDL always; download Full Library when the rider wants lanes / then-next; restart so `AppComponentFactory` can switch to `OsmandApplication`.
- **Rejected: monolithic embedded flavor (~370 MB)** — unacceptable default install for a BLE companion.
- **Rejected: Full Library in the base APK with only ABI filters** — still ~150 MB+ before first launch.
- **Rejected: self-hosted `.so` CDN outside Play modules** — brittle version coupling to OsmAnd AARs.

## Consequences

- Rich nav needs Play (or `bundletool --local-testing` for sideload); plain `installDebug` is base-only unless the module is fused/tested via bundletool.
- After module install, the process must restart so `MotoHudAppComponentFactory` can load `MotoHudOsmandApp`.
- Emulators without arm64 cannot run the Full Library module; they keep using AIDL.
