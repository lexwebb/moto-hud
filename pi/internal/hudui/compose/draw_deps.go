package compose

import (
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

// DrawDeps supplies layout helpers for compose (ADR 0012). Vector fragments not yet on scene use RawSVG call sites.
type DrawDeps struct {
	ManeuverPaths func(protocol.Maneuver) string
	RibbonSVG     func(nav protocol.NavMessage, w, h int) string
	Fit           func(face scene.Face, s string, maxW int) string
	WrapRoad      func(s string, maxW, maxLines int) []string
	RoadBlockH    func(lineCount int) int
	HasMinimap    func(protocol.NavMessage) bool
	MinimapSVG    func(mm *protocol.MinimapMessage, w, h int) string
	HasLanes      func(protocol.NavMessage) bool
	LaneStripSVG  func(lanes []protocol.LaneInfo, maxW int) string
	// TextSVG emits SVG for full-frame / templ bodies until BodySVG migrates to scene.
	TextSVG func(id string, faceSize string, x, baseline int, anchor, s string) string
}

// NavSVGDeps is deprecated; use DrawDeps.
type NavSVGDeps = DrawDeps
