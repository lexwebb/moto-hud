package compose_test

import (
	"strings"
	"testing"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

func TestNavClassicPlanLayers(t *testing.T) {
	in := compose.Input{
		Screen: compose.ScreenNav,
		Nav: protocol.NavMessage{
			Active: true, DistanceM: 200, Road: "High St", EtaMin: 5,
			Maneuver: protocol.ManeuverLeft,
		},
		NavSVG: stubDrawDeps(),
	}
	sp, err := compose.BuildPlan(in)
	if err != nil {
		t.Fatal(err)
	}
	markup := compose.FrameBodyForTest(in, sp)
	for _, frag := range []string{`id="maneuver"`, `id="distance"`, `id="ribbon"`, `id="mode"`} {
		if !strings.Contains(markup, frag) {
			t.Fatalf("frame body missing %s", frag)
		}
	}
	if len(sp.Layers) < 4 {
		t.Fatalf("layers %d", len(sp.Layers))
	}
	var distLayer bool
	for _, l := range sp.Layers {
		if l.ID == hudui.NodeDistance {
			distLayer = true
			if l.Slot.Empty() || l.Patch == nil {
				t.Fatalf("distance layer not patchable: %+v", l)
			}
		}
	}
	if !distLayer {
		t.Fatal("missing distance layer")
	}
}

func stubDrawDeps() compose.DrawDeps {
	return compose.DrawDeps{
		ManeuverNodes: func(protocol.Maneuver) []scene.Node {
			return []scene.Node{scene.Path{D: "M0,0", Filled: false, StrokeWidth: 1}}
		},
		RibbonNodes: func(protocol.NavMessage, int, int) []scene.Node { return nil },
		TextSVG: func(id, faceSize string, x, baseline int, anchor, s string) string {
			return "<text id=\"" + id + "\">" + s + "</text>"
		},
		Fit: func(_ scene.Face, s string, _ int) string { return s },
		WrapRoad: func(s string, _, maxLines int) []string {
			if maxLines < 1 {
				return nil
			}
			return []string{s}
		},
		RoadBlockH: func(n int) int { return n * 16 },
	}
}
