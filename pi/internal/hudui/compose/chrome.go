package compose

import (
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
