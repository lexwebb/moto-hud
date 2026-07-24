package blehub

import (
	"context"
	"encoding/json"
	"log"

	"moto-hud/pi/internal/protocol"
	"moto-hud/pi/internal/transport"
)

// Server is a BLE GATT peripheral (or stub).
type Server interface {
	Start(ctx context.Context) error
}

// StubServer logs that BLE is unavailable and relies on HTTP transport.
type StubServer struct {
	Hub *transport.Hub
}

func (s *StubServer) Start(ctx context.Context) error {
	log.Printf("ble: stub peripheral (device would be %s); use HTTP injector on :8787", protocol.DeviceName)
	s.Hub.SetLinked(false)
	go func() {
		ch := s.Hub.SubscribeCmd()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				b, _ := json.Marshal(msg)
				log.Printf("ble(stub): would notify cmd %s", b)
			}
		}
	}()
	return nil
}
