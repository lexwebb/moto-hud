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
	svg, err := BuildPixelSVG(screen, nav, media, bleLinked)
	if err != nil {
		return fallbackFrame(fmt.Sprintf("svg: %v", err))
	}
	img, err := RasterizeSVG(svg)
	if err != nil {
		return fallbackFrame(fmt.Sprintf("raster: %v", err))
	}
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
	if connected {
		return `<path d="M2,9 L6,2 L6,6 L10,6 L6,10 L6,6" fill="none" stroke="#000" stroke-width="1.6" stroke-linejoin="miter"/>` +
			`<rect x="12" y="4" width="4" height="4" fill="#000"/>`
	}
	return `<line x1="2" y1="2" x2="10" y2="10" stroke="#000" stroke-width="1.6"/>` +
		`<line x1="10" y1="2" x2="2" y2="10" stroke="#000" stroke-width="1.6"/>`
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

// maneuverPaths — design kit glyphs on a 40×40 grid (square caps, filled heads).
func maneuverPaths(m protocol.Maneuver) string {
	switch m {
	case protocol.ManeuverLeft:
		return `<line x1="20" y1="34" x2="20" y2="17"/><line x1="20" y1="17" x2="8" y2="17"/>` +
			`<polygon points="1,17 8,12.7 8,21.3" fill="#000" stroke="none"/>`
	case protocol.ManeuverRight:
		return `<line x1="20" y1="34" x2="20" y2="17"/><line x1="20" y1="17" x2="32" y2="17"/>` +
			`<polygon points="39,17 32,21.3 32,12.7" fill="#000" stroke="none"/>`
	case protocol.ManeuverSlightLeft:
		return `<line x1="21" y1="34" x2="21" y2="24"/><line x1="21" y1="24" x2="11" y2="9"/>` +
			`<polygon points="7,4 15.2,4.3 10.4,11.5" fill="#000" stroke="none"/>`
	case protocol.ManeuverSlightRight:
		return `<line x1="19" y1="34" x2="19" y2="24"/><line x1="19" y1="24" x2="29" y2="9"/>` +
			`<polygon points="33,4 29.6,11.5 24.8,4.3" fill="#000" stroke="none"/>`
	case protocol.ManeuverStraight:
		return `<line x1="20" y1="34" x2="20" y2="14"/>` +
			`<polygon points="20,7 24.3,14 15.7,14" fill="#000" stroke="none"/>`
	case protocol.ManeuverUTurn:
		return `<path fill="none" d="M14,34 V16 A6,6 0 0 1 26,16 V26"/>` +
			`<polygon points="26,33 21.7,26 30.3,26" fill="#000" stroke="none"/>`
	case protocol.ManeuverRoundabout:
		return `<line x1="20" y1="34" x2="20" y2="27"/><circle cx="20" cy="17" r="9" fill="none"/>` +
			`<polygon points="29,8 22,12.3 22,3.7" fill="#000" stroke="none"/>`
	case protocol.ManeuverArrive:
		return `<line x1="13" y1="34" x2="13" y2="7"/><polygon points="13,7 29,11 13,17" fill="#000" stroke="none"/>`
	case protocol.ManeuverDepart:
		return `<circle cx="20" cy="27" r="4" fill="#000" stroke="none"/>` +
			`<line x1="20" y1="22" x2="20" y2="8"/>` +
			`<polygon points="20,4 24.3,11 15.7,11" fill="#000" stroke="none"/>`
	default:
		return `<circle cx="20" cy="20" r="10" fill="none"/><line x1="20" y1="12" x2="20" y2="22"/><circle cx="20" cy="27" r="1.5" fill="#000" stroke="none"/>`
	}
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
