package hud

import (
	"fmt"
	"strings"

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
	meta := mustFace(pixelfont.Size6x12)
	body := mustFace(pixelfont.Size8x16)
	hero := mustFace(pixelfont.Size16x32)
	mw := mainWidth()
	link := linkMarkSVG(bleLinked)

	headerBottom := pad + meta.Metrics.CellH
	divY := headerBottom + gapSm
	contentTop := divY + gapMd
	contentBot := Height - pad

	if !nav.Active {
		title := "MOTO HUD"
		msg := "Waiting for route..."
		blockH := body.Metrics.CellH*2 + gapMd
		top := contentTop + (contentBot-contentTop-blockH)/2
		b1 := top + body.Metrics.Ascent
		b2 := top + body.Metrics.CellH + gapMd + body.Metrics.Ascent
		var c strings.Builder
		c.WriteString(textSVG("", body, mw/2, b1, "middle", title))
		c.WriteString(textSVG("", body, mw/2, b2, "middle", msg))
		return chromeShell("NAV", link, c.String(), "MEDIA", "-", "STATUS")
	}

	dist := nav.DistanceText
	if dist == "" {
		dist = formatDistance(nav.DistanceM)
	}
	road := nav.Road
	if road == "" {
		road = nav.Instruction
	}
	eta := ""
	if nav.EtaMin > 0 {
		eta = formatETA(nav.EtaMin)
	}

	heroH := glyphSize
	if hero.Metrics.CellH > heroH {
		heroH = hero.Metrics.CellH
	}
	roadH := body.Metrics.CellH
	ribbonH := ribbonDefaultH
	etaH := 0
	if eta != "" {
		etaH = body.Metrics.CellH
	}
	// hero + road + optional ETA + ribbon band (design NavActiveRibbon stretch)
	stackH := heroH + gapLg + roadH + gapLg + ribbonH
	if etaH > 0 {
		stackH += etaH + gapSm
	}
	avail := contentBot - contentTop
	extra := avail - stackH
	if extra < 0 {
		extra = 0
		// Prefer shrinking ribbon over clipping hero/road.
		shrink := stackH - avail
		if shrink > 0 && ribbonH > 24 {
			ribbonH -= shrink
			if ribbonH < 24 {
				ribbonH = 24
			}
		}
	}

	heroTop := contentTop + extra/6
	roadTop := heroTop + heroH + gapLg + extra/6
	y := roadTop + roadH + gapLg + extra/6
	etaTop := y
	if etaH > 0 {
		y += etaH + gapSm
	}
	ribbonTop := contentBot - ribbonH
	if y > ribbonTop {
		ribbonTop = y
		ribbonH = contentBot - ribbonTop
		if ribbonH < 20 {
			ribbonH = 20
			ribbonTop = contentBot - ribbonH
		}
	}

	dist = fit(hero, dist, mw-glyphSize-gapMd)
	road = fit(body, road, mw)
	eta = fit(body, eta, mw)

	glyphY := heroTop + (heroH-glyphSize)/2
	distBaseline := heroTop + (heroH-hero.Metrics.CellH)/2 + hero.Metrics.Ascent
	roadBaseline := roadTop + body.Metrics.Ascent

	pts, turnIdx := schematicRibbonForManeuver(nav.Maneuver)

	var c strings.Builder
	fmt.Fprintf(&c, `<g id="maneuver" transform="translate(-2,%d)" fill="#000" stroke="#000" stroke-width="3" stroke-linecap="square" stroke-linejoin="miter">%s</g>`,
		glyphY, maneuverPaths(nav.Maneuver))
	c.WriteString(textSVG("distance", hero, mw, distBaseline, "end", dist))
	c.WriteString(textSVG("road", body, 0, roadBaseline, "start", road))
	if etaH > 0 {
		c.WriteString(textSVG("eta", body, 0, etaTop+body.Metrics.Ascent, "start", eta))
	}
	fmt.Fprintf(&c, `<g id="ribbon" transform="translate(0,%d)">%s</g>`,
		ribbonTop, roadRibbonSVG(pts, turnIdx, mw, ribbonH))

	return chromeShell("NAV", link, c.String(), "MODE", "-", "MODE")
}

func buildMediaBody(media protocol.MediaMessage, bleLinked bool) map[string]string {
	meta := mustFace(pixelfont.Size6x12)
	body := mustFace(pixelfont.Size8x16)
	titleFace := mustFace(pixelfont.Size12x24)
	mw := mainWidth()
	link := linkMarkSVG(bleLinked)

	playing := "PAUSED"
	action := "PLAY"
	if media.Playing {
		playing = "PLAYING"
		action = "PAUSE"
	}
	title := media.Title
	if title == "" || title == "-" {
		title = "No track"
	}
	artist := media.Artist

	headerBottom := pad + meta.Metrics.CellH
	contentTop := headerBottom + gapSm + gapMd
	contentBot := Height - pad

	playing = fit(meta, playing, mw)
	title = fit(titleFace, title, mw)
	artist = fit(body, artist, mw)

	blockH := meta.Metrics.CellH + gapSm + titleFace.Metrics.CellH + gapSm + body.Metrics.CellH
	top := contentTop + (contentBot-contentTop-blockH)/2
	y1 := top + meta.Metrics.Ascent
	y2 := top + meta.Metrics.CellH + gapSm + titleFace.Metrics.Ascent
	y3 := top + meta.Metrics.CellH + gapSm + titleFace.Metrics.CellH + gapSm + body.Metrics.Ascent

	var c strings.Builder
	c.WriteString(textSVG("playing", meta, 0, y1, "start", playing))
	c.WriteString(textSVG("title", titleFace, 0, y2, "start", title))
	c.WriteString(textSVG("artist", body, 0, y3, "start", artist))
	return chromeShell("MEDIA", link, c.String(), "SKIP", action, "SKIP")
}

func buildStatusBody(bleLinked, navActive bool) map[string]string {
	body := mustFace(pixelfont.Size8x16)
	meta := mustFace(pixelfont.Size6x12)
	mw := mainWidth()
	link := linkMarkSVG(bleLinked)

	ble, nav := "DOWN", "OFF"
	if bleLinked {
		ble = "UP"
	}
	if navActive {
		nav = "ON"
	}

	headerBottom := pad + meta.Metrics.CellH
	contentTop := headerBottom + gapSm + gapMd
	contentBot := Height - pad
	rowH := body.Metrics.CellH + gapMd
	rows := 3
	blockH := rowH*rows - gapMd
	top := contentTop + (contentBot-contentTop-blockH)/2

	var c strings.Builder
	labels := []string{"LINK", "NAV", "PKTS"}
	vals := []string{ble, nav, "OK"}
	for i := 0; i < rows; i++ {
		baseline := top + i*rowH + body.Metrics.Ascent
		c.WriteString(textSVG("", body, 0, baseline, "start", labels[i]))
		c.WriteString(textSVG("", body, mw, baseline, "end", vals[i]))
	}
	return chromeShell("STATUS", link, c.String(), "MODE", "REDRAW", "MODE")
}
