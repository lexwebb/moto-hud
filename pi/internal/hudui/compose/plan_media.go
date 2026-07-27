package compose

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

func planMedia(in Input) (plan.ScreenPlan, error) {
	geom := mediaLayoutGeom()
	playing := "PAUSED"
	if in.Media.Playing {
		playing = "PLAYING"
	}
	title := in.Media.Title
	if title == "" || title == "-" {
		title = "No track"
	}
	if in.NavSVG.Fit != nil {
		title = in.NavSVG.Fit(scene.Face12x24, title, geom.mw)
	}
	body := scenetempl.Render(screens.MediaBody(playing, title, in.Media.Artist, geom.mw, geom.yPlaying, geom.yTitle, geom.yArtist))
	k := Keys{}
	deps := in.NavSVG
	media := in.Media
	titleCopy := title
	artistCopy := media.Artist
	layers := []plan.Layer{
		{ID: hudui.NodeMediaState, Tier: hudui.TierSlow, Key: k.MediaPlaying(media), Slot: geom.playingSlot},
		{
			ID: hudui.NodeMediaTitle, Tier: hudui.TierPartialOK, Key: k.Hash(media.Title), Slot: geom.titleSlot,
			Patch: func() (scene.Document, error) {
				return patchMediaTitleDoc(titleCopy, geom.titleSlot.Dx(), geom.titleSlot.Dy(), mediaPatchDeps{deps: deps, role: "title"})
			},
		},
		{
			ID: hudui.NodeMediaArtist, Tier: hudui.TierPartialOK, Key: k.Hash(media.Artist), Slot: geom.artistSlot,
			Patch: func() (scene.Document, error) {
				return patchMediaArtistDoc(artistCopy, geom.artistSlot.Dx(), geom.artistSlot.Dy(), mediaPatchDeps{deps: deps, role: "artist"})
			},
		},
	}
	return finalizePlan(in, k.MediaScreen(media), staticChromeKey(), body, layers), nil
}

type mediaLayout struct {
	mw                               int
	yPlaying, yTitle, yArtist        int
	playingSlot, titleSlot, artistSlot image.Rectangle
}

func mediaLayoutGeom() mediaLayout {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	title, _ := pixelfont.Load(pixelfont.Size12x24)
	mw := token.MainWidth()
	mainX := token.Pad
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	blockH := meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.CellH
	top := contentTop + (contentBot-contentTop-blockH)/2
	return mediaLayout{
		mw: mw,
		yPlaying: top + meta.Metrics.Ascent,
		yTitle:   top + meta.Metrics.CellH + token.GapSm + title.Metrics.Ascent,
		yArtist:  top + meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.Ascent,
		playingSlot: image.Rect(mainX, top, mainX+mw, top+meta.Metrics.CellH),
		titleSlot:   image.Rect(mainX, top+meta.Metrics.CellH+token.GapSm, mainX+mw, top+meta.Metrics.CellH+token.GapSm+title.Metrics.CellH),
		artistSlot:  image.Rect(mainX, top+meta.Metrics.CellH+token.GapSm+title.Metrics.CellH+token.GapSm, mainX+mw, top+blockH),
	}
}

type mediaPatchDeps struct {
	deps DrawDeps
	role string // title | artist
}

func (m mediaPatchDeps) TextSVG(id, faceSize string, x, baseline int, anchor, s string) string {
	return m.deps.TextSVG(id, faceSize, x, baseline, anchor, s)
}

func (m mediaPatchDeps) FitTitle(s string, maxW int) string {
	if m.deps.Fit != nil {
		return m.deps.Fit(scene.Face12x24, s, maxW)
	}
	return s
}

func (m mediaPatchDeps) FitBody(s string, maxW int) string {
	if m.deps.Fit != nil {
		return m.deps.Fit(scene.Face8x16, s, maxW)
	}
	return s
}

func (m mediaPatchDeps) TitleBaseline() int {
	title, _ := pixelfont.Load(pixelfont.Size12x24)
	return title.Metrics.Ascent
}

func (m mediaPatchDeps) BodyBaseline() int {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	return body.Metrics.Ascent
}
