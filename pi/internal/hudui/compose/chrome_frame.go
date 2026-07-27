package compose

import (
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/token"
)

// frameBodyNodes is the full panel body slot for frame.svg (chrome + main column).
func frameBodyNodes(in Input, mainCol []scene.Node) []scene.Node {
	ch := chromeFor(in)
	pad, _, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY := chromeGeom()
	w, h := token.Width, token.Height

	var out []scene.Node
	out = append(out,
		scene.Line{X1: pad, Y1: divY, X2: ruleX, Y2: divY},
		scene.Line{X1: ruleX, Y1: pad, X2: ruleX, Y2: h - pad},
	)

	var main []scene.Node
	main = append(main, scene.Text{ID: "mode", Face: scene.Face6x12, X: 0, Baseline: headerBaseline, Anchor: "start", Value: ch.mode})
	main = append(main, scene.Group{ID: "link", DX: linkX, DY: linkY})
	main = append(main, mainCol...)
	out = append(out, scene.Group{ID: "main", DX: pad, Children: main})

	out = append(out,
		scene.Text{ID: "leg_prev", Face: scene.Face6x12, X: w - pad, Baseline: legTop, Anchor: "end", Value: ch.legPrev},
		scene.Text{ID: "leg_action", Face: scene.Face6x12, X: w - pad, Baseline: legMid, Anchor: "end", Value: ch.legAction},
		scene.Text{ID: "leg_next", Face: scene.Face6x12, X: w - pad, Baseline: legBot, Anchor: "end", Value: ch.legNext},
	)
	return out
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
