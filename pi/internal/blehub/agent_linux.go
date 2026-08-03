//go:build linux && ble

package blehub

import (
	"log"

	"github.com/godbus/dbus/v5"
)

const agentPath = dbus.ObjectPath("/org/motohud/agent")

// justWorksAgent auto-accepts LE pairing (NoInputNoOutput).
// Android often requests a bond on connect; without an agent BlueZ rejects it.
type justWorksAgent struct{}

func (justWorksAgent) Release() *dbus.Error { return nil }

func (justWorksAgent) RequestPinCode(device dbus.ObjectPath) (string, *dbus.Error) {
	return "0000", nil
}

func (justWorksAgent) DisplayPinCode(device dbus.ObjectPath, pincode string) *dbus.Error {
	return nil
}

func (justWorksAgent) RequestPasskey(device dbus.ObjectPath) (uint32, *dbus.Error) {
	return 0, nil
}

func (justWorksAgent) DisplayPasskey(device dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	return nil
}

func (justWorksAgent) RequestConfirmation(device dbus.ObjectPath, passkey uint32) *dbus.Error {
	return nil
}

func (justWorksAgent) RequestAuthorization(device dbus.ObjectPath) *dbus.Error {
	return nil
}

func (justWorksAgent) AuthorizeService(device dbus.ObjectPath, uuid string) *dbus.Error {
	return nil
}

func (justWorksAgent) Cancel() *dbus.Error { return nil }

// agentConn must stay open for the process lifetime so BlueZ can call us.
var agentConn *dbus.Conn

func registerJustWorksAgent() {
	if agentConn != nil {
		return
	}
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Printf("ble: agent bus: %v", err)
		return
	}
	if err := conn.Export(justWorksAgent{}, agentPath, "org.bluez.Agent1"); err != nil {
		_ = conn.Close()
		log.Printf("ble: export agent: %v", err)
		return
	}
	mgr := conn.Object("org.bluez", "/org/bluez")
	if err := mgr.Call("org.bluez.AgentManager1.RegisterAgent", 0, agentPath, "NoInputNoOutput").Err; err != nil {
		_ = conn.Close()
		log.Printf("ble: RegisterAgent: %v", err)
		return
	}
	if err := mgr.Call("org.bluez.AgentManager1.RequestDefaultAgent", 0, agentPath).Err; err != nil {
		log.Printf("ble: RequestDefaultAgent: %v", err)
		// Still keep the registered agent; default may already be set.
	}
	agentConn = conn
	log.Printf("ble: pairing agent ready (Just Works)")
}
