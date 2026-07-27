package compose

import (
	"image"

	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

func patchDistanceDoc(nav protocol.NavMessage, w, h int, deps DrawDeps) (scene.Document, error) {
	dist := nav.DistanceText
	if dist == "" {
		dist = formatDistance(nav.DistanceM)
	}
	dist = compactDistanceText(dist)
	dist = deps.Fit(scene.Face16x32, dist, w)
	hero := svg.MustLoadFace(scene.Face16x32)
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text("distance", scene.Face16x32, w, hero.Metrics.Ascent+2, "end", dist)
	})
	return doc, nil
}

func patchETADoc(nav protocol.NavMessage, w, h int, deps DrawDeps) (scene.Document, error) {
	if nav.EtaMin <= 0 {
		return scene.Patch(w, h, nil), nil
	}
	eta := deps.Fit(scene.Face8x16, formatETA(nav.EtaMin), w)
	body := svg.MustLoadFace(scene.Face8x16)
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text("eta", scene.Face8x16, 0, body.Metrics.Ascent, "start", eta)
	})
	return doc, nil
}

func patchRoadDoc(nav protocol.NavMessage, slot image.Rectangle, deps DrawDeps) (scene.Document, error) {
	road := nav.Road
	if road == "" {
		road = nav.Instruction
	}
	body := svg.MustLoadFace(scene.Face8x16)
	maxLines := slot.Dy() / body.Metrics.CellH
	if maxLines < 1 {
		maxLines = 1
	}
	if maxLines > 3 {
		maxLines = 3
	}
	lines := deps.WrapRoad(road, slot.Dx(), maxLines)
	w, h := slot.Dx(), slot.Dy()
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		for i, ln := range lines {
			b.Text("road", scene.Face8x16, 0, i*body.Metrics.CellH+body.Metrics.Ascent, "start", ln)
		}
	})
	return doc, nil
}

func patchMediaTitleDoc(title string, w, h int, deps mediaPatchDeps) (scene.Document, error) {
	title = deps.FitTitle(title, w)
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text("title", scene.Face12x24, 0, deps.TitleBaseline(), "start", title)
	})
	return doc, nil
}

func patchMediaArtistDoc(artist string, w, h int, deps mediaPatchDeps) (scene.Document, error) {
	artist = deps.FitBody(artist, w)
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text("artist", scene.Face8x16, 0, deps.BodyBaseline(), "start", artist)
	})
	return doc, nil
}

func patchLinkDoc(linked bool, w, h int, linkFn func(bool) string) (scene.Document, error) {
	if linkFn == nil {
		linkFn = LinkMarkFragment
	}
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Raw(linkFn(linked))
	})
	return doc, nil
}

func patchStatusValueDoc(id, value string, w, h int, deps DrawDeps) (scene.Document, error) {
	body := svg.MustLoadFace(scene.Face8x16)
	doc := scene.Patch(w, h, func(b *scene.Builder) {
		b.Text(id, scene.Face8x16, w, body.Metrics.Ascent, "end", value)
	})
	return doc, nil
}
