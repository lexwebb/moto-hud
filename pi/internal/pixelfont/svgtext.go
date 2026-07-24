package pixelfont

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var textElemRe = regexp.MustCompile(`(?s)<text\b([^>]*)>(.*?)</text>`)

// ReplaceSVGText turns every <text> into a <g> of 1×1 pixel rects using Spleen BDFs.
// Layout attributes (x, y=baseline, text-anchor, data-spleen / font-size) drive placement.
// The result is identical in browser preview and canvas rasterization.
func ReplaceSVGText(svg string) (string, error) {
	var firstErr error
	out := textElemRe.ReplaceAllStringFunc(svg, func(full string) string {
		m := textElemRe.FindStringSubmatch(full)
		if m == nil {
			return full
		}
		attrs, content := m[1], htmlUnescape(strings.TrimSpace(m[2]))
		if content == "" {
			return ""
		}
		size, err := sizeFromAttrs(attrs)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return full
		}
		face, err := Load(size)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return full
		}
		x := attrFloat(attrs, "x", 0)
		y := attrFloat(attrs, "y", 0) // SVG text y = baseline
		anchor := attrString(attrs, "text-anchor", "start")
		fill := attrString(attrs, "fill", "#000")
		id := attrString(attrs, "id", "")

		penX := int(x + 0.5)
		if anchor == "end" {
			penX = int(x+0.5) - face.Measure(content)
		} else if anchor == "middle" {
			penX = int(x+0.5) - face.Measure(content)/2
		}
		baseline := int(y + 0.5)

		var b strings.Builder
		if id != "" {
			fmt.Fprintf(&b, `<g id="%s" data-spleen="%s" data-baseline="%d">`, escapeAttr(id), size, baseline)
		} else {
			fmt.Fprintf(&b, `<g data-spleen="%s" data-baseline="%d">`, size, baseline)
		}
		b.WriteString(face.StringToSVG(content, penX, baseline, fill))
		b.WriteString(`</g>`)
		return b.String()
	})
	return out, firstErr
}

// StringToSVG emits 1×1 rects for s. baselineY is SVG/image Y-down baseline.
func (f *Face) StringToSVG(s string, penX, baselineY int, fill string) string {
	var b strings.Builder
	pen := penX
	for _, r := range s {
		g, ok := f.glyphs[r]
		if !ok {
			g, ok = f.glyphs[f.replace]
			if !ok {
				continue
			}
		}
		top := baselineY - (g.h + g.yOff)
		left := pen + g.xOff
		for gy := 0; gy < g.h; gy++ {
			for gx := 0; gx < g.w; gx++ {
				if !g.on(gx, gy) {
					continue
				}
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="1" height="1" fill="%s"/>`, left+gx, top+gy, fill)
			}
		}
		pen += g.advance
	}
	return b.String()
}

func sizeFromAttrs(attrs string) (Size, error) {
	if s := attrString(attrs, "data-spleen", ""); s != "" {
		switch s {
		case "6x12":
			return Size6x12, nil
		case "8x16":
			return Size8x16, nil
		case "12x24":
			return Size12x24, nil
		case "16x32":
			return Size16x32, nil
		default:
			return 0, fmt.Errorf("unknown data-spleen %q", s)
		}
	}
	fs := attrFloat(attrs, "font-size", 16)
	switch int(fs + 0.5) {
	case 12:
		return Size6x12, nil
	case 16:
		return Size8x16, nil
	case 24:
		return Size12x24, nil
	case 32:
		return Size16x32, nil
	// legacy mappings from earlier Liberation sizes
	case 11, 13:
		return Size6x12, nil
	case 17, 18:
		return Size12x24, nil
	case 26, 28:
		return Size16x32, nil
	default:
		return 0, fmt.Errorf("unsupported font-size %g (use 12/16/24/32 or data-spleen)", fs)
	}
}

func attrFloat(attrs, name string, def float64) float64 {
	re := regexp.MustCompile(name + `="([^"]+)"`)
	m := re.FindStringSubmatch(attrs)
	if m == nil {
		return def
	}
	s := strings.TrimSuffix(strings.TrimSpace(m[1]), "px")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func attrString(attrs, name, def string) string {
	re := regexp.MustCompile(name + `="([^"]+)"`)
	m := re.FindStringSubmatch(attrs)
	if m == nil {
		return def
	}
	return m[1]
}

func htmlUnescape(s string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">", "&quot;", "\"", "&amp;", "&", "&apos;", "'").Replace(s)
}

func escapeAttr(s string) string {
	return strings.NewReplacer(`&`, "&amp;", `"`, "&quot;", `<`, "&lt;").Replace(s)
}
