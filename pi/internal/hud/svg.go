package hud

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"

	"moto-hud/pi/internal/hudui/compose"
	"moto-hud/pi/internal/pixelfont"
	"moto-hud/pi/internal/protocol"
)

const (
	Width  = 250
	Height = 122
)

//go:embed assets/frame.svg
var embeddedFS embed.FS

var (
	assetOnce sync.Once
	assetDir  string
)

// Raw SVG fragment placeholders (not XML-escaped).
var rawSVGKeys = map[string]bool{
	"body": true,
}

func SetAssetDir(dir string) {
	assetDir = dir
}

func loadFrameSVG() ([]byte, error) {
	if dir := resolveAssetDir(); dir != "" {
		path := filepath.Join(dir, "frame.svg")
		if b, err := os.ReadFile(path); err == nil {
			return b, nil
		}
	}
	return embeddedFS.ReadFile("assets/frame.svg")
}

func resolveAssetDir() string {
	assetOnce.Do(func() {
		if assetDir != "" {
			return
		}
		candidates := []string{"assets/hud", "../assets/hud", "../../assets/hud"}
		if exe, err := os.Executable(); err == nil {
			candidates = append([]string{
				filepath.Join(filepath.Dir(exe), "assets/hud"),
				filepath.Join(filepath.Dir(exe), "../assets/hud"),
			}, candidates...)
		}
		for _, c := range candidates {
			if st, err := os.Stat(filepath.Join(c, "nav.svg")); err == nil && !st.IsDir() {
				assetDir = c
				return
			}
		}
		assetDir = "assets/hud"
	})
	return assetDir
}

// Render fills an SVG template, converts Terminus <text> to pixel rects, then rasterizes.
func Render(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, bleLinked bool) *image.Gray {
	in := ComposeInput(screen, nav, media, bleLinked)
	sp, err := compose.BuildPlan(in)
	if err != nil {
		return fallbackFrame(fmt.Sprintf("plan: %v", err))
	}
	svg, err := BuildPixelSVG(screen, nav, media, bleLinked)
	if err != nil {
		return fallbackFrame(fmt.Sprintf("svg: %v", err))
	}
	img, err := RasterizeSVG(svg)
	if err != nil {
		return fallbackFrame(fmt.Sprintf("raster: %v", err))
	}
	applyLinkLayer(img, sp)
	return img
}

// BuildSVG returns filled SVG markup with <text> placeholders resolved (designer form).
func BuildSVG(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, bleLinked bool) ([]byte, error) {
	vars := map[string]string{}
	switch screen {
	case ScreenNav:
		vars = buildNavBody(nav, bleLinked)
	case ScreenMedia:
		vars = buildMediaBody(media, bleLinked)
	case ScreenStatus:
		vars = buildStatusBody(bleLinked, nav.Active)
	default:
		vars = buildNavBody(nav, bleLinked)
	}
	raw, err := loadFrameSVG()
	if err != nil {
		return nil, err
	}
	return []byte(applyVars(string(raw), vars)), nil
}

// BuildPixelSVG is BuildSVG then Terminus text→1×1 rects (browser + Pi share this).
func BuildPixelSVG(screen Screen, nav protocol.NavMessage, media protocol.MediaMessage, bleLinked bool) ([]byte, error) {
	svg, err := BuildSVG(screen, nav, media, bleLinked)
	if err != nil {
		return nil, err
	}
	out, err := pixelfont.ReplaceSVGText(string(svg))
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// BuildPixelSVGFromVars rasterizes a pre-filled template vars map (hudui plan path).
func BuildPixelSVGFromVars(vars map[string]string) ([]byte, error) {
	raw, err := loadFrameSVG()
	if err != nil {
		return nil, err
	}
	svg := applyVars(string(raw), vars)
	out, err := pixelfont.ReplaceSVGText(svg)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func applyVars(svg string, vars map[string]string) string {
	for k, v := range vars {
		if rawSVGKeys[k] {
			continue
		}
		svg = strings.ReplaceAll(svg, "{{"+k+"}}", escapeXML(v))
	}
	for k := range rawSVGKeys {
		if v, ok := vars[k]; ok {
			svg = strings.ReplaceAll(svg, "{{"+k+"}}", v)
		}
	}
	return svg
}

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(s)
}

func linkMarkSVG(connected bool) string {
	return compose.LinkMarkFragment(connected)
}

// progressTicks draws 5 coarse distance ticks (design ProgressTicks).
func progressTicks(distanceM int) string {
	filled := 0
	switch {
	case distanceM <= 20:
		filled = 5
	case distanceM <= 50:
		filled = 4
	case distanceM <= 100:
		filled = 3
	case distanceM <= 200:
		filled = 2
	case distanceM <= 500:
		filled = 1
	}
	var b strings.Builder
	for i := 0; i < 5; i++ {
		x := i * 12
		if i < filled {
			fmt.Fprintf(&b, `<rect x="%d" y="0" width="8" height="8" fill="#000"/>`, x)
		} else {
			fmt.Fprintf(&b, `<rect x="%d" y="0" width="8" height="8" fill="none" stroke="#000" stroke-width="1"/>`, x)
		}
	}
	return b.String()
}

// RasterizeSVG renders a pixel SVG (shapes + Terminus rect glyphs) via canvas,
// then hard-thresholds to 1-bit so stroked glyphs (arrows) don't keep AA gray.
func RasterizeSVG(svg []byte) (*image.Gray, error) {
	return RasterizeSVGAt(svg, Width, Height)
}

// RasterizeSVGAt is RasterizeSVG at an arbitrary pane size (e.g. minimap lab).
func RasterizeSVGAt(svg []byte, w, h int) (*image.Gray, error) {
	if w <= 0 {
		w = Width
	}
	if h <= 0 {
		h = Height
	}
	c, err := canvas.ParseSVG(bytes.NewReader(svg))
	if err != nil {
		return nil, err
	}
	const scale = 1.0
	rgba := rasterizer.Draw(c, canvas.DPMM(scale), canvas.DefaultColorSpace)
	return toGray1Bit(rgba, w, h), nil
}

func toGray1Bit(src image.Image, w, h int) *image.Gray {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y*sh/h
		if sy >= b.Max.Y {
			sy = b.Max.Y - 1
		}
		for x := 0; x < w; x++ {
			sx := b.Min.X + x*sw/w
			if sx >= b.Max.X {
				sx = b.Max.X - 1
			}
			r, g, bl, _ := src.At(sx, sy).RGBA()
			lum := (299*r + 587*g + 114*bl) / 1000 >> 8
			y8 := uint8(255)
			if lum < 200 {
				y8 = 0
			}
			out.SetGray(x, y, color.Gray{Y: y8})
		}
	}
	return out
}

func fallbackFrame(msg string) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, Width, Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{Y: 255}}, image.Point{}, draw.Src)
	_ = msg
	return img
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "..."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
