package compose

import (
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// DrawDeps supplies layout helpers and vector draw funcs for compose (ADR 0012).
type DrawDeps struct {
	ManeuverNodes func(protocol.Maneuver) []scene.Node
	RibbonNodes   func(nav protocol.NavMessage, w, h int) []scene.Node
	Fit           func(face scene.Face, s string, maxW int) string
	WrapRoad      func(s string, maxW, maxLines int) []string
	RoadBlockH    func(lineCount int) int
	HasMinimap    func(protocol.NavMessage) bool
	MinimapNodes  func(mm *protocol.MinimapMessage, w, h int) []scene.Node
	// Junction IR (parallel to minimap; gated by PreferJunctionTemplates in hud).
	HasJunction    func(protocol.NavMessage) bool
	JunctionNodes  func(nav protocol.NavMessage, w, h int) []scene.Node
	HasLanes      func(protocol.NavMessage) bool
	LaneStripNodes func(lanes []protocol.LaneInfo, maxW int) []scene.Node
	TextSVG       func(id string, faceSize string, x, baseline int, anchor, s string) string
}

// NavSVGDeps is deprecated; use DrawDeps.
type NavSVGDeps = DrawDeps
