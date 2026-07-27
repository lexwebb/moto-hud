// Package scenetempl bridges templ components to hudui/scene nodes (ADR 0010 + 0012).
package scenetempl

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"

	"moto-hud/pi/internal/hudui/scene"
)

type builderKey struct{}

func withBuilder(ctx context.Context, b *scene.Builder) context.Context {
	return context.WithValue(ctx, builderKey{}, b)
}

func builderFrom(ctx context.Context) *scene.Builder {
	return ctx.Value(builderKey{}).(*scene.Builder)
}

// Render runs a templ component that appends scene nodes via this package's primitives.
func Render(c templ.Component) []scene.Node {
	var b scene.Builder
	_ = c.Render(withBuilder(context.Background(), &b), io.Discard)
	return b.Nodes()
}

// Nodes splices an existing node list into the tree being built.
func Nodes(nodes []scene.Node) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		builderFrom(ctx).Append(nodes...)
		return nil
	})
}

func Text(id string, face scene.Face, x, baseline int, anchor, value string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		builderFrom(ctx).Text(id, face, x, baseline, anchor, value)
		return nil
	})
}

func Line(x1, y1, x2, y2 int) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		builderFrom(ctx).Line(x1, y1, x2, y2)
		return nil
	})
}

func Raw(markup string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		builderFrom(ctx).Raw(markup)
		return nil
	})
}

// EmptyGroup is a translated group with no children (e.g. link slot filled by a layer).
func EmptyGroup(id string, dx, dy int) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		builderFrom(ctx).Append(scene.Group{ID: id, DX: dx, DY: dy})
		return nil
	})
}

// Group collects templ child components into a translated scene group.
func Group(id string, dx, dy int) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		parent := builderFrom(ctx)
		child := templ.GetChildren(ctx)
		if child == nil {
			child = templ.NopComponent
		}
		var sub scene.Builder
		if err := child.Render(withBuilder(ctx, &sub), io.Discard); err != nil {
			return err
		}
		parent.Append(scene.Group{ID: id, DX: dx, DY: dy, Children: sub.Nodes()})
		return nil
	})
}

// ManeuverAt wraps maneuver glyph nodes in the design-kit group offset.
func ManeuverAt(glyphY int, nodes []scene.Node) templ.Component {
	if len(nodes) == 0 {
		return templ.NopComponent
	}
	return Nodes([]scene.Node{scene.Group{ID: "maneuver", DX: -2, DY: glyphY, Children: nodes}})
}

// Maneuver emits the nav glyph group when paths markup is non-empty (deprecated).
func Maneuver(glyphY int, paths string) templ.Component {
	if paths == "" {
		return templ.NopComponent
	}
	markup := fmt.Sprintf(
		`<g id="maneuver" transform="translate(-2,%d)" fill="#000" stroke="#000" stroke-width="3" stroke-linecap="square" stroke-linejoin="miter">%s</g>`,
		glyphY, paths,
	)
	return Raw(markup)
}
