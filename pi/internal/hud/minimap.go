package hud

import (
	"fmt"
	"image"
	"math"
	"strings"

	"moto-hud/pi/internal/protocol"
)

// View radius: zoom to content so the junction fills the pane, but never
// pull out past ~50 m or in closer than ~25 m.
const (
	minimapRadiusMax = 50.0
	minimapRadiusMin = 25.0
)

// HasMinimap reports whether nav carries a minimap route snapshot.
func HasMinimap(nav protocol.NavMessage) bool {
	return hasMinimap(nav)
}

func hasMinimap(nav protocol.NavMessage) bool {
	return nav.Minimap != nil && len(nav.Minimap.Route) >= 2
}

func minimapViewRadius(mm *protocol.MinimapMessage) float64 {
	maxR := 0.0
	note := func(x, y float64) {
		r := math.Hypot(x, y)
		if r > maxR {
			maxR = r
		}
	}
	for _, p := range mm.Route {
		note(p.X, p.Y)
	}
	if mm.Rider != nil {
		note(mm.Rider.X, mm.Rider.Y)
	}
	for _, way := range mm.Context {
		for _, p := range way {
			if math.Hypot(p.X, p.Y) <= minimapRadiusMax {
				note(p.X, p.Y)
			}
		}
	}
	maxR += 6
	if maxR < minimapRadiusMin {
		maxR = minimapRadiusMin
	}
	if maxR > minimapRadiusMax {
		maxR = minimapRadiusMax
	}
	return maxR
}

// minimapSVG draws a tube-map junction snapshot: meter octilinear (H/V/45°),
// then Bresenham 1×1 rects so 1-bit AA can't fatten dashes. Route is thicker.
func minimapSVG(mm *protocol.MinimapMessage, w, h int) string {
	return minimapSVGLayers(mm, w, h, true, true, true)
}

// minimapSVGLayers draws selected layers (context dashes, solid route, turn/rider marks).
func minimapSVGLayers(mm *protocol.MinimapMessage, w, h int, context, route, marks bool) string {
	if w <= 0 {
		w = 70
	}
	if h <= 0 {
		h = 80
	}
	if mm == nil || len(mm.Route) < 2 {
		cx := w / 2
		return fmt.Sprintf(
			`<line x1="%d" y1="4" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`,
			cx, cx, h-4,
		)
	}

	const pad = 3.0
	cx := float64(w) / 2
	cy := float64(h) / 2
	radius := minimapViewRadius(mm)
	scale := math.Min((float64(w)-pad*2)/(2*radius), (float64(h)-pad*2)/(2*radius))

	toScreen := func(x, y float64) (int, int, bool) {
		if math.Hypot(x, y) > radius*1.2 {
			return 0, 0, false
		}
		sx := int(math.Round(cx + x*scale))
		sy := int(math.Round(cy - y*scale))
		if sx < -1 || sx > w || sy < -1 || sy > h {
			return sx, sy, false
		}
		return sx, sy, true
	}

	var b strings.Builder

	if context {
		b.WriteString(`<g id="context">`)
		for _, way := range mm.Context {
			pix := tubePixels(schematizeTube(way, 8), toScreen)
			b.WriteString(dashedPixelRects(pix, 3, 3))
		}
		b.WriteString(`</g>`)
	}

	if route {
		r := ensurePoint(schematizeTube(mm.Route, 5), protocol.RibbonPoint{X: 0, Y: 0})
		routePix := tubePixels(r, toScreen)
		b.WriteString(solidPixelStroke(routePix, 3))
	}

	if marks {
		fmt.Fprintf(&b, `<rect id="turn" x="%d" y="%d" width="4" height="4" fill="#000"/>`,
			int(cx)-2, int(cy)-2)
		if mm.Rider != nil {
			if rx, ry, ok := toScreen(mm.Rider.X, mm.Rider.Y); ok {
				fmt.Fprintf(&b, `<rect id="rider" x="%d" y="%d" width="5" height="5" fill="#000"/>`, rx-2, ry-2)
			}
		}
	}
	return b.String()
}

// RenderMinimap rasterizes the junction pane alone (for lab / golden tests).
func RenderMinimap(mm *protocol.MinimapMessage, w, h int) (*image.Gray, error) {
	return RenderMinimapLayers(mm, w, h, true, true, true)
}

