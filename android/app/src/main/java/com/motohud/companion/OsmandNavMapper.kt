package com.motohud.companion

import net.osmand.aidlapi.navigation.ADirectionInfo

/**
 * Maps OsmAnd [ADirectionInfo] / TurnType ints onto our nav protocol.
 *
 * TurnType values from OsmAnd (net.osmand.router.TurnType):
 * C=1 TL=2 TSLL=3 TSHL=4 TR=5 TSLR=6 TSHR=7 KL=8 KR=9 TU=10 TRU=11
 * OFFR=12 RNDB=13 RNLB=14
 */
object OsmandNavMapper {

    fun toNavState(info: ADirectionInfo): NavState {
        val dist = info.distanceTo
        val turn = info.turnType
        // OsmAnd seeds the callback with (-1, -1) when idle / no next turn.
        if (dist < 0 || turn < 0) {
            return NavState(active = false, instruction = "Navigation ended")
        }
        val maneuver = ManeuverParser.fromOsmandTurnType(turn)
        val instruction = ManeuverParser.instructionForOsmandTurnType(turn)
        return NavState(
            active = true,
            instruction = instruction,
            distanceM = dist,
            distanceText = ManeuverParser.formatDistanceMeters(dist),
            road = "",
            etaMin = 0,
            maneuver = maneuver,
        )
    }
}
