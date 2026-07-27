package compose

import (
	"fmt"
	"image"

	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

func patchSVG(w, h int, body string) ([]byte, error) {
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="100%%" height="100%%" fill="#fff"/>%s</svg>`,
		w, h, w, h, body)
	return []byte(svg), nil
}

func patchDistanceSVG(nav protocol.NavMessage, w, h int, deps NavSVGDeps) ([]byte, error) {
	dist := nav.DistanceText
	if dist == "" {
		dist = formatDistance(nav.DistanceM)
	}
	dist = compactDistanceText(dist)
	dist = deps.Fit("16x32", dist, w)
	hero, _ := pixelfont.Load(pixelfont.Size16x32)
	baseline := hero.Metrics.Ascent + 2
	body := deps.TextSVG("distance", "16x32", w, baseline, "end", dist)
	raw, err := patchSVG(w, h, body)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func patchETASVG(nav protocol.NavMessage, w, h int, deps NavSVGDeps) ([]byte, error) {
	if nav.EtaMin <= 0 {
		return patchSVG(w, h, "")
	}
	eta := deps.Fit("8x16", formatETA(nav.EtaMin), w)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	line := deps.TextSVG("eta", "8x16", 0, body.Metrics.Ascent, "start", eta)
	raw, err := patchSVG(w, h, line)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func patchRoadSVG(nav protocol.NavMessage, slot image.Rectangle, deps NavSVGDeps) ([]byte, error) {
	road := nav.Road
	if road == "" {
		road = nav.Instruction
	}
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	maxLines := slot.Dy() / body.Metrics.CellH
	if maxLines < 1 {
		maxLines = 1
	}
	if maxLines > 3 {
		maxLines = 3
	}
	lines := deps.WrapRoad(road, slot.Dx(), maxLines)
	var frag string
	for i, ln := range lines {
		frag += deps.TextSVG("road", "8x16", 0, i*body.Metrics.CellH+body.Metrics.Ascent, "start", ln)
	}
	raw, err := patchSVG(slot.Dx(), slot.Dy(), frag)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func patchMediaTitleSVG(title string, w, h int, deps mediaPatchDeps) ([]byte, error) {
	title = deps.FitTitle(title, w)
	line := deps.TextSVG("title", "12x24", 0, deps.TitleBaseline(), "start", title)
	raw, err := patchSVG(w, h, line)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func patchMediaArtistSVG(artist string, w, h int, deps mediaPatchDeps) ([]byte, error) {
	artist = deps.FitBody(artist, w)
	line := deps.TextSVG("artist", "8x16", 0, deps.BodyBaseline(), "start", artist)
	raw, err := patchSVG(w, h, line)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func patchLinkSVG(linked bool, w, h int, linkFn func(bool) string) ([]byte, error) {
	if linkFn == nil {
		linkFn = LinkMarkFragment
	}
	return patchSVG(w, h, linkFn(linked))
}

func patchStatusValueSVG(id, value string, w, h int, deps NavSVGDeps) ([]byte, error) {
	if deps.TextSVG == nil {
		return patchSVG(w, h, "")
	}
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	line := deps.TextSVG(id, "8x16", w, body.Metrics.Ascent, "end", value)
	raw, err := patchSVG(w, h, line)
	if err != nil {
		return nil, err
	}
	return pixelfontReplace(raw)
}

func pixelfontReplace(svg []byte) ([]byte, error) {
	out, err := pixelfont.ReplaceSVGText(string(svg))
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
