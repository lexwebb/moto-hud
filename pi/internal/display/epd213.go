package display

import "image"

// Waveshare 2.13" B/W V3/V4 controller memory is portrait 122×250.
const (
	epd213MemW = 122
	epd213MemH = 250
)

// packEPD213 rotates a landscape 250×122 gray frame into the portrait 1-bit
// buffer Waveshare expects (1 = white, 0 = black).
func packEPD213(src *image.Gray) []byte {
	bit := floydSteinberg(src)
	sb := bit.Bounds()
	sw, sh := sb.Dx(), sb.Dy() // expect 250×122
	rowBytes := (epd213MemW + 7) / 8
	buf := make([]byte, rowBytes*epd213MemH)
	for i := range buf {
		buf[i] = 0xFF // white
	}
	// CW rotate: dstX=srcY, dstY=srcW-1-srcX → 122×250
	for sy := 0; sy < sh; sy++ {
		for sx := 0; sx < sw; sx++ {
			dx := sy
			dy := sw - 1 - sx
			if dx < 0 || dx >= epd213MemW || dy < 0 || dy >= epd213MemH {
				continue
			}
			if bit.GrayAt(sb.Min.X+sx, sb.Min.Y+sy).Y >= 128 {
				continue // white already
			}
			bi := dy*rowBytes + dx/8
			buf[bi] &^= 0x80 >> uint(dx%8)
		}
	}
	return buf
}
