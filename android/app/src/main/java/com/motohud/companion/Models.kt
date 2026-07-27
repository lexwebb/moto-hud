package com.motohud.companion

import org.json.JSONArray
import org.json.JSONObject

data class RibbonPoint(
    val x: Double,
    val y: Double,
)

/** One lane, left-to-right. Directions use protocol maneuver strings. */
data class LaneInfo(
    val directions: List<String>,
    val active: Boolean,
)

data class ThenNext(
    val maneuver: String = "unknown",
    val distanceM: Int = 0,
    val distanceText: String = "",
    val instruction: String = "",
    val road: String = "",
)

data class NavState(
    val active: Boolean = false,
    val instruction: String = "",
    val distanceM: Int = 0,
    val distanceText: String = "",
    val road: String = "",
    val etaMin: Int = 0,
    val remainingM: Int = 0,
    val maneuver: String = "unknown",
    val lanes: List<LaneInfo> = emptyList(),
    val thenNext: ThenNext? = null,
    val ribbonPoints: List<RibbonPoint> = emptyList(),
    val ribbonTurn: Int = -1,
) {
    fun toJson(): ByteArray = JSONObject().apply {
        put("type", "nav")
        put("active", active)
        put("instruction", instruction)
        put("distance_m", distanceM)
        put("distance_text", distanceText)
        put("road", road)
        put("eta_min", etaMin)
        if (remainingM > 0) put("remaining_m", remainingM)
        put("maneuver", maneuver)
        if (lanes.isNotEmpty()) {
            put("lanes", JSONArray().also { arr ->
                lanes.forEach { lane ->
                    arr.put(JSONObject().apply {
                        put("directions", JSONArray(lane.directions))
                        put("active", lane.active)
                    })
                }
            })
        }
        thenNext?.let { tn ->
            put("then_next", JSONObject().apply {
                put("maneuver", tn.maneuver)
                put("distance_m", tn.distanceM)
                put("distance_text", tn.distanceText)
                put("instruction", tn.instruction)
                put("road", tn.road)
            })
        }
        if (ribbonPoints.size >= 2) {
            put("ribbon_points", JSONArray().also { arr ->
                ribbonPoints.forEach { p ->
                    arr.put(JSONObject().put("x", p.x.toInt()).put("y", p.y.toInt()))
                }
            })
            put("ribbon_turn", ribbonTurn)
        }
    }.toString().toByteArray(Charsets.UTF_8)
}

data class MediaState(
    val playing: Boolean = false,
    val title: String = "",
    val artist: String = "",
) {
    fun toJson(): ByteArray = JSONObject().apply {
        put("type", "media")
        put("playing", playing)
        put("title", title)
        put("artist", artist)
    }.toString().toByteArray(Charsets.UTF_8)
}

object ManeuverParser {
    // OsmAnd TurnType int constants (stable across free/plus).
    const val OSMAND_C = 1
    const val OSMAND_TL = 2
    const val OSMAND_TSLL = 3
    const val OSMAND_TSHL = 4
    const val OSMAND_TR = 5
    const val OSMAND_TSLR = 6
    const val OSMAND_TSHR = 7
    const val OSMAND_KL = 8
    const val OSMAND_KR = 9
    const val OSMAND_TU = 10
    const val OSMAND_TRU = 11
    const val OSMAND_OFFR = 12
    const val OSMAND_RNDB = 13
    const val OSMAND_RNLB = 14

    fun fromText(text: String): String {
        val t = text.lowercase()
        return when {
            "roundabout" in t || "rotary" in t -> "roundabout"
            "u-turn" in t || "u turn" in t -> "u_turn"
            "slight left" in t || "keep left" in t -> "slight_left"
            "slight right" in t || "keep right" in t -> "slight_right"
            "turn left" in t || t.startsWith("left") || " on the left" in t -> "left"
            "turn right" in t || t.startsWith("right") || " on the right" in t -> "right"
            "straight" in t || "continue" in t || "head " in t -> "straight"
            "arrive" in t || "destination" in t -> "arrive"
            "depart" in t || "start" in t -> "depart"
            else -> "unknown"
        }
    }

    /** Map OsmAnd TurnType.getValue() → protocol maneuver string. */
    fun fromOsmandTurnType(turnType: Int): String = when (turnType) {
        OSMAND_C -> "straight"
        OSMAND_TL, OSMAND_TSHL -> "left"
        OSMAND_TSLL, OSMAND_KL -> "slight_left"
        OSMAND_TR, OSMAND_TSHR -> "right"
        OSMAND_TSLR, OSMAND_KR -> "slight_right"
        OSMAND_TU, OSMAND_TRU -> "u_turn"
        OSMAND_RNDB, OSMAND_RNLB -> "roundabout"
        OSMAND_OFFR -> "unknown"
        else -> "unknown"
    }

