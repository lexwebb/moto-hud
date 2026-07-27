package hud

import "moto-hud/pi/internal/protocol"

// Local aliases for junction IR — types live in protocol (draft / ADR 0013).
// Kind strings match plan vocabulary; keep call sites on these constants.

type (
	JunctionSide    = protocol.JunctionSideArm
	JunctionMessage = protocol.JunctionMessage
)

// Kind / drive string constants (protocol stores them as plain strings).
const (
	JunctionSimple          = "simple"
	JunctionTJunction       = "t_junction"
	JunctionCrossroads      = "crossroads"
	JunctionFork            = "fork"
	JunctionMerge           = "merge"
	JunctionDualCarriageway = "dual_carriageway"
	JunctionRoundabout      = "roundabout"
	JunctionRampExit        = "ramp_exit"
	JunctionRampEnter       = "ramp_enter"
	JunctionUTurn           = "u_turn"
	JunctionArrive          = "arrive"
	JunctionDepart          = "depart"
	JunctionDriveRight      = "right"
	JunctionDriveLeft       = "left"
)

func junctionFromProtocol(j *protocol.JunctionMessage) *JunctionMessage {
	return j
}
