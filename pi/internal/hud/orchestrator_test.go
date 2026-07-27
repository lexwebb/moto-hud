package hud_test

import (
	"testing"

	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/protocol"
)

func TestRefreshOrchestratorDistanceBucketPartial(t *testing.T) {
	o := hud.NewRefreshOrchestrator()
	nav := protocol.NavMessage{
		Active:       true,
		DistanceM:    120,
		DistanceText: "120 m",
		Road:         "High St",
		Maneuver:     protocol.ManeuverLeft,
	}
	r1 := o.Plan(hud.ScreenNav, nav, protocol.MediaMessage{}, true, true)
	if r1.Mode != hud.RefreshFullFrame {
		t.Fatalf("first frame mode %v", r1.Mode)
	}

	nav.DistanceM = 115
	nav.DistanceText = "115 m"
	r2 := o.Plan(hud.ScreenNav, nav, protocol.MediaMessage{}, true, false)
	if r2.Mode != hud.RefreshNone {
		t.Fatalf("same bucket want none got %v", r2.Mode)
	}

	nav.DistanceM = 70
	nav.DistanceText = "70 m"
	r3 := o.Plan(hud.ScreenNav, nav, protocol.MediaMessage{}, true, false)
	if r3.Mode != hud.RefreshSpatialPatch {
		t.Fatalf("bucket change want spatial got %v dirty=%v", r3.Mode, r3.DirtyIDs)
	}
	if len(r3.DirtyIDs) != 1 || r3.DirtyIDs[0] != hudui.NodeDistance {
		t.Fatalf("dirty ids %v", r3.DirtyIDs)
	}
}

func TestRefreshOrchestratorRoadChangeFullFrame(t *testing.T) {
	o := hud.NewRefreshOrchestrator()
	nav := protocol.NavMessage{
		Active: true, DistanceM: 200, Road: "A", Maneuver: protocol.ManeuverStraight,
	}
	_ = o.Plan(hud.ScreenNav, nav, protocol.MediaMessage{}, true, true)
	nav.Road = "B"
	r := o.Plan(hud.ScreenNav, nav, protocol.MediaMessage{}, true, false)
	if r.Mode != hud.RefreshFullFrame {
		t.Fatalf("road change want full got %v", r.Mode)
	}
}
