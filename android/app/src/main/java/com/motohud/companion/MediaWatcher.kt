package com.motohud.companion

import android.content.ComponentName
import android.content.Context
import android.media.MediaMetadata
import android.media.session.MediaController
import android.media.session.MediaSessionManager
import android.media.session.PlaybackState

class MediaWatcher(private val context: Context) : MediaSessionManager.OnActiveSessionsChangedListener {

    private val msm = context.getSystemService(Context.MEDIA_SESSION_SERVICE) as MediaSessionManager
    private var controller: MediaController? = null

    private val callback = object : MediaController.Callback() {
        override fun onMetadataChanged(metadata: MediaMetadata?) = publish()
        override fun onPlaybackStateChanged(state: PlaybackState?) = publish()
    }

    fun start() {
        val cn = ComponentName(context, NavNotificationListener::class.java)
        try {
            msm.addOnActiveSessionsChangedListener(this, cn)
            onActiveSessionsChanged(msm.getActiveSessions(cn))
        } catch (e: SecurityException) {
            HudBus.setStatus("Media: grant notification access")
        }
    }

    fun stop() {
        controller?.unregisterCallback(callback)
        controller = null
        try {
            msm.removeOnActiveSessionsChangedListener(this)
        } catch (_: Exception) {
        }
    }

    override fun onActiveSessionsChanged(controllers: MutableList<MediaController>?) {
        controller?.unregisterCallback(callback)
        controller = controllers?.firstOrNull()
        controller?.registerCallback(callback)
        publish()
    }

    fun dispatch(action: String) {
        val c = controller ?: return
        when (action) {
            "play_pause" -> {
                val st = c.playbackState?.state
                if (st == PlaybackState.STATE_PLAYING) c.transportControls.pause()
                else c.transportControls.play()
            }
            "next_track" -> c.transportControls.skipToNext()
            "prev_track" -> c.transportControls.skipToPrevious()
        }
    }

    private fun publish() {
        val c = controller
        val meta = c?.metadata
        val playing = c?.playbackState?.state == PlaybackState.STATE_PLAYING
        HudBus.publishMedia(
            MediaState(
                playing = playing,
                title = meta?.getString(MediaMetadata.METADATA_KEY_TITLE).orEmpty(),
                artist = meta?.getString(MediaMetadata.METADATA_KEY_ARTIST).orEmpty(),
            )
        )
    }
}
