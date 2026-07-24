package display

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// Display shows a 250x122 grayscale frame.
type Display interface {
	Show(img *image.Gray) error
	Close() error
}

// PNGDisplay writes each frame to a PNG file (dev / mock).
type PNGDisplay struct {
	Path string
}

func NewPNG(path string) *PNGDisplay {
	return &PNGDisplay{Path: path}
}

func (d *PNGDisplay) Show(img *image.Gray) error {
	if err := os.MkdirAll(filepath.Dir(d.Path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(d.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	fmt.Printf("display: wrote %s\n", d.Path)
	return nil
}

func (d *PNGDisplay) Close() error { return nil }
