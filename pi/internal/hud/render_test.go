package hud

import (
	"image"
	"path/filepath"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestRenderIsStrict1Bit(t *testing.T) {
	img := Render(ScreenNav, protocol.NavMessage{
		Active:       true,
		DistanceM:    200,
		DistanceText: "200 m",
		Road:         "High St",
		EtaMin:       12,
		Maneuver:     protocol.ManeuverLeft,
	}, protocol.MediaMessage{}, true)

	if img.Bounds() != image.Rect(0, 0, Width, Height) {
		t.Fatalf("size %v", img.Bounds())
	}
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			v := img.GrayAt(x, y).Y
			if v != 0 && v != 255 {
				t.Fatalf("mid-gray %d at (%d,%d) — arrow/text AA leaked", v, x, y)
			}
		}
	}
}

func TestFontSpecimensRender(t *testing.T) {
	root := findTestRepoRoot(t)
	SetAssetDir(filepath.Join(root, "assets", "hud"))
	cands := ListFontCandidates()
	if len(cands) < 1 {
		t.Skip("no font candidates under assets/fonts/candidates")
	}
	for _, c := range cands {
		img, err := RenderFontSpecimen(c.ID)
		if err != nil {
			t.Fatalf("%s: %v", c.ID, err)
		}
		if img.Bounds() != image.Rect(0, 0, Width, Height) {
			t.Fatalf("%s size %v", c.ID, img.Bounds())
		}
		var black int
		for y := 0; y < Height; y++ {
			for x := 0; x < Width; x++ {
				v := img.GrayAt(x, y).Y
				if v != 0 && v != 255 {
					t.Fatalf("%s mid-gray %d", c.ID, v)
				}
				if v == 0 {
					black++
				}
			}
		}
		if black < 50 {
			t.Fatalf("%s too empty (%d black px)", c.ID, black)
		}
	}
}
