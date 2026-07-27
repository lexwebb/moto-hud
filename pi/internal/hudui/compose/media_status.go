package compose

import (
	"bytes"
	"context"

	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

func mediaLayout() (mw, yPlaying, yTitle, yArtist int) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	title, _ := pixelfont.Load(pixelfont.Size12x24)
	mw = token.MainWidth()
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	blockH := meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.CellH
	top := contentTop + (contentBot-contentTop-blockH)/2
	yPlaying = top + meta.Metrics.Ascent
	yTitle = top + meta.Metrics.CellH + token.GapSm + title.Metrics.Ascent
	yArtist = top + meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.Ascent
	return
}

// MediaBodySVG renders the media main column via templ (strings should already be fitted).
func MediaBodySVG(playing, title, artist string) (string, error) {
	mw, y1, y2, y3 := mediaLayout()
	var buf bytes.Buffer
	if err := screens.MediaBody(playing, title, artist, mw, y1, y2, y3).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MediaBodySVGFromMessage fits fields then renders.
func MediaBodySVGFromMessage(media protocol.MediaMessage, fit func(face *pixelfont.Face, s string, maxW int) string) (string, error) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	titleFace, _ := pixelfont.Load(pixelfont.Size12x24)
	mw := token.MainWidth()
	playing := "PAUSED"
	if media.Playing {
		playing = "PLAYING"
	}
	title := media.Title
	if title == "" || title == "-" {
		title = "No track"
	}
	if fit != nil {
		playing = fit(meta, playing, mw)
		title = fit(titleFace, title, mw)
		artist := fit(body, media.Artist, mw)
		return MediaBodySVG(playing, title, artist)
	}
	return MediaBodySVG(playing, title, media.Artist)
}

func statusLayout() (mw, y1, y2, y3 int) {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	mw = token.MainWidth()
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	rowH := body.Metrics.CellH + token.GapMd
	rows := 3
	blockH := rowH*rows - token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	y1 = top + body.Metrics.Ascent
	y2 = top + rowH + body.Metrics.Ascent
	y3 = top + 2*rowH + body.Metrics.Ascent
	return
}

// StatusBodySVG renders the status rows via templ.
func StatusBodySVG(bleLinked, navActive bool) (string, error) {
	ble, nav := "DOWN", "OFF"
	if bleLinked {
		ble = "UP"
	}
	if navActive {
		nav = "ON"
	}
	mw, y1, y2, y3 := statusLayout()
	var buf bytes.Buffer
	if err := screens.StatusRows(ble, nav, "OK", mw, y1, y2, y3).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
