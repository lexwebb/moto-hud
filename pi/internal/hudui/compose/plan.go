package compose

import (
	"hash/maphash"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
)

var hashSeed = maphash.MakeSeed()

func hashStr(s string) hudui.ChangeKey {
	var h maphash.Hash
	h.SetSeed(hashSeed)
	_, _ = h.WriteString(s)
	return hudui.ChangeKey(h.Sum64())
}

// BuildPlan composes layout, refresh layers, and main-column SVG for the active screen.
func BuildPlan(in Input) (plan.ScreenPlan, error) {
	switch in.Screen {
	case ScreenNav:
		return planNav(in)
	case ScreenMedia:
		return planMedia(in)
	case ScreenStatus:
		return planStatus(in)
	default:
		return planNav(in)
	}
}
