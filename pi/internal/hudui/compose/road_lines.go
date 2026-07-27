package compose

import (
	"strings"

	"moto-hud/pi/internal/pixelfont"
)

func roadLinesSVG(deps DrawDeps, face *pixelfont.Face, x, top int, lines []string) string {
	var b strings.Builder
	for i, ln := range lines {
		b.WriteString(deps.TextSVG("road", "8x16", x, top+i*face.Metrics.CellH+face.Metrics.Ascent, "start", ln))
	}
	return b.String()
}
