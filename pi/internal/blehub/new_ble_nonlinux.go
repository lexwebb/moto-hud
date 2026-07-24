//go:build ble && !linux

package blehub

import "moto-hud/pi/internal/transport"

func New(hub *transport.Hub) Server {
	return &StubServer{Hub: hub}
}
