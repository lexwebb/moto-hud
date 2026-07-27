package hud

import (
	"image"

	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

// MediaSlotSet is the fixed media screen layout (centered block).
type MediaSlotSet struct {
	Playing image.Rectangle
	Title   image.Rectangle
	Artist  image.Rectangle
}

// MediaRefreshSlots returns patch regions for the media screen.
func MediaRefreshSlots() MediaSlotSet {
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
	return MediaSlotSet{
		Playing: image.Rect(mainX, top, mainX+mw, top+meta.Metrics.CellH),
		Title:   image.Rect(mainX, top+meta.Metrics.CellH+token.GapSm, mainX+mw, top+meta.Metrics.CellH+token.GapSm+title.Metrics.CellH),
		Artist:  image.Rect(mainX, top+meta.Metrics.CellH+token.GapSm+title.Metrics.CellH+token.GapSm, mainX+mw, top+blockH),
	}
}

// StatusSlotSet is the diagnostics value column (right-aligned).
type StatusSlotSet struct {
	Values image.Rectangle
}

// StatusRefreshSlots returns the patchable value column on the status screen.
func StatusRefreshSlots() StatusSlotSet {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	mw := token.MainWidth()
	mainX := token.Pad
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	rowH := body.Metrics.CellH + token.GapMd
	rows := 3
	blockH := rowH*rows - token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	return StatusSlotSet{
		Values: image.Rect(mainX+mw/2, top, mainX+mw, top+blockH),
	}
}
