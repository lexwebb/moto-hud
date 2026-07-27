package com.motohud.companion.osmand

import android.content.Context
import android.os.Handler
import android.os.Looper
import android.util.Log
import com.motohud.companion.HudBus
import com.motohud.companion.JunctionBuilder
import com.motohud.companion.ManeuverParser
import com.motohud.companion.NavEngine
import com.motohud.companion.NavSource
import com.motohud.companion.NavState
import com.motohud.companion.OsmandLaneCodec
import com.motohud.companion.ThenNext
import net.osmand.plus.OsmandApplication
import net.osmand.plus.routing.NextDirectionInfo
import net.osmand.router.TurnType

/**
 * In-process OsmAnd Full Library engine: polls RoutingHelper for turn, street,
 * then-next, ETA, and lane bitfields → protocol [NavState].
 *
 * Lives in the on-demand `:osmand` feature module. Region maps still download
 * at runtime via OsmAnd UI (mini basemap is not bundled).
 */
class OsmandEmbeddedNavEngine(private val app: Context) : NavEngine {

    override val id: String = "osmand-embedded"

    private val handler = Handler(Looper.getMainLooper())
    private var running = false
    private val tick = object : Runnable {
        override fun run() {
            if (!running) return
            publishFromRoutingHelper()
            handler.postDelayed(this, POLL_MS)
        }
    }

    override fun start() {
        if (running) return
        running = true
        HudBus.setOsmandBound(true)
        HudBus.setStatus("OsmAnd embedded engine")
        handler.post(tick)
    }

    override fun stop() {
        running = false
        handler.removeCallbacks(tick)
        HudBus.setOsmandBound(false)
    }

    private fun publishFromRoutingHelper() {
        val osmand = app.applicationContext as? OsmandApplication
        if (osmand == null) {
            Log.w(TAG, "Application is not OsmandApplication — restart after installing osmand module")
            return
        }
        val rh = osmand.routingHelper
        if (!rh.isRouteCalculated || !rh.isFollowingMode) {
            HudBus.publishNav(
                NavState(active = false, instruction = "Navigation ended"),
                NavSource.OSMAND,
            )
            return
        }

        if (rh.isDeviatedFromRoute) {
            val deviationM = rh.routeDeviation.toInt().coerceAtLeast(0)
            HudBus.publishNav(
                NavState(
                    active = true,
                    instruction = "Off route",
                    distanceM = deviationM,
                    distanceText = ManeuverParser.formatDistanceMeters(deviationM),
                    maneuver = "unknown",
                ),
                NavSource.OSMAND,
            )
            return
        }

        val next = NextDirectionInfo()
        val ndi = rh.getNextRouteDirectionInfo(next, true)
        if (ndi == null || ndi.distanceTo <= 0 || ndi.directionInfo == null) {
            return
        }
        val di = ndi.directionInfo
        val turn: TurnType = di.turnType
        val lanes = OsmandLaneCodec.decode(turn.lanes)
        val maneuver = ManeuverParser.fromOsmandTurnType(turn.value)
        val road = listOfNotNull(di.streetName, di.ref, di.destinationName)
            .map { it.trim() }
            .filter { it.isNotEmpty() }
            .joinToString(" · ")

        var thenNext: ThenNext? = null
        val after = NextDirectionInfo()
        val ndi2 = rh.getNextRouteDirectionInfoAfter(ndi, after, true)
        if (ndi2 != null && ndi2.distanceTo > 0 && ndi2.directionInfo != null) {
            val t2 = ndi2.directionInfo.turnType
            thenNext = ThenNext(
                maneuver = ManeuverParser.fromOsmandTurnType(t2.value),
                distanceM = ndi2.distanceTo,
                distanceText = ManeuverParser.formatDistanceMeters(ndi2.distanceTo),
                instruction = ManeuverParser.instructionForOsmandTurnType(t2.value),
                road = ndi2.directionInfo.streetName.orEmpty(),
            )
        }

        val leftTime = rh.leftTime
        val etaMin = if (leftTime > 0) (leftTime + 59) / 60 else 0
        val remainingM = rh.leftDistance.coerceAtLeast(0)
        val instruction = di.getDescriptionRoutePart(osmand)
            ?.takeIf { it.isNotBlank() }
            ?: ManeuverParser.instructionForOsmandTurnType(turn.value)

        val junction = try {
            OsmandJunctionExtractor.extract(osmand, rh, di, maneuver)
        } catch (e: Exception) {
            Log.w(TAG, "junction extract failed — emitting minimal IR", e)
            JunctionBuilder.fromManeuver(maneuver)
        }

        HudBus.publishNav(
            NavState(
                active = true,
                instruction = instruction,
                distanceM = ndi.distanceTo,
                distanceText = ManeuverParser.formatDistanceMeters(ndi.distanceTo),
                road = road,
                etaMin = etaMin,
                remainingM = remainingM,
                maneuver = maneuver,
                lanes = lanes,
                thenNext = thenNext,
                junction = junction,
            ),
            NavSource.OSMAND,
        )
    }

    companion object {
        private const val TAG = "OsmandEmbedded"
        private const val POLL_MS = 1000L
    }
}
