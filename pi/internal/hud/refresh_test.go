package hud

import (
	"os"
	"path/filepath"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestRefreshGateThresholds(t *testing.T) {
	g := &RefreshGate{}
	nav := protocol.NavMessage{Active: true, Maneuver: protocol.ManeuverLeft, Road: "A", DistanceM: 800}
	if !g.ShouldRedraw(ScreenNav, nav, false) {
		t.Fatal("first draw should redraw")
	}
	nav.DistanceM = 700
	if g.ShouldRedraw(ScreenNav, nav, false) {
		t.Fatal("same bucket should skip")
	}
	nav.DistanceM = 450
	if !g.ShouldRedraw(ScreenNav, nav, false) {
		t.Fatal("crossing 500->200 bucket should redraw")
	}
	nav.Maneuver = protocol.ManeuverRight
	if !g.ShouldRedraw(ScreenNav, nav, false) {
		t.Fatal("maneuver change should redraw")
	}
	if !g.ShouldRedraw(ScreenMedia, nav, false) {
		t.Fatal("screen change should redraw")
	}
	if g.ShouldRedraw(ScreenMedia, nav, false) {
		t.Fatal("media without force should skip")
	}
	if !g.ShouldRedraw(ScreenMedia, nav, true) {
		t.Fatal("force should redraw")
	}
}

func TestRenderSize(t *testing.T) {
	root := findTestRepoRoot(t)
	SetAssetDir(filepath.Join(root, "assets", "hud"))
	img := Render(ScreenNav, protocol.NavMessage{
		Active: true, DistanceText: "200 m", Road: "High St", Maneuver: protocol.ManeuverLeft,
	}, protocol.MediaMessage{}, true)
	if img.Bounds().Dx() != Width || img.Bounds().Dy() != Height {
		t.Fatalf("got %v", img.Bounds())
	}
	// Should have drawn content (antialiased → some mid grays / dark pixels)
	dark := 0
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			if img.GrayAt(x, y).Y < 200 {
				dark++
			}
		}
	}
	if dark < 50 {
		t.Fatalf("expected drawn content, dark=%d", dark)
	}
}

func findTestRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, "assets", "hud", "nav.svg")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("assets/hud/nav.svg not found")
	return ""
}
