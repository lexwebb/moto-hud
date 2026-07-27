package hud

import (
	"fmt"

	"moto-hud/pi/internal/pixelfont"
)

func mustFace(sz pixelfont.Size) *pixelfont.Face {
	f, err := pixelfont.Load(sz)
	if err != nil {
		panic(err)
	}
	return f
}

// fit truncates s so Measure(s) <= maxW (ASCII ellipsis).
func fit(face *pixelfont.Face, s string, maxW int) string {
	if maxW <= 0 || face.Measure(s) <= maxW {
		return s
	}
	ellipsis := "..."
	ew := face.Measure(ellipsis)
	if ew > maxW {
		return ""
	}
	for len(s) > 0 {
		s = s[:len(s)-1]
		if face.Measure(s+ellipsis) <= maxW {
			return s + ellipsis
		}
	}
	return ellipsis
}

func textSVG(id string, face *pixelfont.Face, x, baseline int, anchor, s string) string {
	if s == "" {
		return ""
	}
	sz := face.Metrics.PixelSize
	attrs := fmt.Sprintf(`x="%d" y="%d" data-pixel="%s" font-size="%d" fill="#000"`, x, baseline, faceSizeAttr(sz), sz)
	if id != "" {
		attrs = fmt.Sprintf(`id="%s" %s`, id, attrs)
	}
	if anchor != "" && anchor != "start" {
		attrs += fmt.Sprintf(` text-anchor="%s"`, anchor)
	}
	return fmt.Sprintf(`<text %s>%s</text>`, attrs, escapeXML(s))
}

func faceSizeAttr(pixelSize int) string {
	switch pixelSize {
	case 12:
		return "6x12"
	case 16:
		return "8x16"
	case 24:
		return "12x24"
	case 32:
		return "16x32"
	default:
		return "8x16"
	}
}
