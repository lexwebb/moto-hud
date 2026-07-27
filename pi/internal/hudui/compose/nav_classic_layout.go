package compose

import (
	"image"

	"github.com/a-h/templ"

	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

type navClassicLayout struct {
	mw, mainX                                                                 int
	glyphY, heroTop, roadTop, etaTop, ribbonTop, ribbonH, distBaseline, etaH int
	stackBottom                                                                int
	dist, eta                                                                 string
	roadLines                                                                  []string
	maneuverSlot, distanceSlot, roadSlot, etaSlot, ribbonSlot                  image.Rectangle
	ribbonNodes, laneNodes                                                     []scene.Node
	laneY                                                                      int
	maneuver                                                                   protocol.Maneuver
}

const laneStripH = 14

func layoutNavClassic(nav protocol.NavMessage, deps DrawDeps) navClassicLayout {
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

	heroH := token.GlyphSz
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

	dist = deps.Fit(scene.Face16x32, dist, mw-token.GlyphSz-token.GapMd)
	eta = deps.Fit(scene.Face8x16, eta, mw)
	glyphY := heroTop + (heroH-token.GlyphSz)/2
	distBaseline := heroTop + (heroH-hero.Metrics.CellH)/2 + hero.Metrics.Ascent

	maneuverSlot := image.Rect(mainX-2, glyphY, mainX-2+token.GlyphSz, glyphY+token.GlyphSz)
	distanceSlot := image.Rect(mainX+token.GlyphSz, heroTop, mainX+mw, heroTop+heroH+(contentBot-contentTop)/6)
	if distanceSlot.Max.Y > contentBot {
		distanceSlot.Max.Y = contentBot
	}
	roadSlot := image.Rect(mainX, roadTop, mainX+mw, ribbonTop-token.GapLg)
	etaSlot := image.Rect(mainX, ribbonTop-token.GapLg-16, mainX+mw, ribbonTop)
	ribbonSlot := image.Rect(mainX, ribbonTop, mainX+mw, contentBot)

	var ribbonNodes []scene.Node
	if deps.RibbonNodes != nil {
		ribbonNodes = deps.RibbonNodes(nav, mw, ribbonH)
	}

	var laneNodes []scene.Node
	laneY := 0
	if deps.HasLanes != nil && deps.HasLanes(nav) && deps.LaneStripNodes != nil {
		laneY = ribbonTop - laneStripH - token.GapSm - 4
		if laneY < y {
			laneY = y
		}
		laneNodes = deps.LaneStripNodes(nav.Lanes, mw)
	}

	return navClassicLayout{
		mw: mw, mainX: mainX,
		glyphY: glyphY, heroTop: heroTop, roadTop: roadTop, etaTop: etaTop,
		ribbonTop: ribbonTop, ribbonH: ribbonH, distBaseline: distBaseline, etaH: etaH,
		stackBottom: y, dist: dist, eta: eta, roadLines: roadLines,
		maneuverSlot: maneuverSlot, distanceSlot: distanceSlot, roadSlot: roadSlot,
		etaSlot: etaSlot, ribbonSlot: ribbonSlot,
		ribbonNodes: ribbonNodes, laneNodes: laneNodes, laneY: laneY, maneuver: nav.Maneuver,
	}
}

// navClassicBodyNodes builds the classic nav main column via screens/nav_classic.templ.
func navClassicBodyNodes(l navClassicLayout, deps DrawDeps) []scene.Node {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	etaBaseline := l.etaTop + body.Metrics.Ascent
	roadBaselines := roadLineBaselines(0, l.roadTop, l.roadLines)

	var maneuverComp templ.Component = templ.NopComponent
	if deps.ManeuverNodes != nil {
		if nodes := deps.ManeuverNodes(l.maneuver); len(nodes) > 0 {
			maneuverComp = scenetempl.ManeuverAt(l.glyphY, nodes)
		}
	}
	var lanesComp templ.Component = templ.NopComponent
	if len(l.laneNodes) > 0 {
		lanesComp = scenetempl.Nodes([]scene.Node{sceneGroupAt(0, l.laneY, l.laneNodes)})
	}
	var ribbonComp templ.Component = templ.NopComponent
	if len(l.ribbonNodes) > 0 {
		ribbonComp = scenetempl.Nodes(l.ribbonNodes)
	}

	return scenetempl.Render(screens.NavClassicBody(
		l.dist, l.distBaseline, l.mw,
		l.roadLines, roadBaselines, l.eta, etaBaseline, l.etaH > 0,
		maneuverComp, lanesComp, ribbonComp, l.ribbonTop,
	))
}

func sceneGroupAt(dx, dy int, children []scene.Node) scene.Group {
	return scene.Group{DX: dx, DY: dy, Children: children}
}
