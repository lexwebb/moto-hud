package hud

import "moto-hud/pi/internal/protocol"

// Distance thresholds that trigger a full e-ink redraw (metres).
var distanceThresholds = []int{500, 200, 100, 50, 20}

type RefreshGate struct {
	lastManeuver protocol.Maneuver
	lastRoad     string
	lastActive   bool
	lastBucket   int
	lastScreen   Screen
	hasLast      bool
}

func bucketForDistance(m int) int {
	for _, t := range distanceThresholds {
		if m >= t {
			return t
		}
	}
	return 0
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
