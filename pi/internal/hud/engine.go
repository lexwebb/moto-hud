package hud

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/protocol"
)

// FrameResult is one composed framebuffer plus refresh metadata (ADR 0010).
type FrameResult struct {
	Image   *image.Gray
	Spatial bool
	Dirty   image.Rectangle
	Patched bool
}

// Engine composes HUD frames with tier-aware partial updates.
type Engine struct {
	Orch  *RefreshOrchestrator
	frame *image.Gray
}

func NewEngine() *Engine {
	return &Engine{Orch: NewRefreshOrchestrator()}
}

// Draw renders or patches the panel for the current state.
func (e *Engine) Draw(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked, force bool) FrameResult {
	plan := e.Orch.Plan(screen, nav, media, linked, force)
	if plan.Mode == RefreshNone {
		if e.frame != nil {
			return FrameResult{Image: e.frame}
		}
		img := Render(screen, nav, media, linked)
		e.frame = cloneGray(img)
		return FrameResult{Image: e.frame}
	}

	if plan.Mode == RefreshSpatialPatch && e.frame != nil {
		if e.applySpatialPatches(screen, plan, nav, media) {
			return FrameResult{
				Image:   e.frame,
				Spatial: true,
				Dirty:   plan.DirtyUnion,
				Patched: true,
			}
		}
	}

	img := Render(screen, nav, media, linked)
	e.frame = cloneGray(img)
	return FrameResult{
		Image:   e.frame,
		Spatial: plan.Mode == RefreshSpatialPatch,
		Dirty:   plan.DirtyUnion,
	}
}

func (e *Engine) applySpatialPatches(screen Screen, plan RefreshPlan, nav protocol.NavMessage, media protocol.MediaMessage) bool {
	for _, id := range plan.DirtyIDs {
		if !isPatchableNode(screen, id) {
			return false
		}
	}
	switch screen {
	case ScreenNav:
		slots := NavRefreshSlots(nav)
		for _, id := range plan.DirtyIDs {
			var err error
			switch id {
			case hudui.NodeDistance:
				err = PatchDistance(e.frame, nav, slots.Distance)
			case hudui.NodeETA:
				err = PatchETA(e.frame, nav, slots.ETA)
			case hudui.NodeRoad:
				err = PatchRoad(e.frame, nav, slots.Road)
			default:
				return false
			}
			if err != nil {
				return false
			}
		}
		return true
	case ScreenMedia:
		slots := MediaRefreshSlots()
		for _, id := range plan.DirtyIDs {
			var err error
			switch id {
			case hudui.NodeMediaTitle:
				err = PatchMediaTitle(e.frame, media, slots.Title)
			case hudui.NodeMediaArtist:
				err = PatchMediaArtist(e.frame, media, slots.Artist)
			default:
				return false
			}
			if err != nil {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isPatchableNode(screen Screen, id hudui.NodeID) bool {
	switch screen {
	case ScreenNav:
		switch id {
		case hudui.NodeDistance, hudui.NodeETA, hudui.NodeRoad:
			return true
		}
	case ScreenMedia:
		switch id {
		case hudui.NodeMediaTitle, hudui.NodeMediaArtist:
			return true
		}
	}
	return false
}

func cloneGray(src *image.Gray) *image.Gray {
	if src == nil {
		return nil
	}
	dst := image.NewGray(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

// RenderEngine is the package-level compositor used by motohud and WASM.
var RenderEngine = NewEngine()

// RenderWithEngine draws via the shared Engine (retains framebuffer for patches).
func RenderWithEngine(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked, force bool) FrameResult {
	return RenderEngine.Draw(screen, nav, media, linked, force)
}