// RenderMinimapLayers rasterizes selected layers onto a white pane.
func RenderMinimapLayers(mm *protocol.MinimapMessage, w, h int, context, route, marks bool) (*image.Gray, error) {
	if w <= 0 {
		w = 70
	}
	if h <= 0 {
		h = 80
	}
	frag := minimapSVGLayers(mm, w, h, context, route, marks)
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="%d" height="%d" fill="#fff"/>%s</svg>`,
		w, h, w, h, w, h, frag,
	)
	return RasterizeSVGAt([]byte(svg), w, h)
}

// MinimapSVGFragment returns the raw SVG shapes for debugging (no outer <svg>).
func MinimapSVGFragment(mm *protocol.MinimapMessage, w, h int) string {
	return minimapSVG(mm, w, h)
}

// MinimapViewRadiusMeters exposes adaptive radius for lab overlays.
func MinimapViewRadiusMeters(mm *protocol.MinimapMessage) float64 {
	return minimapViewRadius(mm)
}

type screenProj func(x, y float64) (sx, sy int, ok bool)

// tubePixels projects, then inserts 45° knees in pixel space.
func tubePixels(pts []protocol.RibbonPoint, project screenProj) [][2]int {
	raw := make([][2]int, 0, len(pts))
	for _, p := range pts {
		sx, sy, ok := project(p.X, p.Y)
		if !ok {
			continue
		}
		if len(raw) > 0 && raw[len(raw)-1][0] == sx && raw[len(raw)-1][1] == sy {
			continue
		}
		raw = append(raw, [2]int{sx, sy})
	}
	if len(raw) < 2 {
		return raw
	}
	out := make([][2]int, 0, len(raw)*2)
	out = append(out, raw[0])
	for i := 1; i < len(raw); i++ {
		a := out[len(out)-1]
		b := raw[i]
		for _, p := range tubeKnee(a, b) {
			if out[len(out)-1] == p {
				continue
			}
			out = append(out, p)
		}
	}
	return out
}

// tubeKnee: H, V, or 45° diagonal as far as possible, then axis finish.
func tubeKnee(a, b [2]int) [][2]int {
	dx, dy := b[0]-a[0], b[1]-a[1]
	adx, ady := absInt(dx), absInt(dy)
	if adx == 0 || ady == 0 || adx == ady {
		return [][2]int{b}
	}
	sx, sy := signInt(dx), signInt(dy)
	if adx > ady {
		return [][2]int{{a[0] + sx*ady, a[1] + sy*ady}, b}
	}
	return [][2]int{{a[0] + sx*adx, a[1] + sy*adx}, b}
}

func dashedPixelRects(pts [][2]int, dash, gap int) string {
	if len(pts) < 2 {
		return ""
	}
	if dash < 1 {
		dash = 1
	}
	var b strings.Builder
	on := true
	left := dash
	for i := 0; i < len(pts)-1; i++ {
		line := bresenham(pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1])
		// Skip first pixel of each segment after the first (shared vertex).
		start := 0
		if i > 0 && len(line) > 0 {
			start = 1
		}
		for _, p := range line[start:] {
			if on {
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="1" height="1" fill="#000"/>`, p[0], p[1])
			}
			left--
			if left <= 0 {
				on = !on
				if on {
					left = dash
				} else {
					left = gap
				}
			}
		}
	}
	return b.String()
}

