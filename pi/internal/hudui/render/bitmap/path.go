package bitmap

import (
	"image"
	"math"

	"moto-hud/pi/internal/hudui/scene"
)

type subpath struct {
	pts    [][2]float64
	closed bool
}

func drawPath(img *image.Gray, st *state, p scene.Path) {
	for _, sp := range parsePath(p.D) {
		if len(sp.pts) == 0 {
			continue
		}
		pts := make([][2]float64, len(sp.pts))
		for i, pt := range sp.pts {
			pts[i] = [2]float64{pt[0] + float64(st.dx), pt[1] + float64(st.dy)}
		}
		if p.Filled {
			fillPolygon(img, pts)
			continue
		}
		w := strokeW(p.StrokeWidth)
		for i := 0; i < len(pts)-1; i++ {
			a, b := pts[i], pts[i+1]
			drawLine(img, int(math.Round(a[0])), int(math.Round(a[1])), int(math.Round(b[0])), int(math.Round(b[1])), w, "")
		}
		if sp.closed && len(pts) > 1 {
			a, b := pts[len(pts)-1], pts[0]
			drawLine(img, int(math.Round(a[0])), int(math.Round(a[1])), int(math.Round(b[0])), int(math.Round(b[1])), w, "")
		}
	}
}

// parsePath supports M/m L/l H/h V/v A/a Z/z (HUD glyphs + u-turn arcs).
func parsePath(d string) []subpath {
	toks := pathTokens(d)
	var out []subpath
	var cur [][2]float64
	var cx, cy float64
	var startX, startY float64
	i := 0
	cmd := byte(0)
	flushOpen := func() {
		if len(cur) > 0 {
			out = append(out, subpath{pts: cur, closed: false})
			cur = nil
		}
	}
	for i < len(toks) {
		t := toks[i]
		if len(t) == 1 && isPathCmd(t[0]) {
			cmd = t[0]
			i++
			continue
		}
		if cmd == 0 {
			i++
			continue
		}
		switch cmd {
		case 'M', 'm':
			if i+1 >= len(toks) {
				flushOpen()
				return out
			}
			x, y := num(toks[i]), num(toks[i+1])
			i += 2
			if cmd == 'm' {
				x += cx
				y += cy
			}
			flushOpen()
			cur = [][2]float64{{x, y}}
			cx, cy = x, y
			startX, startY = x, y
			cmd = mapAfterMove(cmd) // subsequent pairs are implicit LineTo
		case 'L', 'l':
			if i+1 >= len(toks) {
				flushOpen()
				return out
			}
			x, y := num(toks[i]), num(toks[i+1])
			i += 2
			if cmd == 'l' {
				x += cx
				y += cy
			}
			cur = append(cur, [2]float64{x, y})
			cx, cy = x, y
		case 'H', 'h':
			if i >= len(toks) {
				flushOpen()
				return out
			}
			x := num(toks[i])
			i++
			if cmd == 'h' {
				x += cx
			}
			cur = append(cur, [2]float64{x, cy})
			cx = x
		case 'V', 'v':
			if i >= len(toks) {
				flushOpen()
				return out
			}
			y := num(toks[i])
			i++
			if cmd == 'v' {
				y += cy
			}
			cur = append(cur, [2]float64{cx, y})
			cy = y
		case 'A', 'a':
			// A rx ry xrot large sweep x y
			if i+6 >= len(toks) {
				flushOpen()
				return out
			}
			rx, ry := num(toks[i]), num(toks[i+1])
			rot := num(toks[i+2])
			large := num(toks[i+3]) != 0
			sweep := num(toks[i+4]) != 0
			x, y := num(toks[i+5]), num(toks[i+6])
			i += 7
			if cmd == 'a' {
				x += cx
				y += cy
			}
			arc := tessellateArc(cx, cy, rx, ry, rot, large, sweep, x, y)
			if len(arc) > 0 {
				// First point is current; skip duplicate.
				if len(arc) > 1 {
					cur = append(cur, arc[1:]...)
				}
			} else {
				cur = append(cur, [2]float64{x, y})
			}
			cx, cy = x, y
		case 'Z', 'z':
			if len(cur) > 0 {
				out = append(out, subpath{pts: cur, closed: true})
				cur = nil
			}
			cx, cy = startX, startY
			cmd = 0
		default:
			// Unsupported command — skip one number if present.
			i++
		}
	}
	flushOpen()
	return out
}

