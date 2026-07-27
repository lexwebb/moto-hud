package hud

import (
	"fmt"
	"strings"

	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

// Panel geometry — integer pixels only (design spacing scale 2/4/6/8).
const (
	pad       = 4
	gapSm     = 2
	gapMd     = 4
	gapLg     = 6
	legendW   = 40
	ruleW     = 1
	glyphSize = 40
)

func mainWidth() int {
	return Width - pad*2 - gapMd - ruleW - legendW
}

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

func chromeShell(mode, link, content, legPrev, legAction, legNext string) map[string]string {
	meta := mustFace(pixelfont.Size6x12)
	mw := mainWidth()

	headerBaseline := pad + meta.Metrics.Ascent
	headerBottom := pad + meta.Metrics.CellH
	divY := headerBottom + gapSm

	legTop := headerBaseline
	legMid := Height / 2
	legBot := Height - pad - meta.Metrics.Descent

	ruleX := pad + mw + gapMd

	var b strings.Builder
	// Header rule meets the legend rule (T-junction, no gap).
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`, pad, divY, ruleX, divY)
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`,
		ruleX, pad, ruleX, Height-pad)

	fmt.Fprintf(&b, `<g id="main" transform="translate(%d,0)">`, pad)
	b.WriteString(textSVG("mode", meta, 0, headerBaseline, "start", mode))
	fmt.Fprintf(&b, `<g id="link" transform="translate(%d,%d)">%s</g>`, mw-16, pad+1, link)
	b.WriteString(content)
	b.WriteString(`</g>`)

	b.WriteString(textSVG("leg_prev", meta, Width-pad, legTop, "end", legPrev))
	b.WriteString(textSVG("leg_action", meta, Width-pad, legMid, "end", legAction))
	b.WriteString(textSVG("leg_next", meta, Width-pad, legBot, "end", legNext))

	return map[string]string{"body": b.String()}
}

func buildNavBody(nav protocol.NavMessage, bleLinked bool) map[string]string {
	link := linkMarkSVG(bleLinked)
	in := ComposeInput(ScreenNav, nav, protocol.MediaMessage{}, bleLinked)
	sp, err := compose.BuildPlan(in)
	if err != nil {
		return chromeShell("NAV", link, "", "MEDIA", "-", "STATUS")
	}
	legPrev, legAction, legNext := "MEDIA", "-", "STATUS"
	if nav.Active {
		legPrev, legAction, legNext = "MODE", "-", "MODE"
	}
	return chromeShell("NAV", link, sp.BodySVG, legPrev, legAction, legNext)
}

func buildMediaBody(media protocol.MediaMessage, bleLinked bool) map[string]string {
	link := linkMarkSVG(bleLinked)
	action := "PLAY"
	if media.Playing {
		action = "PAUSE"
	}
	in := ComposeInput(ScreenMedia, protocol.NavMessage{}, media, bleLinked)
	sp, err := compose.BuildPlan(in)
	if err != nil {
		return chromeShell("MEDIA", link, "", "SKIP", action, "SKIP")
	}
	return chromeShell("MEDIA", link, sp.BodySVG, "SKIP", action, "SKIP")
}

func buildStatusBody(bleLinked, navActive bool) map[string]string {
	link := linkMarkSVG(bleLinked)
	nav := protocol.NavMessage{Active: navActive}
	in := ComposeInput(ScreenStatus, nav, protocol.MediaMessage{}, bleLinked)
	sp, err := compose.BuildPlan(in)
	if err != nil {
		return chromeShell("STATUS", link, "", "MODE", "REDRAW", "MODE")
	}
	return chromeShell("STATUS", link, sp.BodySVG, "MODE", "REDRAW", "MODE")
}
