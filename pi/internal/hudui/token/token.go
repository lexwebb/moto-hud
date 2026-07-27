// Package token holds design spacing and canvas constants (aligned with design/tokens).
package token

// Canonical HUD canvas (ADR 0002).
const (
	Width  = 250
	Height = 122
)

const (
	Pad     = 4
	GapSm   = 2
	GapMd   = 4
	GapLg   = 6
	LegendW = 40
	RuleW   = 1
	GlyphSz = 40
)

func MainWidth() int {
	return Width - Pad*2 - GapMd - RuleW - LegendW
}