func mapAfterMove(cmd byte) byte {
	if cmd == 'M' {
		return 'L'
	}
	return 'l'
}

func isPathCmd(c byte) bool {
	switch c {
	case 'M', 'm', 'L', 'l', 'Z', 'z', 'H', 'h', 'V', 'v', 'C', 'c', 'S', 's', 'Q', 'q', 'T', 't', 'A', 'a':
		return true
	}
	return false
}

func pathTokens(d string) []string {
	var out []string
	n := len(d)
	for i := 0; i < n; {
		c := d[i]
		if c == ',' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if isPathCmd(c) {
			out = append(out, string(c))
			i++
			continue
		}
		j := i
		if d[j] == '+' || d[j] == '-' {
			j++
		}
		for j < n && ((d[j] >= '0' && d[j] <= '9') || d[j] == '.') {
			j++
		}
		if j > i {
			out = append(out, d[i:j])
			i = j
			continue
		}
		i++
	}
	return out
}

func num(s string) float64 {
	var f float64
	var sign float64 = 1
	i := 0
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		if s[0] == '-' {
			sign = -1
		}
		i++
	}
	for ; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		f = f*10 + float64(s[i]-'0')
	}
	if i < len(s) && s[i] == '.' {
		i++
		place := 0.1
		for ; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
			f += float64(s[i]-'0') * place
			place *= 0.1
		}
	}
	return f * sign
}

// tessellateArc converts an SVG elliptical arc to polyline points (W3C endpoint→center).
// Returns points including the start (x1,y1).
func tessellateArc(x1, y1, rx, ry, phiDeg float64, large, sweep bool, x2, y2 float64) [][2]float64 {
	if rx == 0 || ry == 0 {
		return [][2]float64{{x1, y1}, {x2, y2}}
	}
	rx, ry = math.Abs(rx), math.Abs(ry)

	phi := phiDeg * math.Pi / 180
	cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)

	dx := (x1 - x2) / 2
	dy := (y1 - y2) / 2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Ensure radii are large enough.
	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		s := math.Sqrt(lambda)
		rx *= s
		ry *= s
	}

	rxSq, rySq := rx*rx, ry*ry
	x1pSq, y1pSq := x1p*x1p, y1p*y1p
	num := rxSq*rySq - rxSq*y1pSq - rySq*x1pSq
	den := rxSq*y1pSq + rySq*x1pSq
	if den == 0 {
		return [][2]float64{{x1, y1}, {x2, y2}}
	}
	frac := num / den
	if frac < 0 {
		frac = 0
	}
	coef := math.Sqrt(frac)
	if large == sweep {
		coef = -coef
	}
	cxp := coef * (rx * y1p / ry)
	cyp := coef * (-ry * x1p / rx)

	cx := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	cy := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	theta1 := angle(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dtheta := angle((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)
	if !sweep && dtheta > 0 {
		dtheta -= 2 * math.Pi
	} else if sweep && dtheta < 0 {
		dtheta += 2 * math.Pi
	}

	// ~1px steps along the longer radius.
	steps := int(math.Ceil(math.Abs(dtheta) * math.Max(rx, ry)))
	if steps < 2 {
		steps = 2
	}
	if steps > 64 {
		steps = 64
	}
	out := make([][2]float64, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		th := theta1 + dtheta*t
		x := cosPhi*rx*math.Cos(th) - sinPhi*ry*math.Sin(th) + cx
		y := sinPhi*rx*math.Cos(th) + cosPhi*ry*math.Sin(th) + cy
		out = append(out, [2]float64{x, y})
	}
	return out
}

func angle(ux, uy, vx, vy float64) float64 {
	dot := ux*vx + uy*vy
	n1 := math.Hypot(ux, uy)
	n2 := math.Hypot(vx, vy)
	if n1 == 0 || n2 == 0 {
		return 0
	}
	c := dot / (n1 * n2)
	if c > 1 {
		c = 1
	} else if c < -1 {
		c = -1
	}
	a := math.Acos(c)
	if ux*vy-uy*vx < 0 {
		return -a
	}
	return a
}
