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

// MinimapMessage is deprecated geographic polylines (superseded by JunctionMessage / ADR 0013).
// Kept for unmarshaling old fixtures until the Pi renderer fully migrates.
type MinimapMessage struct {
	Route   []RibbonPoint   `json:"route,omitempty"`
	Context [][]RibbonPoint `json:"context,omitempty"`
	Rider   *RibbonPoint    `json:"rider,omitempty"`
}

// JunctionSideArm is an extra arm on the approach or at the decision node.
type JunctionSideArm struct {
	Side  string `json:"side"`  // left | right
	At    string `json:"at"`    // before | at | after
	Style string `json:"style"` // dashed | solid
}

// JunctionMessage is the semantic turn-scene IR on nav.junction (replaces minimap).
// Shared frame: approach from bottom, ahead toward top. See protocol/junction.ts.
type JunctionMessage struct {
	Kind        string            `json:"kind"`
	Drive       string            `json:"drive,omitempty"` // left | right; omit → right
	Outbound    string            `json:"outbound"`
	Through     bool              `json:"through"`
	Sides       []JunctionSideArm `json:"sides,omitempty"`
	CrossMedian bool              `json:"cross_median,omitempty"` // dual_carriageway
	Exits       int               `json:"exits,omitempty"`        // roundabout 2–6
	Exit        int               `json:"exit,omitempty"`         // roundabout 1-based
	Side        string            `json:"side,omitempty"`         // merge / ramp_enter
}

// LaneInfo is one lane in left-to-right order at the upcoming junction.
// Directions use the same strings as Maneuver (typically a subset).
type LaneInfo struct {
	Directions []string `json:"directions"`
	Active     bool     `json:"active"`
}

// ThenNextMessage is the maneuver after the immediate next turn (optional).
type ThenNextMessage struct {
	Maneuver     Maneuver `json:"maneuver"`
	DistanceM    int      `json:"distance_m,omitempty"`
	DistanceText string   `json:"distance_text,omitempty"`
	Instruction  string   `json:"instruction,omitempty"`
	Road         string   `json:"road,omitempty"`
}

type NavMessage struct {
	Type         string           `json:"type"`
	Active       bool             `json:"active"`
	Instruction  string           `json:"instruction"`
	DistanceM    int              `json:"distance_m"`
	DistanceText string           `json:"distance_text"`
	Road         string           `json:"road"`
	EtaMin       int              `json:"eta_min"`
	RemainingM   int              `json:"remaining_m,omitempty"`
	Maneuver     Maneuver         `json:"maneuver"`
	Lanes        []LaneInfo       `json:"lanes,omitempty"`
	ThenNext     *ThenNextMessage `json:"then_next,omitempty"`
	RibbonPoints []RibbonPoint     `json:"ribbon_points,omitempty"`
	RibbonTurn   int               `json:"ribbon_turn,omitempty"`
	Junction     *JunctionMessage  `json:"junction,omitempty"`
	Minimap      *MinimapMessage   `json:"minimap,omitempty"` // deprecated; prefer Junction
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
