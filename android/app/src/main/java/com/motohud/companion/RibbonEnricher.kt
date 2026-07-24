package com.motohud.companion

import android.Manifest
import android.annotation.SuppressLint
import android.content.Context
import android.content.pm.PackageManager
import android.location.Location
import android.location.LocationManager
import android.util.Log
import androidx.core.content.ContextCompat
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.net.HttpURLConnection
import java.net.URL
import java.util.Locale
import kotlin.math.max

/**
 * Optionally attaches OSRM corridor points to nav. On failure or missing GPS,
 * returns nav unchanged so the Pi falls back to synthetic ribbons.
 *
 * Uses the public OSRM demo (`router.project-osrm.org`) — rate-limited; fine for
 * personal use until it becomes a problem.
 */
class RibbonEnricher(private val context: Context) {

    private var cacheKey: String? = null
    private var cacheResult: OsrmRibbon.Result? = null
    private var prevFix: Location? = null

    suspend fun enrich(nav: NavState): NavState {
        if (!nav.active) {
            cacheKey = null
            cacheResult = null
            return nav
        }
        return withContext(Dispatchers.IO) {
            val loc = lastLocation() ?: return@withContext nav
            val bearing = resolveBearing(loc)
            val key = "${nav.maneuver}|${nav.distanceM / 50}|${(loc.latitude * 2000).toInt()}|${(loc.longitude * 2000).toInt()}"
            cacheResult?.let { cached ->
                if (key == cacheKey) {
                    return@withContext nav.copy(ribbonPoints = cached.points, ribbonTurn = cached.turnIndex)
                }
            }

            val aheadM = max(400.0, nav.distanceM + 150.0)
            val (dLat, dLon) = OsrmRibbon.destinationAhead(loc.latitude, loc.longitude, bearing, aheadM)
            val url = String.format(
                Locale.US,
                "https://router.project-osrm.org/route/v1/driving/%.6f,%.6f;%.6f,%.6f?overview=full&geometries=geojson&steps=true",
                loc.longitude, loc.latitude, dLon, dLat,
            )
            try {
                val body = httpGet(url) ?: return@withContext nav
                val result = OsrmRibbon.corridorFromRouteJson(
                    body, loc.latitude, loc.longitude, bearing, nav.maneuver, nav.distanceM,
                ) ?: return@withContext nav
                cacheKey = key
                cacheResult = result
                nav.copy(ribbonPoints = result.points, ribbonTurn = result.turnIndex)
            } catch (e: Exception) {
                Log.w(TAG, "OSRM ribbon failed", e)
                nav
            }
        }
    }

    @SuppressLint("MissingPermission")
    private fun lastLocation(): Location? {
        if (ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_FINE_LOCATION)
            != PackageManager.PERMISSION_GRANTED
        ) {
            return null
        }
        val lm = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
        val providers = listOf(
            LocationManager.GPS_PROVIDER,
            LocationManager.NETWORK_PROVIDER,
            LocationManager.PASSIVE_PROVIDER,
        )
        var best: Location? = null
        for (p in providers) {
            try {
                val loc = lm.getLastKnownLocation(p) ?: continue
                if (best == null || loc.time > best.time) best = loc
            } catch (_: Exception) {
            }
        }
        return best
    }

    private fun resolveBearing(loc: Location): Float {
        if (loc.hasBearing() && loc.bearing != 0f) {
            prevFix = loc
            return loc.bearing
        }
        val prev = prevFix
        prevFix = loc
        if (prev != null && OsrmRibbon.metersBetween(prev.latitude, prev.longitude, loc.latitude, loc.longitude) > 8) {
            return OsrmRibbon.bearingBetween(prev.latitude, prev.longitude, loc.latitude, loc.longitude)
        }
        return 0f
    }

    private fun httpGet(url: String): String? {
        val conn = (URL(url).openConnection() as HttpURLConnection).apply {
            connectTimeout = 4_000
            readTimeout = 6_000
            requestMethod = "GET"
            setRequestProperty("User-Agent", "MotoHUD/0.1")
        }
        return try {
            val code = conn.responseCode
            if (code !in 200..299) {
                Log.w(TAG, "OSRM HTTP $code")
                null
            } else {
                conn.inputStream.bufferedReader().use { it.readText() }
            }
        } finally {
            conn.disconnect()
        }
    }

    companion object {
        private const val TAG = "RibbonEnricher"
    }
}
