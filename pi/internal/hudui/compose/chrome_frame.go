package compose

import (
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
)

// frameBodyNodes is the full panel body slot for frame.svg (chrome + main column).
func frameBodyNodes(in Input, mainCol []scene.Node) []scene.Node {
	ch := chromeFor(in)
	pad, _, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY := chromeGeom()
	w, h := token.Width, token.Height
	return scenetempl.Render(screens.ChromeFrameBody(
		ch.mode, ch.legPrev, ch.legAction, ch.legNext,
		pad, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY, w, h,
		scenetempl.Nodes(mainCol),
	))
}

// FrameVars returns frame.svg template vars with a scene-serialized body fragment.
func FrameVars(in Input, sp plan.ScreenPlan) (map[string]string, error) {
	frag := svg.Fragment(frameBodyNodes(in, sp.Body))
	return map[string]string{"body": frag}, nil
}

// FrameBodyForTest serializes chrome + main column for tests.
func FrameBodyForTest(in Input, sp plan.ScreenPlan) string {
	return svg.Fragment(frameBodyNodes(in, sp.Body))
}
