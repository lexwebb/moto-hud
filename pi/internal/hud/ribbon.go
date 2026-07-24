package hud

import (
	"fmt"
	"math"
	"strings"

	"moto-hud/pi/internal/protocol"
)

// RoadPoint is a schematic corridor vertex in local units (not lat/lng).
// Y increases "ahead"; rendering flips Y so ahead is toward the top of the band.
type RoadPoint struct {
	X, Y float64
}

const ribbonDefaultH = 36

// roadRibbonSVG draws the design-kit RoadRibbon as raw SVG shapes.
// Empty/short points → dashed vertical "no corridor" line.
func roadRibbonSVG(points []RoadPoint, turnIndex int, w, h int) string {
	if w <= 0 {
		w = mainWidth()
	}
	if h <= 0 {
		h = ribbonDefaultH
	}
	if len(points) < 2 {
		cx := w / 2
		return fmt.Sprintf(
			`<line x1="%d" y1="4" x2="%d" y2="%d" stroke="#000" stroke-width="2" stroke-dasharray="4 5" stroke-linecap="square"/>`,
			cx, cx, h-4,
		)
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

	var b strings.Builder
	fmt.Fprintf(&b,
		`<path d="%s" fill="none" stroke="#000" stroke-width="3" stroke-linejoin="miter" stroke-linecap="square"/>`,
		d.String(),
	)
	if turnIndex >= 0 && turnIndex < len(points) {
		t := points[turnIndex]
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="6" height="6" fill="#000"/>`,
			sx(t.X)-3, sy(t.Y)-3)
	}
	return b.String()
}

// schematicRibbonForManeuver returns canned corridor geometry matching the design kit.
// Real route polylines can replace this later via protocol without changing the SVG drawer.
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
