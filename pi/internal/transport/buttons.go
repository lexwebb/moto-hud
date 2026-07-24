package transport

import (
	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
)

// HandleButtonEvent applies a physical/virtual button event.
// On Media: short Prev/Next skip tracks; long Prev/Next change screen.
// Elsewhere: Prev/Next change screen. Action = context; ActionLong = Nav home.
func (h *Hub) HandleButtonEvent(ev buttons.Event) {
	screen := h.State.CurrentScreen()
	switch ev {
	case buttons.Prev:
		if screen == hud.ScreenMedia {
			h.PublishCmd(protocol.CmdPrevTrack)
			return
		}
		h.State.PrevScreen()
	case buttons.Next:
		if screen == hud.ScreenMedia {
			h.PublishCmd(protocol.CmdNextTrack)
			return
		}
		h.State.NextScreen()
	case buttons.PrevLong:
		h.State.PrevScreen()
	case buttons.NextLong:
		h.State.NextScreen()
	case buttons.Action:
		h.handleAction()
	case buttons.ActionLong:
		h.State.HomeNav()
	}
	h.notifyChange()
}

// HandleMediaSkip forces a skip command (HTTP helper).
func (h *Hub) HandleMediaSkip(next bool) {
	if next {
		h.PublishCmd(protocol.CmdNextTrack)
	} else {
		h.PublishCmd(protocol.CmdPrevTrack)
	}
}
