package bitmap_test

import (
	"testing"

	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/hudui/render/bitmap"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/protocol"
)

func TestRasterize_uturnPathHasInk(t *testing.T) {
	doc := scene.Patch(40, 40, func(b *scene.Builder) {
		b.Append(hud.ManeuverNodes(protocol.ManeuverUTurn)...)
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	ink := 0
	for _, p := range got.Pix {
		if p < 128 {
			ink++
		}
	}
	if ink < 20 {
		t.Fatalf("u-turn too sparse: %d ink px", ink)
	}
}

func TestRasterize_pathHV(t *testing.T) {
	doc := scene.Patch(20, 20, func(b *scene.Builder) {
		b.Path("M2,2 H18 V18 H2 Z", false, 1)
	})
	got, err := bitmap.Rasterize(doc)
	if err != nil {
		t.Fatal(err)
	}
	ink := 0
	for _, p := range got.Pix {
		if p < 128 {
			ink++
		}
	}
	if ink < 40 {
		t.Fatalf("H/V rect stroke too sparse: %d", ink)
	}
}
