package com.motohud.companion.osmand

import android.util.Log
import com.motohud.companion.JunctionArmHint
import com.motohud.companion.JunctionBuildInput
import com.motohud.companion.JunctionBuilder
import com.motohud.companion.JunctionMessage
import net.osmand.ResultMatcher
import net.osmand.binary.BinaryMapIndexReader
import net.osmand.binary.BinaryMapRouteReaderAdapter.RouteRegion
import net.osmand.binary.BinaryMapRouteReaderAdapter.RouteSubregion
import net.osmand.binary.RouteDataObject
import net.osmand.plus.OsmandApplication
import net.osmand.plus.routing.RouteDirectionInfo
import net.osmand.plus.routing.RoutingHelper
import net.osmand.router.RouteSegmentResult
import net.osmand.router.TurnType
import net.osmand.util.MapUtils
import kotlin.math.abs
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin

/**
 * Full Library → [JunctionMessage].
 *
 * Dual detection ports hardened rules from
 * `site/public/emulator/osmand-poc/DumpOsmandJunction.java` (attached + spatial OBF).
 * OBF has no `dual_carriageway` tag.
 */
object OsmandJunctionExtractor {

    private const val TAG = "OsmandJunction"
    private const val PARALLEL_DEG = 18.0
    private const val OPPOSITE_BEARING_MIN = 150.0
    private const val SEP_MIN_M = 10.0
    private const val SEP_MAX_M = 22.0
    private const val APPROACH_LOOKBACK = 6
    private const val SPATIAL_RADIUS_M = 50.0

