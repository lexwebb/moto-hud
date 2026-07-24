// Package pixelfont loads Spleen BDF bitmap fonts and blits 1-bit glyphs
// at integer pixel coordinates only (no scaling, no antialiasing).
package pixelfont

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

//go:embed data/*.bdf
var bdfFS embed.FS

// Size is a fixed Spleen face. Cell is glyph cell WxH; use only these sizes.
type Size int

const (
	Size6x12  Size = 12 // spleen-6x12 — meta / footers
	Size8x16  Size = 16 // spleen-8x16 — secondary lines
	Size12x24 Size = 24 // spleen-12x24 — road / titles
	Size16x32 Size = 32 // spleen-16x32 — distance hero
)

func (s Size) file() string {
	switch s {
	case Size6x12:
		return "data/spleen-6x12.bdf"
	case Size8x16:
		return "data/spleen-8x16.bdf"
	case Size12x24:
		return "data/spleen-12x24.bdf"
	case Size16x32:
		return "data/spleen-16x32.bdf"
	default:
		return ""
	}
}

func (s Size) String() string {
	switch s {
	case Size6x12:
		return "6x12"
	case Size8x16:
		return "8x16"
	case Size12x24:
		return "12x24"
	case Size16x32:
		return "16x32"
	default:
		return "?"
	}
}

// Metrics are confirmed pixel metrics from the BDF (no scaling).
type Metrics struct {
	Name       string
	CellW      int
	CellH      int
	Ascent     int
	Descent    int
	PixelSize  int
	GlyphCount int
}

// Face is a loaded bitmap face.
type Face struct {
	Metrics Metrics
	glyphs  map[rune]glyph
	replace rune
}

type glyph struct {
	advance  int
	w, h     int
	xOff     int
	yOff     int
	bits     []byte
	rowBytes int
}

var cache = map[Size]*Face{}

// Load returns a cached Face for a fixed size.
func Load(size Size) (*Face, error) {
	if f, ok := cache[size]; ok {
		return f, nil
	}
	name := size.file()
	if name == "" {
		return nil, fmt.Errorf("unknown size %d", size)
	}
	raw, err := bdfFS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	f, err := parseBDF(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", size, err)
	}
	cache[size] = f
	return f, nil
}

func parseBDF(raw []byte) (*Face, error) {
	f := &Face{
		glyphs:  make(map[rune]glyph),
		replace: '?',
	}
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		curName                string
		enc                    int
		dwidth                 int
		bbxW, bbxH, bbxX, bbxY int
		inBitmap               bool
		rows                   []string
	)

	flush := func() error {
		if enc == 0 && len(rows) == 0 {
			curName, inBitmap, rows = "", false, nil
			return nil
		}
		if enc < 0 {
			curName, inBitmap, rows = "", false, nil
			return nil
		}
		g, err := decodeBitmap(rows, bbxW, bbxH, bbxX, bbxY, dwidth)
		if err != nil {
			return err
		}
		f.glyphs[rune(enc)] = g
		curName, enc, inBitmap, rows = "", 0, false, nil
		return nil
	}

	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "FONTBOUNDINGBOX "):
			fmt.Sscanf(line, "FONTBOUNDINGBOX %d %d", &f.Metrics.CellW, &f.Metrics.CellH)
		case strings.HasPrefix(line, "PIXEL_SIZE "):
			fmt.Sscanf(line, "PIXEL_SIZE %d", &f.Metrics.PixelSize)
		case strings.HasPrefix(line, "FONT_ASCENT "):
			fmt.Sscanf(line, "FONT_ASCENT %d", &f.Metrics.Ascent)
		case strings.HasPrefix(line, "FONT_DESCENT "):
			fmt.Sscanf(line, "FONT_DESCENT %d", &f.Metrics.Descent)
		case strings.HasPrefix(line, "FAMILY_NAME "):
			f.Metrics.Name = strings.Trim(strings.TrimPrefix(line, "FAMILY_NAME "), "\"")
		case strings.HasPrefix(line, "STARTCHAR "):
			if err := flush(); err != nil {
				return nil, err
			}
			curName = strings.TrimPrefix(line, "STARTCHAR ")
		case strings.HasPrefix(line, "ENCODING "):
			fmt.Sscanf(line, "ENCODING %d", &enc)
		case strings.HasPrefix(line, "DWIDTH "):
			fmt.Sscanf(line, "DWIDTH %d", &dwidth)
		case strings.HasPrefix(line, "BBX "):
			fmt.Sscanf(line, "BBX %d %d %d %d", &bbxW, &bbxH, &bbxX, &bbxY)
		case line == "BITMAP":
			inBitmap = true
			rows = rows[:0]
		case line == "ENDCHAR":
			inBitmap = false
			if err := flush(); err != nil {
				return nil, err
			}
		case inBitmap:
			rows = append(rows, strings.TrimSpace(line))
		}
		_ = curName
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	f.Metrics.GlyphCount = len(f.glyphs)
	if f.Metrics.CellW == 0 || f.Metrics.CellH == 0 {
		return nil, fmt.Errorf("missing FONTBOUNDINGBOX")
	}
	return f, nil
}

func decodeBitmap(rows []string, w, h, xOff, yOff, advance int) (glyph, error) {
	if h > 0 && len(rows) < h {
		return glyph{}, fmt.Errorf("bitmap rows %d < height %d", len(rows), h)
	}
	if h > 0 && len(rows) > h {
		rows = rows[:h]
	}
	rowBytes := (w + 7) / 8
	bits := make([]byte, rowBytes*h)
	for y, hex := range rows {
		for i := 0; i < rowBytes; i++ {
			if len(hex) < (i+1)*2 {
				break
			}
			v, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
			if err != nil {
				return glyph{}, err
			}
			bits[y*rowBytes+i] = byte(v)
		}
	}
	if advance == 0 {
		advance = w
	}
	return glyph{advance: advance, w: w, h: h, xOff: xOff, yOff: yOff, bits: bits, rowBytes: rowBytes}, nil
}

func (g glyph) on(x, y int) bool {
	if x < 0 || y < 0 || x >= g.w || y >= g.h {
		return false
	}
	b := g.bits[y*g.rowBytes+x/8]
	return b&(0x80>>uint(x%8)) != 0
}

// Measure returns pixel width of s (integer advances only).
func (f *Face) Measure(s string) int {
	w := 0
	for _, r := range s {
		g, ok := f.glyphs[r]
		if !ok {
			g = f.glyphs[f.replace]
		}
		w += g.advance
	}
	return w
}

// DrawString blits s in solid black. baselineY is baseline in image Y-down coords.
func (f *Face) DrawString(dst *image.Gray, x, baselineY int, s string) {
	pen := x
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
				px, py := left+gx, top+gy
				if px < dst.Rect.Min.X || py < dst.Rect.Min.Y || px >= dst.Rect.Max.X || py >= dst.Rect.Max.Y {
					continue
				}
				dst.SetGray(px, py, color.Gray{Y: 0})
			}
		}
		pen += g.advance
	}
}

// DrawStringRight draws s right-aligned so its right edge is at rightX.
func (f *Face) DrawStringRight(dst *image.Gray, rightX, baselineY int, s string) {
	f.DrawString(dst, rightX-f.Measure(s), baselineY, s)
}

// AllSizes lists supported pregenerated faces.
func AllSizes() []Size {
	return []Size{Size6x12, Size8x16, Size12x24, Size16x32}
}
