package hud

import (
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// ManeuverNodes is the 40×40 design-kit maneuver glyph as scene primitives.
func ManeuverNodes(m protocol.Maneuver) []scene.Node {
	var b scene.Builder
	switch m {
	case protocol.ManeuverLeft:
		b.Line(20, 34, 20, 17)
		b.Line(20, 17, 8, 17)
		b.Polygon([][2]float64{{1, 17}, {8, 12.7}, {8, 21.3}}, true)
	case protocol.ManeuverRight:
		b.Line(20, 34, 20, 17)
		b.Line(20, 17, 32, 17)
		b.Polygon([][2]float64{{39, 17}, {32, 21.3}, {32, 12.7}}, true)
	case protocol.ManeuverSlightLeft:
		b.Line(21, 34, 21, 24)
		b.Line(21, 24, 11, 9)
		b.Polygon([][2]float64{{7, 4}, {15.2, 4.3}, {10.4, 11.5}}, true)
	case protocol.ManeuverSlightRight:
		b.Line(19, 34, 19, 24)
		b.Line(19, 24, 29, 9)
		b.Polygon([][2]float64{{33, 4}, {29.6, 11.5}, {24.8, 4.3}}, true)
	case protocol.ManeuverStraight:
		b.Line(20, 34, 20, 14)
		b.Polygon([][2]float64{{20, 7}, {24.3, 14}, {15.7, 14}}, true)
	case protocol.ManeuverUTurn:
		b.Path(`M14,34 V16 A6,6 0 0 1 26,16 V26`, false, 1)
		b.Polygon([][2]float64{{26, 33}, {21.7, 26}, {30.3, 26}}, true)
	case protocol.ManeuverRoundabout:
		b.Line(20, 34, 20, 27)
		b.Circle("", 20, 17, 9, false, 1)
		b.Polygon([][2]float64{{29, 8}, {22, 12.3}, {22, 3.7}}, true)
	case protocol.ManeuverArrive:
		b.Line(13, 34, 13, 7)
		b.Polygon([][2]float64{{13, 7}, {29, 11}, {13, 17}}, true)
	case protocol.ManeuverDepart:
		b.Circle("", 20, 27, 4, true, 0)
		b.Line(20, 22, 20, 8)
		b.Polygon([][2]float64{{20, 4}, {24.3, 11}, {15.7, 11}}, true)
	default:
		b.Circle("", 20, 20, 10, false, 1)
		b.Line(20, 12, 20, 22)
		b.Circle("", 20, 27, 1.5, true, 0)
	}
	return b.Nodes()
}

func maneuverPaths(m protocol.Maneuver) string {
	return svg.Fragment(ManeuverNodes(m))
}
