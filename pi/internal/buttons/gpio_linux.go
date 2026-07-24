//go:build linux

package buttons

import (
	"context"
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type pinBtn struct {
	name     string
	pin      gpio.PinIn
	shortEv  Event
	longEv   Event
}

// Start watches GPIO buttons with debounce and long-press.
func Start(ctx context.Context, on Handler) error {
	if _, err := host.Init(); err != nil {
		fmt.Printf("buttons: host init failed (%v); falling back to no GPIO\n", err)
		return nil
	}

	pins := []struct {
		bcm     string
		name    string
		shortEv Event
		longEv  Event
	}{
		{fmt.Sprintf("GPIO%d", GPIOPrev), "prev", Prev, PrevLong},
		{fmt.Sprintf("GPIO%d", GPIONext), "next", Next, NextLong},
		{fmt.Sprintf("GPIO%d", GPIOAction), "action", Action, ActionLong},
	}

	var btns []pinBtn
	for _, p := range pins {
		pin := gpioreg.ByName(p.bcm)
		if pin == nil {
			fmt.Printf("buttons: %s not found\n", p.bcm)
			continue
		}
		if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
			fmt.Printf("buttons: %s In failed: %v\n", p.bcm, err)
			continue
		}
		btns = append(btns, pinBtn{name: p.name, pin: pin, shortEv: p.shortEv, longEv: p.longEv})
		fmt.Printf("buttons: watching %s (%s)\n", p.name, p.bcm)
	}

	go poll(ctx, btns, on)
	return nil
}

func poll(ctx context.Context, btns []pinBtn, on Handler) {
	pressed := make(map[string]time.Time)
	lastFire := make(map[string]time.Time)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			for _, b := range btns {
				low := b.pin.Read() == gpio.Low
				if low {
					if _, ok := pressed[b.name]; !ok {
						pressed[b.name] = now
					}
					continue
				}
				start, ok := pressed[b.name]
				if !ok {
					continue
				}
				delete(pressed, b.name)
				held := now.Sub(start)
				if held < debounce {
					continue
				}
				if t, ok := lastFire[b.name]; ok && now.Sub(t) < debounce {
					continue
				}
				lastFire[b.name] = now
				if held >= longPress {
					on(b.longEv)
				} else {
					on(b.shortEv)
				}
			}
		}
	}
}
