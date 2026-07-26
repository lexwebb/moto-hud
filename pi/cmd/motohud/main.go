package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/platform"
	"moto-hud/pi/internal/protocol"
	"moto-hud/pi/internal/transport"
)

func main() {
	out := flag.String("out", "out/hud.png", "PNG output path (also Inky fallback)")
	httpAddr := flag.String("http", ":8787", "HTTP injector listen address")
	useInky := flag.Bool("inky", false, "Prefer Inky pHAT when available (Linux; -host auto)")
	demo := flag.Bool("demo", false, "Show a static demo nav frame on start")
	assets := flag.String("assets", "", "Path to assets/hud (auto-detected if empty)")
	hostKind := flag.String("host", "auto", "Hardware host: auto|png|inky|waveshare|lcd|emu|test")
	flag.Parse()

	if *assets != "" {
		hud.SetAssetDir(*assets)
	} else if root := detectRepoRoot(); root != "" {
		hud.SetAssetDir(filepath.Join(root, "assets", "hud"))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	state := hud.NewState()
	gate := &hud.RefreshGate{}

	redrawCh := make(chan struct{}, 1)
	requestRedraw := func() {
		select {
		case redrawCh <- struct{}{}:
		default:
		}
	}

	hub := transport.NewHub(state, gate, requestRedraw)

	host, err := platform.Open(platform.Config{
		Kind:     platform.Kind(*hostKind),
		PNGPath:  *out,
		WantInky: *useInky,
		Hub:      hub,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer host.Screen.Close()
	log.Printf("platform: host=%s", effectiveKind(*hostKind, *useInky))

	if err := hub.StartHTTP(ctx, *httpAddr); err != nil {
		log.Fatal(err)
	}
	if err := host.Phone.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := host.Controls.Listen(ctx, hub.HandleButtonEvent); err != nil {
		log.Fatal(err)
	}

	if *demo {
		state.SetNav(protocol.NavMessage{
			Type:         "nav",
			Active:       true,
			Instruction:  "Turn left onto High St",
			DistanceM:    200,
			DistanceText: "200 m",
			Road:         "High St",
			EtaMin:       12,
			Maneuver:     protocol.ManeuverLeft,
		})
		state.SetMedia(protocol.MediaMessage{
			Type: "media", Playing: true, Title: "Born to Run", Artist: "Bruce Springsteen",
		})
		if platform.Kind(*hostKind) != platform.KindEmu && platform.Kind(*hostKind) != platform.KindTest {
			state.SetBLELinked(true)
		}
	}
	requestRedraw()

	go loop(ctx, state, gate, host.Screen, redrawCh)

	<-ctx.Done()
	log.Println("motohud: shutting down")
}

func effectiveKind(k string, inky bool) string {
	if k == "" || k == "auto" {
		if inky {
			return "inky"
		}
		return "png"
	}
	return k
}

func loop(ctx context.Context, state *hud.State, gate *hud.RefreshGate, scr platform.Screen, redrawCh <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-redrawCh:
			maybeShow(state, gate, scr)
		case <-ticker.C:
			maybeShow(state, gate, scr)
		}
	}
}

func maybeShow(state *hud.State, gate *hud.RefreshGate, scr platform.Screen) {
	screen, nav, media, linked, force := state.Snapshot()
	if !gate.ShouldRedraw(screen, nav, force) {
		return
	}
	state.ClearForce()
	img := hud.Render(screen, nav, media, linked)
	if err := scr.Show(img); err != nil {
		log.Printf("display: %v", err)
	}
}

func detectRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if st, err := os.Stat(filepath.Join(dir, "assets", "hud", "frame.svg")); err == nil && !st.IsDir() {
			return dir
		}
		// fallback for older checkouts that still have nav.svg only
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
