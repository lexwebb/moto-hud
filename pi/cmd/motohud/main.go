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

	"moto-hud/pi/internal/blehub"
	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/display"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
	"moto-hud/pi/internal/transport"
)

func main() {
	out := flag.String("out", "out/hud.png", "PNG output path (also Inky fallback)")
	httpAddr := flag.String("http", ":8787", "HTTP injector listen address")
	useInky := flag.Bool("inky", false, "Use Inky pHAT when available (Linux)")
	demo := flag.Bool("demo", false, "Show a static demo nav frame on start")
	assets := flag.String("assets", "", "Path to assets/hud (auto-detected if empty)")
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

	var disp display.Display
	var err error
	if *useInky {
		disp, err = display.NewInky(*out)
	} else {
		disp = display.NewPNG(*out)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer disp.Close()

	redrawCh := make(chan struct{}, 1)
	requestRedraw := func() {
		select {
		case redrawCh <- struct{}{}:
		default:
		}
	}

	hub := transport.NewHub(state, gate, requestRedraw)

	if err := hub.StartHTTP(ctx, *httpAddr); err != nil {
		log.Fatal(err)
	}
	if err := blehub.New(hub).Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := buttons.Start(ctx, func(ev buttons.Event) {
		hub.HandleButtonEvent(ev)
	}); err != nil {
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
	}
	requestRedraw()

	go loop(ctx, state, gate, disp, redrawCh)

	<-ctx.Done()
	log.Println("motohud: shutting down")
}

func loop(ctx context.Context, state *hud.State, gate *hud.RefreshGate, disp display.Display, redrawCh <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-redrawCh:
			maybeShow(state, gate, disp)
		case <-ticker.C:
			maybeShow(state, gate, disp)
		}
	}
}

func maybeShow(state *hud.State, gate *hud.RefreshGate, disp display.Display) {
	screen, nav, media, linked, force := state.Snapshot()
	if !gate.ShouldRedraw(screen, nav, force) {
		return
	}
	state.ClearForce()
	img := hud.Render(screen, nav, media, linked)
	if err := disp.Show(img); err != nil {
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
