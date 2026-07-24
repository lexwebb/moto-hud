package com.motohud.companion

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class HudForegroundService : Service() {

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    private lateinit var ble: BleClient
    private lateinit var media: MediaWatcher
    private var http: HttpHudSink? = null
    private var jobs: Job? = null

    override fun onCreate() {
        super.onCreate()
        createChannel()
        ble = BleClient(this)
        media = MediaWatcher(this)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        startForeground(NOTIF_ID, buildNotification("Connecting…"))
        media.start()
        ble.startScan()
        http = if (LinkPrefs.httpEnabled(this)) {
            HttpHudSink(LinkPrefs.httpBaseUrl(this)).also {
                HudBus.setStatus("HTTP → ${LinkPrefs.httpBaseUrl(this)}")
            }
        } else {
            null
        }
        jobs?.cancel()
        jobs = scope.launch {
            launch {
                HudBus.nav.collect { nav ->
                    dispatchNav(nav)
                    updateNotif()
                }
            }
            launch {
                HudBus.media.collect { m ->
                    dispatchMedia(m)
                }
            }
            launch {
                HudBus.cmds.collect { action ->
                    media.dispatch(action)
                }
            }
            launch {
                HudBus.status.collect { updateNotif() }
            }
            while (isActive) {
                dispatchHeartbeat()
                delay(15_000)
            }
        }
        return START_STICKY
    }

    private suspend fun dispatchNav(nav: NavState) {
        if (ble.connected) ble.writeNav(nav)
        val sink = http ?: return
        withContext(Dispatchers.IO) {
            try {
                sink.writeNav(nav)
            } catch (e: Exception) {
                Log.w(TAG, "HTTP nav failed", e)
                HudBus.setStatus("HTTP nav failed: ${e.message}")
            }
        }
    }

    private suspend fun dispatchMedia(m: MediaState) {
        if (ble.connected) ble.writeMedia(m)
        val sink = http ?: return
        withContext(Dispatchers.IO) {
            try {
                sink.writeMedia(m)
            } catch (e: Exception) {
                Log.w(TAG, "HTTP media failed", e)
            }
        }
    }

    private suspend fun dispatchHeartbeat() {
        if (ble.connected) ble.writeHeartbeat()
        val sink = http ?: return
        withContext(Dispatchers.IO) {
            try {
                sink.writeHeartbeat()
            } catch (_: Exception) {
            }
        }
    }

    override fun onDestroy() {
        jobs?.cancel()
        scope.cancel()
        media.stop()
        ble.close()
        http = null
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun updateNotif() {
        val nm = getSystemService(NOTIFICATION_SERVICE) as NotificationManager
        nm.notify(NOTIF_ID, buildNotification(HudBus.status.value))
    }

    private fun buildNotification(status: String): Notification {
        val open = PendingIntent.getActivity(
            this, 0, Intent(this, MainActivity::class.java),
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        return NotificationCompat.Builder(this, CHANNEL)
            .setContentTitle(getString(R.string.app_name))
            .setContentText(status)
            .setSmallIcon(android.R.drawable.stat_sys_data_bluetooth)
            .setContentIntent(open)
            .setOngoing(true)
            .build()
    }

    private fun createChannel() {
        val ch = NotificationChannel(CHANNEL, "HUD link", NotificationManager.IMPORTANCE_LOW)
        (getSystemService(NOTIFICATION_SERVICE) as NotificationManager).createNotificationChannel(ch)
    }

    companion object {
        private const val TAG = "HudService"
        private const val CHANNEL = "motohud_link"
        private const val NOTIF_ID = 42
    }
}
