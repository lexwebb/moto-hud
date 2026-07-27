package compose

import (
	"moto-hud/pi/internal/hudui/plan"
)

func planNavIdle(in Input) (plan.ScreenPlan, error) {
	k := Keys{}
	body := navIdleBodyScene()
	return finalizePlan(in, k.NavScreen(in.Nav), staticChromeKey(), body, nil), nil
}
