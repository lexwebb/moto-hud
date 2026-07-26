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

// Display HAT Mini (ST7789) pinout (BCM).
const (
	lcdDC  = "GPIO9"
	lcdBL  = "GPIO13"
	lcdSPI = "SPI0.1" // CE1
)

type lcdDisplay struct {
	port spi.PortCloser
	conn spi.Conn
	dc   gpio.PinOut
	bl   gpio.PinOut
}

// NewLCD opens a Pimoroni Display HAT Mini (320×240 ST7789).
// Letterboxes the canonical 250×122 gray frame. Falls back to PNG on failure.
func NewLCD(pngFallback string) (Display, error) {
	if _, err := host.Init(); err != nil {
		fmt.Printf("display: host init failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	port, err := spireg.Open(lcdSPI)
	if err != nil {
		fmt.Printf("display: SPI1 open failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	dc := gpioreg.ByName(lcdDC)
	bl := gpioreg.ByName(lcdBL)
	if dc == nil || bl == nil {
		_ = port.Close()
		fmt.Println("display: missing LCD GPIO pins; PNG fallback")
		return NewPNG(pngFallback), nil
	}
	if err := dc.Out(gpio.Low); err != nil {
		_ = port.Close()
		fmt.Printf("display: DC Out failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	if err := bl.Out(gpio.Low); err != nil {
		_ = port.Close()
		fmt.Printf("display: BL Out failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	conn, err := port.Connect(32*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		_ = port.Close()
		fmt.Printf("display: SPI connect failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	d := &lcdDisplay{port: port, conn: conn, dc: dc, bl: bl}
	time.Sleep(100 * time.Millisecond)
	_ = bl.Out(gpio.High)
	if err := d.init(); err != nil {
		_ = port.Close()
		fmt.Printf("display: LCD init failed (%v); PNG fallback\n", err)
		return NewPNG(pngFallback), nil
	}
	fmt.Println("display: Display HAT Mini LCD ready (320x240 letterbox)")
	return d, nil
}

func (d *lcdDisplay) Show(img *image.Gray) error {
	frame := LetterboxGray(img)
	frame = RotateRGBA180(frame) // Display HAT Mini default orientation
	pix := RGBAToRGB565(frame)
	if err := d.setWindow(0, 0, LCDWidth-1, LCDHeight-1); err != nil {
		return err
	}
	return d.data(pix)
}

func (d *lcdDisplay) Close() error {
	if d.bl != nil {
		_ = d.bl.Out(gpio.Low)
	}
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}

func (d *lcdDisplay) init() error {
	if err := d.cmd(0x01); err != nil { // SWRESET
		return err
	}
	time.Sleep(150 * time.Millisecond)
	if err := d.cmd(0x36); err != nil { // MADCTL
		return err
	}
	if err := d.data([]byte{0x70}); err != nil {
		return err
	}
	if err := d.cmd(0xB2); err != nil {
		return err
	}
	if err := d.data([]byte{0x0C, 0x0C, 0x00, 0x33, 0x33}); err != nil {
		return err
	}
	if err := d.cmd(0x3A); err != nil { // COLMOD RGB565
		return err
	}
	if err := d.data([]byte{0x05}); err != nil {
		return err
	}
	if err := d.cmd(0xB7); err != nil {
		return err
	}
	if err := d.data([]byte{0x14}); err != nil {
		return err
	}
	if err := d.cmd(0xBB); err != nil {
		return err
	}
	if err := d.data([]byte{0x37}); err != nil {
		return err
	}
	if err := d.cmd(0xC0); err != nil {
		return err
	}
	if err := d.data([]byte{0x2C}); err != nil {
		return err
	}
	if err := d.cmd(0xC2); err != nil {
		return err
	}
	if err := d.data([]byte{0x01}); err != nil {
		return err
	}
	if err := d.cmd(0xC3); err != nil {
		return err
	}
	if err := d.data([]byte{0x12}); err != nil {
		return err
	}
	if err := d.cmd(0xC4); err != nil {
		return err
	}
	if err := d.data([]byte{0x20}); err != nil {
		return err
	}
	if err := d.cmd(0xD0); err != nil {
		return err
	}
	if err := d.data([]byte{0xA4, 0xA1}); err != nil {
		return err
	}
	if err := d.cmd(0xC6); err != nil {
		return err
	}
	if err := d.data([]byte{0x0F}); err != nil {
		return err
	}
	if err := d.cmd(0xE0); err != nil {
		return err
	}
	if err := d.data([]byte{0xD0, 0x04, 0x0D, 0x11, 0x13, 0x2B, 0x3F, 0x54, 0x4C, 0x18, 0x0D, 0x0B, 0x1F, 0x23}); err != nil {
		return err
	}
	if err := d.cmd(0xE1); err != nil {
		return err
	}
	if err := d.data([]byte{0xD0, 0x04, 0x0C, 0x11, 0x13, 0x2C, 0x3F, 0x44, 0x51, 0x2F, 0x1F, 0x1F, 0x20, 0x23}); err != nil {
		return err
	}
	if err := d.cmd(0x21); err != nil { // INVON
		return err
	}
	if err := d.cmd(0x11); err != nil { // SLPOUT
		return err
	}
	if err := d.cmd(0x29); err != nil { // DISPON
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (d *lcdDisplay) setWindow(x0, y0, x1, y1 int) error {
	if err := d.cmd(0x2A); err != nil {
		return err
	}
	if err := d.data([]byte{byte(x0 >> 8), byte(x0), byte(x1 >> 8), byte(x1)}); err != nil {
		return err
	}
	if err := d.cmd(0x2B); err != nil {
		return err
	}
	if err := d.data([]byte{byte(y0 >> 8), byte(y0), byte(y1 >> 8), byte(y1)}); err != nil {
		return err
	}
	return d.cmd(0x2C)
}

func (d *lcdDisplay) cmd(c byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.conn.Tx([]byte{c}, nil)
}

func (d *lcdDisplay) data(b []byte) error {
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
