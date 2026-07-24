package hud

import (
	"strings"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestRoadRibbonEmptyIsDashed(t *testing.T) {
	svg := roadRibbonSVG(nil, -1, 200, 36)
	if !strings.Contains(svg, "stroke-dasharray") {
		t.Fatalf("expected dashed placeholder, got %s", svg)
	}
}

func TestRoadRibbonPathAndTurnMark(t *testing.T) {
	pts := []RoadPoint{{110, 0}, {110, 22}, {170, 34}}
	svg := roadRibbonSVG(pts, 1, 200, 36)
	if !strings.Contains(svg, "<path ") {
		t.Fatalf("expected path, got %s", svg)
	}
	if !strings.Contains(svg, "<rect ") {
		t.Fatalf("expected turn rect, got %s", svg)
	}
}

func TestSchematicRibbonForManeuver(t *testing.T) {
	pts, idx := schematicRibbonForManeuver(protocol.ManeuverRight)
	if len(pts) < 2 || idx < 0 {
		t.Fatalf("right: pts=%v idx=%d", pts, idx)
	}
	pts, idx = schematicRibbonForManeuver(protocol.ManeuverUnknown)
	if pts != nil || idx != -1 {
		t.Fatalf("unknown should be empty, got %v %d", pts, idx)
	}
}

func TestBuildNavBodyIncludesRibbon(t *testing.T) {
	vars := buildNavBody(protocol.NavMessage{
		Active:       true,
		DistanceText: "120 m",
		Road:         "Ridge Rd",
		Maneuver:     protocol.ManeuverRight,
		EtaMin:       8,
	}, true)
	body := vars["body"]
	if !strings.Contains(body, `id="ribbon"`) {
		t.Fatalf("missing ribbon in body: %s", body)
	}
	if strings.Contains(body, `id="ticks"`) {
		t.Fatal("progress ticks should be replaced by ribbon on active nav")
	}
}

func TestProgressTicksStillAvailable(t *testing.T) {
	svg := progressTicks(80)
	if !strings.Contains(svg, "<rect ") {
		t.Fatalf("expected tick rects, got %s", svg)
	}
}
