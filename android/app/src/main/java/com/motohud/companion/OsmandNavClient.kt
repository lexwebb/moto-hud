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
import net.osmand.aidlapi.navigation.ANavigationVoiceRouterMessageParams
import net.osmand.aidlapi.navigation.OnVoiceNavigationParams
import net.osmand.aidlapi.search.SearchResult

/**
 * OsmAnd AIDL nav engine: typed turn + distance, optional voice-router hints.
 * Lanes require the embedded Full Library flavor ([OsmandEmbeddedNavEngine]).
 */
class OsmandNavClient(private val app: Context) : NavEngine {

    override val id: String = "osmand-aidl"

    private var api: IOsmAndAidlInterface? = null
    private var navCallbackId = -1L
    private var voiceCallbackId = -1L
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

        override fun onVoiceRouterNotify(params: OnVoiceNavigationParams?) {
            val cmds = params?.commands ?: return
            if (cmds.isEmpty()) return
            // Voice cmds are TTS tokens; join as a soft instruction hint when useful.
            val joined = cmds.filterNotNull().joinToString(" ").trim()
            if (joined.isBlank()) return
            Log.d(TAG, "osmand voice: $joined")
            HudBus.publishNav(
                NavState(active = true, instruction = joined, road = guessRoadFromVoice(joined)),
                NavSource.OSMAND_ENRICH,
            )
        }

        override fun onKeyEvent(params: KeyEvent?) {}
        override fun onLogcatMessage(params: OnLogcatMessageParams?) {}
    }

    private val connection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, service: IBinder?) {
            api = IOsmAndAidlInterface.Stub.asInterface(service)
            bound = true
            HudBus.setOsmandBound(true)
            subscribeNav()
            subscribeVoice()
            Log.i(TAG, "bound to ${name?.packageName}")
            HudBus.setStatus("OsmAnd AIDL bound")
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            Log.w(TAG, "OsmAnd disconnected")
            api = null
            bound = false
            navCallbackId = -1L
            voiceCallbackId = -1L
            HudBus.setOsmandBound(false)
        }
    }

    override fun start() {
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

    override fun stop() {
        unsubscribeNav()
        unsubscribeVoice()
        if (bound) {
            try {
                app.unbindService(connection)
            } catch (_: IllegalArgumentException) {
            }
        }
        api = null
        bound = false
        navCallbackId = -1L
        voiceCallbackId = -1L
        HudBus.setOsmandBound(false)
    }

    private fun subscribeNav() {
        val iface = api ?: return
        try {
            val params = ANavigationUpdateParams().apply {
                setSubscribeToUpdates(true)
                setCallbackId(-1L)
            }
            navCallbackId = iface.registerForNavigationUpdates(params, callback)
            Log.i(TAG, "registerForNavigationUpdates → $navCallbackId")
        } catch (e: RemoteException) {
            Log.e(TAG, "registerForNavigationUpdates failed", e)
        }
    }

    private fun unsubscribeNav() {
        val iface = api ?: return
        if (navCallbackId < 0) return
        try {
            val params = ANavigationUpdateParams().apply {
                setSubscribeToUpdates(false)
                setCallbackId(navCallbackId)
            }
            iface.registerForNavigationUpdates(params, callback)
        } catch (e: RemoteException) {
            Log.w(TAG, "unsubscribe nav failed", e)
        }
    }

    private fun subscribeVoice() {
        val iface = api ?: return
        try {
            val params = ANavigationVoiceRouterMessageParams().apply {
                setSubscribeToUpdates(true)
                setCallbackId(-1L)
            }
            voiceCallbackId = iface.registerForVoiceRouterMessages(params, callback)
            Log.i(TAG, "registerForVoiceRouterMessages → $voiceCallbackId")
        } catch (e: RemoteException) {
            Log.w(TAG, "registerForVoiceRouterMessages failed", e)
        } catch (e: Exception) {
            Log.w(TAG, "voice router unavailable", e)
        }
    }

    private fun unsubscribeVoice() {
        val iface = api ?: return
        if (voiceCallbackId < 0) return
        try {
            val params = ANavigationVoiceRouterMessageParams().apply {
                setSubscribeToUpdates(false)
                setCallbackId(voiceCallbackId)
            }
            iface.registerForVoiceRouterMessages(params, callback)
        } catch (_: Exception) {
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

        fun guessRoadFromVoice(text: String): String {
            // Common TTS patterns: "Turn left onto High Street" / "onto the A40"
            val onto = Regex("""\bonto\s+(?:the\s+)?(.+)$""", RegexOption.IGNORE_CASE).find(text)
            if (onto != null) return onto.groupValues[1].trim().trimEnd('.', ',')
            return ""
        }
    }
}
