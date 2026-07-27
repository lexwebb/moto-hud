package display

import (
	"image"
	"image/color"
	"testing"
)

func TestCanvasRectToEPDBoundsTopLeft(t *testing.T) {
	// Canvas (0,0) → EPD (0, 249)
	r := image.Rect(0, 0, 8, 8)
	epd := canvasRectToEPDBounds(r)
	if epd.Min.X != 0 || epd.Min.Y != 249-7 || epd.Max.Y != 250 {
		t.Fatalf("epd bounds %v", epd)
	}
}

func TestPackEPD213WindowSubset(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 250, 122))
	for i := range src.Pix {
		src.Pix[i] = 255
	}
	src.SetGray(10, 10, color.Gray{Y: 0})

	region := image.Rect(8, 8, 32, 24)
	buf, epdR := packEPD213Window(src, region)
	if epdR.Empty() || len(buf) == 0 {
		t.Fatal("empty region pack")
	}
	full := packEPD213(src)
	// Ink in subset should appear somewhere non-0xFF in subset buffer.
	ink := 0
	for _, b := range buf {
		if b != 0xFF {
			ink++
		}
	}
	if ink == 0 {
		t.Fatal("expected ink in window buffer")
	}
	_ = full
}
