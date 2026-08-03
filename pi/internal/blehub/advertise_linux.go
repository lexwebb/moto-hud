//go:build linux && ble

package blehub

import (
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

// LegacyOK is true when advertising used the legacy path (needed on
// kernel 6.18.34+rpt until 6.18.36+ — raspberrypi/linux#7473).
var LegacyOK bool

// peripheralAdv keeps LE advertising alive for the process lifetime.
type peripheralAdv struct {
	conn    *dbus.Conn
	path    dbus.ObjectPath
	adapter dbus.BusObject
	legacy  bool
}

func startPeripheralAdvertisement(localName string) (*peripheralAdv, error) {
	setAdapterAlias(localName)
	registerJustWorksAgent()

	adv, err := registerBlueZAdvertisement(localName)
	if err == nil {
		return adv, nil
	}

	// Kernel 6.18.34+rpt rejects (or hangs on) MGMT Add Extended Advertising
	// Data. Fixed in 6.18.36+. Do not call btmgmt from this process — it can
	// deadlock with bluetoothd under systemd. Rely on motohud-ble-adv.service
	// (or a manual `btmgmt advertising on`) for the legacy flag.
	// https://github.com/raspberrypi/linux/issues/7473
	LegacyOK = true
	return &peripheralAdv{legacy: true}, nil
}

func registerBlueZAdvertisement(localName string) (*peripheralAdv, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("system bus: %w", err)
	}

	adapterPath, err := findLEAdapter(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	adapter := conn.Object("org.bluez", adapterPath)

	path := dbus.ObjectPath("/org/motohud/advertisement0")
	_, err = prop.Export(conn, path, map[string]map[string]*prop.Prop{
		"org.bluez.LEAdvertisement1": {
			"Type":      {Value: "peripheral"},
			"LocalName": {Value: localName},
			"Timeout":   {Value: uint16(0)},
		},
	})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("export advertisement: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- adapter.Call("org.bluez.LEAdvertisingManager1.RegisterAdvertisement", 0, path, map[string]interface{}{}).Err
	}()
	select {
	case err := <-errCh:
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("register advertisement: %w", err)
		}
	case <-time.After(2 * time.Second):
		_ = conn.Close()
		return nil, fmt.Errorf("register advertisement: timed out")
	}

	return &peripheralAdv{conn: conn, path: path, adapter: adapter}, nil
}

func setAdapterAlias(localName string) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return
	}
	defer conn.Close()
	adapterPath, err := findLEAdapter(conn)
	if err != nil {
		return
	}
	adapter := conn.Object("org.bluez", adapterPath)
	_ = adapter.SetProperty("org.bluez.Adapter1.Alias", dbus.MakeVariant(localName))
	_ = adapter.SetProperty("org.bluez.Adapter1.DiscoverableTimeout", dbus.MakeVariant(uint32(0)))
	_ = adapter.SetProperty("org.bluez.Adapter1.Discoverable", dbus.MakeVariant(true))
	// Android (esp. Fairphone/Pixel) often initiates LE pairing on connect.
	// BlueZ defaults to Pairable=no → phone shows "pairing rejected".
	_ = adapter.SetProperty("org.bluez.Adapter1.PairableTimeout", dbus.MakeVariant(uint32(0)))
	_ = adapter.SetProperty("org.bluez.Adapter1.Pairable", dbus.MakeVariant(true))
}

func (a *peripheralAdv) Stop() error {
	if a == nil || a.legacy {
		return nil
	}
	_ = a.adapter.Call("org.bluez.LEAdvertisingManager1.UnregisterAdvertisement", 0, a.path)
	return a.conn.Close()
}

func findLEAdapter(conn *dbus.Conn) (dbus.ObjectPath, error) {
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	obj := conn.Object("org.bluez", "/")
	if err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects); err != nil {
		return "", fmt.Errorf("list bluez objects: %w", err)
	}
	for path, ifaces := range objects {
		if _, ok := ifaces["org.bluez.LEAdvertisingManager1"]; !ok {
			continue
		}
		if !strings.HasPrefix(string(path), "/org/bluez/") {
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf("no LEAdvertisingManager1 adapter found")
}
