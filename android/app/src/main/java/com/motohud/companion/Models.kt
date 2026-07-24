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
}
