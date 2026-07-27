package compose

import (
	"bytes"
	"context"
	"fmt"

	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

type chromeHints struct {
	mode, legPrev, legAction, legNext string
}

func chromeFor(in Input) chromeHints {
	switch in.Screen {
	case ScreenMedia:
		action := "PLAY"
		if in.Media.Playing {
			action = "PAUSE"
		}
		return chromeHints{mode: "MEDIA", legPrev: "SKIP", legAction: action, legNext: "SKIP"}
	case ScreenStatus:
		return chromeHints{mode: "STATUS", legPrev: "MODE", legAction: "REDRAW", legNext: "MODE"}
	default:
		legPrev, legAction, legNext := "MEDIA", "-", "STATUS"
		if in.Nav.Active {
			legPrev, legAction, legNext = "MODE", "-", "MODE"
		}
		return chromeHints{mode: "NAV", legPrev: legPrev, legAction: legAction, legNext: legNext}
	}
}

func chromeGeom() (pad, mw, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY int) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	pad = token.Pad
	mw = token.MainWidth()
	headerBaseline = pad + meta.Metrics.Ascent
	headerBottom := pad + meta.Metrics.CellH
	divY = headerBottom + token.GapSm
	ruleX = pad + mw + token.GapMd
	legTop = headerBaseline
	legMid = token.Height / 2
	legBot = token.Height - pad - meta.Metrics.Descent
	linkX = mw - 16
	linkY = pad + 1
	return
}

// ChromeBodySVG wraps main-column SVG in the shared chrome shell (templ).
func ChromeBodySVG(mode, linkSVG, content, legPrev, legAction, legNext string) (string, error) {
	pad, mw, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY := chromeGeom()
	var buf bytes.Buffer
	err := screens.ChromeShell(
		mode, legPrev, legAction, legNext,
		pad, mw, divY, ruleX, headerBaseline, legTop, legMid, legBot, linkX, linkY,
		token.Width, token.Height,
		linkSVG, content,
	).Render(context.Background(), &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// FrameVars returns frame.svg template vars for a composed screen plan.
func FrameVars(in Input, sp plan.ScreenPlan) (map[string]string, error) {
	link := ""
	if in.LinkSVG != nil {
		link = in.LinkSVG(in.Linked)
	}
	ch := chromeFor(in)
	body, err := ChromeBodySVG(ch.mode, link, sp.BodySVG, ch.legPrev, ch.legAction, ch.legNext)
	if err != nil {
		return nil, fmt.Errorf("compose: chrome: %w", err)
	}
	return map[string]string{"body": body}, nil
}
