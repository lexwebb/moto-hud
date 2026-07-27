package hud

import (
	"strings"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestLaneStripSVG_activeAndInactive(t *testing.T) {
	svg := laneStripSVG([]protocol.LaneInfo{
		{Directions: []string{"left"}, Active: false},
		{Directions: []string{"straight"}, Active: true},
		{Directions: []string{"right"}, Active: true},
	}, 80)
	if !strings.Contains(svg, `id="lanes"`) {
		t.Fatal("missing lanes group")
	}
	if !strings.Contains(svg, `fill="none"`) {
		t.Fatal("expected outline for inactive lane")
	}
	if strings.Count(svg, `<rect`) != 3 {
		t.Fatalf("expected 3 lane rects, got svg=%s", svg)
	}
}

func TestBuildNavBody_includesLanes(t *testing.T) {
	vars := buildNavBody(protocol.NavMessage{
		Active:       true,
		DistanceM:    120,
		DistanceText: "120 m",
		Road:         "High St",
		Maneuver:     protocol.ManeuverLeft,
		Lanes: []protocol.LaneInfo{
			{Directions: []string{"straight"}, Active: false},
			{Directions: []string{"left"}, Active: true},
		},
	}, true)
	if !strings.Contains(vars["body"], `id="lanes"`) {
		t.Fatal("classic nav body should include lane strip")
	}
}
