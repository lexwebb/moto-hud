package compose

import (
	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
)

func planStatus(in Input) (plan.ScreenPlan, error) {
	body, err := StatusBodySVG(in.Linked, in.Nav.Active)
	if err != nil {
		return plan.ScreenPlan{}, err
	}
	k := Keys{}
	return plan.ScreenPlan{
		BodySVG:     body,
		Descriptors: plan.BuildDescriptors(statusScreenKey(in.Linked, in.Nav.Active), k.Bool(in.Linked), nil),
	}, nil
}

func statusScreenKey(linked, navActive bool) hudui.ChangeKey {
	var key hudui.ChangeKey
	if linked {
		key |= 1
	}
	if navActive {
		key |= 2
	}
	return key
}