    fun instructionForOsmandTurnType(turnType: Int): String = when (turnType) {
        OSMAND_C -> "Continue straight"
        OSMAND_TL -> "Turn left"
        OSMAND_TSLL -> "Turn slightly left"
        OSMAND_TSHL -> "Turn sharply left"
        OSMAND_TR -> "Turn right"
        OSMAND_TSLR -> "Turn slightly right"
        OSMAND_TSHR -> "Turn sharply right"
        OSMAND_KL -> "Keep left"
        OSMAND_KR -> "Keep right"
        OSMAND_TU, OSMAND_TRU -> "Make a U-turn"
        OSMAND_RNDB, OSMAND_RNLB -> "Roundabout"
        OSMAND_OFFR -> "Off route"
        else -> "Navigate"
    }

    fun parseDistanceMeters(text: String): Int {
        val km = Regex("""(\d+(?:[.,]\d+)?)\s*km""", RegexOption.IGNORE_CASE).find(text)
        if (km != null) {
            val v = km.groupValues[1].replace(',', '.').toDoubleOrNull() ?: return 0
            return (v * 1000).toInt()
        }
        val m = Regex("""(\d+)\s*m(?:\b|eter)""", RegexOption.IGNORE_CASE).find(text)
        if (m != null) return m.groupValues[1].toIntOrNull() ?: 0
        val bare = Regex("""^\s*(\d+)\s*$""").find(text.trim())
        return bare?.groupValues?.get(1)?.toIntOrNull() ?: 0
    }

    fun parseEtaMinutes(text: String): Int {
        val hms = Regex(
            """(\d+)\s*h(?:ours?)?\s*(\d+)\s*min""",
            RegexOption.IGNORE_CASE,
        ).find(text)
        if (hms != null) {
            val h = hms.groupValues[1].toIntOrNull() ?: 0
            val m = hms.groupValues[2].toIntOrNull() ?: 0
            return h * 60 + m
        }
        val min = Regex("""(\d+)\s*min""", RegexOption.IGNORE_CASE).find(text)
        if (min != null) return min.groupValues[1].toIntOrNull() ?: 0
        return 0
    }

    /** Format meters for HUD / distance_text (locale-stable). */
    fun formatDistanceMeters(meters: Int): String {
        if (meters < 0) return ""
        if (meters >= 1000) {
            val km = meters / 1000.0
            val s = if (km >= 10) {
                String.format(java.util.Locale.US, "%.0f", km)
            } else {
                String.format(java.util.Locale.US, "%.1f", km)
            }
            return "$s km"
        }
        return "$meters m"
    }
}

/**
 * Decode OsmAnd TurnType lane bitfields into protocol [LaneInfo].
 * Encoding: bit0 = active; bits1-4 primary turn; bits5-9 secondary; bits10-14 tertiary.
 */
object OsmandLaneCodec {
    fun decode(laneValues: IntArray?): List<LaneInfo> {
        if (laneValues == null || laneValues.isEmpty()) return emptyList()
        return laneValues.map { v ->
            // OsmAnd: bit0 = active/recommended for this route.
            val active = (v and 1) == 1
            val dirs = buildList {
                val primary = primaryTurn(v)
                if (primary != 0) add(ManeuverParser.fromOsmandTurnType(primary))
                val secondary = secondaryTurn(v)
                if (secondary != 0) add(ManeuverParser.fromOsmandTurnType(secondary))
                val tertiary = tertiaryTurn(v)
                if (tertiary != 0) add(ManeuverParser.fromOsmandTurnType(tertiary))
            }.ifEmpty { listOf("straight") }
            LaneInfo(directions = dirs.distinct(), active = active)
        }
    }

    // Matches net.osmand.router.TurnType bit packing (getPrimaryTurn / getSecondaryTurn / getTertiaryTurn).
    fun primaryTurn(laneValue: Int): Int = (laneValue shr 1) and 0xF

    fun secondaryTurn(laneValue: Int): Int = (laneValue shr 5) and 0x1F

    fun tertiaryTurn(laneValue: Int): Int = (laneValue shr 10) and 0x1F

    /** Encode helper for tests: active + primary turn type. */
    fun encodeLane(active: Boolean, primary: Int, secondary: Int = 0, tertiary: Int = 0): Int {
        var v = (primary and 0xF) shl 1
        if (active) v = v or 1
        v = v or ((secondary and 0x1F) shl 5)
        v = v or ((tertiary and 0x1F) shl 10)
        return v
    }
}
