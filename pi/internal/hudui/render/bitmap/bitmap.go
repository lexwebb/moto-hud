// Package bitmap rasterizes scene.Document to a 1-bit grayscale image
// (black ink on white) without SVG or tdewolff/canvas.
package bitmap

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"unicode"

	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/pixelfont"
)

var (
	ink   = color.Gray{Y: 0}
	paper = color.Gray{Y: 255}
)

// Rasterize paints doc onto a new Gray image (white background, black ink).
// RawSVG nodes are unsupported — return an error so callers can fall back.
func Rasterize(doc scene.Document) (*image.Gray, error) {
	if doc.Width <= 0 || doc.Height <= 0 {
		return nil, fmt.Errorf("bitmap: bad size %dx%d", doc.Width, doc.Height)
	}
	if err := checkRawSVG(doc.Nodes); err != nil {
		return nil, err
	}
	img := image.NewGray(image.Rect(0, 0, doc.Width, doc.Height))
	fillRect(img, img.Bounds(), paper)
	var st state
	drawNodes(img, &st, doc.Nodes)
	return img, nil
}

func checkRawSVG(nodes []scene.Node) error {
	for _, n := range nodes {
		switch v := n.(type) {
		case scene.RawSVG:
			return fmt.Errorf("bitmap: RawSVG not supported")
		case scene.Group:
			if err := checkRawSVG(v.Children); err != nil {
				return err
			}
		}
	}
	return nil
}

type state struct {
	dx, dy int
}

func drawNodes(img *image.Gray, st *state, nodes []scene.Node) {
	for _, n := range nodes {
		drawNode(img, st, n)
	}
}

func drawNode(img *image.Gray, st *state, n scene.Node) {
	switch v := n.(type) {
	case scene.Text:
		drawText(img, st, v)
	case scene.Group:
		sub := state{dx: st.dx + v.DX, dy: st.dy + v.DY}
		drawNodes(img, &sub, v.Children)
	case scene.Line:
		drawLine(img, st.dx+v.X1, st.dy+v.Y1, st.dx+v.X2, st.dy+v.Y2, strokeW(v.StrokeWidth), v.Dash)
	case scene.Rect:
		r := image.Rect(st.dx+v.X, st.dy+v.Y, st.dx+v.X+v.W, st.dy+v.Y+v.H)
		if v.Filled {
			fillRect(img, r, ink)
		} else {
			strokeRect(img, r, 1)
		}
	case scene.Polyline:
		drawPolyline(img, st, v.Points, strokeW(v.StrokeWidth), false)
	case scene.Polygon:
		pts := floatPoints(st, v.Points)
		if v.Filled {
			fillPolygon(img, pts)
		} else {
			strokePolygon(img, pts, 1)
		}
	case scene.Path:
		drawPath(img, st, v)
	case scene.Circle:
		cx, cy := float64(st.dx)+v.CX, float64(st.dy)+v.CY
		if v.Filled {
			fillCircle(img, cx, cy, v.R)
		} else {
			strokeCircle(img, cx, cy, v.R, strokeW(v.StrokeWidth))
		}
	}
}

func strokeW(w float64) int {
	if w <= 0 {
		return 1
	}
	iw := int(math.Round(w))
	if iw < 1 {
		return 1
	}
	return iw
}

func drawText(img *image.Gray, st *state, t scene.Text) {
	face, err := loadFace(t.Face)
	if err != nil {
		return
	}
	x := st.dx + t.X
	baseline := st.dy + t.Baseline
	switch t.Anchor {
	case "end":
		x -= face.Measure(t.Value)
	case "middle":
		x -= face.Measure(t.Value) / 2
	}
	face.DrawString(img, x, baseline, t.Value)
}

func loadFace(f scene.Face) (*pixelfont.Face, error) {
	var sz pixelfont.Size
	switch f {
	case scene.Face6x12:
		sz = pixelfont.Size6x12
	case scene.Face12x24:
		sz = pixelfont.Size12x24
	case scene.Face16x32:
		sz = pixelfont.Size16x32
	default:
		sz = pixelfont.Size8x16
	}
	return pixelfont.Load(sz)
}

func fillRect(img *image.Gray, r image.Rectangle, c color.Gray) {
	r = r.Intersect(img.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		i := img.PixOffset(r.Min.X, y)
		for x := r.Min.X; x < r.Max.X; x++ {
			img.Pix[i] = c.Y
			i++
		}
	}
}

func strokeRect(img *image.Gray, r image.Rectangle, w int) {
	if r.Empty() || w < 1 {
		return
	}
	// Match SVG stroke centered on the rect edge (expands ~w/2 outside and inside).
	outer := image.Rect(r.Min.X-w, r.Min.Y-w, r.Max.X+w, r.Max.Y+w)
	fillRect(img, outer, ink)
	inner := image.Rect(r.Min.X+w, r.Min.Y+w, r.Max.X-w, r.Max.Y-w)
	if inner.Dx() > 0 && inner.Dy() > 0 {
		fillRect(img, inner, paper)
	}
}

func setInk(img *image.Gray, x, y int) {
	if !image.Pt(x, y).In(img.Bounds()) {
		return
	}
	img.SetGray(x, y, ink)
}

