package hud

import (
	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

// ComposeInput builds compose.Input with hud-backed SVG helpers (single wiring point).
func ComposeInput(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked bool) compose.Input {
	return compose.Input{
		Screen: composeScreenKind(screen),
		Nav:    nav,
		Media:  media,
		Linked: linked,
		NavSVG: navSVGDeps(),
	}
}

func composeScreenKind(s Screen) compose.ScreenKind {
	switch s {
	case ScreenMedia:
		return compose.ScreenMedia
	case ScreenStatus:
		return compose.ScreenStatus
	default:
		return compose.ScreenNav
	}
}

func navSVGDeps() compose.NavSVGDeps {
	return compose.NavSVGDeps{
		ManeuverPaths: maneuverPaths,
		RibbonSVG: func(nav protocol.NavMessage, w, h int) string {
			pts, turnIdx := ribbonForNav(nav)
			return roadRibbonSVG(pts, turnIdx, w, h)
		},
		TextSVG: textSVGFromFaceSize,
		Fit:     fitFromFaceSize,
		WrapRoad: func(s string, maxW, maxLines int) []string {
			body := mustFace(pixelfont.Size8x16)
			return wrapLines(body, abbreviateRoad(s), maxW, maxLines)
		},
		RoadBlockH: func(lineCount int) int {
			body := mustFace(pixelfont.Size8x16)
			return roadBlockHeight(body, lineCount)
		},
		HasMinimap: HasMinimap,
		MinimapSVG: minimapSVG,
	}
}

func textSVGFromFaceSize(id, faceSize string, x, baseline int, anchor, s string) string {
	face := faceFromSize(faceSize)
	if face == nil {
		return ""
	}
	return textSVG(id, face, x, baseline, anchor, s)
}

func fitFromFaceSize(faceSize, s string, maxW int) string {
	face := faceFromSize(faceSize)
	if face == nil {
		return s
	}
	return fit(face, s, maxW)
}

func faceFromSize(size string) *pixelfont.Face {
	switch size {
	case "6x12":
		return mustFace(pixelfont.Size6x12)
	case "8x16":
		return mustFace(pixelfont.Size8x16)
	case "12x24":
		return mustFace(pixelfont.Size12x24)
	case "16x32":
		return mustFace(pixelfont.Size16x32)
	default:
		return mustFace(pixelfont.Size8x16)
	}
}
