package compose

import (
	"moto-hud/pi/internal/hudui/plan"
)

func planNav(in Input) (plan.ScreenPlan, error) {
	if !in.Nav.Active {
		return planNavIdle(in)
	}
	if in.NavSVG.HasMinimap != nil && (in.NavSVG.HasMinimap(in.Nav) || len(in.Nav.RibbonPoints) >= 2) {
		return planNavLive(in)
	}
	return planNavClassic(in)
}
