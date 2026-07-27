package svg

import (
	"fmt"
	"strings"

	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/pixelfont"
)

// PatchBytes serializes a slot document to 1-bit-ready SVG bytes (white fill + pixelfont rects).
func PatchBytes(doc scene.Document) ([]byte, error) {
	if doc.Width <= 0 || doc.Height <= 0 {
		return nil, fmt.Errorf("render/svg: bad patch size %dx%d", doc.Width, doc.Height)
	}
	body := writeNodes(doc.Nodes)
	s := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="100%%" height="100%%" fill="#fff"/>%s</svg>`,
		doc.Width, doc.Height, doc.Width, doc.Height, body)
	out, err := pixelfont.ReplaceSVGText(s)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func writeNodes(nodes []scene.Node) string {
	var b strings.Builder
	for _, n := range nodes {
		writeNode(&b, n)
	}
	return b.String()
}

func writeNode(b *strings.Builder, n scene.Node) {
	switch v := n.(type) {
	case scene.Text:
		writeText(b, v)
	case scene.Group:
		fmt.Fprintf(b, `<g`)
		if v.ID != "" {
			fmt.Fprintf(b, ` id="%s"`, escapeAttr(v.ID))
		}
		fmt.Fprintf(b, ` transform="translate(%d,%d)">`, v.DX, v.DY)
		for _, c := range v.Children {
			writeNode(b, c)
		}
		b.WriteString(`</g>`)
	case scene.RawSVG:
		b.WriteString(v.Markup)
	}
}

func writeText(b *strings.Builder, t scene.Text) {
	sz, pixel := faceFontSize(t.Face)
	attrs := fmt.Sprintf(`x="%d" y="%d" data-pixel="%s" font-size="%d" fill="#000"`, t.X, t.Baseline, pixel, sz)
	if t.ID != "" {
		attrs = fmt.Sprintf(`id="%s" %s`, escapeAttr(t.ID), attrs)
	}
	if t.Anchor != "" && t.Anchor != "start" {
		attrs += fmt.Sprintf(` text-anchor="%s"`, escapeAttr(t.Anchor))
	}
	fmt.Fprintf(b, `<text %s>%s</text>`, attrs, escapeText(t.Value))
}

func faceFontSize(f scene.Face) (fontSize int, dataPixel string) {
	switch f {
	case scene.Face6x12:
		return 12, "6x12"
	case scene.Face12x24:
		return 24, "12x24"
	case scene.Face16x32:
		return 32, "16x32"
	default:
		return 16, "8x16"
	}
}

func escapeAttr(s string) string {
	return strings.NewReplacer(`&`, "&amp;", `"`, "&quot;").Replace(s)
}

func escapeText(s string) string {
	return strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;").Replace(s)
}

// Fragment writes nodes without an SVG wrapper (for future full-frame scene roots).
func Fragment(nodes []scene.Node) string {
	return writeNodes(nodes)
}

// MustLoadFace resolves a scene face to pixelfont metrics (compose layout helpers).
func MustLoadFace(f scene.Face) *pixelfont.Face {
	var sz pixelfont.Size
	switch f {
	case scene.Face6x12:
		sz = pixelfont.Size6x12
	case scene.Face12x24:
		sz = pixelfont.Size12x24
	case scene.Face16x32:
		sz = pixelfont.Size16x32
	default:
		sz = pixelfont.Size8x16
	}
	face, err := pixelfont.Load(sz)
	if err != nil {
		panic(err)
	}
	return face
}
