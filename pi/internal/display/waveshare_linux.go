//go:build linux

package display

import (
	"fmt"
	"image"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// Waveshare 2.13" B/W HAT pinout (BCM).
const (
	waveshareDC   = "GPIO25"
	waveshareRST  = "GPIO17"
	waveshareBUSY = "GPIO24"
)

type waveshareDisplay struct {
	port              spi.PortCloser
	conn              spi.Conn
	dc                gpio.PinOut
	rst               gpio.PinOut
	busy              gpio.PinIn
	hasBase           bool
	partialsSinceFull int
}

// Full refresh after this many partials (Waveshare V4 datasheet recommendation).
const waveshareFullEveryN = 5

// NewWaveshare opens a Waveshare 2.13" black/white e-Paper HAT (V3/V4).
// Falls back to PNG if hardware init fails.
func NewWaveshare(pngFallback string) (Display, error) {
	if _, err := host.Init(); err != nil {
		fmt.Printf("display: host init failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	port, err := spireg.Open("SPI0.0")
	if err != nil {
		fmt.Printf("display: SPI open failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	dc := gpioreg.ByName(waveshareDC)
	rst := gpioreg.ByName(waveshareRST)
	busy := gpioreg.ByName(waveshareBUSY)
	if dc == nil || rst == nil || busy == nil {
		_ = port.Close()
		fmt.Println("display: missing Waveshare GPIO pins; PNG fallback")
		return NewPNG(pngFallback), nil
	}
	if err := dc.Out(gpio.Low); err != nil {
		_ = port.Close()
		fmt.Printf("display: DC Out failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	if err := rst.Out(gpio.High); err != nil {
		_ = port.Close()
		fmt.Printf("display: RST Out failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	if err := busy.In(gpio.PullUp, gpio.NoEdge); err != nil {
		_ = port.Close()
		fmt.Printf("display: BUSY In failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	conn, err := port.Connect(4*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		_ = port.Close()
		fmt.Printf("display: SPI connect failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	d := &waveshareDisplay{port: port, conn: conn, dc: dc, rst: rst, busy: busy}
	if err := d.init(); err != nil {
		_ = port.Close()
		fmt.Printf("display: Waveshare init failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	fmt.Println("display: Waveshare 2.13 B/W e-Paper ready")
	return d, nil
}

func (d *waveshareDisplay) Show(img *image.Gray) error {
	return d.ShowFrame(img, FrameMeta{})
}

func (d *waveshareDisplay) ShowFrame(img *image.Gray, meta FrameMeta) error {
	if meta.Spatial && !meta.Dirty.Empty() && d.hasBase && d.partialsSinceFull < waveshareFullEveryN {
		reg := AlignCanvasEPD(meta.Dirty)
		if winBuf, epdR := packEPD213Window(img, reg); !epdR.Empty() && len(winBuf) > 0 {
			if err := d.showPartialEPDWindow(winBuf, epdR); err != nil {
				return err
			}
			return nil
		}
	}
	buf := packEPD213(img)
	// First frame and every N partials: full refresh (writes both RAM buffers).
	// Otherwise: partial (~0.3s, no flicker). Matches Waveshare epd2in13_V4.
	if !d.hasBase || d.partialsSinceFull >= waveshareFullEveryN {
		return d.showFull(buf)
	}
	return d.showPartial(buf)
}

func (d *waveshareDisplay) showPartialEPDWindow(winBuf []byte, epdR image.Rectangle) error {
	if err := d.rst.Out(gpio.Low); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	if err := d.rst.Out(gpio.High); err != nil {
		return err
	}
	if err := d.cmd(0x3C); err != nil {
		return err
	}
	if err := d.data([]byte{0x80}); err != nil {
		return err
	}
	if err := d.cmd(0x01); err != nil {
		return err
	}
	if err := d.data([]byte{0xF9, 0x00, 0x00}); err != nil {
		return err
	}
	if err := d.cmd(0x11); err != nil {
		return err
	}
	if err := d.data([]byte{0x03}); err != nil {
		return err
	}
	x0, y0 := epdR.Min.X, epdR.Min.Y
	x1, y1 := epdR.Max.X-1, epdR.Max.Y-1
	if err := d.setWindow(x0, y0, x1, y1); err != nil {
		return err
	}
	if err := d.setCursor(x0, y0); err != nil {
		return err
	}
	if err := d.cmd(0x24); err != nil {
		return err
	}
	if err := d.data(winBuf); err != nil {
		return err
	}
	if err := d.turnOnPartial(); err != nil {
		return err
	}
	d.partialsSinceFull++
	return nil
}

func (d *waveshareDisplay) showFull(buf []byte) error {
	// After partials, re-init before a global refresh (Waveshare FAQ).
	if d.partialsSinceFull > 0 {
		if err := d.init(); err != nil {
			return err
		}
	}
	if err := d.cmd(0x24); err != nil {
		return err
	}
	if err := d.data(buf); err != nil {
		return err
	}
	if err := d.cmd(0x26); err != nil {
		return err
	}
	if err := d.data(buf); err != nil {
		return err
	}
	if err := d.turnOn(); err != nil {
		return err
	}
	d.hasBase = true
	d.partialsSinceFull = 0
	return nil
}

func (d *waveshareDisplay) showPartial(buf []byte) error {
	if err := d.rst.Out(gpio.Low); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	if err := d.rst.Out(gpio.High); err != nil {
		return err
	}

	if err := d.cmd(0x3C); err != nil { // BorderWaveform
		return err
	}
	if err := d.data([]byte{0x80}); err != nil {
		return err
	}
	if err := d.cmd(0x01); err != nil { // Driver output control
		return err
	}
	if err := d.data([]byte{0xF9, 0x00, 0x00}); err != nil {
		return err
	}
	if err := d.cmd(0x11); err != nil { // data entry mode
		return err
	}
	if err := d.data([]byte{0x03}); err != nil {
		return err
	}
	if err := d.setWindow(0, 0, epd213MemW-1, epd213MemH-1); err != nil {
		return err
	}
	if err := d.setCursor(0, 0); err != nil {
		return err
	}
	if err := d.cmd(0x24); err != nil {
		return err
	}
	if err := d.data(buf); err != nil {
		return err
	}
	if err := d.turnOnPartial(); err != nil {
		return err
	}
	d.partialsSinceFull++
	return nil
}

func (d *waveshareDisplay) Close() error {
	_ = d.cmd(0x10) // deep sleep
	_ = d.data([]byte{0x01})
	time.Sleep(100 * time.Millisecond)
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}

func (d *waveshareDisplay) init() error {
	if err := d.reset(); err != nil {
		return err
	}
	d.waitBusy()
	if err := d.cmd(0x12); err != nil { // SWRESET
		return err
	}
	d.waitBusy()
	if err := d.cmd(0x01); err != nil { // Driver output control
		return err
	}
	if err := d.data([]byte{0xF9, 0x00, 0x00}); err != nil {
		return err
	}
	if err := d.cmd(0x11); err != nil { // data entry mode
		return err
	}
	if err := d.data([]byte{0x03}); err != nil {
		return err
	}
	if err := d.setWindow(0, 0, epd213MemW-1, epd213MemH-1); err != nil {
		return err
	}
	if err := d.setCursor(0, 0); err != nil {
		return err
	}
	if err := d.cmd(0x3C); err != nil { // BorderWaveform
		return err
	}
	if err := d.data([]byte{0x05}); err != nil {
		return err
	}
	if err := d.cmd(0x21); err != nil { // Display update control
		return err
	}
	if err := d.data([]byte{0x00, 0x80}); err != nil {
		return err
	}
	if err := d.cmd(0x18); err != nil { // temperature sensor
		return err
	}
	if err := d.data([]byte{0x80}); err != nil {
		return err
	}
	d.waitBusy()
	return nil
}

func (d *waveshareDisplay) reset() error {
	if err := d.rst.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	if err := d.rst.Out(gpio.Low); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	if err := d.rst.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	return nil
}

func (d *waveshareDisplay) setWindow(x0, y0, x1, y1 int) error {
	if err := d.cmd(0x44); err != nil {
		return err
	}
	if err := d.data([]byte{byte(x0 >> 3), byte(x1 >> 3)}); err != nil {
		return err
	}
	if err := d.cmd(0x45); err != nil {
		return err
	}
	return d.data([]byte{
		byte(y0), byte(y0 >> 8),
		byte(y1), byte(y1 >> 8),
	})
}

func (d *waveshareDisplay) setCursor(x, y int) error {
	if err := d.cmd(0x4E); err != nil {
		return err
	}
	if err := d.data([]byte{byte(x)}); err != nil {
		return err
	}
	if err := d.cmd(0x4F); err != nil {
		return err
	}
	return d.data([]byte{byte(y), byte(y >> 8)})
}

func (d *waveshareDisplay) turnOn() error {
	if err := d.cmd(0x22); err != nil {
		return err
	}
	if err := d.data([]byte{0xF7}); err != nil {
		return err
	}
	if err := d.cmd(0x20); err != nil {
		return err
	}
	d.waitBusy()
	return nil
}

func (d *waveshareDisplay) turnOnPartial() error {
	if err := d.cmd(0x22); err != nil {
		return err
	}
	if err := d.data([]byte{0xFF}); err != nil {
		return err
	}
	if err := d.cmd(0x20); err != nil {
		return err
	}
	d.waitBusy()
	return nil
}

func (d *waveshareDisplay) waitBusy() {
	deadline := time.Now().Add(40 * time.Second)
	for time.Now().Before(deadline) {
		if d.busy.Read() == gpio.Low {
			time.Sleep(10 * time.Millisecond)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("display: Waveshare BUSY timed out")
}

func (d *waveshareDisplay) cmd(c byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.conn.Tx([]byte{c}, nil)
}

func (d *waveshareDisplay) data(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	const chunk = 4096
	for i := 0; i < len(b); i += chunk {
		end := i + chunk
		if end > len(b) {
			end = len(b)
		}
		if err := d.conn.Tx(b[i:end], nil); err != nil {
			return err
		}
	}
	return nil
}
