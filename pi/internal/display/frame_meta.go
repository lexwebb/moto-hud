package display

import "image"

// FrameMeta accompanies a framebuffer for hosts that support spatial partial updates.
type FrameMeta struct {
	Spatial bool
	Dirty   image.Rectangle
	Patched bool
}

// FramedScreen can receive refresh metadata (emulator debug, future Waveshare windows).
type FramedScreen interface {
	ShowFrame(img *image.Gray, meta FrameMeta) error
}

// ShowFrame calls ShowFrame on FramedScreen or falls back to Show.
func ShowFrame(s Display, img *image.Gray, meta FrameMeta) error {
	if fs, ok := s.(FramedScreen); ok {
		return fs.ShowFrame(img, meta)
	}
	return s.Show(img)
}
