package platform_test

import (
	"context"
	"image"
	"image/color"
	"testing"
	"time"

	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/platform"
	"moto-hud/pi/internal/transport"
)

func TestOpenEmuAndPushButton(t *testing.T) {
	state := hud.NewState()
	gate := &hud.RefreshGate{}
	hub := transport.NewHub(state, gate, func() {})

	host, err := platform.Open(platform.Config{
		Kind:    platform.KindEmu,
		PNGPath: "",
		Hub:     hub,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer host.Screen.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := host.Phone.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := host.Controls.Listen(ctx, hub.HandleButtonEvent); err != nil {
		t.Fatal(err)
	}

	img := image.NewGray(image.Rect(0, 0, hud.Width, hud.Height))
	img.SetGray(0, 0, color.Gray{Y: 0})
	if err := host.Screen.Show(img); err != nil {
		t.Fatal(err)
	}
	mem := host.Screen.(*platform.MemoryScreen)
	if mem.Frame() == nil {
		t.Fatal("expected stored frame")
	}

	ctrl := host.Controls.(*platform.ChannelControls)
	ctrl.Push(buttons.Next)
	deadline := time.After(500 * time.Millisecond)
	for {
		if state.CurrentScreen() == hud.ScreenMedia {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("screen=%v want media", state.CurrentScreen())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
