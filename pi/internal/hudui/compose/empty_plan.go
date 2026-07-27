package compose

import "moto-hud/pi/internal/hudui/plan"

// EmptyPlan is a screen plan with no main-column body (chrome-only fallback).
func EmptyPlan() plan.ScreenPlan {
	return plan.ScreenPlan{}
}
