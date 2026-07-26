package hud

import (
	"strings"
	"unicode"

	"moto-hud/pi/internal/pixelfont"
)

// Whole-word road shortenings (USPS C1 + common UK). Keys are lowercase.
// Only exact token matches — "Northumberland" is never touched by "north".
var roadAbbrev = map[string]string{
	"street": "St", "streets": "Sts",
	"avenue": "Ave", "avenues": "Aves",
	"road": "Rd", "roads": "Rds",
	"boulevard": "Blvd",
	"lane": "Ln", "lanes": "Lns",
	"drive": "Dr", "drives": "Drs",
	"place": "Pl",
	"court": "Ct",
	"circle": "Cir",
	"highway": "Hwy",
	"parkway": "Pkwy",
	"expressway": "Expy",
	"freeway": "Fwy",
	"terrace": "Ter",
	"crescent": "Cres",
	"close": "Cl",
	"junction": "Jct",
	"square": "Sq",
	"trail": "Trl",
	"mount": "Mt",
	"mountain": "Mtn",
	"saint": "St",
	"fort": "Ft",
	"bridge": "Br",
	"crossing": "Xing",
	"north": "N", "south": "S", "east": "E", "west": "W",
	"northeast": "NE", "northwest": "NW",
	"southeast": "SE", "southwest": "SW",
}

// abbreviateRoad shortens common road/place tokens; unknown words kept as-is.
func abbreviateRoad(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return s
	}
	out := make([]string, 0, len(fields))
	for _, w := range fields {
		core, lead, trail := splitPunct(w)
		key := strings.ToLower(core)
		if abbr, ok := roadAbbrev[key]; ok {
			core = abbr
		}
		out = append(out, lead+core+trail)
	}
	return strings.Join(out, " ")
}

func splitPunct(w string) (core, lead, trail string) {
	runes := []rune(w)
	start, end := 0, len(runes)
	for start < end {
		r := runes[start]
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			start++
			continue
		}
		break
	}
	for end > start {
		r := runes[end-1]
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			end--
			continue
		}
		break
	}
	return string(runes[start:end]), string(runes[:start]), string(runes[end:])
}

// wrapLines word-wraps s to at most maxLines of width maxW.
// Oversized tokens soft-split across lines when more lines remain; the last line uses fit().
func wrapLines(face *pixelfont.Face, s string, maxW, maxLines int) []string {
	s = strings.TrimSpace(s)
	if s == "" || maxLines <= 0 || maxW <= 0 {
		return nil
	}
	if maxLines == 1 || face.Measure(s) <= maxW {
		return []string{fit(face, s, maxW)}
	}

	words := strings.Fields(s)
	var lines []string
	for len(words) > 0 && len(lines) < maxLines {
		lastLine := len(lines) == maxLines-1
		if lastLine {
			lines = append(lines, fit(face, strings.Join(words, " "), maxW))
			break
		}

		cur := words[0]
		words = words[1:]
		for len(words) > 0 {
			trial := cur + " " + words[0]
			if face.Measure(trial) > maxW {
				break
			}
			cur += " " + words[0]
			words = words[1:]
		}

		if face.Measure(cur) <= maxW {
			lines = append(lines, cur)
			continue
		}

		// Single token wider than the column: soft-split without ellipsis when
		// another line is available for the remainder.
		pre, rest := splitToWidth(face, cur, maxW)
		if pre == "" {
			lines = append(lines, fit(face, cur, maxW))
			break
		}
		lines = append(lines, pre)
		if rest != "" {
			words = append([]string{rest}, words...)
		}
	}
	return lines
}

// splitToWidth returns the longest prefix of s that fits maxW, and the remainder.
func splitToWidth(face *pixelfont.Face, s string, maxW int) (string, string) {
	if s == "" || face.Measure(s) <= maxW {
		return s, ""
	}
	runes := []rune(s)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if face.Measure(string(runes[:mid])) <= maxW {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo == 0 {
		return "", s
	}
	return string(runes[:lo]), string(runes[lo:])
}

func roadBlockHeight(face *pixelfont.Face, lines int) int {
	if lines <= 0 {
		return 0
	}
	return lines*face.Metrics.CellH + (lines-1)*gapSm
}

func roadLinesSVG(id string, face *pixelfont.Face, x, top int, anchor string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	for i, line := range lines {
		baseline := top + i*(face.Metrics.CellH+gapSm) + face.Metrics.Ascent
		lineID := id
		if i > 0 {
			lineID = ""
		}
		b.WriteString(textSVG(lineID, face, x, baseline, anchor, line))
	}
	return b.String()
}
