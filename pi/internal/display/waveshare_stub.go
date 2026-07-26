//go:build !linux

package display

import "fmt"

// NewWaveshare returns a PNG fallback on non-Linux hosts.
func NewWaveshare(pngFallback string) (Display, error) {
	fmt.Println("display: Waveshare unavailable on this OS; using PNG fallback")
	return NewPNG(pngFallback), nil
}
