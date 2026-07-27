package hud

import (
	"image"

	"moto-hud/pi/internal/hudui"
)

// RefreshPolicy tunes spatial partial decisions (ADR 0010).
type RefreshPolicy struct {
	MaxPartialPixels int
}

func DefaultRefreshPolicy() RefreshPolicy {
	const canvas = Width * Height
	return RefreshPolicy{MaxPartialPixels: canvas * 35 / 100}
}

// RefreshMode is how the display should be updated this frame.
type RefreshMode int

const (
	RefreshNone RefreshMode = iota
	RefreshSpatialPatch
	RefreshFullFrame
)

// RefreshPlan is the orchestrator output for one draw.
type RefreshPlan struct {
	Mode       RefreshMode
	DirtyIDs   []hudui.NodeID
	DirtyUnion image.Rectangle
	FullRender bool
}
