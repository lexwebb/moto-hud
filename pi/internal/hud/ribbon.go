package hud

import (
	"fmt"
	"math"
	"strings"

	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// RoadPoint is a schematic corridor vertex in local units (not lat/lng).
// Y increases "ahead"; rendering flips Y so ahead is toward the top of the band.
type RoadPoint struct {
	X, Y float64
}

const ribbonDefaultH = 36

// roadRibbonSVG draws the design-kit RoadRibbon (legacy string wrapper).
func roadRibbonSVG(points []RoadPoint, turnIndex int, w, h int) string {
	return svg.Fragment(RoadRibbonNodes(points, turnIndex, w, h))
}

// RoadRibbonNodes draws the design-kit RoadRibbon as scene nodes.
func RoadRibbonNodes(points []RoadPoint, turnIndex int, w, h int) []scene.Node {
	if w <= 0 {
		w = mainWidth()
	}
	if h <= 0 {
		h = ribbonDefaultH
	}
	if len(points) < 2 {
		cx := w / 2
		return []scene.Node{
			scene.Line{X1: cx, Y1: 4, X2: cx, Y2: h - 4, StrokeWidth: 2, Dash: "4 5"},
		}
	}

	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	const pad = 5.0
	rangeX := maxX - minX
	if rangeX == 0 {
		rangeX = 1
	}
	rangeY := maxY - minY
	if rangeY == 0 {
		rangeY = 1
	}
	scale := math.Min((float64(w)-pad*2)/rangeX, (float64(h)-pad*2)/rangeY)
	plottedW := rangeX * scale
	offsetX := pad + ((float64(w) - pad*2) - plottedW) / 2
	screenBottom := float64(h) - pad
	sx := func(x float64) float64 { return offsetX + (x-minX)*scale }
	sy := func(y float64) float64 { return screenBottom - (y-minY)*scale }

	var d strings.Builder
	for i, p := range points {
		if i == 0 {
			fmt.Fprintf(&d, "M%.1f,%.1f", sx(p.X), sy(p.Y))
		} else {
			fmt.Fprintf(&d, " L%.1f,%.1f", sx(p.X), sy(p.Y))
		}
	}

	var out []scene.Node
	out = append(out, scene.Path{
		D: d.String(), Filled: false, StrokeWidth: 3,
	})
	if turnIndex >= 0 && turnIndex < len(points) {
		t := points[turnIndex]
		tx := int(sx(t.X) - 3)
		ty := int(sy(t.Y) - 3)
		out = append(out, scene.Rect{ID: "", X: tx, Y: ty, W: 6, H: 6, Filled: true})
	}
	return out
}

// RibbonNodesForNav builds ribbon scene nodes for the active nav message.
func RibbonNodesForNav(nav protocol.NavMessage, w, h int) []scene.Node {
	pts, turnIdx := ribbonForNav(nav)
	return RoadRibbonNodes(pts, turnIdx, w, h)
}

// ribbonForNav prefers phone-supplied corridor points; otherwise a canned schematic.
func ribbonForNav(nav protocol.NavMessage) (points []RoadPoint, turnIndex int) {
	if len(nav.RibbonPoints) >= 2 {
		pts := make([]RoadPoint, len(nav.RibbonPoints))
		for i, p := range nav.RibbonPoints {
			pts[i] = RoadPoint{X: p.X, Y: p.Y}
		}
		turn := nav.RibbonTurn
		if turn < 0 || turn >= len(pts) {
			turn = -1
		}
		return pts, turn
	}
	return schematicRibbonForManeuver(nav.Maneuver)
}

// schematicRibbonForManeuver returns canned corridor geometry matching the design kit.
func schematicRibbonForManeuver(m protocol.Maneuver) (points []RoadPoint, turnIndex int) {
	switch m {
	case protocol.ManeuverLeft, protocol.ManeuverSlightLeft:
		return []RoadPoint{{110, 0}, {110, 22}, {50, 34}}, 1
	case protocol.ManeuverRight, protocol.ManeuverSlightRight:
		return []RoadPoint{{110, 0}, {110, 22}, {170, 34}}, 1
	case protocol.ManeuverUTurn:
		return []RoadPoint{{110, 0}, {110, 18}, {150, 18}, {150, 8}}, 2
	case protocol.ManeuverRoundabout:
		return []RoadPoint{{110, 0}, {110, 14}, {135, 22}, {110, 30}}, 1
	case protocol.ManeuverArrive:
		return []RoadPoint{{110, 0}, {110, 28}}, 1
	case protocol.ManeuverDepart, protocol.ManeuverStraight:
		return []RoadPoint{{110, 0}, {110, 36}}, 1
	default:
		return nil, -1
	}
}
