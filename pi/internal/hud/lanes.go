package hud

import (
	"moto-hud/pi/internal/protocol"
)
const (
	laneBoxW = 10
	laneBoxH = 14
	laneGap  = 2
)

func hasLanes(nav protocol.NavMessage) bool {
	return len(nav.Lanes) > 0
}

func primaryLaneDir(dirs []string) protocol.Maneuver {
	if len(dirs) == 0 {
		return protocol.ManeuverStraight
	}
	return protocol.Maneuver(dirs[0])
}

func laneStripWidth(lanes []protocol.LaneInfo, maxW int) int {
	n := len(lanes)
	if n == 0 {
		return 0
	}
	boxW := laneBoxW
	totalW := n*boxW + (n-1)*laneGap
	if totalW > maxW {
		boxW = (maxW - (n-1)*laneGap) / n
		if boxW < 6 {
			boxW = 6
		}
		totalW = n*boxW + (n-1)*laneGap
	}
	return totalW
}
