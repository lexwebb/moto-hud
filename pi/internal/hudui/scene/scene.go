// Package scene is the HUD display list (ADR 0012): integer coords, 1-bit black/white.
package scene

// Face selects pixelfont glyph metrics for Text nodes.
type Face string

const (
	Face6x12  Face = "6x12"
	Face8x16  Face = "8x16"
	Face12x24 Face = "12x24"
	Face16x32 Face = "16x32"
)

// Document is a slot-sized patch or sub-tree (width × height in canvas pixels).
type Document struct {
	Width, Height int
	Nodes         []Node
}

// Node is one draw primitive.
type Node interface {
	node()
}

// Text is a pixelfont-backed label (lowered to SVG <text data-pixel> by render/svg).
type Text struct {
	ID       string
	Face     Face
	X        int
	Baseline int
	Anchor   string // start, end, middle
	Value    string
}

func (Text) node() {}

// Group applies a translate then draws children.
type Group struct {
	ID       string
	DX, DY   int
	Children []Node
}

func (Group) node() {}

// RawSVG is a legacy escape hatch; prefer typed vector nodes.
type RawSVG struct {
	Markup string
}

func (RawSVG) node() {}

// Line is a stroke segment (chrome rules, glyphs).
type Line struct {
	X1, Y1, X2, Y2 int
	StrokeWidth    float64 // 0 → 1px in SVG backend
	Dash           string  // SVG stroke-dasharray, e.g. "4 5"
}

func (Line) node() {}

// Rect is an axis-aligned box (lane tiles, minimap pixels, ribbon turn mark).
type Rect struct {
	ID     string
	X, Y   int
	W, H   int
	Filled bool // false → fill="none" with stroke
}

func (Rect) node() {}

// Polyline is an open path through integer pixel centers.
type Polyline struct {
	Points      [][2]int
	Filled      bool
	StrokeWidth float64 // 0 → 1
}

func (Polyline) node() {}

// Polygon is a closed shape (maneuver arrow heads).
type Polygon struct {
	Points [][2]float64 // x,y pairs
	Filled bool
}

func (Polygon) node() {}

// Path is SVG path data (ribbon corridor, U-turn arc).
type Path struct {
	D           string
	Filled      bool
	StrokeWidth float64 // 0 → 1 when not filled
}

func (Path) node() {}

// Circle is a circle or dot mark.
type Circle struct {
	ID          string
	CX, CY, R   float64
	Filled      bool
	StrokeWidth float64
}

func (Circle) node() {}

// Builder collects nodes for a document.
type Builder struct {
	nodes []Node
}

func (b *Builder) Text(id string, face Face, x, baseline int, anchor, value string) {
	if value == "" {
		return
	}
	b.nodes = append(b.nodes, Text{ID: id, Face: face, X: x, Baseline: baseline, Anchor: anchor, Value: value})
}

func (b *Builder) Group(id string, dx, dy int, fn func(*Builder)) {
	if fn == nil {
		return
	}
	var sub Builder
	fn(&sub)
	b.nodes = append(b.nodes, Group{ID: id, DX: dx, DY: dy, Children: sub.nodes})
}

func (b *Builder) Raw(markup string) {
	if markup == "" {
		return
	}
	b.nodes = append(b.nodes, RawSVG{Markup: markup})
}

func (b *Builder) Line(x1, y1, x2, y2 int) {
	b.nodes = append(b.nodes, Line{X1: x1, Y1: y1, X2: x2, Y2: y2})
}

func (b *Builder) LineStyled(x1, y1, x2, y2 int, strokeWidth float64, dash string) {
	b.nodes = append(b.nodes, Line{X1: x1, Y1: y1, X2: x2, Y2: y2, StrokeWidth: strokeWidth, Dash: dash})
}

func (b *Builder) Rect(id string, x, y, w, h int, filled bool) {
	b.nodes = append(b.nodes, Rect{ID: id, X: x, Y: y, W: w, H: h, Filled: filled})
}

func (b *Builder) Polyline(points [][2]int, filled bool, strokeWidth float64) {
	if len(points) == 0 {
		return
	}
	b.nodes = append(b.nodes, Polyline{Points: points, Filled: filled, StrokeWidth: strokeWidth})
}

func (b *Builder) Polygon(points [][2]float64, filled bool) {
	if len(points) < 3 {
		return
	}
	b.nodes = append(b.nodes, Polygon{Points: points, Filled: filled})
}

func (b *Builder) Path(d string, filled bool, strokeWidth float64) {
	if d == "" {
		return
	}
	b.nodes = append(b.nodes, Path{D: d, Filled: filled, StrokeWidth: strokeWidth})
}

func (b *Builder) Circle(id string, cx, cy, r float64, filled bool, strokeWidth float64) {
	b.nodes = append(b.nodes, Circle{ID: id, CX: cx, CY: cy, R: r, Filled: filled, StrokeWidth: strokeWidth})
}

// Append copies nodes from another builder.
func (b *Builder) Append(nodes ...Node) {
	b.nodes = append(b.nodes, nodes...)
}

func (b *Builder) Nodes() []Node {
	return b.nodes
}

// Patch builds a white-backed slot document.
func Patch(w, h int, fn func(*Builder)) Document {
	var b Builder
	if fn != nil {
		fn(&b)
	}
	return Document{Width: w, Height: h, Nodes: b.Nodes()}
}
