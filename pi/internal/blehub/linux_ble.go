//go:build linux && ble

package blehub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"tinygo.org/x/bluetooth"

	"moto-hud/pi/internal/protocol"
	"moto-hud/pi/internal/transport"
)

// BlueZServer advertises MotoHUD and accepts nav/media writes.
type BlueZServer struct {
	Hub *transport.Hub
}

func New(hub *transport.Hub) Server {
	return &BlueZServer{Hub: hub}
}

func (s *BlueZServer) Start(ctx context.Context) error {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		log.Printf("ble: enable failed (%v); stubbing", err)
		return (&StubServer{Hub: s.Hub}).Start(ctx)
	}

	svcUUID := mustUUID(protocol.ServiceUUID)
	navUUID := mustUUID(protocol.NavUUID)
	mediaUUID := mustUUID(protocol.MediaUUID)
	cmdUUID := mustUUID(protocol.CmdUUID)
	hbUUID := mustUUID(protocol.HeartbeatUUID)

	var cmdChar bluetooth.Characteristic

	err := adapter.AddService(&bluetooth.Service{
		UUID: svcUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  navUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					var n protocol.NavMessage
					if json.Unmarshal(value, &n) == nil {
						s.Hub.ApplyNav(n)
						s.Hub.SetLinked(true)
					}
				},
			},
			{
				UUID:  mediaUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					var m protocol.MediaMessage
					if json.Unmarshal(value, &m) == nil {
						s.Hub.ApplyMedia(m)
						s.Hub.SetLinked(true)
					}
				},
			},
			{
				Handle: &cmdChar,
				UUID:   cmdUUID,
				Flags:  bluetooth.CharacteristicNotifyPermission | bluetooth.CharacteristicReadPermission,
			},
			{
				UUID:  hbUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					s.Hub.SetLinked(true)
				},
			},
		},
	})
	if err != nil {
		log.Printf("ble: AddService failed (%v); stubbing", err)
		return (&StubServer{Hub: s.Hub}).Start(ctx)
	}

	// Bypass tinygo DefaultAdvertisement: it hardcodes Type=broadcast, which
	// BlueZ on Pi Zero W rejects (Invalid Parameters 0x0d). Register a
	// connectable peripheral advertisement ourselves (name only; Android scans
	// by device name, and a 128-bit ServiceUUID won't fit with the name).
	adv, err := startPeripheralAdvertisement(protocol.DeviceName)
	if err != nil {
		log.Printf("ble: advertise start failed (%v); stubbing", err)
		return (&StubServer{Hub: s.Hub}).Start(ctx)
	}
	if LegacyOK {
		log.Printf("ble: GATT up as %s (legacy adv — ensure btmgmt advertising on / motohud-ble-adv.service)", protocol.DeviceName)
	} else {
		log.Printf("ble: advertising as %s", protocol.DeviceName)
	}
	s.Hub.SetLinked(false)

	go func() {
		ch := s.Hub.SubscribeCmd()
		for {
			select {
			case <-ctx.Done():
				_ = adv.Stop()
				return
			case msg := <-ch:
				b, _ := json.Marshal(msg)
				_, _ = cmdChar.Write(b)
			case <-time.After(30 * time.Second):
				// keep process alive
			}
		}
	}()
	return nil
}

func mustUUID(s string) bluetooth.UUID {
	u, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}
