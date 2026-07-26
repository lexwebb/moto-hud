package protocol

// BLE UUIDs shared by Pi and Android companion.
const (
	ServiceUUID     = "6e400001-b5a3-f393-e0a9-e50e24dcca9e"
	NavUUID         = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	MediaUUID       = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
	CmdUUID         = "6e400004-b5a3-f393-e0a9-e50e24dcca9e"
	HeartbeatUUID   = "6e400005-b5a3-f393-e0a9-e50e24dcca9e"
	DeviceName      = "MotoHUD"
)

type Maneuver string

const (
	ManeuverLeft        Maneuver = "left"
	ManeuverRight       Maneuver = "right"
	ManeuverStraight    Maneuver = "straight"
	ManeuverSlightLeft  Maneuver = "slight_left"
	ManeuverSlightRight Maneuver = "slight_right"
	ManeuverUTurn       Maneuver = "u_turn"
	ManeuverRoundabout  Maneuver = "roundabout"
	ManeuverArrive      Maneuver = "arrive"
	ManeuverDepart      Maneuver = "depart"
	ManeuverUnknown     Maneuver = "unknown"
)

// RibbonPoint is a corridor vertex in local units (not lat/lng).
// Y increases ahead of the rider; X is lateral (right positive).
type RibbonPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MinimapMessage is a top-down junction snapshot in meters.
// Origin ≈ next turn; +Y along the inbound approach (rider usually at negative Y).
type MinimapMessage struct {
	Route   []RibbonPoint   `json:"route,omitempty"`
	Context [][]RibbonPoint `json:"context,omitempty"`
	Rider   *RibbonPoint    `json:"rider,omitempty"`
}

type NavMessage struct {
	Type         string          `json:"type"`
	Active       bool            `json:"active"`
	Instruction  string          `json:"instruction"`
	DistanceM    int             `json:"distance_m"`
	DistanceText string          `json:"distance_text"`
	Road         string          `json:"road"`
	EtaMin       int             `json:"eta_min"`
	Maneuver     Maneuver        `json:"maneuver"`
	RibbonPoints []RibbonPoint   `json:"ribbon_points,omitempty"`
	RibbonTurn   int             `json:"ribbon_turn,omitempty"`
	Minimap      *MinimapMessage `json:"minimap,omitempty"`
}

type MediaMessage struct {
	Type    string `json:"type"`
	Playing bool   `json:"playing"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
}

type CmdAction string

const (
	CmdPlayPause  CmdAction = "play_pause"
	CmdNextTrack  CmdAction = "next_track"
	CmdPrevTrack  CmdAction = "prev_track"
)

type CmdMessage struct {
	Type   string    `json:"type"`
	Action CmdAction `json:"action"`
}

type HeartbeatMessage struct {
	Type string `json:"type"`
	Ts   int64  `json:"ts"`
}
