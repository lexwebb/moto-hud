package compose

import (
	"image"

	"moto-hud/pi/internal/hudui"
	"moto-hud/pi/internal/hudui/plan"
	"moto-hud/pi/internal/hudui/token"
	"moto-hud/pi/internal/pixelfont"
)

type statusLayout struct {
	mw                                         int
	y1, y2, y3                                 int
	linkVal, navVal, pktsVal                   string
	linkSlot, navSlot, pktsSlot                image.Rectangle
}

func statusLayoutGeom(linked, navActive bool) statusLayout {
	body, _ := pixelfont.Load(pixelfont.Size8x16)
	mw, y1, y2, y3 := statusBodyLayout()
	mainX := token.Pad
	rowH := body.Metrics.CellH + token.GapMd
	top := y1 - body.Metrics.Ascent
	linkVal, navVal := "DOWN", "OFF"
	if linked {
		linkVal = "UP"
	}
	if navActive {
		navVal = "ON"
	}
	pktsVal := "OK"
	// Value column slots (right-aligned text in each row).
	valW := 48
	valX := mainX + mw - valW
	return statusLayout{
		mw: mw, y1: y1, y2: y2, y3: y3,
		linkVal: linkVal, navVal: navVal, pktsVal: pktsVal,
		linkSlot:  image.Rect(valX, top, mainX+mw, top+body.Metrics.CellH),
		navSlot:   image.Rect(valX, top+rowH, mainX+mw, top+rowH+body.Metrics.CellH),
		pktsSlot:  image.Rect(valX, top+2*rowH, mainX+mw, top+2*rowH+body.Metrics.CellH),
	}
}

func planStatus(in Input) (plan.ScreenPlan, error) {
	geom := statusLayoutGeom(in.Linked, in.Nav.Active)
	body, err := StatusBodySVG(in.Linked, in.Nav.Active)
	if err != nil {
		return plan.ScreenPlan{}, err
	}
	k := Keys{}
	deps := in.NavSVG
	linked := in.Linked
	navActive := in.Nav.Active
	layers := []plan.Layer{
		{
			ID: hudui.NodeStatusLink, Tier: hudui.TierPartialOK, Key: k.StatusLink(linked), Slot: geom.linkSlot,
			Patch: func() ([]byte, error) {
				v := "DOWN"
				if linked {
					v = "UP"
				}
				return patchStatusValueSVG("status_link", v, geom.linkSlot.Dx(), geom.linkSlot.Dy(), deps)
			},
		},
		{
			ID: hudui.NodeStatusNav, Tier: hudui.TierPartialOK, Key: k.StatusNav(navActive), Slot: geom.navSlot,
			Patch: func() ([]byte, error) {
				v := "OFF"
				if navActive {
					v = "ON"
				}
				return patchStatusValueSVG("status_nav", v, geom.navSlot.Dx(), geom.navSlot.Dy(), deps)
			},
		},
		{
			ID: hudui.NodeStatusPkts, Tier: hudui.TierPartialOK, Key: k.StatusPkts(), Slot: geom.pktsSlot,
			Patch: func() ([]byte, error) {
				return patchStatusValueSVG("status_pkts", "OK", geom.pktsSlot.Dx(), geom.pktsSlot.Dy(), deps)
			},
		},
	}
	return plan.ScreenPlan{
		BodySVG:     body,
		Descriptors: plan.BuildDescriptors(Keys{}.Hash("status_screen"), Keys{}.Hash("status_chrome"), layers),
		Layers:      layers,
	}, nil
}
