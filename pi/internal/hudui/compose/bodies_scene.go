package compose

import (
	"moto-hud/pi/internal/hudui/scene"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

func mediaBodyLayout() (mw, yPlaying, yTitle, yArtist int) {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	title, _ := pixelfont.Load(pixelfont.Size12x24)
	mw = token.MainWidth()
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	blockH := meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.CellH
	top := contentTop + (contentBot-contentTop-blockH)/2
	return mw, top + meta.Metrics.Ascent,
		top + meta.Metrics.CellH + token.GapSm + title.Metrics.Ascent,
		top + meta.Metrics.CellH + token.GapSm + title.Metrics.CellH + token.GapSm + body.Metrics.Ascent
}

func mediaBodyScene(playing, title, artist string) []scene.Node {
	_, y1, y2, y3 := mediaBodyLayout()
	var b scene.Builder
	b.Text("playing", scene.Face6x12, 0, y1, "start", playing)
	b.Text("title", scene.Face12x24, 0, y2, "start", title)
	b.Text("artist", scene.Face8x16, 0, y3, "start", artist)
	return b.Nodes()
}

func statusBodyLayout() (mw, y1, y2, y3 int) {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	mw = token.MainWidth()
	headerBottom := token.Pad + meta.Metrics.CellH
	contentTop := headerBottom + token.GapSm + token.GapMd
	contentBot := token.Height - token.Pad
	rowH := body.Metrics.CellH + token.GapMd
	rows := 3
	blockH := rowH*rows - token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	return mw, top + body.Metrics.Ascent, top + rowH + body.Metrics.Ascent, top + 2*rowH + body.Metrics.Ascent
}

func statusBodyScene(bleLinked, navActive bool) []scene.Node {
	ble, nav := "DOWN", "OFF"
	if bleLinked {
		ble = "UP"
	}
	if navActive {
		nav = "ON"
	}
	mw, y1, y2, y3 := statusBodyLayout()
	var b scene.Builder
	b.Text("", scene.Face8x16, 0, y1, "start", "LINK")
	b.Text("status_link", scene.Face8x16, mw, y1, "end", ble)
	b.Text("", scene.Face8x16, 0, y2, "start", "NAV")
	b.Text("status_nav", scene.Face8x16, mw, y2, "end", nav)
	b.Text("", scene.Face8x16, 0, y3, "start", "PKTS")
	b.Text("status_pkts", scene.Face8x16, mw, y3, "end", "OK")
	return b.Nodes()
}

func navIdleBodyScene() []scene.Node {
	meta, _ := pixelfont.Load(pixelfont.Size6x12)
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	mw := token.MainWidth()
	pad := token.Pad
	headerBottom := pad + meta.Metrics.CellH
	divY := headerBottom + token.GapSm
	contentTop := divY + token.GapMd
	contentBot := token.Height - pad
	blockH := body.Metrics.CellH*2 + token.GapMd
	top := contentTop + (contentBot-contentTop-blockH)/2
	b1 := top + body.Metrics.Ascent
	b2 := top + body.Metrics.CellH + token.GapMd + body.Metrics.Ascent
	mid := mw / 2
	var b scene.Builder
	b.Text("title", scene.Face8x16, mid, b1, "middle", "MOTO HUD")
	b.Text("msg", scene.Face8x16, mid, b2, "middle", "Waiting for route...")
	return b.Nodes()
}
