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
	return plan.ScreenPlan{
		BodySVG:     body,
		Descriptors: plan.BuildDescriptors(k.NavScreen(in.Nav, in.Linked), k.Bool(in.Linked), nil),
	}, nil
}