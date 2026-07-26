package com.motohud.companion

import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.os.Build
import android.os.IBinder
import android.os.RemoteException
import android.util.Log
import android.view.KeyEvent
import net.osmand.aidlapi.IOsmAndAidlCallback
import net.osmand.aidlapi.IOsmAndAidlInterface
import net.osmand.aidlapi.gpx.AGpxBitmap
import net.osmand.aidlapi.logcat.OnLogcatMessageParams
import net.osmand.aidlapi.navigation.ADirectionInfo
import net.osmand.aidlapi.navigation.ANavigationUpdateParams
import net.osmand.aidlapi.navigation.OnVoiceNavigationParams
import net.osmand.aidlapi.search.SearchResult

/**
 * Binds to OsmAnd's AIDL service and subscribes to typed turn updates.
 * Prefers free OsmAnd, then OsmAnd+.
 */
class OsmandNavClient(private val app: Context) {

    private var api: IOsmAndAidlInterface? = null
    private var callbackId = -1L
    private var bound = false

    private val callback = object : IOsmAndAidlCallback.Stub() {
        override fun onSearchComplete(resultSet: List<SearchResult>?) {}
        override fun onUpdate() {}
        override fun onAppInitialized() {}
        override fun onGpxBitmapCreated(bitmap: AGpxBitmap?) {}

        override fun updateNavigationInfo(directionInfo: ADirectionInfo?) {
            val info = directionInfo ?: return
            val nav = OsmandNavMapper.toNavState(info)
            Log.d(TAG, "osmand nav: $nav (turn=${info.turnType} dist=${info.distanceTo})")
            HudBus.publishNav(nav, NavSource.OSMAND)
        }

        override fun onContextMenuButtonClicked(buttonId: Int, pointId: String?, layerId: String?) {}
        override fun onVoiceRouterNotify(params: OnVoiceNavigationParams?) {}
        override fun onKeyEvent(params: KeyEvent?) {}
        override fun onLogcatMessage(params: OnLogcatMessageParams?) {}
    }

    private val connection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, service: IBinder?) {
            api = IOsmAndAidlInterface.Stub.asInterface(service)
            bound = true
            HudBus.setOsmandBound(true)
            subscribe()
            Log.i(TAG, "bound to ${name?.packageName}")
            HudBus.setStatus("OsmAnd AIDL bound")
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            Log.w(TAG, "OsmAnd disconnected")
            api = null
            bound = false
            callbackId = -1L
            HudBus.setOsmandBound(false)
        }
    }

    fun start() {
        if (bound) return
        val pkg = installedOsmandPackage()
        if (pkg == null) {
            Log.i(TAG, "OsmAnd not installed; Maps notification scrape only")
            HudBus.setOsmandBound(false)
            return
        }
        val intent = Intent(SERVICE_ACTION).setPackage(pkg)
        var flags = Context.BIND_AUTO_CREATE
        if (Build.VERSION.SDK_INT >= 34) {
            flags = flags or Context.BIND_ALLOW_ACTIVITY_STARTS
        }
        val ok = app.bindService(intent, connection, flags)
        if (!ok) {
            Log.w(TAG, "bindService failed for $pkg")
            HudBus.setOsmandBound(false)
        }
    }

    fun stop() {
        unsubscribe()
        if (bound) {
            try {
                app.unbindService(connection)
            } catch (_: IllegalArgumentException) {
            }
        }
        api = null
        bound = false
        callbackId = -1L
        HudBus.setOsmandBound(false)
    }

    private fun subscribe() {
        val iface = api ?: return
        try {
            val params = ANavigationUpdateParams().apply {
                setSubscribeToUpdates(true)
                setCallbackId(-1L)
            }
            callbackId = iface.registerForNavigationUpdates(params, callback)
            Log.i(TAG, "registerForNavigationUpdates → $callbackId")
        } catch (e: RemoteException) {
            Log.e(TAG, "registerForNavigationUpdates failed", e)
        }
    }

    private fun unsubscribe() {
        val iface = api ?: return
        if (callbackId < 0) return
        try {
            val params = ANavigationUpdateParams().apply {
                setSubscribeToUpdates(false)
                setCallbackId(callbackId)
            }
            iface.registerForNavigationUpdates(params, callback)
        } catch (e: RemoteException) {
            Log.w(TAG, "unsubscribe failed", e)
        }
    }

    private fun installedOsmandPackage(): String? {
        val pm = app.packageManager
        for (pkg in PACKAGES) {
            try {
                pm.getPackageInfo(pkg, 0)
                return pkg
            } catch (_: PackageManager.NameNotFoundException) {
            }
        }
        return null
    }

    companion object {
        private const val TAG = "OsmandNav"
        const val SERVICE_ACTION = "net.osmand.aidl.OsmandAidlServiceV2"
        val PACKAGES = listOf("net.osmand", "net.osmand.plus")
    }
}
