package hud

import (
	"image"

	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/protocol"
)

// NavSlotSet holds conservative bounding boxes for classic nav layout.
type NavSlotSet struct {
	Maneuver image.Rectangle
	Distance image.Rectangle
	Road     image.Rectangle
	ETA      image.Rectangle
	Ribbon   image.Rectangle
}

// NavRefreshSlots returns refresh regions for the active nav screen.
func NavRefreshSlots(nav protocol.NavMessage) NavSlotSet {
	if !nav.Active {
		return idleRefreshSlots()
	}
	if HasMinimap(nav) || len(nav.RibbonPoints) >= 2 {
		return NavSlotSet{}
	}
	return classicNavRefreshSlots()
}

func idleRefreshSlots() NavSlotSet {
	mw := token.MainWidth()
	mainX := token.Pad
	top := token.Pad + 12 + token.GapSm + token.GapMd + 20
	return NavSlotSet{
		Road: image.Rect(mainX, top, mainX+mw, top+40),
	}
}

func classicNavRefreshSlots() NavSlotSet {
	mw := token.MainWidth()
	mainX := token.Pad
	contentTop := token.Pad + 12 + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	heroH := token.GlyphSz
	if heroH < 32 {
		heroH = 32
	}
	heroTop := contentTop
	heroBot := contentTop + heroH + (contentBot-contentTop)/6
	if heroBot > contentBot {
		heroBot = contentBot
	}
	ribbonH := 40
	if ribbonH > contentBot-contentTop-heroH {
		ribbonH = contentBot - contentTop - heroH
	}
	ribbonTop := contentBot - ribbonH

	return NavSlotSet{
		Maneuver: image.Rect(mainX-2, heroTop, mainX+token.GlyphSz, heroTop+token.GlyphSz),
		Distance: image.Rect(mainX+token.GlyphSz, heroTop, mainX+mw, heroBot),
		Road:     image.Rect(mainX, heroBot+token.GapLg, mainX+mw, ribbonTop-token.GapLg),
		ETA:      image.Rect(mainX, ribbonTop-token.GapLg-16, mainX+mw, ribbonTop),
		Ribbon:   image.Rect(mainX, ribbonTop, mainX+mw, contentBot),
	}
}
