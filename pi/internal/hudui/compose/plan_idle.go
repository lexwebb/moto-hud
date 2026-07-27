package compose

import (
	"moto-hud/pi/internal/hudui/plan"
)

func planNavIdle(in Input) (plan.ScreenPlan, error) {
	body, err := NavIdleBodySVG()
	if err != nil {
		return plan.ScreenPlan{}, err
	}
	k := Keys{}
	return finalizePlan(in, k.NavScreen(in.Nav), staticChromeKey(), body, nil), nil
}