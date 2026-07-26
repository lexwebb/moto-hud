package buttons

import "time"

type Event int

const (
	Prev Event = iota
	Next
	Action
	ActionLong
	PrevLong
	NextLong
)

// Handler receives debounced button events.
type Handler func(Event)

const (
	debounce  = 40 * time.Millisecond
	longPress = 800 * time.Millisecond
)

// Default GPIO BCM numbers (avoid Inky SPI: 17 busy, 27 reset, 22 dc).
const (
	GPIOPrev   = 5
	GPIONext   = 6
	GPIOAction = 13
	// GPIOActionLCD is Display HAT Mini button X — BCM 13 is the LCD backlight.
	GPIOActionLCD = 16
)

var actionGPIO = GPIOAction

// SetActionGPIO overrides the Action button BCM pin (e.g. LCD host → 16).
func SetActionGPIO(bcm int) {
	actionGPIO = bcm
}

// ActionGPIO returns the BCM pin currently used for Action.
func ActionGPIO() int {
	return actionGPIO
}
