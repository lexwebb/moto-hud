package com.motohud.companion

import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test

class HudBusNavMergeTest {

    @Before
    fun reset() {
        HudBus.setOsmandBound(false)
        HudBus.publishNav(NavState(active = false), NavSource.OSMAND)
    }

    @Test
    fun enrichDrivesHudWhenAidlBoundButNotOwningYet() {
        HudBus.setOsmandBound(true)
        HudBus.publishNav(
            NavState(
                active = true,
                maneuver = "right",
                distanceM = 40,
                distanceText = "40 m • Turn right",
                road = "40 m • Turn right",
                instruction = "Turn right and go 80 m",
            ),
            NavSource.OSMAND_ENRICH,
        )
        val nav = HudBus.nav.value
        assertEquals(true, nav.active)
        assertEquals(40, nav.distanceM)
        assertEquals("40 m", nav.distanceText)
        assertEquals("right", nav.maneuver)
        assertEquals("", nav.road)
    }

    @Test
    fun enrichPrefersCloserNotificationDistanceOverAidl() {
        HudBus.setOsmandBound(true)
        HudBus.publishNav(
            NavState(
                active = true,
                maneuver = "roundabout",
                distanceM = 733,
                distanceText = "733 m",
                instruction = "Roundabout",
            ),
            NavSource.OSMAND,
        )
        HudBus.publishNav(
            NavState(
                active = true,
                maneuver = "right",
                distanceM = 40,
                distanceText = "40 m",
                instruction = "40 m • Turn right and go",
                road = "40 m • Turn right and go",
                etaMin = 21,
            ),
            NavSource.OSMAND_ENRICH,
        )
        val nav = HudBus.nav.value
        assertEquals(40, nav.distanceM)
        assertEquals("40 m", nav.distanceText)
        assertEquals("right", nav.maneuver)
        assertEquals(21, nav.etaMin)
        // Distance-banner scrapes must not become the road name.
        assertEquals("", nav.road)
    }

    @Test
    fun aidlDoesNotClobberCloserImminentTurn() {
        HudBus.setOsmandBound(true)
        HudBus.publishNav(
            NavState(active = true, maneuver = "roundabout", distanceM = 733, distanceText = "733 m"),
            NavSource.OSMAND,
        )
        HudBus.publishNav(
            NavState(active = true, maneuver = "right", distanceM = 40, distanceText = "40 m"),
            NavSource.OSMAND_ENRICH,
        )
        HudBus.publishNav(
            NavState(
                active = true,
                maneuver = "roundabout",
                distanceM = 720,
                distanceText = "720 m",
                road = "High St",
                etaMin = 20,
            ),
            NavSource.OSMAND,
        )
        val nav = HudBus.nav.value
        assertEquals(40, nav.distanceM)
        assertEquals("right", nav.maneuver)
        assertEquals("High St", nav.road)
        assertEquals(20, nav.etaMin)
        assertEquals("roundabout", nav.thenNext?.maneuver)
        assertEquals(720, nav.thenNext?.distanceM)
    }
}
