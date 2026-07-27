package hud

import (
	"image"

	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/hudui/plan"
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
	in := ComposeInput(screen, nav, media, linked)
	rp, sp, err := e.Orch.PlanFromCompose(in, force)
	if err != nil {
		img := Render(screen, nav, media, linked)
		e.frame = cloneGray(img)
		return FrameResult{Image: e.frame}
	}

	if rp.Mode == RefreshNone {
		if e.frame != nil {
			return FrameResult{Image: e.frame}
		}
		img := renderFromPlan(in, sp)
		e.frame = cloneGray(img)
		return FrameResult{Image: e.frame}
	}

	if rp.Mode == RefreshSpatialPatch && e.frame != nil {
		ok := true
		for _, id := range rp.DirtyIDs {
			layer, found := sp.LayerByID(id)
			if !found || layer.Patch == nil {
				ok = false
				break
			}
			if err := PatchLayer(e.frame, layer); err != nil {
				ok = false
				break
			}
		}
		if ok {
			return FrameResult{
				Image:   e.frame,
				Spatial: true,
				Dirty:   rp.DirtyUnion,
				Patched: true,
			}
		}
	}

	img := renderFromPlan(in, sp)
	e.frame = cloneGray(img)
	return FrameResult{
		Image:   e.frame,
		Spatial: rp.Mode == RefreshSpatialPatch,
		Dirty:   rp.DirtyUnion,
	}
}

func renderFromPlan(in compose.Input, sp plan.ScreenPlan) *image.Gray {
	vars, err := compose.FrameVars(in, sp)
	if err != nil {
		return Render(composeScreen(in.Screen), in.Nav, in.Media, in.Linked)
	}
	svg, err := BuildPixelSVGFromVars(vars)
	if err != nil {
		return Render(composeScreen(in.Screen), in.Nav, in.Media, in.Linked)
	}
	img, err := RasterizeSVG(svg)
	if err != nil {
		return Render(composeScreen(in.Screen), in.Nav, in.Media, in.Linked)
	}
	return img
}

func composeScreen(k compose.ScreenKind) Screen {
	switch k {
	case compose.ScreenMedia:
		return ScreenMedia
	case compose.ScreenStatus:
		return ScreenStatus
	default:
		return ScreenNav
	}
}

// RenderEngine is the package-level compositor used by motohud and WASM.
var RenderEngine = NewEngine()

// RenderWithEngine draws via the shared Engine (retains framebuffer for patches).
func RenderWithEngine(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked, force bool) FrameResult {
	return RenderEngine.Draw(screen, nav, media, linked, force)
}

func cloneGray(src *image.Gray) *image.Gray {
	if src == nil {
		return nil
	}
	dst := image.NewGray(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}
