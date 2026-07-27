package hud

import (
	"image"

	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

// DistancePatchSVG is a small SVG for patching the hero distance slot (white bg + text).
func DistancePatchSVG(nav protocol.NavMessage, w, h int) ([]byte, error) {
	hero := mustFace(pixelfont.Size16x32)
	dist := nav.DistanceText
	if dist == "" {
		dist = formatDistance(nav.DistanceM)
	}
	dist = fit(hero, dist, w)
	baseline := hero.Metrics.Ascent + 2
	body := textSVG("distance", hero, w, baseline, "end", dist)
	svg, err := patchSVG(w, h, body)
	if err != nil {
		return nil, err
	}
	return svg, nil
}

// PatchDistance blits a rerasterized distance readout into an existing frame.
func PatchDistance(img *image.Gray, nav protocol.NavMessage, slot image.Rectangle) error {
	svg, err := DistancePatchSVG(nav, slot.Dx(), slot.Dy())
	if err != nil {
		return err
	}
	return blitPatch(img, slot, svg)
}
