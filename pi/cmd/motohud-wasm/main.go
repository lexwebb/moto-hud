//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"image/png"
	"syscall/js"

	"moto-hud/pi/internal/buttons"
	"moto-hud/pi/internal/hud"
	"moto-hud/pi/internal/protocol"
	"moto-hud/pi/internal/transport"
)

var (
	state = hud.NewState()
	gate  = &hud.RefreshGate{}
	hub   = transport.NewHub(state, gate, func() {})
)

func main() {
	state.SetBLELinked(true)
	js.Global().Set("MotoHUD", js.ValueOf(map[string]any{
		"applyNav":          js.FuncOf(applyNav),
		"applyMedia":        js.FuncOf(applyMedia),
		"button":            js.FuncOf(button),
		"renderPNG":         js.FuncOf(renderPNG),
		"renderMinimapPNG":  js.FuncOf(renderMinimapPNG),
		"minimapSVG":        js.FuncOf(minimapSVGJS),
		"screen":            js.FuncOf(screenName),
		"width":             hud.Width,
		"height":            hud.Height,
	}))
	select {}
}

func applyNav(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errVal("missing json")
	}
	var msg protocol.NavMessage
	if err := json.Unmarshal([]byte(args[0].String()), &msg); err != nil {
		return errVal(err.Error())
	}
	hub.ApplyNav(msg)
	return okVal()
}

func applyMedia(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errVal("missing json")
	}
	var msg protocol.MediaMessage
	if err := json.Unmarshal([]byte(args[0].String()), &msg); err != nil {
		return errVal(err.Error())
	}
	hub.ApplyMedia(msg)
	return okVal()
}

func button(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errVal("missing event")
	}
	switch args[0].String() {
	case "prev":
		hub.HandleButtonEvent(buttons.Prev)
	case "next":
		hub.HandleButtonEvent(buttons.Next)
	case "action":
		hub.HandleButtonEvent(buttons.Action)
	case "prev_long":
		hub.HandleButtonEvent(buttons.PrevLong)
	case "next_long":
		hub.HandleButtonEvent(buttons.NextLong)
	case "action_long":
		hub.HandleButtonEvent(buttons.ActionLong)
	default:
		return errVal("unknown button")
	}
	return okVal()
}

func renderPNG(this js.Value, args []js.Value) any {
	screen, nav, media, linked, _ := state.Snapshot()
	fr := hud.RenderWithEngine(screen, nav, media, linked, true)
	img := fr.Image
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return errVal(err.Error())
	}
	data := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(dst, data)
	return dst
}

func renderMinimapPNG(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errVal("missing minimap json")
	}
	var mm protocol.MinimapMessage
	if err := json.Unmarshal([]byte(args[0].String()), &mm); err != nil {
		return errVal(err.Error())
	}
	w, h := 70, 80
	if len(args) >= 3 {
		if args[1].Truthy() {
			w = args[1].Int()
		}
		if args[2].Truthy() {
			h = args[2].Int()
		}
	}
	context, route, marks := true, true, true
	if len(args) >= 4 && args[3].Type() == js.TypeString {
		switch args[3].String() {
		case "context":
			route, marks = false, false
		case "route":
			context, marks = false, false
		case "marks":
			context, route = false, false
		case "route+marks":
			context = false
		case "all", "":
			// defaults
		default:
			return errVal("layer must be all|context|route|marks|route+marks")
		}
	}
	img, err := hud.RenderMinimapLayers(&mm, w, h, context, route, marks)
	if err != nil {
		return errVal(err.Error())
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return errVal(err.Error())
	}
	data := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(dst, data)
	return dst
}

func minimapSVGJS(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errVal("missing minimap json")
	}
	var mm protocol.MinimapMessage
	if err := json.Unmarshal([]byte(args[0].String()), &mm); err != nil {
		return errVal(err.Error())
	}
	w, h := 70, 80
	if len(args) >= 3 {
		if args[1].Truthy() {
			w = args[1].Int()
		}
		if args[2].Truthy() {
			h = args[2].Int()
		}
	}
	return hud.MinimapSVGFragment(&mm, w, h)
}

func screenName(this js.Value, args []js.Value) any {
	return state.CurrentScreen().String()
}

func okVal() js.Value  { return js.ValueOf(map[string]any{"ok": true}) }
func errVal(m string) js.Value { return js.ValueOf(map[string]any{"ok": false, "error": m}) }
