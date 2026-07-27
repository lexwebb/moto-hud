package hud

import (
	"fmt"
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
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="100%%" height="100%%" fill="#fff"/>%s</svg>`,
		w, h, w, h, body)
	out, err := pixelfont.ReplaceSVGText(svg)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// PatchDistance blits a rerasterized distance readout into an existing frame.
func PatchDistance(img *image.Gray, nav protocol.NavMessage, slot image.Rectangle) error {
	if img == nil || slot.Empty() {
		return fmt.Errorf("hud: bad patch target")
	}
	w, h := slot.Dx(), slot.Dy()
	svg, err := DistancePatchSVG(nav, w, h)
	if err != nil {
		return err
	}
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
