package hud

import (
	"bytes"
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

var (
	assetOnce sync.Once
	assetDir  string
)

func SetAssetDir(dir string) {
	assetDir = dir
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

// Render fills an SVG template, converts Spleen <text> to pixel rects, then rasterizes.
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
	name := "nav.svg"
	vars := map[string]string{}
	switch screen {
	case ScreenNav:
		name = "nav.svg"
		vars = navVars(nav)
	case ScreenMedia:
		name = "media.svg"
		vars = mediaVars(media)
	case ScreenStatus:
		name = "status.svg"
		vars = statusVars(bleLinked, nav.Active)
	}
	raw, err := os.ReadFile(filepath.Join(resolveAssetDir(), name))
	if err != nil {
		return nil, err
	}
	return []byte(applyVars(string(raw), vars)), nil
}

// BuildPixelSVG is BuildSVG then Spleen text→1×1 rects (browser + Pi share this).
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

func applyVars(svg string, vars map[string]string) string {
	for k, v := range vars {
		if k == "maneuver_paths" {
			continue
		}
		svg = strings.ReplaceAll(svg, "{{"+k+"}}", escapeXML(v))
	}
	if paths, ok := vars["maneuver_paths"]; ok {
		svg = strings.ReplaceAll(svg, "{{maneuver_paths}}", paths)
	}
	return svg
}

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(s)
}

func navVars(nav protocol.NavMessage) map[string]string {
	dist := nav.DistanceText
	if dist == "" {
		dist = formatDistance(nav.DistanceM)
	}
	road := nav.Road
	if road == "" {
		road = truncate(nav.Instruction, 22)
	} else {
		road = truncate(road, 22)
	}
	eta := ""
	if nav.EtaMin > 0 {
		eta = formatETA(nav.EtaMin)
	}
	idle := ""
	if !nav.Active {
		idle = "NAV IDLE"
	}
	return map[string]string{
		"distance": dist, "road": road, "eta": eta, "idle": idle,
		"maneuver_paths": maneuverPaths(nav.Maneuver),
	}
}

func mediaVars(media protocol.MediaMessage) map[string]string {
	playing := "PAUSED"
	if media.Playing {
		playing = "PLAYING"
	}
	return map[string]string{
		"playing": playing,
		"title":   truncate(media.Title, 22),
		"artist":  truncate(media.Artist, 24),
	}
}

func statusVars(bleLinked, navActive bool) map[string]string {
	ble, nav := "BLE: DOWN", "NAV: OFF"
	if bleLinked {
		ble = "BLE: UP"
	}
	if navActive {
		nav = "NAV: ON"
	}
	return map[string]string{"ble": ble, "nav": nav}
}

func maneuverPaths(m protocol.Maneuver) string {
	switch m {
	case protocol.ManeuverLeft, protocol.ManeuverSlightLeft:
		return `<path fill="none" d="M40 10 L40 28 L12 28"/><polyline fill="none" points="22,18 12,28 22,38"/>`
	case protocol.ManeuverRight, protocol.ManeuverSlightRight:
		return `<path fill="none" d="M12 10 L12 28 L40 28"/><polyline fill="none" points="30,18 40,28 30,38"/>`
	case protocol.ManeuverStraight:
		return `<path fill="none" d="M26 44 L26 12"/><polyline fill="none" points="16,22 26,12 36,22"/>`
	case protocol.ManeuverUTurn:
		return `<path fill="none" d="M18 44 L18 20 Q18 8 30 8 Q42 8 42 20 L42 28"/><polyline fill="none" points="34,20 42,28 50,20"/>`
	case protocol.ManeuverRoundabout:
		return `<circle fill="none" cx="28" cy="28" r="14"/><polyline fill="none" points="28,8 28,0 34,6"/>`
	case protocol.ManeuverArrive:
		return `<circle cx="28" cy="28" r="10" fill="#000"/><circle cx="28" cy="28" r="4" fill="#fff"/>`
	case protocol.ManeuverDepart:
		return `<path fill="none" d="M12 28 L40 28"/><polyline fill="none" points="30,18 40,28 30,38"/>`
	default:
		return `<circle cx="28" cy="28" r="12" fill="none"/>`
	}
}

// RasterizeSVG renders a pixel SVG (shapes + Spleen rect glyphs) via canvas,
// then hard-thresholds to 1-bit so stroked glyphs (arrows) don't keep AA gray.
func RasterizeSVG(svg []byte) (*image.Gray, error) {
	c, err := canvas.ParseSVG(bytes.NewReader(svg))
	if err != nil {
		return nil, err
	}
	const scale = 1.0 // already pixel geometry — no supersample
	rgba := rasterizer.Draw(c, canvas.DPMM(scale), canvas.DefaultColorSpace)
	return toGray1Bit(rgba, Width, Height), nil
}

// toGray1Bit nearest-neighbor samples to w×h and snaps every pixel to 0 or 255.
// Bias (< 200 → black) so soft stroke fringes from the canvas rasterizer fill in.
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
	return s[:n-1] + "…"
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
