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
  "remaining_m": 5400,
  "maneuver": "left",
  "lanes": [
    {"directions": ["straight"], "active": false},
    {"directions": ["left", "straight"], "active": true}
  ],
  "then_next": {
    "maneuver": "right",
    "distance_m": 400,
    "distance_text": "400 m",
    "instruction": "Turn right",
    "road": "Bridge Rd"
  },
  "ribbon_points": [
    {"x": 0, "y": 0},
    {"x": 0, "y": 120},
    {"x": -40, "y": 180}
  ],
  "ribbon_turn": 2,
  "junction": {
    "kind": "crossroads",
    "drive": "left",
    "outbound": "right",
    "through": true,
    "sides": [
      { "side": "left", "at": "before", "style": "dashed" },
      { "side": "left", "at": "at", "style": "dashed" }
    ]
  }
}
```

`maneuver` enum: `left`, `right`, `straight`, `slight_left`, `slight_right`, `u_turn`, `roundabout`, `arrive`, `depart`, `unknown`

Optional `lanes` (left→right): each lane lists allowed `directions` (same strings as `maneuver`) and whether it is `active` for the route. Omitted when the engine cannot provide lane guidance (e.g. MapKit, stock AIDL). Optional `then_next` is the following maneuver. Optional `remaining_m` is distance to destination.

Optional `ribbon_points` / `ribbon_turn`: local-unit corridor vertices (Y ahead, X right). Secondary path when `junction` is absent and the Pi cannot synthesize from `maneuver`. The Android companion may still fill these from a short public-OSRM probe.

Optional `junction` (**replaces** `minimap` — do not send geographic polylines): semantic turn-scene IR. Shared frame: approach from bottom, ahead toward top; proportions belong to the renderer. TypeScript: [`junction.ts`](junction.ts). Decision: [ADR 0013](../docs/adr/0013-junction-ir-replaces-minimap.md).

#### `junction` shared fields

| Field | Meaning |
|-------|---------|
| `kind` | Template family (required when `junction` present) |
| `drive` | `right` \| `left`; omit → `right` |
| `outbound` | Our exit relative to approach: `left` / `right` / `slight_*` / `straight` / `u_turn` |
| `through` | Main corridor continues past the decision |
| `sides[]` | Extra arms: `side` `left`\|`right`, `at` `before`\|`at`\|`after`, `style` `dashed`\|`solid` |

#### `junction` kinds (v1 draw + reserved)

| Kind | Extra fields | Diagram idea | v1 draw |
|------|--------------|--------------|---------|
| `simple` | optional `sides` | Spine + outbound kink (+ stubs) | yes |
| `t_junction` | — | T bar; `through` false | yes |
| `crossroads` | — | + solid route, dashed others | yes |
| `fork` | outbound slight L/R | Y; route solid, other dashed | yes |
| `merge` | optional `side` | Side joins spine | yes |
| `dual_carriageway` | optional `cross_median` | Twin parallels; median gap if crossing | yes |
| `roundabout` | `exits` (2–6), `exit` (1-based) | Ring + ticks; ours solid | yes |
| `ramp_exit` | `through` true | Mainline + diverging slip | yes |
| `ramp_enter` | optional `side` | Slip → mainline | yes |
| `u_turn` | — | Canned U (flip with `drive`) | yes |
| `arrive` / `depart` | — | End/start mark on short spine | yes |
| `jughandle`, `interchange`, `gyratory` | — | Reserved; render as `simple` | no |

`cross_median`: set when `kind=dual_carriageway` and outbound is a hard left/right (not slight/straight). Dual detection is **inferred** on the companion (same-name opposite oneway, or parallel opposite oneway); OBF has no `dual_carriageway` tag.

Phone produces rich `junction` when possible (OsmAnd Full Library). When omitted, Pi synthesizes `{ kind, outbound }` from `maneuver` (`left`/`right`→`simple`, `slight_*`→`fork`, `roundabout`→`roundabout` exits=4 exit=2, etc.); `sides` empty. Unknown `kind` → `simple` fallback.

#### OsmAnd TurnType / classifier → `kind`

Maps through protocol `maneuver` first ([`ManeuverParser.fromOsmandTurnType`](../android/app/src/main/java/com/motohud/companion/Models.kt)); rich producers then upgrade `kind` from topology.

| OsmAnd `TurnType` | `maneuver` | Default / synthesized `kind` | Rich upgrade when |
|-------------------|------------|------------------------------|-------------------|
| C (1) | `straight` | `simple` | dual → `dual_carriageway`; arms → `crossroads` |
| TL (2), TSHL (4) | `left` | `simple` | T / cross / dual / `cross_median` |
| TR (5), TSHR (7) | `right` | `simple` | same |
| TSLL (3), KL (8) | `slight_left` | `fork` | dual context may still be `dual_carriageway` |
| TSLR (6), KR (9) | `slight_right` | `fork` | same |
| TU (10), TRU (11) | `u_turn` | `u_turn` | — |
| RNDB (13), RNLB (14) | `roundabout` | `roundabout` | fill `exits`/`exit` when known |
| OFFR (12) | `unknown` | `simple` | — |
| (arrive / depart) | `arrive` / `depart` | `arrive` / `depart` | — |
| `*_link` highway diverge | (turn class) | `ramp_exit` | attached mainline continues |
| `*_link` highway merge | (turn class) | `ramp_enter` | slip joins spine |
| classifier: L+R+through arms | — | `crossroads` | — |
| classifier: one side, no through | — | `t_junction` | — |
| classifier: dual inferred | — | `dual_carriageway` | `cross_median` if hard L/R |

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
| 1a | **OsmAnd embedded** (`embedded` flavor) | Full Library `RoutingHelper` poll | Lanes, then-next, ETA, street, remaining |
| 1b | **OsmAnd AIDL** (`aidl` flavor) | `registerForNavigationUpdates` + voice + OsmAnd notif enrich | Typed turn/distance; soft road/ETA from notifications |
| 2 | Google Maps / Maps Go | `NotificationListenerService` text scrape | Fallback when OsmAnd is absent or not actively navigating |

While OsmAnd reports `active` navigation, Maps notification updates are ignored.

### OsmAnd tips

- Install [OsmAnd](https://play.google.com/store/apps/details?id=net.osmand) (`net.osmand`) or OsmAnd+ (`net.osmand.plus`).
- Start HUD link after OsmAnd is installed so the companion can bind `net.osmand.aidl.OsmandAidlServiceV2`.
- `TurnType` ints map to protocol `maneuver` (e.g. TL→`left`, TSLL/KL→`slight_left`, RNDB→`roundabout`).

### Google Maps tips

If nav fields are empty, disable Google Maps **Live Updates** / **Live info** notification categories in Android system settings for Maps, then restart navigation.

## Road ribbon / junction

Active nav with live geometry uses a **two-column** layout: left = junction diagram or corridor, right = compacted hero distance + road + ETA (no maneuver arrows).

- Prefer optional `junction`: semantic IR → idealized template by `kind` (immediate decision only).
- Else synthesize a minimal junction from `maneuver`.
- Else `ribbon_points` schematic corridor.
- Else synthetic kink from `maneuver` in the classic stack (glyph + distance + road + bottom ribbon); unknown → dashed placeholder.

`nav.minimap` (meter polylines) is **removed** from the live HUD path. Emulator/lab may still classify offline; production rich IR comes from OsmAnd Full Library, not AIDL.
