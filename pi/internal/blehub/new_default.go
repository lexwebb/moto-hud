//go:build !ble

package blehub

import "moto-hud/pi/internal/transport"

// New returns the stub BLE server unless built with -tags ble on Linux.
func New(hub *transport.Hub) Server {
	return &StubServer{Hub: hub}
}
