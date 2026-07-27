package com.motohud.companion

import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class JunctionBuilderTest {

    @Test
    fun fromManeuver_matchesProtocolDefaults() {
        assertEquals("arrive", JunctionBuilder.fromManeuver("arrive").kind)
        assertEquals("depart", JunctionBuilder.fromManeuver("depart").kind)
        assertEquals("u_turn", JunctionBuilder.fromManeuver("u_turn").kind)
        val rb = JunctionBuilder.fromManeuver("roundabout")
        assertEquals("roundabout", rb.kind)
        assertEquals(4, rb.exits)
        assertEquals(2, rb.exit)
        assertEquals("fork", JunctionBuilder.fromManeuver("slight_left").kind)
        assertEquals("slight_left", JunctionBuilder.fromManeuver("slight_left").outbound)
        assertEquals("simple", JunctionBuilder.fromManeuver("left").kind)
        assertEquals("simple", JunctionBuilder.fromManeuver("straight").kind)
        assertTrue(JunctionBuilder.fromManeuver("straight").through)
    }

    @Test
    fun build_dualCarriageway_crossMedian() {
        val msg = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = "left",
                dual = true,
                drive = "left",
            ),
        )
        assertEquals("dual_carriageway", msg.kind)
        assertTrue(msg.crossMedian)
        assertEquals("left", msg.outbound)
        assertEquals("left", msg.drive)
    }

    @Test
    fun build_dualUpgradesSlight() {
        val msg = JunctionBuilder.build(
            JunctionBuildInput(maneuver = "slight_right", dual = true),
        )
        assertEquals("dual_carriageway", msg.kind)
        assertFalse(msg.crossMedian)
        assertEquals("slight_right", msg.outbound)
    }

    @Test
    fun build_crossroadsAndTJunction() {
        val cross = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = "left",
                arms = listOf(
                    JunctionArmHint("left", isOutbound = true),
                    JunctionArmHint("right"),
                    JunctionArmHint("through"),
                ),
            ),
        )
        assertEquals("crossroads", cross.kind)
        assertTrue(cross.through)

        val tee = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = "right",
                arms = listOf(
                    JunctionArmHint("right", isOutbound = true),
                    JunctionArmHint("left"),
                ),
            ),
        )
        assertEquals("t_junction", tee.kind)
        assertFalse(tee.through)
    }

    @Test
    fun build_rampExitAndEnter() {
        val exit = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = "slight_right",
                approachHighway = "motorway",
                turnHighway = "motorway_link",
            ),
        )
        // dual false; ramp checked before slight→fork
        assertEquals("ramp_exit", exit.kind)
        assertTrue(exit.through)

        val enter = JunctionBuilder.build(
            JunctionBuildInput(
                maneuver = "slight_left",
                approachHighway = "trunk_link",
                turnHighway = "trunk",
            ),
        )
        assertEquals("ramp_enter", enter.kind)
        assertEquals("left", enter.side)
    }

    @Test
    fun build_forkWithoutDual() {
        val msg = JunctionBuilder.build(JunctionBuildInput(maneuver = "slight_left"))
        assertEquals("fork", msg.kind)
        assertEquals("slight_left", msg.outbound)
    }

    @Test
    fun navState_toJson_includesJunction() {
        val json = String(
            NavState(
                active = true,
                maneuver = "left",
                junction = JunctionBuilder.build(
                    JunctionBuildInput(
                        maneuver = "left",
                        dual = true,
                        arms = listOf(JunctionArmHint("left", isOutbound = true), JunctionArmHint("right")),
                    ),
                ),
            ).toJson(),
            Charsets.UTF_8,
        )
        val j = JSONObject(json).getJSONObject("junction")
        assertEquals("dual_carriageway", j.getString("kind"))
        assertEquals("left", j.getString("outbound"))
        assertTrue(j.getBoolean("cross_median"))
        assertTrue(j.has("sides"))
    }

    @Test
    fun isMajorAndLinkHelpers() {
        assertTrue(JunctionBuilder.isMajorHighway("primary"))
        assertTrue(JunctionBuilder.isMajorHighway("primary_link"))
        assertTrue(JunctionBuilder.isLink("primary_link"))
        assertFalse(JunctionBuilder.isLink("primary"))
        assertFalse(JunctionBuilder.isMajorHighway("residential"))
    }
}
