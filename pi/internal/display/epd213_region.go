package display

import "image"

const hudCanvasW = 250

// canvasRectToEPDBounds maps a HUD canvas rectangle to EPD memory coordinates (122×250 portrait).
// EPD mapping matches packEPD213: dx = sy, dy = hudCanvasW - 1 - sx.
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

// packEPD213Window packs only the EPD pixels for canvasRegion (must be non-empty).
func packEPD213Window(src *image.Gray, canvasRegion image.Rectangle) ([]byte, image.Rectangle) {
	epdR := canvasRectToEPDBounds(canvasRegion)
	if epdR.Empty() {
		return nil, image.Rectangle{}
	}
	w := epdR.Dx()
	h := epdR.Dy()
	rowBytes := (w + 7) / 8
	buf := make([]byte, rowBytes*h)
	for i := range buf {
		buf[i] = 0xFF
	}
	sw := src.Bounds().Dx()
	sb := src.Bounds()
	for dy := epdR.Min.Y; dy < epdR.Max.Y; dy++ {
		for dx := epdR.Min.X; dx < epdR.Max.X; dx++ {
			sx := sw - 1 - dy
			sy := dx
			if sx < sb.Min.X || sx >= sb.Max.X || sy < sb.Min.Y || sy >= sb.Max.Y {
				continue
			}
			if sx < canvasRegion.Min.X || sx >= canvasRegion.Max.X || sy < canvasRegion.Min.Y || sy >= canvasRegion.Max.Y {
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

// AlignCanvasEPD expands a canvas dirty rect to 8px columns (HUD X axis).
func AlignCanvasEPD(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}
	x0 := r.Min.X &^ 7
	x1 := (r.Max.X + 7) &^ 7
	if x1 < r.Max.X {
		x1 += 8
	}
	if x1 > hudCanvasW {
		x1 = hudCanvasW
	}
	return image.Rect(x0, r.Min.Y, x1, r.Max.Y)
}
