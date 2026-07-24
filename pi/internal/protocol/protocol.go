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

type NavMessage struct {
	Type         string   `json:"type"`
	Active       bool     `json:"active"`
	Instruction  string   `json:"instruction"`
	DistanceM    int      `json:"distance_m"`
	DistanceText string   `json:"distance_text"`
	Road         string   `json:"road"`
	EtaMin       int      `json:"eta_min"`
	Maneuver     Maneuver `json:"maneuver"`
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
