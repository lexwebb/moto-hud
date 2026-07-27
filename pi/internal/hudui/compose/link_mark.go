package compose

import (
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
)

const linkStroke = 1.6

// LinkMarkNodes is the BLE link glyph (16×12 logical units, drawn at chrome link slot).
func LinkMarkNodes(connected bool) []scene.Node {
	var b scene.Builder
	if connected {
		b.Path("M2,9 L6,2 L6,6 L10,6 L6,10 L6,6", false, linkStroke)
		b.Rect("", 12, 4, 4, 4, true)
	} else {
		b.LineStyled(2, 2, 10, 10, linkStroke, "")
		b.LineStyled(10, 2, 2, 10, linkStroke, "")
	}
	return b.Nodes()
}

// LinkMarkFragment serializes LinkMarkNodes for legacy SVG string call sites.
func LinkMarkFragment(connected bool) string {
	return svg.Fragment(LinkMarkNodes(connected))
}
