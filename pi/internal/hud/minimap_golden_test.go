package hud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"moto-hud/pi/internal/protocol"
)

const minimapPaneW = 70
const minimapPaneH = 80

type minimapFixture struct {
	ID         string                   `json:"id"`
	Title      string                   `json:"title"`
	Notes      string                   `json:"notes"`
	AlongM     float64                  `json:"along_m"`
	Acceptance []string                 `json:"acceptance"`
	Minimap    *protocol.MinimapMessage `json:"minimap"`
	Nav        *protocol.NavMessage     `json:"nav,omitempty"`
}

func TestMinimapGolden(t *testing.T) {
	dir := filepath.Join("testdata", "minimap")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	update := os.Getenv("UPDATE_MINIMAP_GOLDEN") == "1"
	n := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" || e.Name() == "index.json" {
			continue
		}
		n++
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			var fix minimapFixture
			if err := json.Unmarshal(raw, &fix); err != nil {
				t.Fatal(err)
			}
			mm := fix.Minimap
			if mm == nil && fix.Nav != nil {
				mm = fix.Nav.Minimap
			}
			if mm == nil || len(mm.Route) < 2 {
				t.Fatal("fixture missing minimap.route")
			}
			img, err := RenderMinimap(mm, minimapPaneW, minimapPaneH)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			gotPNG := mustEncodePNG(t, img)
			base := strings.TrimSuffix(name, ".json")
			goldenPath := filepath.Join(dir, base+".png")
			svgPath := filepath.Join(dir, base+".svg")
			frag := MinimapSVGFragment(mm, minimapPaneW, minimapPaneH)
			svgDoc := fmt.Sprintf(
				`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">%s</svg>`+"\n",
				minimapPaneW, minimapPaneH, minimapPaneW, minimapPaneH, frag,
			)

			if update {
				if err := os.WriteFile(goldenPath, gotPNG, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(svgPath, []byte(svgDoc), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("missing golden %s — run UPDATE_MINIMAP_GOLDEN=1 go test ./internal/hud/ -run MinimapGolden", goldenPath)
			}
			wantImg, err := png.Decode(bytes.NewReader(want))
			if err != nil {
				t.Fatal(err)
			}
			if !grayEqual(img, wantImg) {
				bad := filepath.Join(dir, base+".got.png")
				_ = os.WriteFile(bad, gotPNG, 0o644)
				t.Fatalf("raster mismatch vs golden; wrote %s", bad)
			}
		})
	}
	if n == 0 {
		t.Fatal("no minimap fixtures found in testdata/minimap")
	}
}

func mustEncodePNG(t *testing.T, img *image.Gray) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func grayEqual(a *image.Gray, b image.Image) bool {
	ab := a.Bounds()
	bb := b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return false
	}
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			ag := a.GrayAt(ab.Min.X+x, ab.Min.Y+y).Y
			r, g, bl, _ := b.At(bb.Min.X+x, bb.Min.Y+y).RGBA()
			bg := uint8((299*r + 587*g + 114*bl) / 1000 >> 8)
			if ag != bg {
				return false
			}
		}
	}
	return true
}
