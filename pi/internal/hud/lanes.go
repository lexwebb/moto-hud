package hud

import (
	"fmt"
	"strings"

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

// laneStripSVG draws a 1-bit left-to-right lane strip (active = filled, inactive = outline).
func laneStripSVG(lanes []protocol.LaneInfo, maxW int) string {
	if len(lanes) == 0 {
		return ""
	}
	n := len(lanes)
	boxW, boxH, gap := laneBoxW, laneBoxH, laneGap
	totalW := n*boxW + (n-1)*gap
	if totalW > maxW && n > 0 {
		boxW = (maxW - (n-1)*gap) / n
		if boxW < 6 {
			boxW = 6
		}
	}
	var b strings.Builder
	b.WriteString(`<g id="lanes" fill="#000" stroke="#000" stroke-width="1" stroke-linecap="square">`)
	x := 0
	for _, lane := range lanes {
		if lane.Active {
			fmt.Fprintf(&b, `<rect x="%d" y="0" width="%d" height="%d"/>`, x, boxW, boxH)
		} else {
			fmt.Fprintf(&b, `<rect x="%d" y="0" width="%d" height="%d" fill="none"/>`, x, boxW, boxH)
		}
		dir := primaryLaneDir(lane.Directions)
		cx := x + boxW/2
		// Direction tick sits just above the box so it stays visible on filled lanes.
		switch dir {
		case protocol.ManeuverLeft, protocol.ManeuverSlightLeft:
			fmt.Fprintf(&b, `<polyline fill="none" points="%d,-1 %d,-4 %d,-1"/>`, cx+3, cx-2, cx+3)
		case protocol.ManeuverRight, protocol.ManeuverSlightRight:
			fmt.Fprintf(&b, `<polyline fill="none" points="%d,-1 %d,-4 %d,-1"/>`, cx-3, cx+2, cx-3)
		default:
			fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d"/>`, cx, -1, cx, -5)
		}
		x += boxW + gap
	}
	b.WriteString(`</g>`)
	return b.String()
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
