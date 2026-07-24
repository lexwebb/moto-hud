//go:build linux

package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/inky"
	"periph.io/x/host/v3"
)

type inkyDisplay struct {
	dev  *inky.Dev
	port spi.PortCloser
}

// NewInky opens the Inky pHAT over SPI. Falls back to PNG if hardware init fails.
func NewInky(pngFallback string) (Display, error) {
	if _, err := host.Init(); err != nil {
		fmt.Printf("display: host init failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	port, err := spireg.Open("SPI0.0")
	if err != nil {
		fmt.Printf("display: SPI open failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	dc := gpioreg.ByName("GPIO22")
	reset := gpioreg.ByName("GPIO27")
	busy := gpioreg.ByName("GPIO17")
	if dc == nil || reset == nil || busy == nil {
		_ = port.Close()
		fmt.Println("display: missing GPIO pins; PNG fallback")
		return NewPNG(pngFallback), nil
	}
	dev, err := inky.New(port, dc, reset, busy, &inky.Opts{
		Model:       inky.PHAT,
		ModelColor:  inky.Black,
		BorderColor: inky.Black,
	})
	if err != nil {
		_ = port.Close()
		fmt.Printf("display: inky.New failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	_ = dc.Out(gpio.Low)
	return &inkyDisplay{dev: dev, port: port}, nil
}

func (d *inkyDisplay) Show(img *image.Gray) error {
	// E-ink is 1-bit; dither antialiased gray so edges look smoother than a hard threshold.
	bit := floydSteinberg(img)
	bounds := bit.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, bit, bounds.Min, draw.Src)
	return d.dev.Draw(bounds, dst, image.Point{})
}

func floydSteinberg(src *image.Gray) *image.Gray {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	buf := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf[y*w+x] = float64(src.GrayAt(b.Min.X+x, b.Min.Y+y).Y)
		}
	}
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			old := buf[i]
			neu := 255.0
			if old < 128 {
				neu = 0
			}
			out.SetGray(x, y, color.Gray{Y: uint8(neu)})
			err := old - neu
			if x+1 < w {
				buf[i+1] += err * 7 / 16
			}
			if y+1 < h {
				if x > 0 {
					buf[(y+1)*w+(x-1)] += err * 3 / 16
				}
				buf[(y+1)*w+x] += err * 5 / 16
				if x+1 < w {
					buf[(y+1)*w+(x+1)] += err * 1 / 16
				}
			}
		}
	}
	return out
}

func (d *inkyDisplay) Close() error {
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}
