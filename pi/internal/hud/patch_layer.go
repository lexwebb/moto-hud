package hud

import (
	"errors"
	"image"

	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/render/bitmap"
)

// PatchLayer rasterizes a plan layer patch into the framebuffer at its template-defined slot.
func PatchLayer(img *image.Gray, layer plan.Layer) error {
	if layer.Patch == nil || layer.Slot.Empty() {
		return errors.New("hud: layer not patchable")
	}
	doc, err := layer.Patch()
	if err != nil {
		return err
	}
	patch, err := bitmap.Rasterize(doc)
	if err != nil {
		return err
	}
	return blitGray(img, layer.Slot, patch)
}

func blitGray(img *image.Gray, slot image.Rectangle, patch *image.Gray) error {
	if img == nil || patch == nil || slot.Empty() {
		return errors.New("hud: bad patch target")
	}
	w, h := slot.Dx(), slot.Dy()
	b := img.Bounds()
	pb := patch.Bounds()
	for y := 0; y < h; y++ {
		dy := slot.Min.Y + y
		if dy < b.Min.Y || dy >= b.Max.Y {
			continue
		}
		sy := pb.Min.Y + y
		if sy >= pb.Max.Y {
			continue
		}
		for x := 0; x < w; x++ {
			dx := slot.Min.X + x
			if dx < b.Min.X || dx >= b.Max.X {
				continue
			}
			sx := pb.Min.X + x
			if sx >= pb.Max.X {
				continue
			}
			img.SetGray(dx, dy, patch.GrayAt(sx, sy))
		}
	}
	return nil
}
