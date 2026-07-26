package display

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	// LCDWidth/LCDHeight are Display HAT Mini panel pixels (ST7789).
	LCDWidth  = 320
	LCDHeight = 240
)

// LetterboxGray centers a 250×122 gray HUD frame on a 320×240 RGB canvas
// as ink-on-paper (black/white), matching the 1-bit instrument look.
func LetterboxGray(src *image.Gray) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, LCDWidth, LCDHeight))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	if src == nil {
		return dst
	}
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	ox := (LCDWidth - sw) / 2
	oy := (LCDHeight - sh) / 2
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			v := src.GrayAt(sb.Min.X+x, sb.Min.Y+y).Y
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			if v < 128 {
				c = color.RGBA{A: 255}
			}
			dst.SetRGBA(ox+x, oy+y, c)
		}
	}
	return dst
}

// RotateRGBA180 returns a new image rotated 180° (Display HAT Mini default).
func RotateRGBA180(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.SetRGBA(w-1-x, h-1-y, src.RGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// RGBAToRGB565 packs an RGBA image to big-endian RGB565 bytes.
func RGBAToRGB565(src *image.RGBA) []byte {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]byte, w*h*2)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.RGBAAt(b.Min.X+x, b.Min.Y+y)
			rgb := uint16((uint16(c.R)&0xF8)<<8) | (uint16(c.G)&0xFC)<<3 | (uint16(c.B) >> 3)
			out[i] = byte(rgb >> 8)
			out[i+1] = byte(rgb)
			i += 2
		}
	}
	return out
}
