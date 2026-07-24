//go:build !linux

package display

import "fmt"

// NewInky returns a PNG fallback on non-Linux hosts.
func NewInky(pngFallback string) (Display, error) {
	fmt.Println("display: Inky unavailable on this OS; using PNG fallback")
	return NewPNG(pngFallback), nil
}