    fun extract(
        app: OsmandApplication,
        rh: RoutingHelper,
        di: RouteDirectionInfo,
        maneuver: String,
    ): JunctionMessage {
        val drive = driveSide(app)
        val route = rh.route
        val segs = route?.immutableAllSegments
        if (segs.isNullOrEmpty()) {
            Log.d(TAG, "thin: no segments → minimal junction for $maneuver")
            return JunctionBuilder.fromManeuver(maneuver, drive)
        }

        val turnLoc = rh.getLocationFromRouteDirection(di)
        val turnIdx = findTurnSegmentIndex(segs, di, turnLoc)
        if (turnIdx < 0) {
            Log.d(TAG, "thin: turn segment not found → minimal junction for $maneuver")
            return JunctionBuilder.fromManeuver(maneuver, drive).let { enrichRoundabout(it, di.turnType) }
        }

        val turnSeg = segs[turnIdx]
        val approachSeg = if (turnIdx > 0) segs[turnIdx - 1] else turnSeg
        val approachHwy = approachSeg.`object`?.highway
        val turnHwy = turnSeg.`object`?.highway
        val approachBearing = approachSeg.bearingEnd

        val arms = collectCardinalArms(turnSeg, approachBearing, maneuver)
        val dual = inferDual(app, segs, turnIdx)
        val (exits, exit) = roundaboutMeta(di.turnType)

        val msg = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = maneuver,
                drive = drive,
                dual = dual.hit,
                approachHighway = approachHwy,
                turnHighway = turnHwy,
                arms = arms,
                roundaboutExits = exits,
                roundaboutExit = exit,
            ),
        )
        Log.i(
            TAG,
            "junction kind=${msg.kind} outbound=${msg.outbound} through=${msg.through} " +
                "dual=${dual.hit}${if (dual.hit) "(${dual.reason})" else ""} " +
                "arms=${arms.size} sides=${msg.sides.size} hwy=$turnHwy",
        )
        return msg
    }

    private fun enrichRoundabout(base: JunctionMessage, turn: TurnType?): JunctionMessage {
        if (base.kind != "roundabout" || turn == null || !turn.isRoundAbout) return base
        val (exits, exit) = roundaboutMeta(turn)
        return base.copy(exits = exits, exit = exit)
    }

    private fun roundaboutMeta(turn: TurnType?): Pair<Int, Int> {
        if (turn == null || !turn.isRoundAbout) return 0 to 0
        val exit = turn.exitOut.coerceAtLeast(1)
        val others = turn.otherTurnAngles?.size ?: 0
        val exits = (others + 1).coerceIn(2, 6).let { if (it < exit) exit.coerceAtMost(6) else it }
        return exits to exit.coerceAtMost(exits)
    }

    private fun driveSide(app: OsmandApplication): String {
        return try {
            val region = app.settings.DRIVING_REGION.get()
            if (region != null && region.leftHandDriving) "left" else "right"
        } catch (_: Exception) {
            "right"
        }
    }

    private fun findTurnSegmentIndex(
        segs: List<RouteSegmentResult>,
        di: RouteDirectionInfo,
        turnLoc: net.osmand.Location?,
    ): Int {
        val want = di.turnType?.value ?: return -1
        var best = -1
        var bestDist = Double.MAX_VALUE
        for (i in segs.indices) {
            val tt = segs[i].turnType ?: continue
            if (tt.value != want) continue
            if (turnLoc == null) return i
            val sp = segs[i].startPoint ?: continue
            val d = MapUtils.getDistance(sp.latitude, sp.longitude, turnLoc.latitude, turnLoc.longitude)
            if (d < bestDist) {
                bestDist = d
                best = i
            }
        }
        if (best >= 0) return best
        // Fallback: first segment with any turn type ahead of current.
        for (i in segs.indices) {
            if (segs[i].turnType != null) return i
        }
        return -1
    }

    private fun collectCardinalArms(
        turnSeg: RouteSegmentResult,
        approachBearing: Float,
        maneuver: String,
    ): List<JunctionArmHint> {
        val outSide = when (maneuver) {
            "left", "slight_left" -> "left"
            "right", "slight_right" -> "right"
            "straight" -> "through"
            else -> null
        }
        val arms = mutableListOf<JunctionArmHint>()
        val seen = mutableSetOf<String>()

        fun add(side: String, hwy: String?, isOutbound: Boolean) {
            if (side != "left" && side != "right" && side != "through") return
            val key = "$side:at"
            if (!seen.add(key) && !isOutbound) return
            arms.add(JunctionArmHint(side = side, at = "at", highway = hwy, isOutbound = isOutbound))
        }

        // Routed outbound always present.
        if (outSide != null) {
            add(outSide, turnSeg.`object`?.highway, isOutbound = true)
        }

        val pointIdx = turnSeg.startPointIndex
        val attached = turnSeg.getAttachedRoutes(pointIdx)
        if (attached != null) {
            for (a in attached) {
                val brg = a.bearingBegin
                val side = relativeSide(approachBearing, brg)
                add(side, a.`object`?.highway, isOutbound = false)
            }
        }

        // Through continuation: end bearing of turn segment if it continues past the node.
        val throughBrg = turnSeg.bearingEnd
        if (abs(bearingDelta(throughBrg, approachBearing)) < 35f) {
            add("through", turnSeg.`object`?.highway, isOutbound = maneuver == "straight")
        }

        return arms
    }

    private fun relativeSide(approachBearing: Float, armBearing: Float): String {
        val d = bearingDelta(armBearing, approachBearing)
        val a = abs(d)
        return when {
            a < 35f -> "through"
            a > 145f -> "back"
            d > 0f -> "right"
            else -> "left"
        }
    }

    // --- Dual inference (DumpOsmandJunction port) ---

    private data class DualHit(val hit: Boolean, val reason: String? = null) {
        companion object {
            fun yes(reason: String) = DualHit(true, reason)
            fun no() = DualHit(false)
        }
    }

    private data class Arm(
        val src: String,
        val segIdx: Int,
        val name: String?,
        val highway: String?,
        val oneway: Int,
        val bearing: Float,
        val lat: Double,
        val lon: Double,
        val midLat: Double,
        val midLon: Double,
    )

    private fun inferDual(
        app: OsmandApplication,
        segs: List<RouteSegmentResult>,
        turnIdx: Int,
    ): DualHit {
        val turnObj = segs[turnIdx].`object` ?: return DualHit.no()
        val hwy = turnObj.highway
        if (!JunctionBuilder.isMajorHighway(hwy) || JunctionBuilder.isLink(hwy)) {
            return DualHit.no()
        }
        var hit = scoreArms(collectArms(segs, turnIdx, 0), turnIdx, allowBearingSep = true)
        if (hit.hit) return hit
        hit = inferDualSpatial(app, segs[turnIdx])
        if (hit.hit) return hit
        return scoreArms(collectArms(segs, turnIdx, APPROACH_LOOKBACK), turnIdx, allowBearingSep = false)
    }

    private fun collectArms(segs: List<RouteSegmentResult>, turnIdx: Int, lookback: Int): List<Arm> {
        val arms = mutableListOf<Arm>()
        val from = (turnIdx - lookback).coerceAtLeast(0)
        for (i in from..turnIdx) {
            val s = segs[i]
            arms.add(armOf(s, "route:$i", i))
            val start = minOf(s.startPointIndex, s.endPointIndex)
            val end = maxOf(s.startPointIndex, s.endPointIndex)
            for (pi in start..end) {
                val att = s.getAttachedRoutes(pi) ?: continue
                for (a in att) {
                    arms.add(armOf(a, "att:$i@$pi", i))
                }
            }
        }
        return arms
    }

    private fun armOf(s: RouteSegmentResult, src: String, segIdx: Int): Arm {
        val o = s.`object`
        val start = s.startPoint
        val end = s.endPoint
        val midLat = (start.latitude + end.latitude) / 2.0
        val midLon = (start.longitude + end.longitude) / 2.0
        return Arm(
            src = src,
            segIdx = segIdx,
            name = safeName(o),
            highway = o?.highway,
            oneway = o?.oneway ?: 0,
            bearing = s.bearingBegin,
            lat = start.latitude,
            lon = start.longitude,
            midLat = midLat,
            midLon = midLon,
        )
    }

    private fun scoreArms(arms: List<Arm>, turnIdx: Int, allowBearingSep: Boolean): DualHit {
        var best = DualHit.no()
        for (a in arms.indices) {
            for (b in a + 1 until arms.size) {
                val x = arms[a]
                val y = arms[b]
                if (!allowBearingSep && x.segIdx != turnIdx && y.segIdx != turnIdx) continue
                val cand = scorePair(x, y, allowBearingSep) ?: continue
                val candAtTurn = x.segIdx == turnIdx || y.segIdx == turnIdx
                val bestAtTurn = best.hit && best.reason?.contains("@turn") == true
                if (!best.hit || (candAtTurn && !bestAtTurn)) {
                    best = if (candAtTurn) DualHit.yes(cand + "@turn") else DualHit.yes(cand)
                }
            }
        }
        return best
    }

    private fun scorePair(x: Arm, y: Arm, allowBearingSep: Boolean): String? {
        if (!JunctionBuilder.isMajorHighway(x.highway) && !JunctionBuilder.isMajorHighway(y.highway)) {
            return null
        }
        if (JunctionBuilder.isLink(x.highway) != JunctionBuilder.isLink(y.highway) && sameName(x.name, y.name)) {
            if (!oppositeOneway(x, y)) return null
        }
        val named = sameName(x.name, y.name)
        val oppOne = oppositeOneway(x, y)
        val oppBrg = oppositeBearing(x.bearing, y.bearing)
        if (!oppOne && !oppBrg) return null
        if (!parallelOrAnti(x.bearing, y.bearing)) return null

        val latSep = lateralSeparationM(x, y)
        val nodeDist = MapUtils.getDistance(x.lat, x.lon, y.lat, y.lon)

        if (oppOne && (named || (JunctionBuilder.isMajorHighway(x.highway) && JunctionBuilder.isMajorHighway(y.highway)))) {
            if (nodeDist <= SEP_MAX_M || (latSep in SEP_MIN_M..SEP_MAX_M)) {
                val label = if (named) x.name else "major"
                return "opposite_oneway:$label"
            }
        }
        if (allowBearingSep && oppBrg && latSep in SEP_MIN_M..SEP_MAX_M) {
            if (x.oneway == 0 && y.oneway == 0) return null
            if (named) return "opposite_bearing_sep:${x.name}"
            if (JunctionBuilder.isMajorHighway(x.highway) && JunctionBuilder.isMajorHighway(y.highway)) {
                return "major_opposite_bearing_sep"
            }
        }
        return null
    }

    private fun inferDualSpatial(app: OsmandApplication, seg: RouteSegmentResult): DualHit {
        return try {
            val selfObj = seg.`object` ?: return DualHit.no()
            val self = armOf(seg, "route", -1)
            if (!JunctionBuilder.isMajorHighway(self.highway) || JunctionBuilder.isLink(self.highway)) {
                return DualHit.no()
            }
            val p = seg.startPoint ?: return DualHit.no()
            val readers = app.resourceManager.routingMapFiles ?: return DualHit.no()
            val nearby = loadNearbyRoutes(readers, p.latitude, p.longitude, SPATIAL_RADIUS_M)
            val selfId = selfObj.id
            for (other in nearby) {
                if (other.id == selfId) continue
                if (!isCarHighway(other.highway) || JunctionBuilder.isLink(other.highway)) continue
                val o = armFromRdo(other, "spatial") ?: continue
                if (self.oneway == 0 && o.oneway != 0) continue
                if (!(self.oneway != 0 && o.oneway != 0 && oppositeBearing(self.bearing, o.bearing)) &&
                    shareAnyNode(selfObj, other)
                ) {
                    continue
                }
                val cand = scorePair(self, o, allowBearingSep = true)
                if (cand != null) return DualHit.yes("spatial:$cand")
            }
            DualHit.no()
        } catch (e: Exception) {
            Log.w(TAG, "spatial dual scan failed (attached-only dual still applies)", e)
            DualHit.no()
        }
    }

    private fun loadNearbyRoutes(
        readers: Array<BinaryMapIndexReader>,
        lat: Double,
        lon: Double,
        radiusM: Double,
    ): List<RouteDataObject> {
        val dLat = radiusM / 111320.0
        val dLon = radiusM / (111320.0 * cos(Math.toRadians(lat)))
        val left = MapUtils.get31TileNumberX(lon - dLon)
        val right = MapUtils.get31TileNumberX(lon + dLon)
        val top = MapUtils.get31TileNumberY(lat + dLat)
        val bottom = MapUtils.get31TileNumberY(lat - dLat)
        val out = mutableListOf<RouteDataObject>()
        val matcher = object : ResultMatcher<RouteDataObject> {
            override fun publish(obj: RouteDataObject): Boolean {
                out.add(obj)
                return false
            }

            override fun isCancelled(): Boolean = false
        }
        val req = BinaryMapIndexReader.buildSearchRouteRequest(left, right, top, bottom, matcher)
        for (reader in readers) {
            val toLoad = ArrayList<RouteSubregion>()
            for (reg: RouteRegion in reader.routingIndexes) {
                val roots = ArrayList(reg.subregions)
                toLoad.addAll(reader.searchRouteIndexTree(req, roots))
            }
            reader.loadRouteIndexData(toLoad, matcher)
        }
        return out
    }

    private fun armFromRdo(o: RouteDataObject, src: String): Arm? {
        if (o.pointsLength < 2) return null
        val lat0 = MapUtils.get31LatitudeY(o.getPoint31YTile(0))
        val lon0 = MapUtils.get31LongitudeX(o.getPoint31XTile(0))
        val i1 = minOf(1, o.pointsLength - 1)
        val lat1 = MapUtils.get31LatitudeY(o.getPoint31YTile(i1))
        val lon1 = MapUtils.get31LongitudeX(o.getPoint31XTile(i1))
        var brg = bearingBetween(lat0, lon0, lat1, lon1).toFloat()
        if (o.oneway < 0) {
            brg = bearingBetween(lat1, lon1, lat0, lon0).toFloat()
        }
        val iMid = minOf(maxOf(1, o.pointsLength / 8), o.pointsLength - 1)
        val latM = MapUtils.get31LatitudeY(o.getPoint31YTile(iMid))
        val lonM = MapUtils.get31LongitudeX(o.getPoint31XTile(iMid))
        return Arm(src, -1, safeName(o), o.highway, o.oneway, brg, lat0, lon0, latM, lonM)
    }

    private fun shareAnyNode(a: RouteDataObject, b: RouteDataObject): Boolean {
        for (i in 0 until a.pointsLength) {
            val ax = a.getPoint31XTile(i)
            val ay = a.getPoint31YTile(i)
            for (j in 0 until b.pointsLength) {
                if (ax == b.getPoint31XTile(j) && ay == b.getPoint31YTile(j)) return true
            }
        }
        return false
    }

    private fun lateralSeparationM(a: Arm, b: Arm): Double {
        val ctA = crossTrackM(a.lat, a.lon, a.bearing, b.midLat, b.midLon)
        val ctB = crossTrackM(b.lat, b.lon, b.bearing, a.midLat, a.midLon)
        return (ctA + ctB) / 2.0
    }

    private fun crossTrackM(
        lat0: Double,
        lon0: Double,
        bearingDeg: Float,
        lat: Double,
        lon: Double,
    ): Double {
        val dist = MapUtils.getDistance(lat0, lon0, lat, lon)
        if (dist < 1) return 0.0
        val brgTo = bearingBetween(lat0, lon0, lat, lon)
        val delta = bearingDelta(bearingDeg, brgTo.toFloat())
        return abs(sin(Math.toRadians(delta.toDouble())) * dist)
    }

    private fun bearingBetween(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
        val φ1 = Math.toRadians(lat1)
        val φ2 = Math.toRadians(lat2)
        val Δλ = Math.toRadians(lon2 - lon1)
        val y = sin(Δλ) * cos(φ2)
        val x = cos(φ1) * sin(φ2) - sin(φ1) * cos(φ2) * cos(Δλ)
        return Math.toDegrees(atan2(y, x))
    }

    private fun oppositeOneway(a: Arm, b: Arm): Boolean {
        if (a.oneway == 0 || b.oneway == 0) return false
        return oppositeBearing(a.bearing, b.bearing)
    }

    private fun oppositeBearing(a: Float, b: Float): Boolean =
        abs(bearingDelta(a, b)) >= OPPOSITE_BEARING_MIN

    private fun parallelOrAnti(a: Float, b: Float): Boolean {
        val d = abs(bearingDelta(a, b).toDouble())
        return d <= PARALLEL_DEG || d >= (180.0 - PARALLEL_DEG)
    }

    private fun bearingDelta(a: Float, b: Float): Float {
        var d = a - b
        while (d > 180f) d -= 360f
        while (d < -180f) d += 360f
        return d
    }

    private fun sameName(a: String?, b: String?): Boolean {
        if (a.isNullOrBlank() || b.isNullOrBlank()) return false
        return a.equals(b, ignoreCase = true)
    }

    private fun isCarHighway(hwy: String?): Boolean {
        if (hwy == null) return false
        if (JunctionBuilder.isMajorHighway(hwy)) return true
        val base = hwy.removeSuffix("_link")
        return base == "tertiary" || base == "unclassified" || base == "residential"
    }

    private fun safeName(o: RouteDataObject?): String? {
        if (o == null) return null
        return try {
            o.name?.takeIf { it.isNotBlank() }
        } catch (_: Throwable) {
            null
        }
    }
}
