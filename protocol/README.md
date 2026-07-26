# Moto HUD Protocol

JSON messages over BLE GATT (phone ↔ Pi) and the same JSON over the **dev HTTP injector** on the Pi.

Device advertises as **MotoHUD**.

## UUIDs

See [uuids.json](uuids.json). Shared constants also live in:

- Go: `pi/internal/protocol/protocol.go`
- Kotlin: `android/.../Protocol.kt`

| Role | Direction | Characteristic |
|------|-----------|----------------|
| `nav` | Phone → Pi (write) | navigation state |
| `media` | Phone → Pi (write) | now playing |
| `cmd` | Pi → Phone (notify) | media commands from buttons |
| `heartbeat` | Phone → Pi (write) | keep-alive (~15s); also used as link presence |

## Messages

### `nav`

```json
{
  "type": "nav",
  "active": true,
  "instruction": "Turn left onto High St",
  "distance_m": 200,
  "distance_text": "200 m",
  "road": "High St",
  "eta_min": 12,
  "maneuver": "left",
  "ribbon_points": [
    {"x": 0, "y": 0},
    {"x": 0, "y": 120},
    {"x": -40, "y": 180}
  ],
  "ribbon_turn": 2
}
```

`maneuver` enum: `left`, `right`, `straight`, `slight_left`, `slight_right`, `u_turn`, `roundabout`, `arrive`, `depart`, `unknown`

Optional `ribbon_points` / `ribbon_turn`: local-unit corridor vertices (Y ahead, X right). When present (≥2 points) and `minimap` is absent, the Pi draws that corridor in the live two-column layout; otherwise it falls back to a synthetic kink from `maneuver`. The Android companion fills these from a short public-OSRM probe.

Optional `minimap` (preferred when available): top-down **junction snapshot** in meters. Origin ≈ next turn; +Y along the inbound approach (rider usually at negative Y). The Pi fits orthographically (no perspective): dashed `context`, solid `route`, turn mark at origin, `rider` blob.

```json
"minimap": {
  "route": [{"x": 0, "y": -40}, {"x": 0, "y": 0}, {"x": 18, "y": 25}],
  "context": [
    [{"x": -20, "y": -30}, {"x": -18, "y": 40}]
  ],
  "rider": {"x": 0, "y": -35}
}
```

The browser emulator builds this from the ride polyline + baked OSM highways around the next maneuver; Android does not send it yet.

### `media`

```json
{
  "type": "media",
  "playing": true,
  "title": "Song Title",
  "artist": "Artist Name"
}
```

### `cmd`

```json
{
  "type": "cmd",
  "action": "play_pause"
}
```

`action` enum: `play_pause`, `next_track`, `prev_track`

### `heartbeat`

```json
{
  "type": "heartbeat",
  "ts": 1710000000
}
```

Phone writes heartbeat while linked. There is **no timeout / auto-unlink** on the Pi yet — `linked` flips when the BLE stack reports connect/disconnect (or stays true for HTTP-only bring-up).

## Dev HTTP injector (Pi)

When `motohud` is started with `-http :8787`, the hub exposes:

| Method | Path | Body | Effect |
|--------|------|------|--------|
| `GET` | `/health` | — | `ok` |
| `POST` | `/nav` | `nav` JSON | `ApplyNav` |
| `POST` | `/media` | `media` JSON | `ApplyMedia` |
| `POST` | `/button` | plain text `prev` / `next` / `action` / `*_long` | `HandleButtonEvent` |
| `GET` | `/frame.png` | — | current 1-bit frame |
| `GET` | `/preview/`, `/emulator/` | — | static web tools |

The Android companion can **also** POST `/nav` and `/media` to this base URL (emulator host loopback `http://10.0.2.2:8787`) while BLE is optional. HTTP has no `/heartbeat` and does not carry `cmd` notifies back to the phone — use BLE for button→media control, or press buttons via `/button` / the emulator UI.

Payload size: keep messages small (notification-sized). BLE writes use no-response; aim well under a single ATT MTU (~20–180 bytes typical without negotiation).

## Nav sources (companions)

Wire format is **engine-agnostic** ([ADR 0006](../docs/adr/0006-engine-agnostic-nav-android-osmand-ios-mapkit.md)): Android → OsmAnd (AIDL now; Full Library planned); iOS → MapKit; optional fields like `lanes` omitted when unavailable.

| Priority | Source | How | Notes |
|----------|--------|-----|--------|
| 1 | **OsmAnd** (free or +) | AIDL `registerForNavigationUpdates` → `ADirectionInfo` | Typed `turnType` + `distanceTo`; no text scrape. Road name / ETA not in this callback yet. |
| 2 | Google Maps / Maps Go | `NotificationListenerService` text scrape | Fallback when OsmAnd is absent or not actively navigating. |

While OsmAnd reports `active` navigation, Maps notification updates are ignored.

### OsmAnd tips

- Install [OsmAnd](https://play.google.com/store/apps/details?id=net.osmand) (`net.osmand`) or OsmAnd+ (`net.osmand.plus`).
- Start HUD link after OsmAnd is installed so the companion can bind `net.osmand.aidl.OsmandAidlServiceV2`.
- `TurnType` ints map to protocol `maneuver` (e.g. TL→`left`, TSLL/KL→`slight_left`, RNDB→`roundabout`).

### Google Maps tips

If nav fields are empty, disable Google Maps **Live Updates** / **Live info** notification categories in Android system settings for Maps, then restart navigation.

## Road ribbon / minimap

Active nav with live geometry uses a **two-column** layout: left = corridor or turn snapshot, right = compacted hero distance + road + ETA (no maneuver arrows).

- Prefer optional `minimap`: top-down orthographic snapshot of the **next turn** (dashed OSM context, solid route, rider + turn marks). Frame stays locked to the junction; the rider blob moves as you approach.
- Else `ribbon_points` schematic corridor.
- Else synthetic kink from `maneuver` in the classic stack (glyph + distance + road + bottom ribbon); unknown → dashed placeholder.

OSRM (phone) uses the public demo server for `ribbon_points` only today. The emulator proves `minimap` offline via a baked OSM extract beside the Whitehall route.
