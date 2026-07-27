package compose

import (
	"fmt"
	"image"
	"strings"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

const glyphSize = 40

// planNavClassic builds the standard nav screen (glyph + distance + road + ribbon).
func planNavClassic(in Input) (plan.ScreenPlan, error) {
	deps := in.NavSVG
	if deps.TextSVG == nil {
		return plan.ScreenPlan{}, fmt.Errorf("compose: NavSVGDeps required")
	}
	nav := in.Nav
	k := Keys{}
	mw := token.MainWidth()
	mainX := token.Pad

	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	hero, _ := pixelfont.Load(pixelfont.Size16x32)

	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad

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
	etaH := 0
	if eta != "" {
		etaH = body.Metrics.CellH
	}
	ribbonH := 40

	roadMaxLines := 2
	roadLines := deps.WrapRoad(road, mw, roadMaxLines)
	roadH := deps.RoadBlockH(len(roadLines))

	stackH := heroH + token.GapLg + roadH + token.GapLg + ribbonH
	if etaH > 0 {
		stackH += etaH + token.GapSm
	}
	avail := contentBot - contentTop
	extra := avail - stackH
	if extra < 0 {
		extra = 0
		shrink := stackH - avail
		if shrink > 0 && ribbonH > 24 {
			ribbonH -= shrink
			if ribbonH < 24 {
				ribbonH = 24
			}
			stackH = heroH + token.GapLg + roadH + token.GapLg + ribbonH
			if etaH > 0 {
				stackH += etaH + token.GapSm
			}
		}
		if avail < stackH && roadMaxLines > 1 {
			roadMaxLines = 1
			roadLines = deps.WrapRoad(road, mw, 1)
			roadH = deps.RoadBlockH(len(roadLines))
			stackH = heroH + token.GapLg + roadH + token.GapLg + ribbonH
			if etaH > 0 {
				stackH += etaH + token.GapSm
			}
		}
		extra = avail - stackH
		if extra < 0 {
			extra = 0
		}
	}

	heroTop := contentTop + extra/6
	roadTop := heroTop + heroH + token.GapLg + extra/6
	y := roadTop + roadH + token.GapLg + extra/6
	etaTop := y
	if etaH > 0 {
		y += etaH + token.GapSm
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

	dist = deps.Fit("16x32", dist, mw-glyphSize-token.GapMd)
	eta = deps.Fit("8x16", eta, mw)
	glyphY := heroTop + (heroH-glyphSize)/2
	distBaseline := heroTop + (heroH-hero.Metrics.CellH)/2 + hero.Metrics.Ascent

	maneuverSlot := image.Rect(mainX-2, glyphY, mainX-2+glyphSize, glyphY+glyphSize)
	distanceSlot := image.Rect(mainX+glyphSize, heroTop, mainX+mw, heroTop+heroH+(contentBot-contentTop)/6)
	if distanceSlot.Max.Y > contentBot {
		distanceSlot.Max.Y = contentBot
	}
	roadSlot := image.Rect(mainX, roadTop, mainX+mw, ribbonTop-token.GapLg)
	etaSlot := image.Rect(mainX, ribbonTop-token.GapLg-16, mainX+mw, ribbonTop)
	ribbonSlot := image.Rect(mainX, ribbonTop, mainX+mw, contentBot)

	ribbonInner := ""
	if deps.RibbonSVG != nil {
		ribbonInner = deps.RibbonSVG(nav, mw, ribbonH)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<g id="maneuver" transform="translate(-2,%d)" fill="#000" stroke="#000" stroke-width="3" stroke-linecap="square" stroke-linejoin="miter">%s</g>`,
		glyphY, deps.ManeuverPaths(nav.Maneuver))
	b.WriteString(deps.TextSVG("distance", "16x32", mw, distBaseline, "end", dist))
	b.WriteString(roadLinesSVG(deps, body, 0, roadTop, roadLines))
	if etaH > 0 {
		b.WriteString(deps.TextSVG("eta", "8x16", 0, etaTop+body.Metrics.Ascent, "start", eta))
	}
	fmt.Fprintf(&b, `<g id="ribbon" transform="translate(0,%d)">%s</g>`, ribbonTop, ribbonInner)

	navCopy := nav
	layers := []plan.Layer{
		{ID: hudui.NodeManeuver, Tier: hudui.TierSlow, Key: k.Maneuver(nav), Slot: maneuverSlot},
		{
			ID: hudui.NodeDistance, Tier: hudui.TierPartialOK, Key: k.DistanceBucket(nav.DistanceM), Slot: distanceSlot,
			Patch: func() ([]byte, error) {
				return patchDistanceSVG(navCopy, distanceSlot.Dx(), distanceSlot.Dy(), deps)
			},
		},
		{
			ID: hudui.NodeRoad, Tier: hudui.TierPartialOK, Key: k.Road(nav), Slot: roadSlot,
			Patch: func() ([]byte, error) {
				return patchRoadSVG(navCopy, roadSlot, deps)
			},
		},
		{
			ID: hudui.NodeETA, Tier: hudui.TierPartialOK, Key: k.ETA(nav), Slot: etaSlot,
			Patch: func() ([]byte, error) {
				return patchETASVG(navCopy, etaSlot.Dx(), etaSlot.Dy(), deps)
			},
		},
		{ID: hudui.NodeRibbon, Tier: hudui.TierSlow, Key: k.Ribbon(nav), Slot: ribbonSlot},
	}

	return plan.ScreenPlan{
		BodySVG:     b.String(),
		Descriptors: plan.BuildDescriptors(k.NavScreen(nav, in.Linked), Keys{}.Bool(in.Linked), layers),
		Layers:      layers,
	}, nil
}

func roadLinesSVG(deps NavSVGDeps, face *pixelfont.Face, x, top int, lines []string) string {
	var b strings.Builder
	for i, ln := range lines {
		b.WriteString(deps.TextSVG("road", "8x16", x, top+i*face.Metrics.CellH+face.Metrics.Ascent, "start", ln))
	}
	return b.String()
}

func formatDistance(m int) string {
	if m >= 1000 {
		return itoa(m/1000) + "." + itoa((m%1000)/100) + " km"
	}
	return itoa(m) + " m"
}

func formatETA(min int) string {
	if min >= 60 {
		return "ETA " + itoa(min/60) + "h " + itoa(min%60) + "m"
	}
	return "ETA " + itoa(min) + " min"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
