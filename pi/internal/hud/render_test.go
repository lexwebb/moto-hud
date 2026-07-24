package hud

import (
	"image"
	"image/color"
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

func TestToGray1BitSnaps(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	src.Set(1, 0, color.RGBA{R: 180, G: 180, B: 180, A: 255}) // fringe → black (<200)
	src.Set(0, 1, color.RGBA{R: 220, G: 220, B: 220, A: 255}) // near-white → white
	src.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	out := toGray1Bit(src, 2, 2)
	want := []uint8{0, 0, 255, 255}
	i := 0
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if got := out.GrayAt(x, y).Y; got != want[i] {
				t.Fatalf("(%d,%d)=%d want %d", x, y, got, want[i])
			}
			i++
		}
	}
}

func TestFontSpecimensRender(t *testing.T) {
	SetAssetDir("/Users/lex/src/moto-hud/assets/hud")
	cands := ListFontCandidates()
	if len(cands) < 2 {
		t.Fatalf("expected candidates, got %d", len(cands))
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
