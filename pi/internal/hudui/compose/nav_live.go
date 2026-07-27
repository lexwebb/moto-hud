package compose

import (
	"fmt"
	"image"
	"strings"

	"github.com/a-h/templ"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/scenetempl"
	"moto-hud/pi/internal/hudui/screens"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

// planNavLive: left corridor (minimap or ribbon), right column distance + road + ETA.
func planNavLive(in Input) (plan.ScreenPlan, error) {
	deps := in.NavSVG
	if deps.TextSVG == nil {
		return plan.ScreenPlan{}, fmt.Errorf("compose: DrawDeps required")
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
	dist = compactDistanceText(dist)
	road := nav.Road
	if road == "" {
		road = nav.Instruction
	}
	eta := ""
	if nav.EtaMin > 0 {
		eta = formatETA(nav.EtaMin)
	}

	leftW := (mw * 44) / 100
	if leftW < 72 {
		leftW = 72
	}
	rightX := leftW + token.GapMd
	rightW := mw - rightX
	if rightW < 60 {
		rightW = 60
		rightX = mw - rightW
		leftW = rightX - token.GapMd
	}
	ribbonH := contentBot - contentTop
	if ribbonH < 20 {
		ribbonH = 20
	}

	dist = deps.Fit(scene.Face16x32, dist, rightW)
	eta = deps.Fit(scene.Face8x16, eta, rightW)

	roadMaxLines := 3
	roadLines := deps.WrapRoad(road, rightW, roadMaxLines)
	roadH := deps.RoadBlockH(len(roadLines))

	distTop := contentTop
	distBaseline := distTop + hero.Metrics.Ascent
	roadTop := distTop + hero.Metrics.CellH + token.GapMd
	etaTop := roadTop + roadH + token.GapSm
	if eta != "" {
		if etaTop+body.Metrics.CellH > contentBot {
			roadMaxLines = 2
			roadLines = deps.WrapRoad(road, rightW, roadMaxLines)
			roadH = deps.RoadBlockH(len(roadLines))
			etaTop = roadTop + roadH + token.GapSm
		}
	}

	var leftNodes []scene.Node
	// Junction IR templates (PreferJunctionTemplates) sit beside the meter minimap —
	// do not remove MinimapNodes until the flag is the live default.
	if deps.HasJunction != nil && deps.HasJunction(nav) && deps.JunctionNodes != nil {
		leftNodes = deps.JunctionNodes(nav, leftW, ribbonH)
	} else if deps.HasMinimap != nil && deps.HasMinimap(nav) && deps.MinimapNodes != nil {
		leftNodes = deps.MinimapNodes(nav.Minimap, leftW, ribbonH)
	} else if deps.RibbonNodes != nil {
		leftNodes = deps.RibbonNodes(nav, leftW, ribbonH)
	}
	var leftCol templ.Component = templ.NopComponent
	if len(leftNodes) > 0 {
		leftCol = scenetempl.Nodes([]scene.Node{scene.Group{ID: "ribbon", DY: contentTop, Children: leftNodes}})
	}

	var lanesCol templ.Component = templ.NopComponent
	if deps.HasLanes != nil && deps.HasLanes(nav) && deps.LaneStripNodes != nil {
		laneY := contentBot - laneStripH - 2
		lanesCol = scenetempl.Nodes([]scene.Node{scene.Group{DX: rightX, DY: laneY, Children: deps.LaneStripNodes(nav.Lanes, rightW)}})
	}
	roadBaselines := roadLineBaselines(rightX, roadTop, roadLines)
	etaBaseline := etaTop + body.Metrics.Ascent
	bodyNodes := scenetempl.Render(screens.NavLiveBody(
		leftCol, lanesCol, dist, distBaseline, mw, rightX,
		roadLines, roadBaselines, eta, etaBaseline, eta != "",
	))
	distanceSlot := image.Rect(mainX+rightX, distTop, mainX+mw, distTop+hero.Metrics.CellH)
	roadBottom := contentBot
	if eta != "" {
		roadBottom = etaTop - token.GapSm
	}
	roadSlot := image.Rect(mainX+rightX, roadTop, mainX+mw, roadBottom)
	var etaSlot image.Rectangle
	if eta != "" {
		etaSlot = image.Rect(mainX+rightX, etaTop, mainX+mw, etaTop+body.Metrics.CellH)
	}

	ribbonSlot := image.Rect(mainX, contentTop, mainX+leftW, contentTop+ribbonH)
	navCopy := nav
	layers := []plan.Layer{
		{ID: hudui.NodeRibbon, Tier: hudui.TierSlow, Key: k.Ribbon(nav), Slot: ribbonSlot},
		{
			ID: hudui.NodeDistance, Tier: hudui.TierPartialOK, Key: k.DistanceBucket(nav.DistanceM), Slot: distanceSlot,
			Patch: func() (scene.Document, error) {
				return patchDistanceDoc(navCopy, distanceSlot.Dx(), distanceSlot.Dy(), deps)
			},
		},
		{
			ID: hudui.NodeRoad, Tier: hudui.TierPartialOK, Key: k.Road(nav), Slot: roadSlot,
			Patch: func() (scene.Document, error) {
				return patchRoadDoc(navCopy, roadSlot, deps)
			},
		},
	}
	if eta != "" {
		layers = append(layers, plan.Layer{
			ID: hudui.NodeETA, Tier: hudui.TierPartialOK, Key: k.ETA(nav), Slot: etaSlot,
			Patch: func() (scene.Document, error) {
				return patchETADoc(navCopy, etaSlot.Dx(), etaSlot.Dy(), deps)
			},
		})
	}

	return finalizePlan(in, k.NavScreen(nav), staticChromeKey(), bodyNodes, layers), nil
}

// CompactDistanceText drops spaces in distance strings for the live nav column.
func CompactDistanceText(s string) string {
	return compactDistanceText(s)
}

func compactDistanceText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
