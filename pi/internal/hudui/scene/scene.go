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

// RawSVG is a temporary escape for not-yet-ported vector fragments (ribbon, link glyph, …).
type RawSVG struct {
	Markup string
}

func (RawSVG) node() {}

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
