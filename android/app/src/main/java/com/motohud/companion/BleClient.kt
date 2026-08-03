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
import android.bluetooth.le.ScanResult
import android.bluetooth.le.ScanSettings
import android.content.Context
import android.os.Handler
import android.os.Looper
import android.util.Log
import org.json.JSONObject
import java.util.UUID

@SuppressLint("MissingPermission")
class BleClient(private val context: Context) : HudSink {

    private val adapter: BluetoothAdapter? =
        (context.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager).adapter
    private val mainHandler = Handler(Looper.getMainLooper())

    private var gatt: BluetoothGatt? = null
    private var navChar: BluetoothGattCharacteristic? = null
    private var mediaChar: BluetoothGattCharacteristic? = null
    private var cmdChar: BluetoothGattCharacteristic? = null
    private var hbChar: BluetoothGattCharacteristic? = null

    @Volatile
    private var wantLink: Boolean = false

    @Volatile
    var connected: Boolean = false
        private set

    private val scanCallback = object : ScanCallback() {
        override fun onScanResult(callbackType: Int, result: ScanResult) {
            if (!wantLink) return
            val name = result.device.name ?: result.scanRecord?.deviceName
            if (name != Protocol.DEVICE_NAME) return
            stopScan()
            HudBus.setStatus("Connecting to ${result.device.address}")
            // Close any prior GATT before a new connect attempt.
            gatt?.close()
            gatt = result.device.connectGatt(context, false, gattCallback, BluetoothDevice.TRANSPORT_LE)
        }

        override fun onScanFailed(errorCode: Int) {
            val why = when (errorCode) {
                SCAN_FAILED_ALREADY_STARTED -> "already started"
                SCAN_FAILED_APPLICATION_REGISTRATION_FAILED -> "app registration failed"
                SCAN_FAILED_INTERNAL_ERROR -> "internal error"
                SCAN_FAILED_FEATURE_UNSUPPORTED -> "unsupported"
                SCAN_FAILED_OUT_OF_HARDWARE_RESOURCES -> "out of hardware resources"
                6 -> "scanning too frequently"
                else -> "code $errorCode"
            }
            HudBus.setStatus("BLE scan failed: $why")
        }
    }

    private val gattCallback = object : BluetoothGattCallback() {
        override fun onConnectionStateChange(g: BluetoothGatt, status: Int, newState: Int) {
            Log.i(TAG, "connection status=$status newState=$newState")
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                if (status != BluetoothGatt.GATT_SUCCESS) {
                    HudBus.setStatus("BLE connect error $status")
                    cleanupGatt(g)
                    rescanIfWanted()
                    return
                }
                HudBus.setStatus("BLE connected, discovering…")
                // Many stacks drop the link if discoverServices runs too early.
                mainHandler.postDelayed({
                    if (gatt !== g || !wantLink) return@postDelayed
                    runCatching { g.requestConnectionPriority(BluetoothGatt.CONNECTION_PRIORITY_HIGH) }
                    if (!g.discoverServices()) {
                        HudBus.setStatus("discoverServices rejected")
                        g.disconnect()
                    }
                }, 600)
                // If discovery never completes (common with flaky BlueZ), bail and rescan.
                mainHandler.postDelayed({
                    if (gatt === g && wantLink && !connected) {
                        Log.w(TAG, "service discovery timed out")
                        HudBus.setStatus("Discovery timeout — retrying")
                        runCatching { g.disconnect() }
                    }
                }, 8_000)
            } else if (newState == BluetoothProfile.STATE_DISCONNECTED) {
                connected = false
                clearChars()
                HudBus.setStatus(
                    if (status == BluetoothGatt.GATT_SUCCESS) "BLE disconnected"
                    else "BLE disconnected (status $status)",
                )
                cleanupGatt(g)
                rescanIfWanted()
            }
        }

        override fun onServicesDiscovered(g: BluetoothGatt, status: Int) {
            Log.i(TAG, "servicesDiscovered status=$status")
            if (g !== gatt) return
            if (status != BluetoothGatt.GATT_SUCCESS) {
                HudBus.setStatus("Service discovery failed: $status")
                g.disconnect()
                return
            }
            val svc = g.getService(UUID.fromString(Protocol.SERVICE_UUID))
            if (svc == null) {
                val uuids = g.services?.joinToString { it.uuid.toString() }.orEmpty()
                Log.w(TAG, "service missing; have=[$uuids]")
                HudBus.setStatus("Service not found ($uuids)")
                g.disconnect()
                return
            }
            navChar = svc.getCharacteristic(UUID.fromString(Protocol.NAV_UUID))
            mediaChar = svc.getCharacteristic(UUID.fromString(Protocol.MEDIA_UUID))
            cmdChar = svc.getCharacteristic(UUID.fromString(Protocol.CMD_UUID))
            hbChar = svc.getCharacteristic(UUID.fromString(Protocol.HEARTBEAT_UUID))
            cmdChar?.let { enableNotify(g, it) }
            connected = true
            HudBus.setStatus("HUD ready")
            // Flush current nav (incl. idle) — StateFlow may already have been
            // collected before GATT was up, so the Pi would keep a stale frame.
            writeNav(HudBus.nav.value)
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
        wantLink = true
        val scanner = adapter?.bluetoothLeScanner
        if (scanner == null) {
            HudBus.setStatus("Bluetooth unavailable")
            return
        }
        if (adapter?.isEnabled != true) {
            HudBus.setStatus("Bluetooth is off")
            return
        }
        // Stop any prior scan first — otherwise Android returns
        // SCAN_FAILED_ALREADY_STARTED (1) and the UI looks stuck.
        runCatching { scanner.stopScan(scanCallback) }

        HudBus.setStatus("Scanning for ${Protocol.DEVICE_NAME}…")
        // Do not filter on Service UUID: the Pi advertises LocalName "MotoHUD"
        // only (no UUID in the 31-byte ADV / legacy btmgmt path). Match by name
        // in onScanResult.
        val settings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
            .build()
        scanner.startScan(emptyList(), settings, scanCallback)
    }

    fun stopScan() {
        adapter?.bluetoothLeScanner?.stopScan(scanCallback)
    }

    fun close() {
        wantLink = false
        mainHandler.removeCallbacksAndMessages(null)
        stopScan()
        gatt?.disconnect()
        gatt?.close()
        gatt = null
        connected = false
        clearChars()
    }

    override fun writeNav(nav: NavState) {
        write(navChar, nav.toJson())
    }

    override fun writeMedia(media: MediaState) {
        write(mediaChar, media.toJson())
    }

    override fun writeHeartbeat() {
        val body = JSONObject().put("type", "heartbeat").put("ts", System.currentTimeMillis() / 1000).toString()
        write(hbChar, body.toByteArray(Charsets.UTF_8))
    }

    private fun rescanIfWanted() {
        if (!wantLink) return
        mainHandler.postDelayed({
            if (wantLink && !connected && gatt == null) startScan()
        }, 750)
    }

    private fun cleanupGatt(g: BluetoothGatt) {
        runCatching { g.close() }
        if (gatt === g) gatt = null
    }

    private fun clearChars() {
        navChar = null
        mediaChar = null
        cmdChar = null
        hbChar = null
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
