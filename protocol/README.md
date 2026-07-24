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

Optional `ribbon_points` / `ribbon_turn`: local-unit corridor vertices (Y ahead, X right). When present (≥2 points), the Pi draws that corridor; otherwise it falls back to a synthetic kink from `maneuver`. The Android companion fills these from a short public-OSRM probe using GPS + the Maps notification distance/maneuver (no Maps polyline API). Omit or leave empty when GPS/OSRM is unavailable.

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

## Google Maps tips

If nav fields are empty, disable Google Maps **Live Updates** / **Live info** notification categories in Android system settings for Maps, then restart navigation.

## Road ribbon

Active nav draws a **RoadRibbon** under the road name: a bold kinked corridor, not a map. Prefer optional `ribbon_points` from the phone (OSRM corridor guess + GPS). If missing, geometry is **synthetic from `maneuver`** on the Pi (`schematicRibbonForManeuver`); unknown maneuver → dashed placeholder. Progress ticks are omitted while the ribbon is shown.

OSRM uses the public demo server (`router.project-osrm.org`) with no API key — rate-limited and fine for personal use; revisit if it becomes a problem (self-host or another provider).
