package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
)

// Hub fans in nav/media updates and fans out cmd notifications.
type Hub struct {
	State *hud.State
	Gate  *hud.RefreshGate

	mu       sync.Mutex
	cmdSubs  []chan protocol.CmdMessage
	onChange func()
}

func NewHub(state *hud.State, gate *hud.RefreshGate, onChange func()) *Hub {
	return &Hub{State: state, Gate: gate, onChange: onChange}
}

func (h *Hub) OnChange(fn func()) {
	h.onChange = fn
}

func (h *Hub) notifyChange() {
	if h.onChange != nil {
		h.onChange()
	}
}

func (h *Hub) ApplyNav(n protocol.NavMessage) {
	h.State.SetNav(n)
	h.notifyChange()
}

func (h *Hub) ApplyMedia(m protocol.MediaMessage) {
	h.State.SetMedia(m)
	h.Gate.MarkContentChanged()
	h.State.RequestRedraw()
	h.notifyChange()
}

func (h *Hub) SetLinked(linked bool) {
	h.State.SetBLELinked(linked)
	h.Gate.MarkContentChanged()
	h.State.RequestRedraw()
	h.notifyChange()
}

func (h *Hub) PublishCmd(action protocol.CmdAction) {
	msg := protocol.CmdMessage{Type: "cmd", Action: action}
	h.mu.Lock()
	subs := append([]chan protocol.CmdMessage(nil), h.cmdSubs...)
	h.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
	b, _ := json.Marshal(msg)
	log.Printf("cmd: %s", b)
}

func (h *Hub) SubscribeCmd() chan protocol.CmdMessage {
	ch := make(chan protocol.CmdMessage, 8)
	h.mu.Lock()
	h.cmdSubs = append(h.cmdSubs, ch)
	h.mu.Unlock()
	return ch
}

// StartHTTP serves a mock/dev injector and status API on addr (e.g. :8787).
func (h *Hub) StartHTTP(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/nav", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var n protocol.NavMessage
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.ApplyNav(n)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/media", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var m protocol.MediaMessage
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.ApplyMedia(m)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/button", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		switch string(body) {
		case "prev":
			h.HandleButtonEvent(buttons.Prev)
		case "next":
			h.HandleButtonEvent(buttons.Next)
		case "prev_long":
			h.HandleButtonEvent(buttons.PrevLong)
		case "next_long":
			h.HandleButtonEvent(buttons.NextLong)
		case "action":
			h.HandleButtonEvent(buttons.Action)
		case "action_long":
			h.HandleButtonEvent(buttons.ActionLong)
		case "skip_prev":
			h.HandleMediaSkip(false)
		case "skip_next":
			h.HandleMediaSkip(true)
		default:
			http.Error(w, "unknown button", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/frame.svg", func(w http.ResponseWriter, r *http.Request) {
		screen, nav, media, linked, _ := h.State.Snapshot()
		svg, err := hud.BuildPixelSVG(screen, nav, media, linked)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(svg)
	})
	// Same 1-bit raster the Pi / PNG mock use — browser won't AA stroked arrows.
	mux.HandleFunc("/frame.png", func(w http.ResponseWriter, r *http.Request) {
		screen, nav, media, linked, _ := h.State.Snapshot()
		img := hud.Render(screen, nav, media, linked)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		if err := png.Encode(w, img); err != nil {
			log.Printf("transport: frame.png encode: %v", err)
		}
	})
	mux.HandleFunc("/fonts.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(hud.ListFontCandidates())
	})
	mux.HandleFunc("/font-specimen.png", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			id = "terminus-bold"
		}
		img, err := hud.RenderFontSpecimen(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		if err := png.Encode(w, img); err != nil {
			log.Printf("transport: font-specimen encode: %v", err)
		}
	})
	if root := findRepoRoot(); root != "" {
		mux.Handle("/preview/", http.StripPrefix("/preview/", http.FileServer(http.Dir(filepath.Join(root, "web/preview")))))
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(root, "assets")))))
	}
	mux.HandleFunc("/cmd/stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no flush", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		ch := h.SubscribeCmd()
		defer func() {
			h.mu.Lock()
			for i, c := range h.cmdSubs {
				if c == ch {
					h.cmdSubs = append(h.cmdSubs[:i], h.cmdSubs[i+1:]...)
					break
				}
			}
			h.mu.Unlock()
		}()
		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-ch:
				b, _ := json.Marshal(msg)
				fmt.Fprintf(w, "data: %s\n\n", b)
				flusher.Flush()
			}
		}
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("transport: HTTP injector on http://%s (preview /preview/  frame /frame.png|/frame.svg)", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("transport: HTTP error: %v", err)
		}
	}()
	return nil
}

func (h *Hub) handleAction() {
	switch h.State.CurrentScreen() {
	case hud.ScreenMedia:
		h.PublishCmd(protocol.CmdPlayPause)
	case hud.ScreenStatus:
		h.State.RequestRedraw()
	default:
		// Nav: no-op for short press
	}
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if st, err := os.Stat(filepath.Join(dir, "assets", "hud", "nav.svg")); err == nil && !st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
