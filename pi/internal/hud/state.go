package hud

import (
	"moto-hud/pi/internal/protocol"
	"sync"
)

type Screen int

const (
	ScreenNav Screen = iota
	ScreenMedia
	ScreenStatus
)

func (s Screen) String() string {
	switch s {
	case ScreenNav:
		return "NAV"
	case ScreenMedia:
		return "MEDIA"
	case ScreenStatus:
		return "STATUS"
	default:
		return "?"
	}
}

type State struct {
	mu sync.RWMutex

	Screen      Screen
	Nav         protocol.NavMessage
	Media       protocol.MediaMessage
	BLELinked   bool
	ForceRedraw bool
}

func NewState() *State {
	return &State{
		Screen: ScreenNav,
		Nav: protocol.NavMessage{
			Type:         "nav",
			Active:       false,
			Instruction:  "Waiting for nav",
			DistanceText: "--",
			Road:         "",
			Maneuver:     protocol.ManeuverUnknown,
		},
		Media: protocol.MediaMessage{
			Type:    "media",
			Playing: false,
			Title:   "-",
			Artist:  "",
		},
	}
}

func (s *State) Snapshot() (Screen, protocol.NavMessage, protocol.MediaMessage, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Screen, s.Nav, s.Media, s.BLELinked, s.ForceRedraw
}

func (s *State) ClearForce() {
	s.mu.Lock()
	s.ForceRedraw = false
	s.mu.Unlock()
}

func (s *State) SetNav(n protocol.NavMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n.Type = "nav"
	s.Nav = n
}

func (s *State) SetMedia(m protocol.MediaMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m.Type = "media"
	s.Media = m
}

func (s *State) SetBLELinked(linked bool) {
	s.mu.Lock()
	s.BLELinked = linked
	s.mu.Unlock()
}

func (s *State) NextScreen() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Screen = (s.Screen + 1) % 3
	s.ForceRedraw = true
}

func (s *State) PrevScreen() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Screen = (s.Screen + 2) % 3
	s.ForceRedraw = true
}

func (s *State) HomeNav() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Screen = ScreenNav
	s.ForceRedraw = true
}

func (s *State) RequestRedraw() {
	s.mu.Lock()
	s.ForceRedraw = true
	s.mu.Unlock()
}

func (s *State) CurrentScreen() Screen {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Screen
}
