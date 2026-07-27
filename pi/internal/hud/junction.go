package hud

import (
	"fmt"
	"image"
	"math"

	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// PreferJunctionTemplates selects the semantic IR drawer for the live left column
// when nav carries a JunctionMessage. Default false keeps the production meter
// minimap path (MinimapNodes) untouched.
//
// Hook points:
//   - compose/nav_live.go left corridor (via DrawDeps.HasJunction / JunctionNodes)
//   - this file's JunctionNodes templates (replaces RDP/octilinear meter fit eventually)
//   - site junction-lab preview (JS mirror of the same IR shape)
// See also: minimap.go (legacy meter path) and testdata/junction/README.md.
var PreferJunctionTemplates = false

const (
	junctionStroke    = 2.0
	junctionDash      = "3 3"
	junctionThick     = 3.0
)

// HasJunction reports whether nav carries usable semantic junction IR.
func HasJunction(nav protocol.NavMessage) bool {
	return JunctionFromNav(nav) != nil
}

// JunctionFromNav returns IR from nav.junction when present.
func JunctionFromNav(nav protocol.NavMessage) *JunctionMessage {
	if nav.Junction == nil {
		return nil
	}
	j := junctionFromProtocol(nav.Junction)
	if j == nil || j.Kind == "" {
		return nil
	}
	return j
}

// JunctionNodesForNav picks nav.junction or synthesizes from maneuver (Pi fallback).
func JunctionNodesForNav(nav protocol.NavMessage, w, h int) []scene.Node {
	j := JunctionFromNav(nav)
	if j == nil {
		j = SynthesizeJunctionFromManeuver(nav.Maneuver)
	}
	return JunctionNodes(j, w, h)
}

// JunctionNodes draws an idealized template from semantic IR (not meter polylines).
// Unknown kinds fall back to simple. Pane defaults match the live left column (~70×80).
func JunctionNodes(j *JunctionMessage, w, h int) []scene.Node {
	if w <= 0 {
		w = 70
	}
	if h <= 0 {
		h = 80
	}
	if j == nil || j.Kind == "" {
		cx := w / 2
		return []scene.Node{scene.Line{X1: cx, Y1: 4, X2: cx, Y2: h - 4, StrokeWidth: junctionStroke}}
	}

	kind := j.Kind
	switch kind {
	case JunctionSimple:
		return drawSimple(j, w, h)
	case JunctionTJunction:
		return drawTJunction(j, w, h)
	case JunctionCrossroads:
		return drawCrossroads(j, w, h)
	case JunctionFork:
		return drawFork(j, w, h)
	case JunctionMerge:
		return drawMerge(j, w, h)
	case JunctionDualCarriageway:
		return drawDualCarriageway(j, w, h)
	case JunctionRoundabout:
		return drawRoundabout(j, w, h)
	case JunctionRampExit:
		return drawRampExit(j, w, h)
	case JunctionRampEnter:
		return drawRampEnter(j, w, h)
	case JunctionUTurn:
		return drawUTurn(j, w, h)
	case JunctionArrive:
		return drawArrive(j, w, h)
	case JunctionDepart:
		return drawDepart(j, w, h)
	default:
		// Unknown → simple fallback (plan).
		fb := *j
		fb.Kind = JunctionSimple
		return drawSimple(&fb, w, h)
	}
}

// JunctionSVGFragment returns raw SVG shapes for lab/WASM (no outer <svg>).
func JunctionSVGFragment(j *JunctionMessage, w, h int) string {
	return svg.Fragment(JunctionNodes(j, w, h))
}

// RenderJunction rasterizes the junction pane alone (lab / unit tests).
func RenderJunction(j *JunctionMessage, w, h int) (*image.Gray, error) {
	if w <= 0 {
		w = 70
	}
	if h <= 0 {
		h = 80
	}
	frag := JunctionSVGFragment(j, w, h)
	svgDoc := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="%d" height="%d" fill="#fff"/>%s</svg>`,
		w, h, w, h, w, h, frag,
	)
	return RasterizeSVGAt([]byte(svgDoc), w, h)
}

// SynthesizeJunctionFromManeuver builds a minimal IR when the phone omits junction.
// Mirrors protocol/junction.ts synthesizeJunctionFromManeuver (drive omit → right).
func SynthesizeJunctionFromManeuver(m protocol.Maneuver) *JunctionMessage {
	switch m {
	case protocol.ManeuverArrive:
		return &JunctionMessage{Kind: JunctionArrive, Outbound: "straight", Through: false}
	case protocol.ManeuverDepart:
		return &JunctionMessage{Kind: JunctionDepart, Outbound: "straight", Through: true}
	case protocol.ManeuverUTurn:
		return &JunctionMessage{Kind: JunctionUTurn, Outbound: "u_turn", Through: false}
	case protocol.ManeuverRoundabout:
		return &JunctionMessage{Kind: JunctionRoundabout, Outbound: "right", Through: false, Exits: 4, Exit: 2}
	case protocol.ManeuverSlightLeft:
		return &JunctionMessage{Kind: JunctionFork, Outbound: "slight_left", Through: false}
	case protocol.ManeuverSlightRight:
		return &JunctionMessage{Kind: JunctionFork, Outbound: "slight_right", Through: false}
	case protocol.ManeuverStraight:
		return &JunctionMessage{Kind: JunctionSimple, Outbound: "straight", Through: true}
	case protocol.ManeuverLeft:
		return &JunctionMessage{Kind: JunctionSimple, Outbound: "left", Through: false}
	case protocol.ManeuverRight:
		return &JunctionMessage{Kind: JunctionSimple, Outbound: "right", Through: false}
	default:
		return &JunctionMessage{Kind: JunctionSimple, Outbound: "straight", Through: true}
	}
}

func drawSimple(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	ob := normalizeOutbound(j.Outbound)

	// Approach spine (bottom → node).
	b.LineStyled(cx, bot, cx, cy, junctionThick, "")

	switch {
	case ob == "straight" || ob == "":
		if j.Through || ob == "straight" || ob == "" {
			b.LineStyled(cx, cy, cx, top, junctionThick, "")
		}
	case isLeftish(ob):
		ex := cx - (w * 2 / 5)
		if isSlight(ob) {
			ex = cx - (w / 5)
			b.LineStyled(cx, cy, ex, top+8, junctionThick, "")
		} else {
			b.LineStyled(cx, cy, ex, cy, junctionThick, "")
			if j.Through {
				b.LineStyled(cx, cy, cx, top, junctionStroke, junctionDash)
			}
		}
	case isRightish(ob):
		ex := cx + (w * 2 / 5)
		if isSlight(ob) {
			ex = cx + (w / 5)
			b.LineStyled(cx, cy, ex, top+8, junctionThick, "")
		} else {
			b.LineStyled(cx, cy, ex, cy, junctionThick, "")
			if j.Through {
				b.LineStyled(cx, cy, cx, top, junctionStroke, junctionDash)
			}
		}
	case ob == "u_turn":
		driveLeft := j.Drive == JunctionDriveLeft
		ux := cx + 14
		if driveLeft {
			ux = cx - 14
		}
		b.Path(fmt.Sprintf("M%d,%d V%d A8,8 0 0 1 %d,%d V%d", cx, cy, cy-10, ux, cy-10, cy+4), false, junctionThick)
	}

	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

func drawCrossroads(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	left, right := 8, w-8
	ob := normalizeOutbound(j.Outbound)

	// Unused arms dashed; route solid.
	arm := func(x1, y1, x2, y2 int, ours bool) {
		if ours {
			b.LineStyled(x1, y1, x2, y2, junctionThick, "")
		} else {
			b.LineStyled(x1, y1, x2, y2, junctionStroke, junctionDash)
		}
	}

	// Always draw four arms from center.
	arm(cx, bot, cx, cy, true) // inbound always ours
	arm(cx, cy, cx, top, ob == "straight" || ob == "")
	arm(cx, cy, left, cy, isLeftish(ob) && !isSlight(ob))
	arm(cx, cy, right, cy, isRightish(ob) && !isSlight(ob))
	if isSlight(ob) && isLeftish(ob) {
		arm(cx, cy, cx-(w/5), top+6, true)
	}
	if isSlight(ob) && isRightish(ob) {
		arm(cx, cy, cx+(w/5), top+6, true)
	}

	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

func drawDualCarriageway(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	sep := 7
	driveLeft := j.Drive == JunctionDriveLeft
	// Our carriageway sits on the traffic side.
	ourX, oppX := cx-sep, cx+sep
	if !driveLeft {
		ourX, oppX = cx+sep, cx-sep
	}
	ob := normalizeOutbound(j.Outbound)

	if j.CrossMedian && (isLeftish(ob) || isRightish(ob)) {
		// Gap in median around the node; route crosses then exits or continues.
		gap := 10
		b.LineStyled(ourX, bot, ourX, cy+gap, junctionThick, "")
		b.LineStyled(oppX, bot, oppX, cy+gap, junctionStroke, junctionDash)
		b.LineStyled(ourX, cy+gap, oppX, cy-gap, junctionThick, "")
		if isLeftish(ob) {
			b.LineStyled(oppX, cy-gap, 8, cy-gap, junctionThick, "")
			if j.Through {
				b.LineStyled(oppX, cy-gap, oppX, top, junctionStroke, junctionDash)
			}
		} else {
			b.LineStyled(oppX, cy-gap, w-8, cy-gap, junctionThick, "")
			if j.Through {
				b.LineStyled(oppX, cy-gap, oppX, top, junctionStroke, junctionDash)
			}
		}
	} else {
		// Parallel twins: we stay on our side; opposite dashed.
		b.LineStyled(ourX, bot, ourX, top, junctionThick, "")
		b.LineStyled(oppX, bot, oppX, top, junctionStroke, junctionDash)
		if isLeftish(ob) && !j.CrossMedian {
			b.LineStyled(ourX, cy, 8, cy, junctionThick, "")
		} else if isRightish(ob) && !j.CrossMedian {
			b.LineStyled(ourX, cy, w-8, cy, junctionThick, "")
		}
	}

	appendSides(&b, j.Sides, ourX, cy, w, h)
	b.Rect("turn", ourX-2, cy-2, 4, 4, true)
	return b.Nodes()
}

func drawRoundabout(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := float64(w)/2, float64(h)/2-2
	r := math.Min(float64(w), float64(h)) * 0.28
	if r < 10 {
		r = 10
	}
	exits := j.Exits
	if exits < 2 {
		exits = 4
	}
	if exits > 6 {
		exits = 6
	}
	exit := j.Exit
	if exit < 1 {
		exit = 1
	}
	if exit > exits {
		exit = exits
	}

	// Entry from bottom.
	b.LineStyled(int(cx), h-6, int(cx), int(cy+r), junctionThick, "")
	b.Circle("ring", cx, cy, r, false, junctionStroke)

	// Exit spokes: entry is south (angle π/2 in screen Y-down… use math angles with Y up).
	// Angle 0 = east; we want entry at south (−π/2 from +X in Y-up → screen +Y is down so south is +π/2).
	driveLeft := j.Drive == JunctionDriveLeft
	for i := 1; i <= exits; i++ {
		// Number exits clockwise (right-drive) or counter-clockwise (left-drive) from entry.
		var ang float64
		step := 2 * math.Pi / float64(exits)
		if driveLeft {
			ang = math.Pi/2 - float64(i)*step // CCW from south
		} else {
			ang = math.Pi/2 + float64(i)*step // CW from south
		}
		// Screen: +Y down → flip sin.
		ex := cx + math.Cos(ang)*(r+14)
		ey := cy - math.Sin(ang)*(r+14)
		ix := cx + math.Cos(ang)*r
		iy := cy - math.Sin(ang)*r
		ours := i == exit
		if ours {
			b.LineStyled(int(ix), int(iy), int(ex), int(ey), junctionThick, "")
		} else {
			b.LineStyled(int(ix), int(iy), int(ex), int(ey), junctionStroke, junctionDash)
		}
	}

	b.Rect("turn", int(cx)-2, int(cy+r)-2, 4, 4, true)
	return b.Nodes()
}

// drawTJunction: approach into a T bar; outbound left/right solid, other arm dashed.
func drawTJunction(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	bot, left, right := h-6, 8, w-8
	ob := normalizeOutbound(j.Outbound)

	b.LineStyled(cx, bot, cx, cy, junctionThick, "")
	goLeft := isLeftish(ob) || (!isRightish(ob) && ob != "straight")
	if goLeft {
		b.LineStyled(cx, cy, left, cy, junctionThick, "")
		b.LineStyled(cx, cy, right, cy, junctionStroke, junctionDash)
	} else {
		b.LineStyled(cx, cy, right, cy, junctionThick, "")
		b.LineStyled(cx, cy, left, cy, junctionStroke, junctionDash)
	}
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawFork: Y split; our outbound solid, alternate arm dashed.
func drawFork(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2+4
	top, bot := 6, h-6
	ob := normalizeOutbound(j.Outbound)
	lx, ly := cx-(w/5), top+4
	rx, ry := cx+(w/5), top+4
	if !isSlight(ob) {
		lx, rx = cx-(w*2/5), cx+(w*2/5)
		ly, ry = cy-8, cy-8
	}

	b.LineStyled(cx, bot, cx, cy, junctionThick, "")
	leftOurs := isLeftish(ob)
	if !leftOurs && !isRightish(ob) {
		leftOurs = true // default slight_left bias when straight
	}
	if leftOurs {
		b.LineStyled(cx, cy, lx, ly, junctionThick, "")
		b.LineStyled(cx, cy, rx, ry, junctionStroke, junctionDash)
	} else {
		b.LineStyled(cx, cy, rx, ry, junctionThick, "")
		b.LineStyled(cx, cy, lx, ly, junctionStroke, junctionDash)
	}
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawMerge: side road joins the continuing spine (side from field or outbound).
func drawMerge(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	side := joinSide(j)

	b.LineStyled(cx, bot, cx, top, junctionThick, "")
	sx := 8
	if side == "right" {
		sx = w - 8
	}
	b.LineStyled(sx, bot-2, cx, cy, junctionThick, "")
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawRampExit: mainline continues (through); slip diverges on our outbound.
func drawRampExit(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	ob := normalizeOutbound(j.Outbound)
	side := joinSide(j)

	b.LineStyled(cx, bot, cx, top, junctionThick, "")
	sx := 8
	sy := top + 10
	if side == "right" || isRightish(ob) {
		sx = w - 8
	}
	if isSlight(ob) {
		sy = top + 4
		if side == "left" || isLeftish(ob) {
			sx = cx - (w / 5)
		} else {
			sx = cx + (w / 5)
		}
	}
	b.LineStyled(cx, cy, sx, sy, junctionThick, "")
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawRampEnter: slip joins into continuing mainline.
func drawRampEnter(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2
	top, bot := 6, h-6
	side := joinSide(j)

	b.LineStyled(cx, bot, cx, top, junctionThick, "")
	sx := 8
	if side == "right" {
		sx = w - 8
	}
	// Slip from mid-side into node (highway-ish, not full bottom approach).
	b.LineStyled(sx, cy+12, cx, cy, junctionThick, "")
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawUTurn: canned U loop; flips with drive side.
func drawUTurn(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx, cy := w/2, h/2+6
	bot := h - 6
	driveLeft := j.Drive == JunctionDriveLeft
	ux := cx + 14
	if driveLeft {
		ux = cx - 14
	}

	b.LineStyled(cx, bot, cx, cy, junctionThick, "")
	b.Path(fmt.Sprintf("M%d,%d V%d A8,8 0 0 1 %d,%d V%d", cx, cy, cy-14, ux, cy-14, cy+8), false, junctionThick)
	appendSides(&b, j.Sides, cx, cy, w, h)
	b.Rect("turn", cx-2, cy-2, 4, 4, true)
	return b.Nodes()
}

// drawArrive: short approach ending at a destination mark.
func drawArrive(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx := w / 2
	bot := h - 6
	endY := h/2 - 4

	b.LineStyled(cx, bot, cx, endY, junctionThick, "")
	b.Circle("dest", float64(cx), float64(endY), 5, false, junctionStroke)
	b.Rect("turn", cx-2, endY-2, 4, 4, true)
	_ = j
	return b.Nodes()
}

// drawDepart: start mark then spine ahead.
func drawDepart(j *JunctionMessage, w, h int) []scene.Node {
	var b scene.Builder
	cx := w / 2
	top, startY := 6, h/2+10

	b.Circle("origin", float64(cx), float64(startY), 5, false, junctionStroke)
	b.LineStyled(cx, startY, cx, top, junctionThick, "")
	b.Rect("turn", cx-2, startY-2, 4, 4, true)
	_ = j
	return b.Nodes()
}

// joinSide picks merge/ramp side from explicit field, else outbound, else drive default.
func joinSide(j *JunctionMessage) string {
	if j.Side == "left" || j.Side == "right" {
		return j.Side
	}
	ob := normalizeOutbound(j.Outbound)
	if isLeftish(ob) {
		return "left"
	}
	if isRightish(ob) {
		return "right"
	}
	// Right-drive: ramps usually on the right; left-drive: left.
	if j.Drive == JunctionDriveLeft {
		return "left"
	}
	return "right"
}

func appendSides(b *scene.Builder, sides []JunctionSide, cx, cy, w, h int) {
	for _, s := range sides {
		x := cx
		if s.Side == "left" {
			x = 8
		} else if s.Side == "right" {
			x = w - 8
		} else {
			continue
		}
		y := cy
		switch s.At {
		case "before":
			y = cy + (h / 5)
		case "after":
			y = cy - (h / 5)
		}
		dash := ""
		if s.Style != "solid" {
			dash = junctionDash
		}
		b.LineStyled(cx, y, x, y, junctionStroke, dash)
	}
}

func normalizeOutbound(o string) string {
	switch o {
	case "left", "right", "slight_left", "slight_right", "straight", "u_turn":
		return o
	default:
		return "straight"
	}
}

func isLeftish(o string) bool  { return o == "left" || o == "slight_left" }
func isRightish(o string) bool { return o == "right" || o == "slight_right" }
func isSlight(o string) bool   { return o == "slight_left" || o == "slight_right" }
