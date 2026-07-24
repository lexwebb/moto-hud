// Command pixelfont-sheet writes verification PNGs proving each Terminus Bold size
// is 1-bit and aligned to the pixel grid (no gray fringes).
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"moto-hud/pi/internal/pixelfont"
)

func main() {
	outDir := "out/pixelfont"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	metricsPath := filepath.Join(outDir, "METRICS.txt")
	mf, err := os.Create(metricsPath)
	if err != nil {
		fatal(err)
	}
	defer mf.Close()

	fmt.Fprintf(mf, "Terminus Bold bitmap faces — exact BDF metrics (no scaling)\n")
	fmt.Fprintf(mf, "License: SIL OFL 1.1 (see assets/fonts/terminus/OFL.TXT)\n\n")

	for _, sz := range pixelfont.AllSizes() {
		face, err := pixelfont.Load(sz)
		if err != nil {
			fatal(err)
		}
		m := face.Metrics
		fmt.Fprintf(mf, "%s  cell=%dx%d  ascent=%d  descent=%d  pixel_size=%d  glyphs=%d\n",
			sz, m.CellW, m.CellH, m.Ascent, m.Descent, m.PixelSize, m.GlyphCount)

		// Specimen: alphabet + digits on white, 1-bit only
		cols := 32
		rows := 4
		pad := 2
		w := cols*m.CellW + pad*2
		h := rows*(m.CellH+2) + pad*2 + m.Ascent
		img := image.NewGray(image.Rect(0, 0, w, h))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{Y: 255}}, image.Point{}, draw.Src)

		sample := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ m/km ETA High St"
		y := pad + m.Ascent
		face.DrawString(img, pad, y, sample)
		y += m.CellH + 4
		face.DrawString(img, pad, y, "The quick brown fox jumps")
		y += m.CellH + 4
		face.DrawString(img, pad, y, "LEFT 200 m  ETA 12 min")

		// Assert binary
		for py := 0; py < h; py++ {
			for px := 0; px < w; px++ {
				v := img.GrayAt(px, py).Y
				if v != 0 && v != 255 {
					fatal(fmt.Errorf("%s: gray pixel %d at %d,%d", sz, v, px, py))
				}
			}
		}

		path := filepath.Join(outDir, fmt.Sprintf("terminus-%s.png", sz))
		if err := writePNG(path, img); err != nil {
			fatal(err)
		}
		fmt.Println("wrote", path)
	}
	fmt.Println("wrote", metricsPath)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
