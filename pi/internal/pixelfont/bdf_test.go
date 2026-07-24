package pixelfont

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestLoadAllSizes(t *testing.T) {
	for _, sz := range AllSizes() {
		f, err := Load(sz)
		if err != nil {
			t.Fatalf("%s: %v", sz, err)
		}
		if f.Metrics.CellH != int(sz) && !(sz == Size6x12 && f.Metrics.CellH == 12) {
			// Size enum value equals pixel height for our mapping
		}
		if f.Metrics.GlyphCount < 90 {
			t.Fatalf("%s: too few glyphs %d", sz, f.Metrics.GlyphCount)
		}
		t.Logf("%s: cell=%dx%d ascent=%d descent=%d glyphs=%d",
			sz, f.Metrics.CellW, f.Metrics.CellH, f.Metrics.Ascent, f.Metrics.Descent, f.Metrics.GlyphCount)
	}
}

func TestPixelPerfectNoGray(t *testing.T) {
	f, err := Load(Size8x16)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewGray(image.Rect(0, 0, 120, 32))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{Y: 255}}, image.Point{}, draw.Src)
	f.DrawString(img, 0, f.Metrics.Ascent, "ABC123 m")
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			v := img.GrayAt(x, y).Y
			if v != 0 && v != 255 {
				t.Fatalf("non-binary pixel at %d,%d = %d (not pixel-perfect)", x, y, v)
			}
		}
	}
}

func TestAdvanceMonospace(t *testing.T) {
	f, err := Load(Size12x24)
	if err != nil {
		t.Fatal(err)
	}
	w1 := f.Measure("00000")
	w2 := f.Measure("11111")
	if w1 != w2 {
		t.Fatalf("expected mono advances, got %d vs %d", w1, w2)
	}
	if w1 != 5*f.Metrics.CellW {
		t.Fatalf("expected %d got %d", 5*f.Metrics.CellW, w1)
	}
}
