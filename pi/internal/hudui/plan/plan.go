package plan

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/scene"
)

// Layer is one refresh-tracked region produced by a screen template/layout.
type Layer struct {
	ID    hudui.NodeID
	Tier  hudui.Tier
	Key   hudui.ChangeKey
	Slot  image.Rectangle // absolute canvas coordinates (0..250 × 0..122)
	Patch func() (scene.Document, error)
}

// ScreenPlan is the output of a screen compose pass (layout + SVG body + refresh metadata).
type ScreenPlan struct {
	BodySVG    string // main-column fragment (inside chrome translate)
	Descriptors []hudui.Descriptor
	Layers     []Layer
}

// LayerByID finds a layer for spatial patching.
func (p *ScreenPlan) LayerByID(id hudui.NodeID) (Layer, bool) {
	for _, l := range p.Layers {
		if l.ID == id {
			return l, true
		}
	}
	return Layer{}, false
}

// BuildDescriptors returns orchestrator-ready descriptors (includes chrome/screen keys).
func BuildDescriptors(screenKey, chromeKey hudui.ChangeKey, layers []Layer) []hudui.Descriptor {
	out := make([]hudui.Descriptor, 0, len(layers)+2)
	out = append(out,
		hudui.Descriptor{ID: hudui.NodeScreen, Tier: hudui.TierStatic, Key: screenKey},
		hudui.Descriptor{ID: hudui.NodeChrome, Tier: hudui.TierStatic, Key: chromeKey},
	)
	for _, l := range layers {
		out = append(out, hudui.Descriptor{
			ID: l.ID, Tier: l.Tier, Key: l.Key, Slot: l.Slot,
		})
	}
	return out
}
