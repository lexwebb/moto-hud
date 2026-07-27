package compose

import (
	"bytes"
	"context"

	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

// NavIdleBodySVG renders the inactive-nav main column via templ (ADR 0009).
func NavIdleBodySVG() (string, error) {
	meta, err := pixelfont.Load(pixelfont.Size6x12)
	if err != nil {
		return "", err
	}
	body, err := pixelfont.Load(pixelfont.Size8x16)
	if err != nil {
		return "", err
	}
	mw := token.MainWidth()
	pad := token.Pad
	headerBottom := pad + meta.Metrics.CellH
	divY := headerBottom + token.GapSm
	contentTop := divY + token.GapMd
	contentBot := token.Height - pad
	blockH := body.Metrics.CellH*2 + token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	b1 := top + body.Metrics.Ascent
	b2 := top + body.Metrics.CellH + token.GapMd + body.Metrics.Ascent

	var buf bytes.Buffer
	if err := screens.NavIdleCentered("MOTO HUD", "Waiting for route...", mw, b1, b2).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
