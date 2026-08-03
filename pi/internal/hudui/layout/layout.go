package layout

import "image"

// Align cross-axis alignment.
type Align int

const (
	AlignStart Align = iota
	AlignEnd
	AlignCenter
	AlignBaseline
)

// Justify main-axis distribution.
type Justify int

const (
	JustifyStart Justify = iota
	JustifyEnd
	JustifyCenter
	JustifySpaceBetween
)

// Box is a flex container with integer pixel geometry.
type Box struct {
	X, Y, W, H int
	Pad        int
	Gap        int
	DirRow     bool
	Justify    Justify
	Align      Align
	Children   []Child
}

// Child is one flex item.
type Child struct {
	W, H   int // 0 = intrinsic / fill
	Flex   int // grow weight on main axis
	Offset image.Point
	Size   image.Point
}

// Layout runs a simplified flex pass; returns child placements relative to box origin.
func (b Box) Layout(measure func(i int) (w, h int)) []Child {
	out := make([]Child, len(b.Children))
	innerW := b.W - 2*b.Pad
	innerH := b.H - 2*b.Pad
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	mainIsW := b.DirRow
	var totalMain, totalCross, flexSum int
	sizes := make([]image.Point, len(b.Children))
	for i := range b.Children {
		w, h := b.Children[i].W, b.Children[i].H
		if w == 0 || h == 0 {
			mw, mh := measure(i)
			if w == 0 {
				w = mw
			}
			if h == 0 {
				h = mh
			}
		}
		sizes[i] = image.Pt(w, h)
		if mainIsW {
			totalMain += w
			if h > totalCross {
				totalCross = h
			}
		} else {
			totalMain += h
			if w > totalCross {
				totalCross = w
			}
		}
		flexSum += b.Children[i].Flex
	}
	gap := b.Gap
	if len(b.Children) > 1 {
		totalMain += gap * (len(b.Children) - 1)
	}

	extra := 0
	if mainIsW {
		extra = innerW - totalMain
	} else {
		extra = innerH - totalMain
	}
	if extra < 0 {
		extra = 0
	}
	if flexSum > 0 && extra > 0 {
		for i := range b.Children {
			if b.Children[i].Flex > 0 {
				share := extra * b.Children[i].Flex / flexSum
				if mainIsW {
					sizes[i].X += share
				} else {
					sizes[i].Y += share
				}
			}
		}
	}

	main := b.Pad
	if mainIsW {
		switch b.Justify {
		case JustifyEnd:
			main = b.Pad + innerW - sumMain(sizes, gap, true)
		case JustifyCenter:
			main = b.Pad + (innerW-sumMain(sizes, gap, true))/2
		}
	} else {
		switch b.Justify {
		case JustifyEnd:
			main = b.Pad + innerH - sumMain(sizes, gap, false)
		case JustifyCenter:
			main = b.Pad + (innerH-sumMain(sizes, gap, false))/2
		}
	}

	for i := range b.Children {
		w, h := sizes[i].X, sizes[i].Y
		x, y := b.X, b.Y
		if mainIsW {
			x += main
			switch b.Align {
			case AlignEnd:
				y += b.Pad + innerH - h
			case AlignCenter:
				y += b.Pad + (innerH-h)/2
			case AlignBaseline:
				y += b.Pad
			default:
				y += b.Pad
			}
			out[i] = Child{Offset: image.Pt(x, y), Size: image.Pt(w, h)}
			main += w + gap
		} else {
			y += main
			switch b.Align {
			case AlignEnd:
				x += b.Pad + innerW - w
			case AlignCenter:
				x += b.Pad + (innerW-w)/2
			default:
				x += b.Pad
			}
			out[i] = Child{Offset: image.Pt(x, y), Size: image.Pt(w, h)}
			main += h + gap
		}
	}
	return out
}

func sumMain(sizes []image.Point, gap int, row bool) int {
	n := 0
	for i, s := range sizes {
		if row {
			n += s.X
		} else {
			n += s.Y
		}
		if i+1 < len(sizes) {
			n += gap
		}
	}
	return n
}

// UnionRects returns the minimal rectangle covering all non-empty rects.
func UnionRects(rects []image.Rectangle) image.Rectangle {
	var u image.Rectangle
	first := true
	for _, r := range rects {
		if r.Empty() {
			continue
		}
		if first {
			u = r
			first = false
		} else {
			u = u.Union(r)
		}
	}
	return u
}

// AlignEPD expands r to 8px on both axes for Waveshare window updates.
// After CW rotate into EPD memory, EPD X = canvas Y (byte-addressed), so Y
// alignment is required; X is padded for symmetry with the display packer.
func AlignEPD(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}
	const canvasH = 122
	x0 := r.Min.X &^ 7
	x1 := (r.Max.X + 7) &^ 7
	y0 := r.Min.Y &^ 7
	y1 := (r.Max.Y + 7) &^ 7
	if x1 > 250 {
		x1 = 250
	}
	if y1 > canvasH {
		y1 = canvasH
	}
	return image.Rect(x0, y0, x1, y1)
}