// drawLine draws a thick stroke with square caps (integer Bresenham + brush).
func drawLine(img *image.Gray, x0, y0, x1, y1, width int, dash string) {
	if width < 1 {
		width = 1
	}
	pattern, period := parseDash(dash)
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	x, y := x0, y0
	dist := 0
	for {
		if dashOn(pattern, period, dist) {
			stampBrush(img, x, y, width)
		}
		if x == x1 && y == y1 {
			break
		}
		e2 := 2 * err
		stepped := false
		if e2 >= dy {
			err += dy
			x += sx
			stepped = true
		}
		if e2 <= dx {
			err += dx
			y += sy
			stepped = true
		}
		if stepped {
			dist++
		}
	}
}

func parseDash(dash string) (on []bool, period int) {
	dash = strings.TrimSpace(dash)
	if dash == "" {
		return nil, 0
	}
	parts := strings.FieldsFunc(dash, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	var lens []int
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			continue
		}
		lens = append(lens, n)
	}
	if len(lens) == 0 {
		return nil, 0
	}
	if len(lens)%2 == 1 {
		lens = append(lens, lens[len(lens)-1])
	}
	total := 0
	for _, n := range lens {
		total += n
	}
	if total == 0 {
		return nil, 0
	}
	on = make([]bool, total)
	i := 0
	inkSeg := true
	for _, n := range lens {
		for j := 0; j < n; j++ {
			on[i] = inkSeg
			i++
		}
		inkSeg = !inkSeg
	}
	return on, total
}

func dashOn(pattern []bool, period, dist int) bool {
	if period == 0 || pattern == nil {
		return true
	}
	return pattern[dist%period]
}

func stampBrush(img *image.Gray, cx, cy, width int) {
	if width <= 1 {
		setInk(img, cx, cy)
		return
	}
	// Axis-aligned square brush of `width` pixels, roughly centered on (cx, cy).
	x0 := cx - width/2
	y0 := cy - width/2
	fillRect(img, image.Rect(x0, y0, x0+width, y0+width), ink)
}

func drawPolyline(img *image.Gray, st *state, pts [][2]int, width int, closed bool) {
	if len(pts) < 2 {
		return
	}
	for i := 0; i < len(pts)-1; i++ {
		a, b := pts[i], pts[i+1]
		drawLine(img, st.dx+a[0], st.dy+a[1], st.dx+b[0], st.dy+b[1], width, "")
	}
	if closed {
		a, b := pts[len(pts)-1], pts[0]
		drawLine(img, st.dx+a[0], st.dy+a[1], st.dx+b[0], st.dy+b[1], width, "")
	}
}

func floatPoints(st *state, pts [][2]float64) [][2]float64 {
	out := make([][2]float64, len(pts))
	for i, p := range pts {
		out[i] = [2]float64{p[0] + float64(st.dx), p[1] + float64(st.dy)}
	}
	return out
}

func strokePolygon(img *image.Gray, pts [][2]float64, width int) {
	if len(pts) < 2 {
		return
	}
	for i := 0; i < len(pts); i++ {
		a := pts[i]
		b := pts[(i+1)%len(pts)]
		drawLine(img, int(math.Round(a[0])), int(math.Round(a[1])), int(math.Round(b[0])), int(math.Round(b[1])), width, "")
	}
}

func fillPolygon(img *image.Gray, pts [][2]float64) {
	if len(pts) < 3 {
		return
	}
	minY, maxY := pts[0][1], pts[0][1]
	for _, p := range pts[1:] {
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	y0 := int(math.Floor(minY))
	y1 := int(math.Ceil(maxY))
	b := img.Bounds()
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}
	if y1 > b.Max.Y {
		y1 = b.Max.Y
	}
	n := len(pts)
	for y := y0; y < y1; y++ {
		var xs []float64
		yy := float64(y) + 0.5
		for i := 0; i < n; i++ {
			p0, p1 := pts[i], pts[(i+1)%n]
			yA, yB := p0[1], p1[1]
			if (yA <= yy && yB > yy) || (yB <= yy && yA > yy) {
				t := (yy - yA) / (yB - yA)
				xs = append(xs, p0[0]+t*(p1[0]-p0[0]))
			}
		}
		// insertion sort
		for i := 1; i < len(xs); i++ {
			v := xs[i]
			j := i
			for j > 0 && xs[j-1] > v {
				xs[j] = xs[j-1]
				j--
			}
			xs[j] = v
		}
		for i := 0; i+1 < len(xs); i += 2 {
			x0 := int(math.Ceil(xs[i]))
			x1 := int(math.Floor(xs[i+1]))
			for x := x0; x <= x1; x++ {
				setInk(img, x, y)
			}
		}
	}
}

func fillCircle(img *image.Gray, cx, cy, r float64) {
	if r <= 0 {
		return
	}
	rr := r * r
	x0 := int(math.Floor(cx - r))
	x1 := int(math.Ceil(cx + r))
	y0 := int(math.Floor(cy - r))
	y1 := int(math.Ceil(cy + r))
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= rr {
				setInk(img, x, y)
			}
		}
	}
}

func strokeCircle(img *image.Gray, cx, cy, r float64, width int) {
	if r <= 0 {
		return
	}
	if width < 1 {
		width = 1
	}
	outer := r + float64(width)/2
	inner := r - float64(width)/2
	if inner < 0 {
		inner = 0
	}
	oo, ii := outer*outer, inner*inner
	x0 := int(math.Floor(cx - outer))
	x1 := int(math.Ceil(cx + outer))
	y0 := int(math.Floor(cy - outer))
	y1 := int(math.Ceil(cy + outer))
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			d := dx*dx + dy*dy
			if d <= oo && d >= ii {
				setInk(img, x, y)
			}
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
