package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"moto-hud/pi/internal/protocol"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1:8787", "motohud HTTP base URL")
	scenario := flag.String("scenario", "approach", "approach|arrive|idle|media")
	flag.Parse()

	switch *scenario {
	case "approach":
		runApproach(*addr)
	case "arrive":
		postNav(*addr, protocol.NavMessage{
			Type: "nav", Active: true, Instruction: "Arrive at destination",
			DistanceM: 0, DistanceText: "0 m", Road: "Destination", Maneuver: protocol.ManeuverArrive,
		})
	case "idle":
		postNav(*addr, protocol.NavMessage{
			Type: "nav", Active: false, Instruction: "Waiting for nav",
			DistanceText: "--", Maneuver: protocol.ManeuverUnknown,
		})
	case "media":
		postMedia(*addr, protocol.MediaMessage{
			Type: "media", Playing: true, Title: "Born to Run", Artist: "Bruce Springsteen",
		})
	default:
		fmt.Fprintf(os.Stderr, "unknown scenario %q\n", *scenario)
		os.Exit(2)
	}
}

func runApproach(addr string) {
	steps := []int{800, 500, 200, 100, 50, 20}
	for _, d := range steps {
		postNav(addr, protocol.NavMessage{
			Type:         "nav",
			Active:       true,
			Instruction:  "Turn left onto High St",
			DistanceM:    d,
			DistanceText: fmt.Sprintf("%d m", d),
			Road:         "High St",
			EtaMin:       12,
			Maneuver:     protocol.ManeuverLeft,
		})
		time.Sleep(1500 * time.Millisecond)
	}
}

func postNav(addr string, n protocol.NavMessage) {
	post(addr+"/nav", n)
}

func postMedia(addr string, m protocol.MediaMessage) {
	post(addr+"/media", m)
}

func post(url string, v any) {
	b, _ := json.Marshal(v)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST %s: %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	fmt.Printf("POST %s -> %s\n", url, resp.Status)
}
