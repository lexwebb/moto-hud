package compose

import (
	"moto-hud/pi/internal/hudui/render/svg"
	"moto-hud/pi/internal/hudui/scene"
)

func roadLineBaselines(x, top int, lines []string) []int {
	face := svg.MustLoadFace(scene.Face8x16)
	out := make([]int, len(lines))
	for i := range lines {
		out[i] = top + i*face.Metrics.CellH + face.Metrics.Ascent
	}
	return out
}

func appendRoadLines(b *scene.Builder, x, top int, lines []string) {
	baselines := roadLineBaselines(x, top, lines)
	for i, ln := range lines {
		b.Text("road", scene.Face8x16, x, baselines[i], "start", ln)
	}
}
