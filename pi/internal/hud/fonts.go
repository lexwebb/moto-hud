package hud

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"sync"

	"moto-hud/pi/internal/pixelfont"
)

// FontCandidate is a BDF family under assets/fonts/candidates/ for A/B preview.
type FontCandidate struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Notes string `json:"notes"`
}

type roleFace struct {
	file  string // relative to candidate dir
	scale int    // pixel multiply (1 = native)
}

type candidateSpec struct {
	FontCandidate
	dir      string
	distance roleFace
	road     roleFace
	body     roleFace
	meta     roleFace
}

var (
	candOnce sync.Once
	candRoot string
	cands    []candidateSpec
)

func candidatesRoot() string {
	candOnce.Do(func() {
		hud := resolveAssetDir() // .../assets/hud
		candRoot = filepath.Clean(filepath.Join(hud, "..", "fonts", "candidates"))
		cands = []candidateSpec{
			{
				FontCandidate: FontCandidate{
					ID:    "spleen",
					Name:  "Spleen",
					URL:   "https://github.com/fcambus/spleen",
					Notes: "Previous face — thin at large sizes.",
				},
				dir:      "spleen",
				distance: roleFace{"spleen-16x32.bdf", 1},
				road:     roleFace{"spleen-12x24.bdf", 1},
				body:     roleFace{"spleen-8x16.bdf", 1},
				meta:     roleFace{"spleen-6x12.bdf", 1},
			},
			{
				FontCandidate: FontCandidate{
					ID:    "terminus-bold",
					Name:  "Terminus Bold (selected)",
					URL:   "https://terminus-font.sourceforge.net/",
					Notes: "HUD default — heavier stems; exact 12/16/24/32 sizes.",
				},
				dir:      "terminus-bold",
				distance: roleFace{"ter-u32b.bdf", 1},
				road:     roleFace{"ter-u24b.bdf", 1},
				body:     roleFace{"ter-u16b.bdf", 1},
				meta:     roleFace{"ter-u12b.bdf", 1},
			},
			{
				FontCandidate: FontCandidate{
					ID:    "tamzen-bold",
					Name:  "Tamzen Bold",
					URL:   "https://github.com/sunaku/tamzen-font",
					Notes: "No native 24/32 — hero/road are 8×16 and 6×12 pixel-doubled.",
				},
				dir:      "tamzen-bold",
				distance: roleFace{"Tamzen8x16b.bdf", 2},
				road:     roleFace{"Tamzen10x20b.bdf", 1},
				body:     roleFace{"Tamzen8x16b.bdf", 1},
				meta:     roleFace{"Tamzen6x12b.bdf", 1},
			},
			{
				FontCandidate: FontCandidate{
					ID:    "gohu-bold",
					Name:  "GohuFont Bold",
					URL:   "https://font.gohu.org/",
					Notes: "Single 14px bold face; scaled ×1/×2/×2 for meta/body/hero (chunky).",
				},
				dir:      "gohu-bold",
				distance: roleFace{"gohufont-uni-14b.bdf", 2},
				road:     roleFace{"gohufont-uni-14b.bdf", 2},
				body:     roleFace{"gohufont-uni-14b.bdf", 1},
				meta:     roleFace{"gohufont-uni-14b.bdf", 1},
			},
		}
	})
	return candRoot
}

// ListFontCandidates returns preview metadata (JSON-friendly).
func ListFontCandidates() []FontCandidate {
	_ = candidatesRoot()
	out := make([]FontCandidate, len(cands))
	for i, c := range cands {
		out[i] = c.FontCandidate
	}
	return out
}

func loadRole(c candidateSpec, role roleFace) (*pixelfont.Face, int, error) {
	path := filepath.Join(candidatesRoot(), c.dir, role.file)
	if _, err := os.Stat(path); err != nil {
		return nil, 0, err
	}
	f, err := pixelfont.LoadFile(path)
	if err != nil {
		return nil, 0, err
	}
	return f, role.scale, nil
}

// RenderFontSpecimen draws a 250×122 nav-like mock with the given candidate id.
func RenderFontSpecimen(id string) (*image.Gray, error) {
	_ = candidatesRoot()
	var spec *candidateSpec
	for i := range cands {
		if cands[i].ID == id {
			spec = &cands[i]
			break
		}
	}
	if spec == nil {
		return nil, fmt.Errorf("unknown font candidate %q", id)
	}

	dist, ds, err := loadRole(*spec, spec.distance)
	if err != nil {
		return nil, err
	}
	road, rs, err := loadRole(*spec, spec.road)
	if err != nil {
		return nil, err
	}
	body, bs, err := loadRole(*spec, spec.body)
	if err != nil {
		return nil, err
	}
	meta, ms, err := loadRole(*spec, spec.meta)
	if err != nil {
		return nil, err
	}

	img := image.NewGray(image.Rect(0, 0, Width, Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{Y: 255}}, image.Point{}, draw.Src)

	// Rough left-arrow block (solid, not stroked) so AA isn't a factor.
	for y := 18; y < 46; y++ {
		for x := 40; x < 48; x++ {
			img.SetGray(x, y, color.Gray{Y: 0})
		}
	}
	for y := 28; y < 36; y++ {
		for x := 16; x < 48; x++ {
			img.SetGray(x, y, color.Gray{Y: 0})
		}
	}
	for i := 0; i < 12; i++ {
		for t := -3; t <= 3; t++ {
			img.SetGray(28-i, 32-i+t, color.Gray{Y: 0})
			img.SetGray(28-i, 32+i+t, color.Gray{Y: 0})
		}
	}

	dist.DrawStringRightScaled(img, 244, 30, "200 m", ds)
	road.DrawStringScaled(img, 8, 78, "High St", rs)
	body.DrawStringScaled(img, 8, 102, "ETA 12 min", bs)
	meta.DrawStringScaled(img, 8, 116, spec.ID, ms)
	return img, nil
}
