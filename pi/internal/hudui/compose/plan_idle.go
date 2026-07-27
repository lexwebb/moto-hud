package compose

import (
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

func planNavIdle(in Input) (plan.ScreenPlan, error) {
	k := Keys{}
	mw, b1, b2 := navIdleBaselines()
	body := scenetempl.Render(screens.NavIdleCentered("MOTO HUD", "Waiting for route...", mw, b1, b2))
	return finalizePlan(in, k.NavScreen(in.Nav), staticChromeKey(), body, nil), nil
}

func navIdleBaselines() (mw, baseline1, baseline2 int) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	mw = token.MainWidth()
	pad := token.Pad
	headerBottom := pad + meta.Metrics.CellH
	divY := headerBottom + token.GapSm
	contentTop := divY + token.GapMd
	contentBot := token.Height - pad
	blockH := body.Metrics.CellH*2 + token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	return mw, top + body.Metrics.Ascent, top + body.Metrics.CellH + token.GapMd + body.Metrics.Ascent
}
