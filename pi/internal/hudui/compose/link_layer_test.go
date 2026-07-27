package compose_test

import (
	"testing"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/compose"
)

func TestEveryPlanIncludesLinkLayer(t *testing.T) {
	cases := []compose.Input{
		{Screen: compose.ScreenNav, NavSVG: stubDrawDeps()},
		{Screen: compose.ScreenMedia, NavSVG: stubDrawDeps()},
		{Screen: compose.ScreenStatus, NavSVG: stubDrawDeps()},
	}
	for _, in := range cases {
		sp, err := compose.BuildPlan(in)
		if err != nil {
			t.Fatalf("%v: %v", in.Screen, err)
		}
		l, ok := sp.LayerByID(hudui.NodeLink)
		if !ok {
			t.Fatalf("screen %v missing link layer", in.Screen)
		}
		if l.Slot.Empty() || l.Patch == nil {
			t.Fatalf("link layer not patchable: %+v", l)
		}
	}
}
