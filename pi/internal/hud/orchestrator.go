package hud

import (
	"hash/maphash"
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/layout"
	"moto-hud/pi/internal/protocol"
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

// RefreshOrchestrator tracks per-node change keys and chooses update mode.
type RefreshOrchestrator struct {
	Policy RefreshPolicy
	last   map[hudui.NodeID]hudui.ChangeKey
	slots  map[hudui.NodeID]image.Rectangle
	screen Screen
}

func NewRefreshOrchestrator() *RefreshOrchestrator {
	return &RefreshOrchestrator{
		Policy: DefaultRefreshPolicy(),
		last:   make(map[hudui.NodeID]hudui.ChangeKey),
		slots:  make(map[hudui.NodeID]image.Rectangle),
	}
}

// Plan compares incoming state to last frame and returns draw mode + dirty nodes.
func (o *RefreshOrchestrator) Plan(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked bool, force bool) RefreshPlan {
	descs := refreshDescriptors(screen, nav, media, linked)
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
		if hadSlot && d.ID == hudui.NodeDistance && prevSlot != d.Slot {
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

func refreshDescriptors(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, linked bool) []hudui.Descriptor {
	switch screen {
	case ScreenNav:
		return navRefreshDescriptors(nav, linked)
	case ScreenMedia:
		return mediaRefreshDescriptors(media, linked)
	case ScreenStatus:
		return statusRefreshDescriptors(nav.Active, linked)
	default:
		return navRefreshDescriptors(nav, linked)
	}
}

func navRefreshDescriptors(nav protocol.NavMessage, linked bool) []hudui.Descriptor {
	slots := NavRefreshSlots(nav)
	return []hudui.Descriptor{
		{ID: hudui.NodeScreen, Tier: hudui.TierStatic, Key: keyScreenNav(nav, linked)},
		{ID: hudui.NodeChrome, Tier: hudui.TierStatic, Key: keyBool(linked)},
		{ID: hudui.NodeManeuver, Tier: hudui.TierSlow, Slot: slots.Maneuver, Key: keyManeuver(nav)},
		{ID: hudui.NodeDistance, Tier: hudui.TierPartialOK, Slot: slots.Distance, Key: keyDistance(nav)},
		{ID: hudui.NodeRoad, Tier: hudui.TierSlow, Slot: slots.Road, Key: keyRoad(nav)},
		{ID: hudui.NodeETA, Tier: hudui.TierFast, Slot: slots.ETA, Key: keyETA(nav)},
		{ID: hudui.NodeRibbon, Tier: hudui.TierSlow, Slot: slots.Ribbon, Key: keyRibbon(nav)},
	}
}

func mediaRefreshDescriptors(media protocol.MediaMessage, linked bool) []hudui.Descriptor {
	return []hudui.Descriptor{
		{ID: hudui.NodeScreen, Tier: hudui.TierStatic, Key: keyMedia(media, linked)},
		{ID: hudui.NodeChrome, Tier: hudui.TierStatic, Key: keyBool(linked)},
	}
}

func statusRefreshDescriptors(navActive, linked bool) []hudui.Descriptor {
	var k hudui.ChangeKey
	if linked {
		k |= 1
	}
	if navActive {
		k |= 2
	}
	return []hudui.Descriptor{
		{ID: hudui.NodeScreen, Tier: hudui.TierStatic, Key: k},
		{ID: hudui.NodeChrome, Tier: hudui.TierStatic, Key: keyBool(linked)},
	}
}

func keyBool(b bool) hudui.ChangeKey {
	if b {
		return 1
	}
	return 0
}

func keyManeuver(nav protocol.NavMessage) hudui.ChangeKey {
	return refreshHashStr(string(nav.Maneuver)) | hudui.ChangeKey(boolKey(nav.Active) << 8)
}

func keyDistance(nav protocol.NavMessage) hudui.ChangeKey {
	return hudui.ChangeKey(BucketForDistance(nav.DistanceM))
}

func keyRoad(nav protocol.NavMessage) hudui.ChangeKey {
	return refreshHashStr(nav.Road) ^ refreshHashStr(nav.Instruction)
}

func keyETA(nav protocol.NavMessage) hudui.ChangeKey {
	return hudui.ChangeKey(nav.EtaMin)
}

func keyRibbon(nav protocol.NavMessage) hudui.ChangeKey {
	k := hudui.ChangeKey(len(nav.RibbonPoints)) | hudui.ChangeKey(nav.RibbonTurn<<8)
	if nav.Minimap != nil {
		k ^= hudui.ChangeKey(len(nav.Minimap.Route) << 4)
	}
	return k
}

func keyScreenNav(nav protocol.NavMessage, linked bool) hudui.ChangeKey {
	return keyBool(nav.Active) | hudui.ChangeKey(refreshHashStr(string(nav.Maneuver))<<4) | keyBool(linked)<<12
}

func keyMedia(m protocol.MediaMessage, linked bool) hudui.ChangeKey {
	return refreshHashStr(m.Title) ^ refreshHashStr(m.Artist) ^ hudui.ChangeKey(boolKey(m.Playing)<<1) ^ keyBool(linked)<<2
}

func boolKey(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

var refreshHashSeed = maphash.MakeSeed()

func refreshHashStr(s string) hudui.ChangeKey {
	var h maphash.Hash
	h.SetSeed(refreshHashSeed)
	_, _ = h.WriteString(s)
	return hudui.ChangeKey(h.Sum64())
}
