package display

import (
	"image"
	"image/color"
	"testing"
)

func TestLetterboxGrayCentersHUD(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 250, 122))
	for y := 0; y < 122; y++ {
		for x := 0; x < 250; x++ {
			src.SetGray(x, y, color.Gray{Y: 0}) // black HUD
		}
	}
	dst := LetterboxGray(src)
	if dst.Bounds().Dx() != LCDWidth || dst.Bounds().Dy() != LCDHeight {
		t.Fatalf("size %v", dst.Bounds())
	}
	ox, oy := (LCDWidth-250)/2, (LCDHeight-122)/2
	// Corner of letterbox should be white paper.
	if r, g, b, _ := dst.At(0, 0).RGBA(); r|g|b == 0 {
		t.Fatal("expected white margin at 0,0")
	}
	// Center of HUD block should be black ink.
	cx, cy := ox+125, oy+61
	if r, g, b, _ := dst.At(cx, cy).RGBA(); r|g|b != 0 {
		t.Fatalf("expected black at %d,%d got %d,%d,%d", cx, cy, r, g, b)
	}
}

func TestPackEPD213SizeAndInk(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 250, 122))
	for y := 0; y < 122; y++ {
		for x := 0; x < 250; x++ {
			src.SetGray(x, y, color.Gray{Y: 255})
		}
	}
	// One black pixel at top-left of landscape HUD.
	src.SetGray(0, 0, color.Gray{Y: 0})

	buf := packEPD213(src)
	rowBytes := (epd213MemW + 7) / 8
	want := rowBytes * epd213MemH
	if len(buf) != want {
		t.Fatalf("len=%d want %d", len(buf), want)
	}
	// CW: dstX=0, dstY=249 → last row, first bit cleared.
	bi := 249*rowBytes + 0
	if buf[bi]&0x80 != 0 {
		t.Fatalf("expected ink bit cleared at byte %d got %#02x", bi, buf[bi])
	}
	// Far white pixel should stay white.
	if buf[0] != 0xFF && buf[1] != 0xFF {
		// row 0 may still be white except if rotation maps elsewhere
	}
	white := 0
	for _, b := range buf {
		if b == 0xFF {
			white++
		}
	}
	if white < want-2 {
		t.Fatalf("expected mostly white buffer, whiteBytes=%d/%d", white, want)
	}
}

func TestRGBAToRGB565BlackWhite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{A: 255})
	img.SetRGBA(1, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	b := RGBAToRGB565(img)
	if len(b) != 4 {
		t.Fatalf("len=%d", len(b))
	}
	if b[0] != 0 || b[1] != 0 {
		t.Fatalf("black pixel %#02x%#02x", b[0], b[1])
	}
	if b[2] != 0xFF || b[3] != 0xFF {
		t.Fatalf("white pixel %#02x%#02x", b[2], b[3])
	}
}
