package hud

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestMinimapSVGHasContextRouteAndRider(t *testing.T) {
	rider := protocol.RibbonPoint{X: 0, Y: -40}
	mm := &protocol.MinimapMessage{
		Route: []protocol.RibbonPoint{
			{X: 0, Y: -50}, {X: 0, Y: 0}, {X: 20, Y: 30},
		},
		Context: [][]protocol.RibbonPoint{
			{{X: -20, Y: -30}, {X: -18, Y: 20}},
			{{X: 22, Y: -10}, {X: 25, Y: 40}},
		},
		Rider: &rider,
	}
	svg := minimapSVG(mm, 70, 90)
	if !strings.Contains(svg, `id="route"`) {
		t.Fatalf("expected solid route, got %s", svg)
	}
	if !strings.Contains(svg, `id="rider"`) {
		t.Fatalf("expected rider blob, got %s", svg)
	}
	if !strings.Contains(svg, `id="turn"`) {
		t.Fatalf("expected turn mark, got %s", svg)
	}
	if strings.Contains(svg, "stroke=") {
		t.Fatalf("expected fill-only rects, got stroked geometry: %s", svg)
	}
}

func TestBuildNavBodyLiveUsesMinimap(t *testing.T) {
	rider := protocol.RibbonPoint{X: 1, Y: -35}
	vars := buildNavBody(protocol.NavMessage{
		Active:       true,
		DistanceText: "≈ 100 m",
		Road:         "Whitehall",
		Maneuver:     protocol.ManeuverRight,
		EtaMin:       6,
		Minimap: &protocol.MinimapMessage{
			Route:   []protocol.RibbonPoint{{X: 0, Y: -40}, {X: 0, Y: 0}, {X: 15, Y: 25}},
			Context: [][]protocol.RibbonPoint{{{X: -15, Y: -20}, {X: -12, Y: 30}}},
			Rider:   &rider,
		},
	}, true)
	body := vars["body"]
	if strings.Contains(body, `id="maneuver"`) {
		t.Fatal("minimap live layout should not include maneuver glyph")
	}
	if !strings.Contains(body, `id="route"`) {
		t.Fatal("expected solid route")
	}
}

func TestMinimapSVGCentersTurn(t *testing.T) {
	mm := &protocol.MinimapMessage{
		Route: []protocol.RibbonPoint{
			{X: 0, Y: -50}, {X: 0, Y: 0}, {X: 20, Y: 30},
		},
	}
	const w, h = 70, 90
	svg := minimapSVG(mm, w, h)
	needle := fmt.Sprintf(`id="turn" x="%d" y="%d"`, w/2-2, h/2-2)
	if !strings.Contains(svg, needle) {
		t.Fatalf("turn not centered, want %s in %s", needle, svg)
	}
}

func TestSchematizeTubeIsOctilinear(t *testing.T) {
	pts := []protocol.RibbonPoint{
		{X: 0, Y: 0}, {X: 10, Y: 3},
	}
	out := schematizeTube(pts, 1)
	if len(out) < 2 {
		t.Fatal("expected snapped polyline")
	}
	dx := out[1].X - out[0].X
	dy := out[1].Y - out[0].Y
	ang := math.Atan2(dy, dx)
	step := math.Pi / 4
	nearest := math.Round(ang/step) * step
	if math.Abs(ang-nearest) > 1e-6 {
		t.Fatalf("leg not octilinear: ang=%.3f", ang)
	}
}

func TestTubeKneeAllowsDiagonal(t *testing.T) {
	got := tubeKnee([2]int{0, 0}, [2]int{10, 4})
	if len(got) != 2 || got[0] != ([2]int{4, 4}) {
		t.Fatalf("expected 45° then axis finish, got %v", got)
	}
}

func TestMinimapViewRadiusZoomsIn(t *testing.T) {
	mm := &protocol.MinimapMessage{
		Route: []protocol.RibbonPoint{
			{X: 0, Y: -12}, {X: 0, Y: 0}, {X: 10, Y: 8},
		},
	}
	r := minimapViewRadius(mm)
	if r > 30 {
		t.Fatalf("expected zoom-in radius ≤30, got %.1f", r)
	}
	if r < minimapRadiusMin {
		t.Fatalf("radius below minimum: %.1f", r)
	}
}

func TestMinimapViewRadiusCapsAt50(t *testing.T) {
	mm := &protocol.MinimapMessage{
		Route: []protocol.RibbonPoint{
			{X: 0, Y: -80}, {X: 0, Y: 0}, {X: 60, Y: 40},
		},
	}
	r := minimapViewRadius(mm)
	if r != minimapRadiusMax {
		t.Fatalf("expected cap %.0f, got %.1f", minimapRadiusMax, r)
	}
}

func TestRenderMinimapNotBlank(t *testing.T) {
	mm := &protocol.MinimapMessage{
		Route: []protocol.RibbonPoint{{X: 0, Y: -40}, {X: 0, Y: 0}, {X: -30, Y: 0}},
	}
	img, err := RenderMinimap(mm, 70, 80)
	if err != nil {
		t.Fatal(err)
	}
	var black, white int
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y < 128 {
				black++
			} else {
				white++
			}
		}
	}
	if white < 100 || black < 20 {
		t.Fatalf("expected mixed 1-bit pane, got white=%d black=%d", white, black)
	}
}
