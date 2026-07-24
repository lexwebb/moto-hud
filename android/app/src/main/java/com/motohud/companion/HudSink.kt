package com.motohud.companion

import java.net.HttpURLConnection
import java.net.URL

/** Outbound HUD sink (BLE, HTTP, …). */
interface HudSink {
    fun writeNav(nav: NavState)
    fun writeMedia(media: MediaState)
    fun writeHeartbeat()
}

/**
 * Posts nav/media JSON to a local motohud HTTP injector
 * (`motohud -host png -http :8787` → POST /nav, /media).
 */
class HttpHudSink(
    private val baseUrl: String,
) : HudSink {
    private fun post(path: String, body: ByteArray) {
        val url = URL(baseUrl.trimEnd('/') + path)
        val conn = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            doOutput = true
            connectTimeout = 3_000
            readTimeout = 3_000
            setRequestProperty("Content-Type", "application/json")
        }
        try {
            conn.outputStream.use { it.write(body) }
            conn.responseCode
        } finally {
            conn.disconnect()
        }
    }

    override fun writeNav(nav: NavState) {
        post("/nav", nav.toJson())
    }

    override fun writeMedia(media: MediaState) {
        post("/media", media.toJson())
    }

    override fun writeHeartbeat() {
        // Pi HTTP hub has no heartbeat endpoint; link is implicit while posts succeed.
    }
}

object LinkPrefs {
    private const val PREFS = "motohud_link"
    private const val KEY_HTTP_URL = "http_base_url"
    private const val KEY_HTTP_ENABLED = "http_enabled"

    fun httpEnabled(ctx: android.content.Context): Boolean =
        ctx.getSharedPreferences(PREFS, android.content.Context.MODE_PRIVATE)
            .getBoolean(KEY_HTTP_ENABLED, false)

    fun httpBaseUrl(ctx: android.content.Context): String =
        ctx.getSharedPreferences(PREFS, android.content.Context.MODE_PRIVATE)
            .getString(KEY_HTTP_URL, "http://10.0.2.2:8787") ?: "http://10.0.2.2:8787"

    fun setHttp(ctx: android.content.Context, enabled: Boolean, baseUrl: String) {
        ctx.getSharedPreferences(PREFS, android.content.Context.MODE_PRIVATE).edit()
            .putBoolean(KEY_HTTP_ENABLED, enabled)
            .putString(KEY_HTTP_URL, baseUrl.trim())
            .apply()
    }
}
