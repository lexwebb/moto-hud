package compose

import (
	"bytes"
	"context"

	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

func mediaBodyLayout() (mw, yPlaying, yTitle, yArtist int) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	title, _ := pixelfont.Load(pixelfont.Size12x24)
	mw = token.MainWidth()
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	blockH := meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.CellH
	top := contentTop + (contentBot-contentTop-blockH)/2
	return mw, top + meta.Metrics.Ascent,
		top + meta.Metrics.CellH + token.GapSm + title.Metrics.Ascent,
		top + meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.Ascent
}

// MediaBodySVG renders the media main column via templ.
func MediaBodySVG(playing, title, artist string) (string, error) {
	mw, y1, y2, y3 := mediaBodyLayout()
	var buf bytes.Buffer
	if err := screens.MediaBody(playing, title, artist, mw, y1, y2, y3).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func statusBodyLayout() (mw, y1, y2, y3 int) {
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
	return mw, top + body.Metrics.Ascent, top + rowH + body.Metrics.Ascent, top + 2*rowH + body.Metrics.Ascent
}

// StatusBodySVG renders status rows via templ.
func StatusBodySVG(bleLinked, navActive bool) (string, error) {
	ble, nav := "DOWN", "OFF"
	if bleLinked {
		ble = "UP"
	}
	if navActive {
		nav = "ON"
	}
	mw, y1, y2, y3 := statusBodyLayout()
	var buf bytes.Buffer
	if err := screens.StatusRows(ble, nav, "OK", mw, y1, y2, y3).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
