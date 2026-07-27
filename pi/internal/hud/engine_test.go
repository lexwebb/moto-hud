package hud_test

import (
	"testing"

	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
)

func TestEngineDistancePatch(t *testing.T) {
	e := hud.NewEngine()
	nav := protocol.NavMessage{
		Active:       true,
		DistanceM:    200,
		DistanceText: "200 m",
		Road:         "High St",
		Maneuver:     protocol.ManeuverLeft,
	}
	r1 := e.Draw(hud.ScreenNav, nav, protocol.MediaMessage{}, true, true)
	if r1.Image == nil {
		t.Fatal("nil image")
	}

	nav.DistanceM = 140
	nav.DistanceText = "140 m"
	r2 := e.Draw(hud.ScreenNav, nav, protocol.MediaMessage{}, true, false)
	if !r2.Patched {
		t.Fatalf("expected patched frame, spatial=%v dirty=%v", r2.Spatial, r2.Dirty)
	}
}
