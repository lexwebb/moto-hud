package com.motohud.companion

import org.json.JSONArray
import org.json.JSONObject

data class RibbonPoint(
    val x: Double,
    val y: Double,
)

data class NavState(
    val active: Boolean = false,
    val instruction: String = "",
    val distanceM: Int = 0,
    val distanceText: String = "",
    val road: String = "",
    val etaMin: Int = 0,
    val maneuver: String = "unknown",
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
        put("maneuver", maneuver)
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
