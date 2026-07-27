package hud

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/hudui/layout"
	"moto-hud/pi/internal/hudui/plan"
)

// RefreshOrchestrator tracks per-node change keys and chooses update mode.
type RefreshOrchestrator struct {
	Policy RefreshPolicy
	last   map[hudui.NodeID]hudui.ChangeKey
	slots  map[hudui.NodeID]image.Rectangle
	screen compose.ScreenKind
}

func NewRefreshOrchestrator() *RefreshOrchestrator {
	return &RefreshOrchestrator{
		Policy: DefaultRefreshPolicy(),
		last:   make(map[hudui.NodeID]hudui.ChangeKey),
		slots:  make(map[hudui.NodeID]image.Rectangle),
	}
}

// Plan compares template-produced descriptors to the previous frame.
func (o *RefreshOrchestrator) Plan(screen compose.ScreenKind, descs []hudui.Descriptor, force bool) RefreshPlan {
	if force || screen != o.screen {
		o.screen = screen
		o.rememberAll(descs)
		return RefreshPlan{Mode: RefreshFullFrame, FullRender: true, DirtyIDs: allNodeIDs(descs)}
	}

	var dirty []hudui.NodeID
	var rects []image.Rectangle
	allPartialOK := true
	for _, d := range descs {
		prev, ok := o.last[d.ID]
		if !ok || prev != d.Key {
			dirty = append(dirty, d.ID)
			if d.Tier != hudui.TierPartialOK {
				allPartialOK = false
			}
			if !d.Slot.Empty() {
				rects = append(rects, d.Slot)
			}
		}
		prevSlot, hadSlot := o.slots[d.ID]
		if hadSlot && d.Tier == hudui.TierPartialOK && prevSlot != d.Slot {
			allPartialOK = false
		}
	}

	if len(dirty) == 0 {
		return RefreshPlan{Mode: RefreshNone}
	}

	o.rememberAll(descs)

	union := layout.AlignEPD(layout.UnionRects(rects))
	area := union.Dx() * union.Dy()
	if !allPartialOK || union.Empty() || area > o.Policy.MaxPartialPixels {
		return RefreshPlan{
			Mode:       RefreshFullFrame,
			FullRender: true,
			DirtyIDs:   dirty,
			DirtyUnion: union,
		}
	}

	return RefreshPlan{
		Mode:       RefreshSpatialPatch,
		FullRender: false,
		DirtyIDs:   dirty,
		DirtyUnion: union,
	}
}

func (o *RefreshOrchestrator) rememberAll(descs []hudui.Descriptor) {
	for _, d := range descs {
		o.last[d.ID] = d.Key
		if !d.Slot.Empty() {
			o.slots[d.ID] = d.Slot
		}
	}
}

func allNodeIDs(descs []hudui.Descriptor) []hudui.NodeID {
	out := make([]hudui.NodeID, len(descs))
	for i, d := range descs {
		out[i] = d.ID
	}
	return out
}

// PlanFromCompose builds a screen plan and runs refresh planning in one step.
func (o *RefreshOrchestrator) PlanFromCompose(in compose.Input, force bool) (RefreshPlan, plan.ScreenPlan, error) {
	sp, err := compose.BuildPlan(in)
	if err != nil {
		return RefreshPlan{}, plan.ScreenPlan{}, err
	}
	rp := o.Plan(in.Screen, sp.Descriptors, force)
	return rp, sp, nil
}
