# Moto HUD — domain language

Terms meaningful to the product. Implementation details live in code and ADRs.

| Term | Meaning |
|------|---------|
| **Companion** | Phone app that ingests navigation + media and sends HUD state to the Pi over BLE (or dev HTTP). |
| **Hub** | Go service on the Pi: applies nav/media, composites frames, drives the panel, handles buttons. |
| **Canvas** | Logical 250×122 grayscale HUD frame before a display backend maps it to hardware. |
| **Host / backend** | Panel adapter selected with `-host` (`inky`, `waveshare`, `lcd`, `png`, …). |
| **Nav engine** | Platform-specific source of turn-by-turn data that fills the shared `nav` protocol (e.g. OsmAnd on Android, MapKit on iOS). |
| **Maneuver** | Next turn class on the wire: `left`, `right`, `slight_*`, `straight`, `u_turn`, `roundabout`, `arrive`, `depart`, `unknown`. |
| **Lanes** | Optional per-lane guidance array on `nav`; omitted when the engine cannot provide it. |
| **Refresh gate** | Policy that limits e-ink redraws (maneuver/road change, coarse distance buckets, force). |
| **Partial / full refresh** | Waveshare B/W update modes (~0.3s quiet vs ~2s wipe); full after every N partials. |
| **Junction IR** | Semantic turn-scene description (`nav.junction`) drawn as idealized templates by `kind`; replaces geographic minimap polylines ([ADR 0013](docs/adr/0013-junction-ir-replaces-minimap.md)). |
| **Ribbon** | Secondary left-column corridor schematic when neither junction nor synthesizable maneuver exists. |
| **Link** | BLE (or HTTP injector) connection carrying `nav`, `media`, `cmd`, `heartbeat`. |

Decisions: [`docs/adr/`](docs/adr/).
