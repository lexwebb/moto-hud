package com.motohud.companion

import org.junit.Assert.assertEquals
import org.junit.Test

class ManeuverParserTest {

    @Test
    fun fromText_roundaboutAndUTurn() {
        assertEquals("roundabout", ManeuverParser.fromText("Enter the roundabout"))
        assertEquals("roundabout", ManeuverParser.fromText("Take the rotary"))
        assertEquals("u_turn", ManeuverParser.fromText("Make a U-turn"))
        assertEquals("u_turn", ManeuverParser.fromText("U turn ahead"))
    }

    @Test
    fun fromText_slightAndCardinal() {
        assertEquals("slight_left", ManeuverParser.fromText("Keep left onto A1"))
        assertEquals("slight_right", ManeuverParser.fromText("Slight right toward exit"))
        assertEquals("left", ManeuverParser.fromText("Turn left onto High St"))
        assertEquals("right", ManeuverParser.fromText("Turn right onto Bridge Rd"))
        assertEquals("left", ManeuverParser.fromText("Left onto Harbor"))
        assertEquals("right", ManeuverParser.fromText("destination on the right"))
    }

    @Test
    fun fromText_straightArriveDepartUnknown() {
        assertEquals("straight", ManeuverParser.fromText("Continue straight"))
        assertEquals("straight", ManeuverParser.fromText("Head north on Main"))
        assertEquals("arrive", ManeuverParser.fromText("You have arrived"))
        assertEquals("arrive", ManeuverParser.fromText("Destination ahead"))
        assertEquals("depart", ManeuverParser.fromText("Depart from home"))
        assertEquals("unknown", ManeuverParser.fromText("Something odd"))
    }

    @Test
    fun fromOsmandTurnType_mapsKnownCodes() {
        assertEquals("straight", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_C))
        assertEquals("left", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TL))
        assertEquals("slight_left", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TSLL))
        assertEquals("left", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TSHL))
        assertEquals("right", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TR))
        assertEquals("slight_right", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TSLR))
        assertEquals("slight_left", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_KL))
        assertEquals("slight_right", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_KR))
        assertEquals("u_turn", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TU))
        assertEquals("u_turn", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_TRU))
        assertEquals("roundabout", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_RNDB))
        assertEquals("unknown", ManeuverParser.fromOsmandTurnType(ManeuverParser.OSMAND_OFFR))
        assertEquals("unknown", ManeuverParser.fromOsmandTurnType(99))
    }

    @Test
    fun instructionForOsmandTurnType_readable() {
        assertEquals("Turn left", ManeuverParser.instructionForOsmandTurnType(ManeuverParser.OSMAND_TL))
        assertEquals("Keep right", ManeuverParser.instructionForOsmandTurnType(ManeuverParser.OSMAND_KR))
        assertEquals("Roundabout", ManeuverParser.instructionForOsmandTurnType(ManeuverParser.OSMAND_RNDB))
    }

    @Test
    fun formatDistanceMeters() {
        assertEquals("200 m", ManeuverParser.formatDistanceMeters(200))
        assertEquals("1.5 km", ManeuverParser.formatDistanceMeters(1500))
        assertEquals("12 km", ManeuverParser.formatDistanceMeters(12000))
    }

    @Test
    fun parseDistanceMeters_kmAndM() {
        assertEquals(1500, ManeuverParser.parseDistanceMeters("1.5 km"))
        assertEquals(1500, ManeuverParser.parseDistanceMeters("1,5 km"))
        assertEquals(250, ManeuverParser.parseDistanceMeters("250 m"))
        assertEquals(250, ManeuverParser.parseDistanceMeters("250 meter"))
        assertEquals(120, ManeuverParser.parseDistanceMeters("120"))
        assertEquals(0, ManeuverParser.parseDistanceMeters("no distance here"))
    }

    @Test
    fun navState_toJson_shape() {
        val json = String(
            NavState(
                active = true,
                instruction = "Turn left",
                distanceM = 200,
                distanceText = "200 m",
                road = "High St",
                etaMin = 12,
                maneuver = "left",
            ).toJson(),
            Charsets.UTF_8,
        )
        val obj = org.json.JSONObject(json)
        assertEquals("nav", obj.getString("type"))
        assertEquals(true, obj.getBoolean("active"))
        assertEquals(200, obj.getInt("distance_m"))
        assertEquals("200 m", obj.getString("distance_text"))
        assertEquals(12, obj.getInt("eta_min"))
        assertEquals("left", obj.getString("maneuver"))
    }
}
