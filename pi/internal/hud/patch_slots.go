package hud

import (
	"fmt"
	"image"
	"strings"

	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
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

func patchSVG(w, h int, body string) ([]byte, error) {
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

// PatchETA blits the ETA line into a fixed slot.
func PatchETA(img *image.Gray, nav protocol.NavMessage, slot image.Rectangle) error {
	if nav.EtaMin <= 0 {
		return nil
	}
	body := mustFace(pixelfont.Size8x16)
	eta := formatETA(nav.EtaMin)
	eta = fit(body, eta, slot.Dx())
	line := textSVG("eta", body, 0, body.Metrics.Ascent, "start", eta)
	svg, err := patchSVG(slot.Dx(), slot.Dy(), line)
	if err != nil {
		return err
	}
	return blitPatch(img, slot, svg)
}

// PatchRoad blits wrapped road text into a fixed slot (max 2 lines).
func PatchRoad(img *image.Gray, nav protocol.NavMessage, slot image.Rectangle) error {
	body := mustFace(pixelfont.Size8x16)
	road := nav.Road
	if road == "" {
		road = nav.Instruction
	}
	road = abbreviateRoad(road)
	lines := wrapLines(body, road, slot.Dx(), 2)
	var b strings.Builder
	for i, ln := range lines {
		y := i * body.Metrics.CellH
		b.WriteString(textSVG("", body, 0, y+body.Metrics.Ascent, "start", ln))
	}
	svg, err := patchSVG(slot.Dx(), slot.Dy(), b.String())
	if err != nil {
		return err
	}
	return blitPatch(img, slot, svg)
}

// PatchMediaTitle patches the media title line.
func PatchMediaTitle(img *image.Gray, media protocol.MediaMessage, slot image.Rectangle) error {
	titleFace := mustFace(pixelfont.Size12x24)
	title := media.Title
	if title == "" || title == "-" {
		title = "No track"
	}
	title = fit(titleFace, title, slot.Dx())
	line := textSVG("title", titleFace, 0, titleFace.Metrics.Ascent, "start", title)
	svg, err := patchSVG(slot.Dx(), slot.Dy(), line)
	if err != nil {
		return err
	}
	return blitPatch(img, slot, svg)
}

// PatchMediaArtist patches the artist line.
func PatchMediaArtist(img *image.Gray, media protocol.MediaMessage, slot image.Rectangle) error {
	body := mustFace(pixelfont.Size8x16)
	artist := fit(body, media.Artist, slot.Dx())
	line := textSVG("artist", body, 0, body.Metrics.Ascent, "start", artist)
	svg, err := patchSVG(slot.Dx(), slot.Dy(), line)
	if err != nil {
		return err
	}
	return blitPatch(img, slot, svg)
}
