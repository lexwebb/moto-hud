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

	if plan.Mode == RefreshSpatialPatch && e.frame != nil && patchOnlyDistance(plan.DirtyIDs) {
		slots := NavRefreshSlots(nav)
		if err := PatchDistance(e.frame, nav, slots.Distance); err == nil {
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

func patchOnlyDistance(ids []hudui.NodeID) bool {
	if len(ids) != 1 {
		return false
	}
	return ids[0] == hudui.NodeDistance
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
