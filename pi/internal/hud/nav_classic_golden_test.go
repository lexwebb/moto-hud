package hud

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"moto-hud/pi/internal/protocol"
)

func TestNavClassicGolden(t *testing.T) {
	nav := protocol.NavMessage{
		Active:       true,
		DistanceM:    200,
		DistanceText: "200 m",
		Road:         "High St",
		EtaMin:       12,
		Maneuver:     protocol.ManeuverLeft,
	}
	img := Render(ScreenNav, nav, protocol.MediaMessage{}, true)
	dir := filepath.Join("testdata", "nav")
	goldenPath := filepath.Join(dir, "classic-active.png")
	gotPNG := mustEncodePNG(t, img)
	if os.Getenv("UPDATE_NAV_CLASSIC_GOLDEN") == "1" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, gotPNG, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("updated %s", goldenPath)
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("missing golden %s — run UPDATE_NAV_CLASSIC_GOLDEN=1 go test ./internal/hud/ -run NavClassicGolden", goldenPath)
	}
	wantImg, err := png.Decode(bytes.NewReader(want))
	if err != nil {
		t.Fatal(err)
	}
	if !grayEqual(img, wantImg) {
		bad := filepath.Join(dir, "classic-active.got.png")
		_ = os.WriteFile(bad, gotPNG, 0o644)
		t.Fatalf("raster mismatch vs golden; wrote %s", bad)
	}
}
