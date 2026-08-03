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

func TestAlignCanvasEPDPadsYForEPDX(t *testing.T) {
	// Distance-like slot with Y not on an 8px boundary (EPD X after rotate).
	r := image.Rect(48, 18, 240, 58)
	a := AlignCanvasEPD(r)
	if a.Min.Y%8 != 0 || a.Max.Y%8 != 0 {
		t.Fatalf("canvas Y not 8-aligned: %v", a)
	}
	epd := canvasRectToEPDBounds(a)
	if epd.Min.X%8 != 0 || epd.Dx()%8 != 0 {
		t.Fatalf("EPD X not byte-aligned after align: %v", epd)
	}
}

func TestPackEPD213WindowByteAligned(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 250, 122))
	for i := range src.Pix {
		src.Pix[i] = 255
	}
	src.SetGray(10, 10, color.Gray{Y: 0})

	// Unaligned canvas Y — packer must expand EPD X to bytes.
	region := AlignCanvasEPD(image.Rect(8, 10, 32, 24))
	buf, epdR := packEPD213Window(src, region)
	if epdR.Empty() || len(buf) == 0 {
		t.Fatal("empty region pack")
	}
	if epdR.Min.X%8 != 0 || epdR.Dx()%8 != 0 {
		t.Fatalf("EPD window not byte-aligned: %v", epdR)
	}
	rowBytes := epdR.Dx() / 8
	want := rowBytes * epdR.Dy()
	if len(buf) != want {
		t.Fatalf("buf len %d want %d (rowBytes=%d h=%d)", len(buf), want, rowBytes, epdR.Dy())
	}
	ink := 0
	for _, b := range buf {
		if b != 0xFF {
			ink++
		}
	}
	if ink == 0 {
		t.Fatal("expected ink in window buffer")
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
	ink := 0
	for _, b := range buf {
		if b != 0xFF {
			ink++
		}
	}
	if ink == 0 {
		t.Fatal("expected ink in window buffer")
	}
}
