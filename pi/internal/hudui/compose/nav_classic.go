package compose

import (
	"fmt"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/scene"
)

// planNavClassic builds the standard nav screen (glyph + distance + road + ribbon).
func planNavClassic(in Input) (plan.ScreenPlan, error) {
	deps := in.NavSVG
	if deps.TextSVG == nil {
		return plan.ScreenPlan{}, fmt.Errorf("compose: DrawDeps required")
	}
	nav := in.Nav
	k := Keys{}
	lay := layoutNavClassic(nav, deps)
	body, err := NavClassicBodySVG(lay, deps)
	if err != nil {
		return plan.ScreenPlan{}, err
	}

	navCopy := nav
	layers := []plan.Layer{
		{ID: hudui.NodeManeuver, Tier: hudui.TierSlow, Key: k.Maneuver(nav), Slot: lay.maneuverSlot},
		{
			ID: hudui.NodeDistance, Tier: hudui.TierPartialOK, Key: k.DistanceBucket(nav.DistanceM), Slot: lay.distanceSlot,
			Patch: func() (scene.Document, error) {
				return patchDistanceDoc(navCopy, lay.distanceSlot.Dx(), lay.distanceSlot.Dy(), deps)
			},
		},
		{
			ID: hudui.NodeRoad, Tier: hudui.TierPartialOK, Key: k.Road(nav), Slot: lay.roadSlot,
			Patch: func() (scene.Document, error) {
				return patchRoadDoc(navCopy, lay.roadSlot, deps)
			},
		},
	}
	if lay.etaH > 0 {
		layers = append(layers, plan.Layer{
			ID: hudui.NodeETA, Tier: hudui.TierPartialOK, Key: k.ETA(nav), Slot: lay.etaSlot,
			Patch: func() (scene.Document, error) {
				return patchETADoc(navCopy, lay.etaSlot.Dx(), lay.etaSlot.Dy(), deps)
			},
		})
	}
	layers = append(layers, plan.Layer{ID: hudui.NodeRibbon, Tier: hudui.TierSlow, Key: k.Ribbon(nav), Slot: lay.ribbonSlot})

	return finalizePlan(in, k.NavScreen(nav), staticChromeKey(), body, layers), nil
}

func formatDistance(m int) string {
	if m >= 1000 {
		return itoa(m/1000) + "." + itoa((m%1000)/100) + " km"
	}
	return itoa(m) + " m"
}

func formatETA(min int) string {
	if min >= 60 {
		return "ETA " + itoa(min/60) + "h " + itoa(min%60) + "m"
	}
	return "ETA " + itoa(min) + " min"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
