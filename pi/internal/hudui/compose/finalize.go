package compose

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/scene"
)

const linkSlotW = 16
const linkSlotH = 12

func linkSlot() image.Rectangle {
	pad, _, _, _, _, _, _, _, linkX, linkY := chromeGeom()
	x0 := pad + linkX
	y0 := linkY
	return image.Rect(x0, y0, x0+linkSlotW, y0+linkSlotH)
}

func linkLayer(in Input) plan.Layer {
	slot := linkSlot()
	linked := in.Linked
	linkFn := in.LinkSVG
	if linkFn == nil {
		linkFn = LinkMarkFragment
	}
	return plan.Layer{
		ID:   hudui.NodeLink,
		Tier: hudui.TierPartialOK,
		Key:  Keys{}.Bool(linked),
		Slot: slot,
		Patch: func() (scene.Document, error) {
			return patchLinkDoc(linked, slot.Dx(), slot.Dy(), linkFn)
		},
	}
}

// finalizePlan attaches refresh metadata and the shared BLE link layer.
func finalizePlan(in Input, screenKey, chromeKey hudui.ChangeKey, body []scene.Node, layers []plan.Layer) plan.ScreenPlan {
	if layers == nil {
		layers = []plan.Layer{}
	}
	layers = append(layers, linkLayer(in))
	return plan.ScreenPlan{
		Body:        body,
		Layers:      layers,
		Descriptors: plan.BuildDescriptors(screenKey, chromeKey, layers),
	}
}

func staticChromeKey() hudui.ChangeKey {
	return Keys{}.Hash("hud_chrome")
}
