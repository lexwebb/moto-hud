package compose

import (
	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/protocol"
)

// ScreenKind mirrors hud.Screen without importing hud (avoids cycles).
type ScreenKind int

const (
	ScreenNav ScreenKind = iota
	ScreenMedia
	ScreenStatus
)

// Input is everything needed to compose one frame plan.
type Input struct {
	Screen ScreenKind
	Nav    protocol.NavMessage
	Media  protocol.MediaMessage
	Linked bool
	NavSVG DrawDeps // set by hud when building nav plans
	LinkSVG func(linked bool) string
}

// Keys builds change keys from protocol state (layout-agnostic).
type Keys struct{}

func (Keys) Bool(b bool) hudui.ChangeKey {
	if b {
		return 1
	}
	return 0
}

func (Keys) DistanceBucket(m int) hudui.ChangeKey {
	return hudui.ChangeKey(bucketDistance(m))
}

func (Keys) Hash(s string) hudui.ChangeKey {
	return hashStr(s)
}

func (Keys) NavScreen(nav protocol.NavMessage) hudui.ChangeKey {
	k := Keys{}
	return k.Bool(nav.Active) | k.Hash(string(nav.Maneuver))<<4
}

func (Keys) Link(linked bool) hudui.ChangeKey {
	return Keys{}.Bool(linked)
}

func (Keys) Maneuver(nav protocol.NavMessage) hudui.ChangeKey {
	k := Keys{}
	return k.Hash(string(nav.Maneuver)) | hudui.ChangeKey(boolKey(nav.Active)<<8)
}

func (Keys) Road(nav protocol.NavMessage) hudui.ChangeKey {
	return hashStr(nav.Road) ^ hashStr(nav.Instruction)
}

func (Keys) ETA(nav protocol.NavMessage) hudui.ChangeKey {
	return hudui.ChangeKey(nav.EtaMin)
}

func (Keys) Ribbon(nav protocol.NavMessage) hudui.ChangeKey {
	k := hudui.ChangeKey(len(nav.RibbonPoints)) | hudui.ChangeKey(nav.RibbonTurn<<8)
	if nav.Minimap != nil {
		k ^= hudui.ChangeKey(len(nav.Minimap.Route) << 4)
	}
	return k
}

func (Keys) MediaScreen(media protocol.MediaMessage) hudui.ChangeKey {
	return hashStr(media.Title) ^ hashStr(media.Artist) ^ hudui.ChangeKey(boolKey(media.Playing)<<1)
}

func (Keys) MediaPlaying(media protocol.MediaMessage) hudui.ChangeKey {
	if media.Playing {
		return 1
	}
	return 0
}

func (Keys) StatusLink(linked bool) hudui.ChangeKey {
	return Keys{}.Bool(linked)
}

func (Keys) StatusNav(active bool) hudui.ChangeKey {
	return Keys{}.Bool(active)
}

func (Keys) StatusPkts() hudui.ChangeKey {
	return 0
}

func boolKey(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

func bucketDistance(m int) int {
	const step = 50
	if m <= 0 {
		return 0
	}
	return ((m + step/2) / step) * step
}
