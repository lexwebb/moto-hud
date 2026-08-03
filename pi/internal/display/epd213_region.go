package display

import "image"

const (
	hudCanvasW = 250
	hudCanvasH = 122
)

// canvasRectToEPDBounds maps a HUD canvas rectangle to EPD memory coordinates (122×250 portrait).
// EPD mapping matches packEPD213: dx = sy, dy = hudCanvasW - 1 - sx.
// So EPD X = canvas Y (byte-addressed) and EPD Y = 249−canvas X.
func canvasRectToEPDBounds(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}
	sw := hudCanvasW
	type pt struct{ sx, sy int }
	corners := []pt{
		{r.Min.X, r.Min.Y},
		{r.Max.X - 1, r.Min.Y},
		{r.Min.X, r.Max.Y - 1},
		{r.Max.X - 1, r.Max.Y - 1},
	}
	minDX, minDY := epd213MemW, epd213MemH
	maxDX, maxDY := -1, -1
	for _, c := range corners {
		dx, dy := c.sy, sw-1-c.sx
		if dx < minDX {
			minDX = dx
		}
		if dx > maxDX {
			maxDX = dx
		}
		if dy < minDY {
			minDY = dy
		}
		if dy > maxDY {
			maxDY = dy
		}
	}
	if maxDX < 0 {
		return image.Rectangle{}
	}
	return image.Rect(minDX, minDY, maxDX+1, maxDY+1)
}

// packEPD213Window packs EPD pixels for canvasRegion (must be non-empty).
// The EPD X range is forced to 8px (byte) alignment so the SPI stream matches
// setWindow's 0x44 byte addresses — misalignment scrambles vertical stripes.
func packEPD213Window(src *image.Gray, canvasRegion image.Rectangle) ([]byte, image.Rectangle) {
	epdR := canvasRectToEPDBounds(canvasRegion)
	if epdR.Empty() {
		return nil, image.Rectangle{}
	}
	// EPD X is the byte-addressed axis (controller 0x44 / 0x4E).
	x0 := epdR.Min.X &^ 7
	x1 := (epdR.Max.X + 7) &^ 7
	if x1 > epd213MemW {
		x1 = epd213MemW
	}
	epdR = image.Rect(x0, epdR.Min.Y, x1, epdR.Max.Y)

	w := epdR.Dx()
	h := epdR.Dy()
	rowBytes := w / 8
	buf := make([]byte, rowBytes*h)
	for i := range buf {
		buf[i] = 0xFF
	}
	sw := src.Bounds().Dx()
	sb := src.Bounds()
	// Pack every pixel in the aligned EPD window from the framebuffer so row
	// length matches the hardware window (no short / phase-shifted rows).
	for dy := epdR.Min.Y; dy < epdR.Max.Y; dy++ {
		for dx := epdR.Min.X; dx < epdR.Max.X; dx++ {
			sx := sw - 1 - dy
			sy := dx
			if sx < sb.Min.X || sx >= sb.Max.X || sy < sb.Min.Y || sy >= sb.Max.Y {
				continue
			}
			if src.GrayAt(sx, sy).Y >= 128 {
				continue
			}
			relX := dx - epdR.Min.X
			relY := dy - epdR.Min.Y
			bi := relY*rowBytes + relX/8
			buf[bi] &^= 0x80 >> uint(relX%8)
		}
	}
	return buf, epdR
}

// AlignCanvasEPD expands a canvas dirty rect for Waveshare window updates.
// After CW rotate, EPD X = canvas Y, and the controller addresses X in 8px
// bytes — so canvas Y (and X for symmetry) must be 8-aligned.
func AlignCanvasEPD(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}
	x0 := r.Min.X &^ 7
	x1 := (r.Max.X + 7) &^ 7
	y0 := r.Min.Y &^ 7
	y1 := (r.Max.Y + 7) &^ 7
	if x1 > hudCanvasW {
		x1 = hudCanvasW
	}
	if y1 > hudCanvasH {
		y1 = hudCanvasH
	}
	return image.Rect(x0, y0, x1, y1)
}
