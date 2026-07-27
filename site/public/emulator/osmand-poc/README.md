# OsmAnd junction proof (no phone)

Off-device proof that **OsmAnd-java** (same library the embedded Full Library module uses) can supply junction IR inputs: `attachedRoutes`, `highway`, `oneway`, bearings — without the tagged Overpass bake.

## Result (Whitehall → Farringdon)

See [`osmand-junction-dump.json`](osmand-junction-dump.json). Summary from last run:

- **9** turn segments; **7** with non-empty `getAttachedRoutes(start)`
- **`dual_carriageway` tag hits: 0** — tag is not in OBF routing encoding
- **`dual_inferred: 3`** — Whitehall (×2) + Northumberland Avenue via spatial OBF scan
- Side streets / Farringdon / Embankment: no dual (aligned with OSM tags; Embankment is physically dual but untagged)

## Dual inference rules (hardened)

`attachedRoutes` alone miss clean duals (parallel carriageways often share no nodes). Algorithm:

1. **Turn corridor only** — skip residential / `*_link` turn segments.
2. **Local arms** — route segment + attachments at every point index.
3. **Spatial OBF scan** (~50 m) via `buildSearchRouteRequest` / `loadRouteIndexData` — required for Whitehall-class duals.
4. **Match** when:
   - both car highways (exclude footway/cycleway/pedestrian)
   - lateral cross-track sep **10–22 m**
   - opposite travel bearings (≥150°)
   - **and** at least one `oneway≠0` (both-bidirectional parallels → Farringdon FP)
5. **Opposite oneway** means both oneway + opposite bearings (each dual carriageway is typically `oneway=+1` along its own geometry — do **not** require `+1` vs `-1` signs).
6. Never same-direction parallels (Fleet Street).

## Reproduce

Prereqs: JDK 17+, network once for Overpass + MapCreator download.

```bat
:: from site/
node scripts/export-whitehall-osm.mjs

:: download MapCreator if missing (~150MB)
curl -fL -o public/emulator/osmand-poc/OsmAndMapCreator-main.zip https://download.osmand.net/latest-night-build/OsmAndMapCreator-main.zip
:: extract to public/emulator/osmand-poc/mapcreator/

cd public/emulator/osmand-poc/mapcreator
utilities.bat generate-roads ..\whitehall-farringdon.osm

cd ..
dump-osmand-junction.bat
```

## What this means for the plan

| Need | Source on phone (Full Library) | Proven here? |
|------|--------------------------------|--------------|
| Turn class | `TurnType` | Yes (KL/TR/TL/TSLR/…) |
| Side arms | `RouteSegmentResult.getAttachedRoutes` | Yes |
| Highway class / ramps | `RouteDataObject.getHighway()` | Yes (`primary_link`, etc.) |
| Divided road | Spatial OBF + oneway/bearing | **Yes** (no tag) |
| Full OSM bake | Not required | Confirmed unnecessary for junction node |

AIDL path still cannot do this; only embedded OsmAnd.
