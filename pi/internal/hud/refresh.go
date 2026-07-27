package hud

import "moto-hud/pi/internal/protocol"

// distanceStepM is how finely we refresh the panel as distance counts down.
// Matches the ≈ nearest-50 m presentation used for e-ink.
const distanceStepM = 50

type RefreshGate struct {
	lastManeuver protocol.Maneuver
	lastRoad     string
	lastActive   bool
	lastBucket   int
	lastScreen   Screen
	hasLast      bool
}

// bucketForDistance snaps metres to the nearest distanceStepM so the panel
// aims to redraw about every 50 m (when a refresh isn't already in flight).
// BucketForDistance snaps metres for refresh keys and e-ink presentation (~50 m).
func BucketForDistance(m int) int {
	return bucketForDistance(m)
}

func bucketForDistance(m int) int {
	if m <= 0 {
		return 0
	}
	return ((m + distanceStepM/2) / distanceStepM) * distanceStepM
}

// ShouldRedraw returns true when the e-ink panel should do a full refresh.
func (g *RefreshGate) ShouldRedraw(screen Screen, nav protocol.NavMessage, force bool) bool {
	if force || !g.hasLast {
		g.remember(screen, nav)
		return true
	}
	if screen != g.lastScreen {
		g.remember(screen, nav)
		return true
	}
	if screen != ScreenNav {
		// Media/Status: only redraw when forced (button) or caller sets force via content change upstream.
		return false
	}
	if nav.Active != g.lastActive ||
		nav.Maneuver != g.lastManeuver ||
		nav.Road != g.lastRoad {
		g.remember(screen, nav)
		return true
	}
	b := bucketForDistance(nav.DistanceM)
	if b != g.lastBucket {
		g.remember(screen, nav)
		return true
	}
	return false
}

// MarkContentChanged forces next media/status content update to redraw.
func (g *RefreshGate) MarkContentChanged() {
	g.hasLast = false
}

func (g *RefreshGate) remember(screen Screen, nav protocol.NavMessage) {
	g.lastScreen = screen
	g.lastManeuver = nav.Maneuver
	g.lastRoad = nav.Road
	g.lastActive = nav.Active
	g.lastBucket = bucketForDistance(nav.DistanceM)
	g.hasLast = true
}
