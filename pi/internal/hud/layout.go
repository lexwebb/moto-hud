package hud

import (
	"moto-hud/pi/internal/hudui/compose"
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

func frameVarsFromPlan(in compose.Input) map[string]string {
	sp, err := compose.BuildPlan(in)
	if err != nil {
		vars, _ := compose.FrameVars(in, compose.EmptyPlan())
		if vars != nil {
			return vars
		}
		return map[string]string{"body": ""}
	}
	vars, err := compose.FrameVars(in, sp)
	if err != nil {
		return map[string]string{"body": ""}
	}
	return vars
}

func buildNavBody(nav protocol.NavMessage, bleLinked bool) map[string]string {
	in := ComposeInput(ScreenNav, nav, protocol.MediaMessage{}, bleLinked)
	return frameVarsFromPlan(in)
}

func buildMediaBody(media protocol.MediaMessage, bleLinked bool) map[string]string {
	in := ComposeInput(ScreenMedia, protocol.NavMessage{}, media, bleLinked)
	return frameVarsFromPlan(in)
}

func buildStatusBody(bleLinked, navActive bool) map[string]string {
	nav := protocol.NavMessage{Active: navActive}
	in := ComposeInput(ScreenStatus, nav, protocol.MediaMessage{}, bleLinked)
	return frameVarsFromPlan(in)
}
