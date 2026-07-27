package hud

import (
	"fmt"
	"image"
)

func blitPatch(img *image.Gray, slot image.Rectangle, svg []byte) error {
	if img == nil || slot.Empty() {
		return fmt.Errorf("hud: bad patch target")
	}
	w, h := slot.Dx(), slot.Dy()
	patch, err := RasterizeSVGAt(svg, w, h)
	if err != nil {
		return err
	}
	b := img.Bounds()
	for y := 0; y < h; y++ {
		dy := slot.Min.Y + y
		if dy < b.Min.Y || dy >= b.Max.Y {
			continue
		}
		for x := 0; x < w; x++ {
			dx := slot.Min.X + x
			if dx < b.Min.X || dx >= b.Max.X {
				continue
			}
			img.SetGray(dx, dy, patch.GrayAt(x, y))
		}
	}
	return nil
}
