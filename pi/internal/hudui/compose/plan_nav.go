package compose

import (
	"moto-hud/pi/internal/hudui/plan"
)

func planNav(in Input) (plan.ScreenPlan, error) {
	if !in.Nav.Active {
		return planNavIdle(in)
	}
	// Junction IR (flagged) or meter minimap / ribbon → live left column.
	if in.NavSVG.HasJunction != nil && in.NavSVG.HasJunction(in.Nav) {
		return planNavLive(in)
	}
	if in.NavSVG.HasMinimap != nil && (in.NavSVG.HasMinimap(in.Nav) || len(in.Nav.RibbonPoints) >= 2) {
		return planNavLive(in)
	}
	return planNavClassic(in)
}
