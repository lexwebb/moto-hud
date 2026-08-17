package compose

import (
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
)

// frameBodyNodes is the full panel body (chrome + main column), without BLE link ink.
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

// FrameDocument is the full 250×122 panel scene (chrome + main + BLE link ink).
func FrameDocument(in Input, sp plan.ScreenPlan) scene.Document {
	nodes := frameBodyNodes(in, sp.Body)
	_, _, _, _, _, _, _, _, linkX, linkY := chromeGeom()
	var b scene.Builder
	b.Append(nodes...)
	b.Group("ble_link", linkX, linkY, func(gb *scene.Builder) {
		gb.Append(LinkMarkNodes(in.Linked)...)
	})
	return scene.Document{Width: token.Width, Height: token.Height, Nodes: b.Nodes()}
}

// FrameVars returns frame.svg template vars with a scene-serialized body fragment.
func FrameVars(in Input, sp plan.ScreenPlan) (map[string]string, error) {
	doc := FrameDocument(in, sp)
	return map[string]string{"body": svg.Fragment(doc.Nodes)}, nil
}

// FrameBodyForTest serializes chrome + main column for tests.
func FrameBodyForTest(in Input, sp plan.ScreenPlan) string {
	return svg.Fragment(frameBodyNodes(in, sp.Body))
}
