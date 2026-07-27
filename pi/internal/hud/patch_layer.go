package hud

import (
	"errors"
	"image"

	"moto-hud/pi/internal/hudui/plan"
)

// PatchLayer rasterizes a plan layer patch into the framebuffer at its template-defined slot.
func PatchLayer(img *image.Gray, layer plan.Layer) error {
	if layer.Patch == nil || layer.Slot.Empty() {
		return errors.New("hud: layer not patchable")
	}
	svg, err := layer.Patch()
	if err != nil {
		return err
	}
	return blitPatch(img, layer.Slot, svg)
}
