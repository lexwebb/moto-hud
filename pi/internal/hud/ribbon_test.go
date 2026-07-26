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

func TestRibbonForNavPrefersProtocolPoints(t *testing.T) {
	nav := protocol.NavMessage{
		Maneuver: protocol.ManeuverLeft, // schematic would bend left
		RibbonPoints: []protocol.RibbonPoint{
			{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 20, Y: 20},
		},
		RibbonTurn: 2,
	}
	pts, idx := ribbonForNav(nav)
	if len(pts) != 3 || idx != 2 || pts[2].X != 20 {
		t.Fatalf("expected protocol points, got %v idx=%d", pts, idx)
	}
	nav.RibbonPoints = nil
	pts, idx = ribbonForNav(nav)
	schematic, sIdx := schematicRibbonForManeuver(protocol.ManeuverLeft)
	if len(pts) != len(schematic) || idx != sIdx {
		t.Fatalf("expected schematic fallback, got %v %d", pts, idx)
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
	if !strings.Contains(body, `id="maneuver"`) {
		t.Fatal("synthetic ribbon layout should keep maneuver glyph")
	}
	if strings.Contains(body, `id="ticks"`) {
		t.Fatal("progress ticks should be replaced by ribbon on active nav")
	}
}

func TestBuildNavBodyLiveTwoColumn(t *testing.T) {
	vars := buildNavBody(protocol.NavMessage{
		Active:       true,
		DistanceText: "≈ 120 m",
		Road:         "Ridge Rd",
		Maneuver:     protocol.ManeuverRight,
		EtaMin:       8,
		RibbonPoints: []protocol.RibbonPoint{
			{X: 0, Y: 0}, {X: 0, Y: 40}, {X: 20, Y: 60},
		},
		RibbonTurn: 1,
	}, true)
	body := vars["body"]
	if !strings.Contains(body, `id="ribbon"`) {
		t.Fatalf("missing ribbon in live body: %s", body)
	}
	if strings.Contains(body, `id="maneuver"`) {
		t.Fatal("live ribbon layout should not include maneuver glyph")
	}
	if !strings.Contains(body, "≈120m") {
		t.Fatalf("expected compacted distance ≈120m, got: %s", body)
	}
	if !strings.Contains(body, "Ridge Rd") {
		t.Fatal("expected road name in live layout")
	}
}

func TestCompactDistanceText(t *testing.T) {
	cases := map[string]string{
		"≈ 120 m":   "≈120m",
		"≈ 1.2 km":  "≈1.2km",
		"120 m":     "120m",
		"≈120m":     "≈120m",
		"":          "",
	}
	for in, want := range cases {
		if got := compactDistanceText(in); got != want {
			t.Fatalf("compactDistanceText(%q)=%q want %q", in, got, want)
		}
	}
}

func TestProgressTicksStillAvailable(t *testing.T) {
	svg := progressTicks(80)
	if !strings.Contains(svg, "<rect ") {
		t.Fatalf("expected tick rects, got %s", svg)
	}
}
