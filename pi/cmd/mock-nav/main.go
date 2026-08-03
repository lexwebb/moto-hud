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
	scenario := flag.String("scenario", "approach", "approach|arrive|idle|media|junctions")
	flag.Parse()

	switch *scenario {
	case "approach":
		runApproach(*addr)
	case "junctions":
		runJunctionTour(*addr)
	case "arrive":
		postNav(*addr, protocol.NavMessage{
			Type: "nav", Active: true, Instruction: "Arrive at destination",
			DistanceM: 0, DistanceText: "0 m", Road: "Destination", Maneuver: protocol.ManeuverArrive,
			Junction: &protocol.JunctionMessage{Kind: "arrive", Outbound: "straight"},
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
	j := &protocol.JunctionMessage{Kind: "simple", Outbound: "left", Through: false}
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
			Junction:     j,
		})
		time.Sleep(1500 * time.Millisecond)
	}
}

// runJunctionTour steps through several junction kinds so the live left column
// can be eyeballed on hardware (pair with motohud -junction).
func runJunctionTour(addr string) {
	type step struct {
		maneuver protocol.Maneuver
		road     string
		instr    string
		j        protocol.JunctionMessage
		dist     int
	}
	tour := []step{
		{protocol.ManeuverLeft, "High St", "Turn left onto High St",
			protocol.JunctionMessage{Kind: "simple", Outbound: "left"}, 200},
		{protocol.ManeuverRight, "Mill Lane", "Turn right onto Mill Lane",
			protocol.JunctionMessage{Kind: "simple", Outbound: "right"}, 150},
		{protocol.ManeuverStraight, "Cross St", "Continue onto Cross St",
			protocol.JunctionMessage{Kind: "crossroads", Outbound: "straight", Through: true,
				Sides: []protocol.JunctionSideArm{{Side: "left"}, {Side: "right"}}}, 180},
		{protocol.ManeuverLeft, "Side Rd", "Turn left at T-junction",
			protocol.JunctionMessage{Kind: "t_junction", Outbound: "left"}, 120},
		{protocol.ManeuverRoundabout, "Ring Rd", "At roundabout take 2nd exit",
			protocol.JunctionMessage{Kind: "roundabout", Outbound: "left", Exits: 4, Exit: 2}, 90},
		{protocol.ManeuverArrive, "Destination", "Arrive at destination",
			protocol.JunctionMessage{Kind: "arrive", Outbound: "straight"}, 0},
	}
	for _, s := range tour {
		postNav(addr, protocol.NavMessage{
			Type: "nav", Active: true, Instruction: s.instr,
			DistanceM: s.dist, DistanceText: fmt.Sprintf("%d m", s.dist),
			Road: s.road, EtaMin: 8, Maneuver: s.maneuver, Junction: &s.j,
		})
		time.Sleep(3 * time.Second)
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
