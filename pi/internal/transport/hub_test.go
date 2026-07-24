package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
)

func newTestHub() *Hub {
	return NewHub(hud.NewState(), &hud.RefreshGate{}, nil)
}

func TestHandleButtonEventScreenCycle(t *testing.T) {
	h := newTestHub()
	if h.State.CurrentScreen() != hud.ScreenNav {
		t.Fatalf("start screen=%v", h.State.CurrentScreen())
	}
	h.HandleButtonEvent(buttons.Next)
	if h.State.CurrentScreen() != hud.ScreenMedia {
		t.Fatalf("after Next screen=%v want media", h.State.CurrentScreen())
	}
	// Short Next on Media skips tracks — use NextLong to change screen
	h.HandleButtonEvent(buttons.NextLong)
	if h.State.CurrentScreen() != hud.ScreenStatus {
		t.Fatalf("after NextLong screen=%v want status", h.State.CurrentScreen())
	}
	h.HandleButtonEvent(buttons.PrevLong)
	if h.State.CurrentScreen() != hud.ScreenMedia {
		t.Fatalf("after PrevLong screen=%v want media", h.State.CurrentScreen())
	}
	h.HandleButtonEvent(buttons.ActionLong)
	if h.State.CurrentScreen() != hud.ScreenNav {
		t.Fatalf("after ActionLong screen=%v want nav", h.State.CurrentScreen())
	}
}

func TestHandleButtonEventMediaCmds(t *testing.T) {
	h := newTestHub()
	ch := h.SubscribeCmd()
	h.HandleButtonEvent(buttons.Next) // Nav → Media
	if h.State.CurrentScreen() != hud.ScreenMedia {
		t.Fatal("want media screen")
	}

	h.HandleButtonEvent(buttons.Next)
	select {
	case msg := <-ch:
		if msg.Action != protocol.CmdNextTrack {
			t.Fatalf("action=%v want next_track", msg.Action)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected next_track cmd")
	}

	h.HandleButtonEvent(buttons.Prev)
	select {
	case msg := <-ch:
		if msg.Action != protocol.CmdPrevTrack {
			t.Fatalf("action=%v want prev_track", msg.Action)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected prev_track cmd")
	}

	h.HandleButtonEvent(buttons.Action)
	select {
	case msg := <-ch:
		if msg.Action != protocol.CmdPlayPause {
			t.Fatalf("action=%v want play_pause", msg.Action)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected play_pause cmd")
	}

	h.HandleButtonEvent(buttons.NextLong)
	if h.State.CurrentScreen() != hud.ScreenStatus {
		t.Fatalf("after NextLong screen=%v want status", h.State.CurrentScreen())
	}
}

func TestApplyNavAndMedia(t *testing.T) {
	changed := 0
	h := NewHub(hud.NewState(), &hud.RefreshGate{}, func() { changed++ })

	h.ApplyNav(protocol.NavMessage{
		Active:       true,
		DistanceM:    200,
		DistanceText: "200 m",
		Road:         "High St",
		Maneuver:     protocol.ManeuverLeft,
		EtaMin:       12,
	})
	_, nav, _, _, _ := h.State.Snapshot()
	if !nav.Active || nav.DistanceM != 200 || nav.Maneuver != protocol.ManeuverLeft {
		t.Fatalf("nav=%+v", nav)
	}
	if changed != 1 {
		t.Fatalf("changed=%d want 1", changed)
	}

	h.ApplyMedia(protocol.MediaMessage{Playing: true, Title: "Song", Artist: "Artist"})
	_, _, media, _, force := h.State.Snapshot()
	if !media.Playing || media.Title != "Song" {
		t.Fatalf("media=%+v", media)
	}
	if !force {
		t.Fatal("media should force redraw")
	}
	if changed != 2 {
		t.Fatalf("changed=%d want 2", changed)
	}
}

func TestPublishCmdNonBlocking(t *testing.T) {
	h := newTestHub()
	ch := h.SubscribeCmd()
	for i := 0; i < 16; i++ {
		h.PublishCmd(protocol.CmdPlayPause)
	}
	select {
	case <-ch:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected at least one cmd delivered")
	}
}

func TestHTTPNavAndButton(t *testing.T) {
	h := newTestHub()
	mux := http.NewServeMux()
	mux.HandleFunc("/nav", func(w http.ResponseWriter, r *http.Request) {
		var n protocol.NavMessage
		_ = json.NewDecoder(r.Body).Decode(&n)
		h.ApplyNav(n)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/button", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(bytes.TrimSpace(b)) == "next" {
			h.HandleButtonEvent(buttons.Next)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := `{"type":"nav","active":true,"distance_m":50,"distance_text":"50 m","road":"A","maneuver":"straight","eta_min":3}`
	res, err := http.Post(srv.URL+"/nav", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	_, nav, _, _, _ := h.State.Snapshot()
	if !nav.Active || nav.DistanceM != 50 {
		t.Fatalf("nav=%+v", nav)
	}

	res, err = http.Post(srv.URL+"/button", "text/plain", bytes.NewBufferString("next"))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if h.State.CurrentScreen() != hud.ScreenMedia {
		t.Fatalf("screen=%v want media", h.State.CurrentScreen())
	}

	_ = context.Background()
}
