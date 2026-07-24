package com.motohud.companion

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test

class OsrmRibbonTest {

    @Test
    fun corridorFromRouteJson_projectsRightTurn() {
        // Origin at 0,0-ish; geometry goes north then east (right turn).
        val json = """
            {
              "code": "Ok",
              "routes": [{
                "geometry": {
                  "type": "LineString",
                  "coordinates": [
                    [0.0, 0.0],
                    [0.0, 0.001],
                    [0.0, 0.002],
                    [0.001, 0.002]
                  ]
                },
                "legs": [{
                  "steps": [
                    {
                      "distance": 220,
                      "maneuver": {"type": "depart", "modifier": "north", "location": [0.0, 0.0]}
                    },
                    {
                      "distance": 80,
                      "maneuver": {"type": "turn", "modifier": "right", "location": [0.0, 0.002]}
                    },
                    {
                      "distance": 0,
                      "maneuver": {"type": "arrive", "location": [0.001, 0.002]}
                    }
                  ]
                }]
              }]
            }
        """.trimIndent()

        val result = OsrmRibbon.corridorFromRouteJson(
            json,
            originLat = 0.0,
            originLon = 0.0,
            bearingDeg = 0f,
            maneuver = "right",
            distanceM = 220,
        )
        assertNotNull(result)
        assertTrue(result!!.points.size >= 2)
        assertTrue(result.turnIndex in result.points.indices)
        // Ahead (Y) should grow; final point should have positive X (right).
        assertTrue(result.points.last().y > result.points.first().y)
        assertTrue(result.points.last().x > 0)
    }

    @Test
    fun destinationAhead_north() {
        val (lat, lon) = OsrmRibbon.destinationAhead(51.0, -0.1, 0f, 111.32)
        assertTrue(lat > 51.0)
        assertEquals(-0.1, lon, 0.0001)
    }

    @Test
    fun navState_toJson_includesRibbonWhenPresent() {
        val json = String(
            NavState(
                active = true,
                maneuver = "right",
                ribbonPoints = listOf(
                    RibbonPoint(0.0, 0.0),
                    RibbonPoint(0.0, 20.0),
                    RibbonPoint(15.0, 30.0),
                ),
                ribbonTurn = 1,
            ).toJson(),
            Charsets.UTF_8,
        )
        val obj = org.json.JSONObject(json)
        assertEquals(3, obj.getJSONArray("ribbon_points").length())
        assertEquals(1, obj.getInt("ribbon_turn"))
        assertEquals(15, obj.getJSONArray("ribbon_points").getJSONObject(2).getInt("x"))
    }

    @Test
    fun navState_toJson_omitsRibbonWhenEmpty() {
        val json = String(NavState(active = true, maneuver = "left").toJson(), Charsets.UTF_8)
        val obj = org.json.JSONObject(json)
        assertTrue(!obj.has("ribbon_points"))
    }
}
