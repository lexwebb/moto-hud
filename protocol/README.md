# Moto HUD Protocol

JSON messages over BLE GATT. Device advertises as **MotoHUD**.

## UUIDs

See [uuids.json](uuids.json).

| Role | Direction | Characteristic |
|------|-----------|----------------|
| `nav` | Phone → Pi (write) | navigation state |
| `media` | Phone → Pi (write) | now playing |
| `cmd` | Pi → Phone (notify) | media commands from buttons |
| `heartbeat` | either (write/notify) | keep-alive |

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

## Google Maps tips

If nav fields are empty, disable Google Maps **Live Updates** / **Live info** notification categories in Android system settings for Maps, then restart navigation.