func solidPixelStroke(pts [][2]int, thickness int) string {
	if len(pts) < 2 {
		return ""
	}
	if thickness < 1 {
		thickness = 1
	}
	seen := map[[2]int]bool{}
	var b strings.Builder
	b.WriteString(`<g id="route">`)
	paint := func(x, y int) {
		p := [2]int{x, y}
		if seen[p] {
			return
		}
		seen[p] = true
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="1" height="1" fill="#000"/>`, x, y)
	}
	r := thickness / 2
	for i := 0; i < len(pts)-1; i++ {
		line := bresenham(pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1])
		start := 0
		if i > 0 && len(line) > 0 {
			start = 1
		}
		for _, p := range line[start:] {
			for ox := -r; ox <= r; ox++ {
				for oy := -r; oy <= r; oy++ {
					// Diamond-ish brush so diagonals don't get a square blob.
					if absInt(ox)+absInt(oy) <= r {
						paint(p[0]+ox, p[1]+oy)
					}
				}
			}
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}

func bresenham(x0, y0, x1, y1 int) [][2]int {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx, sy := signInt(x1-x0), signInt(y1-y0)
	err := dx + dy
	out := make([][2]int, 0, dx+absInt(dy)+1)
	for {
		out = append(out, [2]int{x0, y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
	return out
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func signInt(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

// schematizeTube: RDP + snap each leg to H/V/45° in meter space.
func schematizeTube(pts []protocol.RibbonPoint, eps float64) []protocol.RibbonPoint {
	if len(pts) < 2 {
		return pts
	}
	simp := simplifyRDP(pts, eps)
	if len(simp) < 2 {
		return simp
	}
	out := make([]protocol.RibbonPoint, 0, len(simp))
	out = append(out, simp[0])
	for i := 1; i < len(simp); i++ {
		prev := out[len(out)-1]
		snapped := snapOctilinear(prev, simp[i])
		if hypot2(snapped.X-prev.X, snapped.Y-prev.Y) < 0.5 {
			continue
		}
		out = append(out, snapped)
	}
	if len(out) < 2 {
		return simp
	}
	return out
}

func snapOctilinear(a, b protocol.RibbonPoint) protocol.RibbonPoint {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dist := math.Hypot(dx, dy)
	if dist < 1e-6 {
		return a
	}
	ang := math.Atan2(dy, dx)
	const step = math.Pi / 4
	snapped := math.Round(ang/step) * step
	return protocol.RibbonPoint{
		X: a.X + dist*math.Cos(snapped),
		Y: a.Y + dist*math.Sin(snapped),
	}
}

func simplifyRDP(pts []protocol.RibbonPoint, eps float64) []protocol.RibbonPoint {
	if len(pts) < 3 || eps <= 0 {
		return pts
	}
	keep := make([]bool, len(pts))
	keep[0] = true
	keep[len(pts)-1] = true
	rdpMark(pts, 0, len(pts)-1, eps*eps, keep)
	out := make([]protocol.RibbonPoint, 0, len(pts))
	for i, k := range keep {
		if k {
			out = append(out, pts[i])
		}
	}
	return out
}

func rdpMark(pts []protocol.RibbonPoint, i, j int, eps2 float64, keep []bool) {
	if j <= i+1 {
		return
	}
	maxD := -1.0
	maxI := i
	ax, ay := pts[i].X, pts[i].Y
	bx, by := pts[j].X, pts[j].Y
	for k := i + 1; k < j; k++ {
		d := pointSegDist2(pts[k].X, pts[k].Y, ax, ay, bx, by)
		if d > maxD {
			maxD = d
			maxI = k
		}
	}
	if maxD > eps2 {
		keep[maxI] = true
		rdpMark(pts, i, maxI, eps2, keep)
		rdpMark(pts, maxI, j, eps2, keep)
	}
}

func pointSegDist2(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	if dx == 0 && dy == 0 {
		return hypot2(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return hypot2(px-(ax+t*dx), py-(ay+t*dy))
}

func hypot2(dx, dy float64) float64 { return dx*dx + dy*dy }

// ensurePoint inserts p into the polyline at the nearest vertex slot if missing.
// Prefer middle insertion when p is the turn origin between approach and departure.
func ensurePoint(pts []protocol.RibbonPoint, p protocol.RibbonPoint) []protocol.RibbonPoint {
	for _, q := range pts {
		if math.Abs(q.X-p.X) < 0.5 && math.Abs(q.Y-p.Y) < 0.5 {
			return pts
		}
	}
	// Prefer a slot between an approach (y<0) and a departure (y>0 or |x| large).
	for i := 0; i < len(pts)-1; i++ {
		a, b := pts[i], pts[i+1]
		if a.Y <= 0 && (b.Y > 0 || math.Abs(b.X) > math.Abs(a.X)+2) {
			out := make([]protocol.RibbonPoint, 0, len(pts)+1)
			out = append(out, pts[:i+1]...)
			out = append(out, p)
			out = append(out, pts[i+1:]...)
			return out
		}
	}
	best, bestD := 0, math.MaxFloat64
	for i, q := range pts {
		d := hypot2(q.X-p.X, q.Y-p.Y)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	out := make([]protocol.RibbonPoint, 0, len(pts)+1)
	out = append(out, pts[:best+1]...)
	out = append(out, p)
	out = append(out, pts[best+1:]...)
	return out
}
