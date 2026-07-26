package display

import (
	"image"
	"image/color"
)

// floydSteinberg dithers antialiased gray to 1-bit black/white.
func floydSteinberg(src *image.Gray) *image.Gray {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	buf := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf[y*w+x] = float64(src.GrayAt(b.Min.X+x, b.Min.Y+y).Y)
		}
	}
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			old := buf[i]
			neu := 255.0
			if old < 128 {
				neu = 0
			}
			out.SetGray(x, y, color.Gray{Y: uint8(neu)})
			err := old - neu
			if x+1 < w {
				buf[i+1] += err * 7 / 16
			}
			if y+1 < h {
				if x > 0 {
					buf[(y+1)*w+(x-1)] += err * 3 / 16
				}
				buf[(y+1)*w+x] += err * 5 / 16
				if x+1 < w {
					buf[(y+1)*w+(x+1)] += err * 1 / 16
				}
			}
		}
	}
	return out
}
