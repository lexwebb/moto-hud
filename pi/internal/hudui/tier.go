package hudui

import "image"

// Tier describes how often a component may drive panel updates (ADR 0010).
type Tier int

const (
	TierStatic Tier = iota
	TierSlow
	TierFast
	TierPartialOK
)

// NodeID identifies a refresh-tracked region on a screen.
type NodeID string

const (
	NodeDistance NodeID = "distance"
	NodeRoad     NodeID = "road"
	NodeETA      NodeID = "eta"
	NodeRibbon   NodeID = "ribbon"
	NodeManeuver NodeID = "maneuver"
	NodeChrome   NodeID = "chrome"
	NodeScreen   NodeID = "screen"
	NodeMediaTitle  NodeID = "media_title"
	NodeMediaArtist NodeID = "media_artist"
	NodeMediaState  NodeID = "media_state"
)

// Slot is a fixed or computed rectangle in canvas coordinates.
type Slot struct {
	Rect image.Rectangle
}

// ChangeKey is a comparable snapshot of props that affect pixels in a slot.
type ChangeKey uint64

// Descriptor binds tier, geometry, and change detection for one region.
type Descriptor struct {
	ID    NodeID
	Tier  Tier
	Slot  image.Rectangle
	Key   ChangeKey
	Dirty bool
}
