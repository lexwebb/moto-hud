// Package platform defines hardware ports so the HUD core can run against
// a real Pi, a PNG/keyboard desktop host, an in-process emulator, or tests.
package platform

import (
	"context"
	"fmt"
	"image"
	"sync"

	"moto-hud/pi/internal/blehub"
	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/display"
	"moto-hud/pi/internal/transport"
)

// Screen is the e-ink (or mock) panel.
type Screen interface {
	Show(img *image.Gray) error
	Close() error
}

// Controls delivers physical or virtual button events.
type Controls interface {
	Listen(ctx context.Context, on buttons.Handler) error
}

// PhoneLink is the BLE (or stub/emulator) path to the companion.
type PhoneLink interface {
	Start(ctx context.Context) error
}

// Host bundles the three ports the app needs to run.
type Host struct {
	Screen   Screen
	Controls Controls
	Phone    PhoneLink
}

// Kind selects which concrete adapters to open.
type Kind string

const (
	KindAuto      Kind = "auto"      // Inky on Linux if requested, else PNG+keyboard+stub BLE
	KindPNG       Kind = "png"       // desktop bring-up
	KindInky      Kind = "inky"      // Inky pHAT (Linux)
	KindWaveshare Kind = "waveshare" // Waveshare 2.13" B/W e-Paper HAT (Linux)
	KindLCD       Kind = "lcd"       // Display HAT Mini ST7789 320×240 letterbox (Linux)
	KindEmu       Kind = "emu"       // in-process: memory screen + channel buttons + loopback link
	KindTest      Kind = "test"      // silent memory screen; no auto button source
)

// Config configures Open.
type Config struct {
	Kind     Kind
	PNGPath  string // used by png/inky-fallback/emu snapshot
	WantInky bool   // KindAuto only
	Hub      *transport.Hub
}

// Open builds a Host for the requested kind.
func Open(cfg Config) (*Host, error) {
	if cfg.PNGPath == "" {
		cfg.PNGPath = "out/hud.png"
	}
	if cfg.Hub == nil {
		return nil, fmt.Errorf("platform: Hub is required")
	}
	kind := cfg.Kind
	if kind == "" || kind == KindAuto {
		if cfg.WantInky {
			kind = KindInky
		} else {
			kind = KindPNG
		}
	}
	switch kind {
	case KindPNG:
		buttons.SetActionGPIO(buttons.GPIOAction)
		return &Host{
			Screen:   display.NewPNG(cfg.PNGPath),
			Controls: StdControls{},
			Phone:    blehub.New(cfg.Hub),
		}, nil
	case KindInky:
		buttons.SetActionGPIO(buttons.GPIOAction)
		scr, err := display.NewInky(cfg.PNGPath)
		if err != nil {
			return nil, err
		}
		return &Host{
			Screen:   scr,
			Controls: StdControls{},
			Phone:    blehub.New(cfg.Hub),
		}, nil
	case KindWaveshare:
		buttons.SetActionGPIO(buttons.GPIOAction)
		scr, err := display.NewWaveshare(cfg.PNGPath)
		if err != nil {
			return nil, err
		}
		return &Host{
			Screen:   scr,
			Controls: StdControls{},
			Phone:    blehub.New(cfg.Hub),
		}, nil
	case KindLCD:
		// Display HAT Mini backlight uses BCM 13; map Action to HAT button X (16).
		buttons.SetActionGPIO(buttons.GPIOActionLCD)
		scr, err := display.NewLCD(cfg.PNGPath)
		if err != nil {
			return nil, err
		}
		return &Host{
			Screen:   scr,
			Controls: StdControls{},
			Phone:    blehub.New(cfg.Hub),
		}, nil
	case KindEmu:
		buttons.SetActionGPIO(buttons.GPIOAction)
		mem := NewMemoryScreen(cfg.PNGPath)
		return &Host{
			Screen:   mem,
			Controls: NewChannelControls(),
			Phone:    NewLoopbackLink(cfg.Hub),
		}, nil
	case KindTest:
		buttons.SetActionGPIO(buttons.GPIOAction)
		return &Host{
			Screen:   NewMemoryScreen(""),
			Controls: NewChannelControls(),
			Phone:    NewLoopbackLink(cfg.Hub),
		}, nil
	default:
		return nil, fmt.Errorf("platform: unknown kind %q", kind)
	}
}

// StdControls uses the build-tagged buttons.Start (GPIO on Linux, keyboard elsewhere).
type StdControls struct{}

func (StdControls) Listen(ctx context.Context, on buttons.Handler) error {
	return buttons.Start(ctx, on)
}

// MemoryScreen keeps the last frame in RAM (optional PNG mirror for debugging).
type MemoryScreen struct {
	mu       sync.Mutex
	Last     *image.Gray
	PNGPath  string
	OnShow   func(*image.Gray)
	png      *display.PNGDisplay
}

func NewMemoryScreen(pngPath string) *MemoryScreen {
	m := &MemoryScreen{PNGPath: pngPath}
	if pngPath != "" {
		m.png = display.NewPNG(pngPath)
	}
	return m
}

func (m *MemoryScreen) Show(img *image.Gray) error {
	m.mu.Lock()
	cp := image.NewGray(img.Bounds())
	copy(cp.Pix, img.Pix)
	m.Last = cp
	cb := m.OnShow
	m.mu.Unlock()
	if cb != nil {
		cb(cp)
	}
	if m.png != nil {
		return m.png.Show(img)
	}
	return nil
}

func (m *MemoryScreen) Frame() *image.Gray {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Last
}

func (m *MemoryScreen) Close() error {
	if m.png != nil {
		return m.png.Close()
	}
	return nil
}

// ChannelControls is a virtual button pad for emulator / tests / WASM JS bridge.
type ChannelControls struct {
	ch chan buttons.Event
}

func NewChannelControls() *ChannelControls {
	return &ChannelControls{ch: make(chan buttons.Event, 8)}
}

func (c *ChannelControls) Listen(ctx context.Context, on buttons.Handler) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-c.ch:
				on(ev)
			}
		}
	}()
	return nil
}

// Push injects a button event (emulator UI or test).
func (c *ChannelControls) Push(ev buttons.Event) {
	select {
	case c.ch <- ev:
	default:
	}
}

// LoopbackLink pretends the phone is connected and logs outbound cmds.
type LoopbackLink struct {
	Hub *transport.Hub
}

func NewLoopbackLink(hub *transport.Hub) *LoopbackLink {
	return &LoopbackLink{Hub: hub}
}

func (l *LoopbackLink) Start(ctx context.Context) error {
	l.Hub.SetLinked(true)
	go func() {
		ch := l.Hub.SubscribeCmd()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				// Emulator / tests observe cmds via Hub.SubscribeCmd themselves.
			}
		}
	}()
	return nil
}
