package scene_test

import (
	"strings"
	"testing"

	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
)

func TestPatchBytes_linkRaw(t *testing.T) {
	doc := scene.Patch(16, 12, func(b *scene.Builder) {
		b.Raw(`<line x1="2" y1="2" x2="10" y2="10" stroke="#000"/>`)
	})
	out, err := svg.PatchBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `<svg`) || !strings.Contains(s, `fill="#fff"`) {
		t.Fatalf("expected patch svg doc: %s", s)
	}
}

func TestPatchBytes_textUsesPixelAttr(t *testing.T) {
	doc := scene.Patch(40, 20, func(b *scene.Builder) {
		b.Text("status_link", scene.Face8x16, 40, 14, "end", "UP")
	})
	out, err := svg.PatchBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `data-pixel="8x16"`) {
		t.Fatalf("missing pixel attr: %s", out)
	}
}
