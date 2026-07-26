package hud

import (
	"strings"
	"testing"

	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

func TestAbbreviateRoad(t *testing.T) {
	cases := map[string]string{
		"Northumberland Avenue": "Northumberland Ave",
		"Whitehall Place":       "Whitehall Pl",
		"High Street":           "High St",
		"South Bridge Road":     "S Br Rd",
		"Ridge Rd":              "Ridge Rd",
		"onto Harbor Boulevard": "onto Harbor Blvd",
		"North Street":          "N St",
	}
	for in, want := range cases {
		if got := abbreviateRoad(in); got != want {
			t.Fatalf("abbreviateRoad(%q)=%q want %q", in, got, want)
		}
	}
}

func TestAbbreviateRoadKeepsCompoundNames(t *testing.T) {
	got := abbreviateRoad("Northumberland")
	if got != "Northumberland" {
		t.Fatalf("got %q", got)
	}
}

func TestWrapLines(t *testing.T) {
	face := mustFace(pixelfont.Size8x16)
	lines := wrapLines(face, "Northumberland Ave", 40, 3)
	if len(lines) < 2 {
		t.Fatalf("expected wrap into multiple lines, got %v", lines)
	}
	for _, line := range lines {
		if face.Measure(line) > 40 {
			t.Fatalf("line too wide: %q (%dpx)", line, face.Measure(line))
		}
	}
	single := wrapLines(face, "High St", 200, 3)
	if len(single) != 1 || single[0] != "High St" {
		t.Fatalf("short name should stay one line, got %v", single)
	}
}

func TestBuildNavBodyAbbreviatesAndWrapsLive(t *testing.T) {
	vars := buildNavBody(protocol.NavMessage{
		Active:       true,
		DistanceText: "≈ 200 m",
		Road:         "Northumberland Avenue",
		Maneuver:     protocol.ManeuverRight,
		EtaMin:       5,
		RibbonPoints: []protocol.RibbonPoint{
			{X: 0, Y: 0}, {X: 0, Y: 40}, {X: 15, Y: 55},
		},
		RibbonTurn: 1,
	}, true)
	body := vars["body"]
	if strings.Contains(body, "Avenue") {
		t.Fatalf("expected Avenue abbreviated, got: %s", body)
	}
	if !strings.Contains(body, "Ave") {
		t.Fatal("expected Ave abbreviation")
	}
	if !strings.Contains(body, "Northumb") {
		t.Fatalf("expected road stem, got: %s", body)
	}
}
