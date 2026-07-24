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
  "maneuver": "left"
}
```

`maneuver` enum: `left`, `right`, `straight`, `slight_left`, `slight_right`, `u_turn`, `roundabout`, `arrive`, `depart`, `unknown`

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

Active nav draws a schematic **RoadRibbon** (design kit) under the road name: a bold kinked corridor, not a map. Geometry is **synthetic from `maneuver`** on the Pi today (`schematicRibbonForManeuver`); unknown maneuver → dashed placeholder. Progress ticks are omitted while the ribbon is shown.

Future: optional `ribbon_points` on `nav` from the phone for real upcoming geometry; drawer already accepts arbitrary points.
