package com.motohud.companion

import android.annotation.SuppressLint
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothGatt
import android.bluetooth.BluetoothGattCallback
import android.bluetooth.BluetoothGattCharacteristic
import android.bluetooth.BluetoothGattDescriptor
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothProfile
import android.bluetooth.le.ScanCallback
import android.bluetooth.le.ScanFilter
import android.bluetooth.le.ScanResult
import android.bluetooth.le.ScanSettings
import android.content.Context
import android.os.ParcelUuid
import android.util.Log
import org.json.JSONObject
import java.util.UUID

@SuppressLint("MissingPermission")
class BleClient(private val context: Context) {

    private val adapter: BluetoothAdapter? =
        (context.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager).adapter

    private var gatt: BluetoothGatt? = null
    private var navChar: BluetoothGattCharacteristic? = null
    private var mediaChar: BluetoothGattCharacteristic? = null
    private var cmdChar: BluetoothGattCharacteristic? = null
    private var hbChar: BluetoothGattCharacteristic? = null

    @Volatile
    var connected: Boolean = false
        private set

    private val scanCallback = object : ScanCallback() {
        override fun onScanResult(callbackType: Int, result: ScanResult) {
            val name = result.device.name ?: result.scanRecord?.deviceName
            if (name != Protocol.DEVICE_NAME) return
            stopScan()
            HudBus.setStatus("Connecting to ${result.device.address}")
            gatt = result.device.connectGatt(context, false, gattCallback, BluetoothDevice.TRANSPORT_LE)
        }

        override fun onScanFailed(errorCode: Int) {
            HudBus.setStatus("BLE scan failed: $errorCode")
        }
    }

    private val gattCallback = object : BluetoothGattCallback() {
        override fun onConnectionStateChange(g: BluetoothGatt, status: Int, newState: Int) {
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                HudBus.setStatus("BLE connected, discovering…")
                g.discoverServices()
            } else if (newState == BluetoothProfile.STATE_DISCONNECTED) {
                connected = false
                HudBus.setStatus("BLE disconnected")
                navChar = null
                mediaChar = null
                cmdChar = null
                hbChar = null
            }
        }

        override fun onServicesDiscovered(g: BluetoothGatt, status: Int) {
            val svc = g.getService(UUID.fromString(Protocol.SERVICE_UUID))
            if (svc == null) {
                HudBus.setStatus("Service not found")
                return
            }
            navChar = svc.getCharacteristic(UUID.fromString(Protocol.NAV_UUID))
            mediaChar = svc.getCharacteristic(UUID.fromString(Protocol.MEDIA_UUID))
            cmdChar = svc.getCharacteristic(UUID.fromString(Protocol.CMD_UUID))
            hbChar = svc.getCharacteristic(UUID.fromString(Protocol.HEARTBEAT_UUID))
            cmdChar?.let { enableNotify(g, it) }
            connected = true
            HudBus.setStatus("HUD ready")
            writeHeartbeat()
        }

        override fun onCharacteristicChanged(
            g: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic,
            value: ByteArray,
        ) {
            handleCmd(value)
        }

        @Deprecated("Deprecated in Java")
        override fun onCharacteristicChanged(g: BluetoothGatt, characteristic: BluetoothGattCharacteristic) {
            characteristic.value?.let { handleCmd(it) }
        }
    }

    fun startScan() {
        val scanner = adapter?.bluetoothLeScanner
        if (scanner == null) {
            HudBus.setStatus("Bluetooth unavailable")
            return
        }
        HudBus.setStatus("Scanning for ${Protocol.DEVICE_NAME}…")
        val filter = ScanFilter.Builder()
            .setServiceUuid(ParcelUuid(UUID.fromString(Protocol.SERVICE_UUID)))
            .build()
        val settings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
            .build()
        scanner.startScan(listOf(filter), settings, scanCallback)
    }

    fun stopScan() {
        adapter?.bluetoothLeScanner?.stopScan(scanCallback)
    }

    fun close() {
        stopScan()
        gatt?.close()
        gatt = null
        connected = false
    }

    fun writeNav(nav: NavState) {
        write(navChar, nav.toJson())
    }

    fun writeMedia(media: MediaState) {
        write(mediaChar, media.toJson())
    }

    fun writeHeartbeat() {
        val body = JSONObject().put("type", "heartbeat").put("ts", System.currentTimeMillis() / 1000).toString()
        write(hbChar, body.toByteArray(Charsets.UTF_8))
    }

    private fun write(ch: BluetoothGattCharacteristic?, payload: ByteArray) {
        val g = gatt ?: return
        val c = ch ?: return
        c.value = payload
        c.writeType = BluetoothGattCharacteristic.WRITE_TYPE_NO_RESPONSE
        g.writeCharacteristic(c)
    }

    private fun enableNotify(g: BluetoothGatt, ch: BluetoothGattCharacteristic) {
        g.setCharacteristicNotification(ch, true)
        val cccd = ch.getDescriptor(UUID.fromString(CCCD))
        cccd?.let {
            it.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
            g.writeDescriptor(it)
        }
    }

    private fun handleCmd(value: ByteArray) {
        try {
            val action = JSONObject(String(value, Charsets.UTF_8)).optString("action")
            if (action.isNotBlank()) {
                Log.d(TAG, "cmd $action")
                HudBus.publishCmd(action)
            }
        } catch (e: Exception) {
            Log.w(TAG, "bad cmd payload", e)
        }
    }

    companion object {
        private const val TAG = "BleClient"
        private const val CCCD = "00002902-0000-1000-8000-00805f9b34fb"
    }
}
