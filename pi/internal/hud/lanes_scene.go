package hud

import (
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// LaneStripNodes draws left-to-right lane boxes and direction ticks in local coordinates.
func LaneStripNodes(lanes []protocol.LaneInfo, maxW int) []scene.Node {
	if len(lanes) == 0 {
		return nil
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
	var b scene.Builder
	x := 0
	for _, lane := range lanes {
		if lane.Active {
			b.Rect("", x, 0, boxW, boxH, true)
		} else {
			b.Rect("", x, 0, boxW, boxH, false)
		}
		dir := primaryLaneDir(lane.Directions)
		cx := x + boxW/2
		switch dir {
		case protocol.ManeuverLeft, protocol.ManeuverSlightLeft:
			b.Polyline([][2]int{{cx + 3, -1}, {cx - 2, -4}, {cx + 3, -1}}, false, 1)
		case protocol.ManeuverRight, protocol.ManeuverSlightRight:
			b.Polyline([][2]int{{cx - 3, -1}, {cx + 2, -4}, {cx - 3, -1}}, false, 1)
		default:
			b.Line(cx, -1, cx, -5)
		}
		x += boxW + gap
	}
	return []scene.Node{scene.Group{ID: "lanes", Children: b.Nodes()}}
}

func laneStripSVG(lanes []protocol.LaneInfo, maxW int) string {
	return svg.Fragment(LaneStripNodes(lanes, maxW))
}
