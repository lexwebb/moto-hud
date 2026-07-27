package hud

import (
	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

// ComposeInput builds compose.Input with hud-backed draw helpers (single wiring point).
func ComposeInput(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked bool) compose.Input {
	return compose.Input{
		Screen:  composeScreenKind(screen),
		Nav:     nav,
		Media:   media,
		Linked:  linked,
		NavSVG:  drawDeps(),
		LinkNodes: compose.LinkMarkNodes,
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

func drawDeps() compose.DrawDeps {
	return compose.DrawDeps{
		ManeuverNodes: ManeuverNodes,
		RibbonNodes: func(nav protocol.NavMessage, w, h int) []scene.Node {
			return RibbonNodesForNav(nav, w, h)
		},
		TextSVG: textSVGFromFaceSize,
		Fit:     fitFromSceneFace,
		WrapRoad: func(s string, maxW, maxLines int) []string {
			body := mustFace(pixelfont.Size8x16)
			return wrapLines(body, abbreviateRoad(s), maxW, maxLines)
		},
		RoadBlockH: func(lineCount int) int {
			body := mustFace(pixelfont.Size8x16)
			return roadBlockHeight(body, lineCount)
		},
		HasMinimap: HasMinimap,
		MinimapNodes: MinimapNodes,
		HasLanes:   hasLanes,
		LaneStripNodes: LaneStripNodes,
	}
}

func textSVGFromFaceSize(id, faceSize string, x, baseline int, anchor, s string) string {
	face := faceFromSize(faceSize)
	if face == nil {
		return ""
	}
	return textSVG(id, face, x, baseline, anchor, s)
}

func fitFromSceneFace(f scene.Face, s string, maxW int) string {
	face := faceFromSceneFace(f)
	if face == nil {
		return s
	}
	return fit(face, s, maxW)
}

func faceFromSceneFace(f scene.Face) *pixelfont.Face {
	switch f {
	case scene.Face6x12:
		return mustFace(pixelfont.Size6x12)
	case scene.Face8x16:
		return mustFace(pixelfont.Size8x16)
	case scene.Face12x24:
		return mustFace(pixelfont.Size12x24)
	case scene.Face16x32:
		return mustFace(pixelfont.Size16x32)
	default:
		return mustFace(pixelfont.Size8x16)
	}
}

func faceFromSize(size string) *pixelfont.Face {
	return faceFromSceneFace(sceneFaceFromSizeAttr(size))
}

func sceneFaceFromSizeAttr(size string) scene.Face {
	switch size {
	case "6x12":
		return scene.Face6x12
	case "12x24":
		return scene.Face12x24
	case "16x32":
		return scene.Face16x32
	default:
		return scene.Face8x16
	}
}
