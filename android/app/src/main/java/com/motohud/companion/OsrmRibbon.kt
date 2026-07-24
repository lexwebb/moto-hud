package com.motohud.companion

import org.json.JSONObject
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.hypot
import kotlin.math.min
import kotlin.math.sin
import kotlin.math.sqrt

/**
 * Builds a short ahead-corridor ribbon from an OSRM route response.
 * Public demo server is used by [RibbonEnricher]; this object is pure JSON → points.
 */
object OsrmRibbon {
    private const val maxPoints = 6

    data class Result(val points: List<RibbonPoint>, val turnIndex: Int)

    fun corridorFromRouteJson(
        json: String,
        originLat: Double,
        originLon: Double,
        bearingDeg: Float,
        maneuver: String,
        distanceM: Int,
    ): Result? {
        val root = JSONObject(json)
        if (root.optString("code") != "Ok") return null
        val routes = root.optJSONArray("routes") ?: return null
        if (routes.length() == 0) return null
        val route = routes.getJSONObject(0)
        val geom = route.optJSONObject("geometry") ?: return null
        val coords = geom.optJSONArray("coordinates") ?: return null
        if (coords.length() < 2) return null

        val turnLonLat = findTurnLocation(route, maneuver, distanceM)
            ?: coords.getJSONArray(min(coords.length() - 1, coords.length() / 2)).let {
                it.getDouble(0) to it.getDouble(1)
            }

        val raw = ArrayList<Pair<Double, Double>>(coords.length())
        for (i in 0 until coords.length()) {
            val c = coords.getJSONArray(i)
            raw.add(c.getDouble(0) to c.getDouble(1))
        }

        val projected = project(originLat, originLon, bearingDeg, raw)
        if (projected.size < 2) return null

        // Keep from slightly behind origin through a little past the turn.
        val turnProj = projectOne(originLat, originLon, bearingDeg, turnLonLat.first, turnLonLat.second)
        val yMax = maxOf(turnProj.y + 40.0, distanceM.toDouble() + 80.0, 80.0)
        val clipped = projected.filter { it.y >= -25.0 && it.y <= yMax }
        val usable = if (clipped.size >= 2) clipped else projected
        val down = downsample(usable, maxPoints)

        var turnIdx = 0
        var best = Double.MAX_VALUE
        down.forEachIndexed { i, p ->
            val d = hypot(p.x - turnProj.x, p.y - turnProj.y)
            if (d < best) {
                best = d
                turnIdx = i
            }
        }
        return Result(down, turnIdx)
    }

    /** Destination ~aheadM along bearing for a short OSRM probe route. */
    fun destinationAhead(lat: Double, lon: Double, bearingDeg: Float, aheadM: Double): Pair<Double, Double> {
        val br = Math.toRadians(bearingDeg.toDouble())
        val north = aheadM * cos(br)
        val east = aheadM * sin(br)
        val dLat = north / 111_320.0
        val dLon = east / (111_320.0 * cos(Math.toRadians(lat)).coerceAtLeast(0.2))
        return lat + dLat to lon + dLon
    }

    private fun findTurnLocation(
        route: JSONObject,
        maneuver: String,
        distanceM: Int,
    ): Pair<Double, Double>? {
        val legs = route.optJSONArray("legs") ?: return null
        if (legs.length() == 0) return null
        val steps = legs.getJSONObject(0).optJSONArray("steps") ?: return null
        val want = osrmModifiers(maneuver)

        var bestMatch: Pair<Double, Double>? = null
        var bestDist = Int.MAX_VALUE
        var cum = 0.0
        var closestByDist: Pair<Double, Double>? = null
        var closestDelta = Double.MAX_VALUE

        for (i in 0 until steps.length()) {
            val step = steps.getJSONObject(i)
            val man = step.optJSONObject("maneuver") ?: continue
            val loc = man.optJSONArray("location") ?: continue
            val lon = loc.getDouble(0)
            val lat = loc.getDouble(1)
            val type = man.optString("type")
            val modifier = man.optString("modifier")

            val delta = kotlin.math.abs(cum - distanceM)
            if (delta < closestDelta && type != "depart") {
                closestDelta = delta
                closestByDist = lon to lat
            }

            if (want != null && type in want.types && (want.modifiers.isEmpty() || modifier in want.modifiers)) {
                val d = kotlin.math.abs(cum - distanceM).toInt()
                if (d < bestDist) {
                    bestDist = d
                    bestMatch = lon to lat
                }
            }
            cum += step.optDouble("distance", 0.0)
        }
        return bestMatch ?: closestByDist
    }

