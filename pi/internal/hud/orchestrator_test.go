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
	in := hud.ComposeInput(hud.ScreenNav, nav, protocol.MediaMessage{}, true)
	_, _, _ = o.PlanFromCompose(in, true)

	nav.DistanceM = 115
	nav.DistanceText = "115 m"
	in.Nav = nav
	rp, _, _ := o.PlanFromCompose(in, false)
	if rp.Mode != hud.RefreshNone {
		t.Fatalf("same bucket want none got %v", rp.Mode)
	}

	nav.DistanceM = 70
	nav.DistanceText = "70 m"
	in.Nav = nav
	rp, _, _ = o.PlanFromCompose(in, false)
	if rp.Mode != hud.RefreshSpatialPatch {
		t.Fatalf("bucket change want spatial got %v dirty=%v", rp.Mode, rp.DirtyIDs)
	}
	if len(rp.DirtyIDs) != 1 || rp.DirtyIDs[0] != hudui.NodeDistance {
		t.Fatalf("dirty ids %v", rp.DirtyIDs)
	}
}

func TestRefreshOrchestratorRoadChangeSpatial(t *testing.T) {
	o := hud.NewRefreshOrchestrator()
	nav := protocol.NavMessage{
		Active: true, DistanceM: 200, Road: "A", Maneuver: protocol.ManeuverStraight,
	}
	in := hud.ComposeInput(hud.ScreenNav, nav, protocol.MediaMessage{}, true)
	_, _, _ = o.PlanFromCompose(in, true)
	nav.Road = "B"
	in.Nav = nav
	rp, _, _ := o.PlanFromCompose(in, false)
	if rp.Mode != hud.RefreshSpatialPatch {
		t.Fatalf("road change want spatial got %v dirty=%v", rp.Mode, rp.DirtyIDs)
	}
}

func TestRefreshOrchestratorStatusNavSpatial(t *testing.T) {
	o := hud.NewRefreshOrchestrator()
	nav := protocol.NavMessage{Active: false}
	in := hud.ComposeInput(hud.ScreenStatus, nav, protocol.MediaMessage{}, true)
	_, _, _ = o.PlanFromCompose(in, true)
	nav.Active = true
	in.Nav = nav
	rp, _, _ := o.PlanFromCompose(in, false)
	if rp.Mode != hud.RefreshSpatialPatch {
		t.Fatalf("nav active change want spatial got %v dirty=%v", rp.Mode, rp.DirtyIDs)
	}
	if len(rp.DirtyIDs) != 1 || rp.DirtyIDs[0] != hudui.NodeStatusNav {
		t.Fatalf("dirty ids %v", rp.DirtyIDs)
	}
}

func TestRefreshOrchestratorLinkSpatial(t *testing.T) {
	o := hud.NewRefreshOrchestrator()
	nav := protocol.NavMessage{Active: true, DistanceM: 100, Road: "A", Maneuver: protocol.ManeuverStraight}
	in := hud.ComposeInput(hud.ScreenNav, nav, protocol.MediaMessage{}, false)
	_, _, _ = o.PlanFromCompose(in, true)
	in.Linked = true
	rp, _, _ := o.PlanFromCompose(in, false)
	if rp.Mode != hud.RefreshSpatialPatch {
		t.Fatalf("link change want spatial got %v dirty=%v", rp.Mode, rp.DirtyIDs)
	}
	if len(rp.DirtyIDs) != 1 || rp.DirtyIDs[0] != hudui.NodeLink {
		t.Fatalf("dirty ids %v", rp.DirtyIDs)
	}
}
