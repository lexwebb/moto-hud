package hud

import (
	"image"
	"strings"
	"testing"

	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

func TestJunctionNodesAllKinds(t *testing.T) {
	cases := []struct {
		name string
		j    *JunctionMessage
	}{
		{
			name: "simple_left",
			j:    &JunctionMessage{Kind: JunctionSimple, Outbound: "left", Through: true, Drive: JunctionDriveLeft},
		},
		{
			name: "t_junction_left",
			j:    &JunctionMessage{Kind: JunctionTJunction, Outbound: "left", Through: false},
		},
		{
			name: "crossroads_right",
			j:    &JunctionMessage{Kind: JunctionCrossroads, Outbound: "right", Through: true},
		},
		{
			name: "fork_slight_left",
			j:    &JunctionMessage{Kind: JunctionFork, Outbound: "slight_left", Through: false},
		},
		{
			name: "merge_right",
			j:    &JunctionMessage{Kind: JunctionMerge, Outbound: "straight", Through: true, Side: "right"},
		},
		{
			name: "dual_cross_median",
			j: &JunctionMessage{
				Kind: JunctionDualCarriageway, Outbound: "right", Through: true,
				Drive: JunctionDriveLeft, CrossMedian: true,
			},
		},
		{
			name: "roundabout_exit2",
			j: &JunctionMessage{
				Kind: JunctionRoundabout, Outbound: "right", Drive: JunctionDriveLeft,
				Exits: 4, Exit: 2,
			},
		},
		{
			name: "ramp_exit_right",
			j:    &JunctionMessage{Kind: JunctionRampExit, Outbound: "slight_right", Through: true},
		},
		{
			name: "ramp_enter_left",
			j:    &JunctionMessage{Kind: JunctionRampEnter, Outbound: "straight", Through: true, Side: "left"},
		},
		{
			name: "u_turn_drive_left",
			j:    &JunctionMessage{Kind: JunctionUTurn, Outbound: "u_turn", Through: false, Drive: JunctionDriveLeft},
		},
		{
			name: "arrive",
			j:    &JunctionMessage{Kind: JunctionArrive, Outbound: "straight", Through: false},
		},
		{
			name: "depart",
			j:    &JunctionMessage{Kind: JunctionDepart, Outbound: "straight", Through: true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := JunctionNodes(tc.j, 70, 80)
			if len(nodes) == 0 {
				t.Fatal("expected nodes")
			}
			img, err := RenderJunction(tc.j, 70, 80)
			if err != nil {
				t.Fatal(err)
			}
			if countGrayInk(img) < 20 {
				t.Fatalf("raster nearly blank (%d ink px)", countGrayInk(img))
			}
			frag := JunctionSVGFragment(tc.j, 70, 80)
			if !strings.Contains(frag, "<") {
				t.Fatalf("empty svg fragment: %q", frag)
			}
		})
	}
}

func TestJunctionUnknownFallsBackToSimple(t *testing.T) {
	j := &JunctionMessage{Kind: "jughandle", Outbound: "left", Through: false}
	nodes := JunctionNodes(j, 70, 80)
	solidHoriz := false
	var walk func([]scene.Node)
	walk = func(ns []scene.Node) {
		for _, n := range ns {
			switch v := n.(type) {
			case scene.Group:
				walk(v.Children)
			case scene.Line:
				if v.Y1 == v.Y2 && v.Dash == "" && v.StrokeWidth >= 2 {
					solidHoriz = true
				}
			}
		}
	}
	walk(nodes)
	if !solidHoriz {
		t.Fatal("unknown kind should render as simple with solid outbound arm")
	}
}

func TestJunctionForkHasSolidAndDashedArms(t *testing.T) {
	j := &JunctionMessage{Kind: JunctionFork, Outbound: "slight_left"}
	nodes := JunctionNodes(j, 70, 80)
	solid, dashed := 0, 0
	var walk func([]scene.Node)
	walk = func(ns []scene.Node) {
		for _, n := range ns {
			switch v := n.(type) {
			case scene.Group:
				walk(v.Children)
			case scene.Line:
				if v.Dash != "" {
					dashed++
				} else if v.StrokeWidth >= junctionThick {
					solid++
				}
			}
		}
	}
	walk(nodes)
	if solid < 2 || dashed < 1 {
		t.Fatalf("fork should have approach+route solid and alternate dashed; solid=%d dashed=%d", solid, dashed)
	}
}

func TestSynthesizeJunctionFromManeuver(t *testing.T) {
	j := SynthesizeJunctionFromManeuver(protocol.ManeuverRoundabout)
	if j.Kind != JunctionRoundabout || j.Exits != 4 || j.Exit != 2 || j.Through {
		t.Fatalf("roundabout synth: %+v", j)
	}
	j2 := SynthesizeJunctionFromManeuver(protocol.ManeuverSlightLeft)
	if j2.Kind != JunctionFork || j2.Through {
		t.Fatalf("slight_left → fork through=false, got %+v", j2)
	}
	j3 := SynthesizeJunctionFromManeuver(protocol.ManeuverLeft)
	if j3.Kind != JunctionSimple || j3.Outbound != "left" || j3.Through {
		t.Fatalf("left → simple through=false, got %+v", j3)
	}
}

func TestPreferJunctionTemplatesDefaultOff(t *testing.T) {
	if PreferJunctionTemplates {
		t.Fatal("production default must keep meter minimap (PreferJunctionTemplates=false)")
	}
}

func countGrayInk(img *image.Gray) int {
	ink := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y < 250 {
				ink++
			}
		}
	}
	return ink
}
