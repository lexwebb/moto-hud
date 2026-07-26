//go:build !linux

package display

import "fmt"

// NewLCD returns a PNG fallback on non-Linux hosts.
func NewLCD(pngFallback string) (Display, error) {
	fmt.Println("display: LCD unavailable on this OS; using PNG fallback")
	return NewPNG(pngFallback), nil
}