    private data class OsrmWant(val types: Set<String>, val modifiers: Set<String>)

    private fun osrmModifiers(maneuver: String): OsrmWant? = when (maneuver) {
        "left" -> OsrmWant(setOf("turn", "end of road"), setOf("left"))
        "right" -> OsrmWant(setOf("turn", "end of road"), setOf("right"))
        "slight_left" -> OsrmWant(setOf("turn", "fork", "off ramp"), setOf("slight left"))
        "slight_right" -> OsrmWant(setOf("turn", "fork", "off ramp"), setOf("slight right"))
        "u_turn" -> OsrmWant(setOf("continue", "turn"), setOf("uturn"))
        "roundabout" -> OsrmWant(setOf("roundabout", "rotary"), emptySet())
        "arrive" -> OsrmWant(setOf("arrive"), emptySet())
        "depart" -> OsrmWant(setOf("depart"), emptySet())
        "straight" -> OsrmWant(setOf("new name", "continue", "notification"), setOf("straight"))
        else -> null
    }

    private fun project(
        lat0: Double,
        lon0: Double,
        bearingDeg: Float,
        coords: List<Pair<Double, Double>>,
    ): List<RibbonPoint> = coords.map { (lon, lat) ->
        projectOne(lat0, lon0, bearingDeg, lon, lat)
    }

    private fun projectOne(
        lat0: Double,
        lon0: Double,
        bearingDeg: Float,
        lon: Double,
        lat: Double,
    ): RibbonPoint {
        val br = Math.toRadians(bearingDeg.toDouble())
        val north = (lat - lat0) * 111_320.0
        val east = (lon - lon0) * 111_320.0 * cos(Math.toRadians(lat0))
        val ahead = north * cos(br) + east * sin(br)
        val right = east * cos(br) - north * sin(br)
        return RibbonPoint(right, ahead)
    }

    private fun downsample(pts: List<RibbonPoint>, max: Int): List<RibbonPoint> {
        if (pts.size <= max) return pts
        if (max < 2) return pts.take(2)
        val out = ArrayList<RibbonPoint>(max)
        out.add(pts.first())
        val mid = max - 2
        for (i in 1..mid) {
            val t = i.toDouble() / (mid + 1)
            val idx = (t * (pts.size - 1)).toInt().coerceIn(1, pts.size - 2)
            out.add(pts[idx])
        }
        out.add(pts.last())
        return out
    }

    /** Bearing from two WGS84 points, degrees clockwise from north. */
    fun bearingBetween(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Float {
        val φ1 = Math.toRadians(lat1)
        val φ2 = Math.toRadians(lat2)
        val Δλ = Math.toRadians(lon2 - lon1)
        val y = sin(Δλ) * cos(φ2)
        val x = cos(φ1) * sin(φ2) - sin(φ1) * cos(φ2) * cos(Δλ)
        val θ = Math.toDegrees(atan2(y, x))
        return ((θ + 360.0) % 360.0).toFloat()
    }

    fun metersBetween(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        val dn = (lat2 - lat1) * 111_320.0
        val de = (lon2 - lon1) * 111_320.0 * cos(Math.toRadians(lat1))
        return sqrt(dn * dn + de * de)
    }
}
