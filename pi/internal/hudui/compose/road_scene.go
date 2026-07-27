package compose

import (
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
)

func appendRoadLines(b *scene.Builder, x, top int, lines []string) {
	face := svg.MustLoadFace(scene.Face8x16)
	for i, ln := range lines {
		b.Text("road", scene.Face8x16, x, top+i*face.Metrics.CellH+face.Metrics.Ascent, "start", ln)
	}
}
