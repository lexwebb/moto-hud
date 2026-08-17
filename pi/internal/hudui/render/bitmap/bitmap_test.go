package bitmap_test

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/hudui/render/bitmap"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/pixelfont"
)

func TestRasterize_textMatchesPixelfont(t *testing.T) {
	face, err := pixelfont.Load(pixelfont.Size16x32)
	if err != nil {
		t.Fatal(err)
	}
	const w, h = 120, 40
	const s = "200 m"
	baseline := face.Metrics.Ascent + 2
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text("distance", scene.Face16x32, w, baseline, "end", s)
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	want := image.NewGray(image.Rect(0, 0, w, h))
	draw.Draw(want, want.Bounds(), &image.Uniform{color.Gray{Y: 255}}, image.Point{}, draw.Src)
	face.DrawString(want, w-face.Measure(s), baseline, s)
	if !grayEqual1bit(got, want) {
		t.Fatalf("text raster ≠ DrawString (%d ink got, %d want)", countInk(got), countInk(want))
	}
}

func TestRasterize_bodyTextHasInk(t *testing.T) {
	face, err := pixelfont.Load(pixelfont.Size8x16)
	if err != nil {
		t.Fatal(err)
	}
	doc := scene.Patch(100, 32, func(b *scene.Builder) {
		b.Text("road", scene.Face8x16, 0, face.Metrics.Ascent, "start", "High St")
		b.Text("eta", scene.Face8x16, 0, face.Metrics.CellH+face.Metrics.Ascent, "start", "12m")
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	if n := countInk(got); n < 40 {
		t.Fatalf("body text too sparse: %d", n)
	}
}

func TestRasterize_filledRect(t *testing.T) {
	doc := scene.Patch(20, 20, func(b *scene.Builder) {
		b.Rect("a", 2, 2, 8, 8, true)
		b.Rect("b", 12, 4, 4, 4, false)
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	if got.GrayAt(5, 5).Y >= 128 {
		t.Fatal("expected filled rect ink")
	}
	if got.GrayAt(12, 4).Y >= 128 {
		t.Fatal("expected stroke rect top-left ink")
	}
	if got.GrayAt(14, 6).Y < 128 {
		t.Fatal("expected hollow interior")
	}
}

func TestRasterize_rejectsRawSVG(t *testing.T) {
	doc := scene.Patch(10, 10, func(b *scene.Builder) {
		b.Raw(`<rect x="0" y="0" width="1" height="1" fill="#000"/>`)
	})
	if _, err := bitmap.Rasterize(doc); err == nil {
		t.Fatal("expected RawSVG error")
	}
}

func TestRasterize_linkMarkHasInk(t *testing.T) {
	doc := scene.Patch(16, 12, func(b *scene.Builder) {
		b.Append(compose.LinkMarkNodes(true)...)
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	if n := countInk(got); n < 10 {
		t.Fatalf("link mark too sparse: %d ink px", n)
	}
}

func countInk(img *image.Gray) int {
	n := 0
	for _, p := range img.Pix {
		if p < 128 {
			n++
		}
	}
	return n
}

func grayEqual1bit(a, b *image.Gray) bool {
	ab, bb := a.Bounds(), b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return false
	}
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			ga := a.GrayAt(ab.Min.X+x, ab.Min.Y+y).Y < 128
			gb := b.GrayAt(bb.Min.X+x, bb.Min.Y+y).Y < 128
			if ga != gb {
				return false
			}
		}
	}
	return true
}
