package com.motohud.companion

/**
 * Builds [JunctionMessage] for `nav.junction`.
 *
 * Thin paths (AIDL / notification scrape) call [fromManeuver] so the wire always
 * carries at least `{kind, outbound}` and the Pi need not guess from maneuver alone.
 * Full Library attaches topology via [build].
 *
 * Kind mapping mirrors protocol/junction.ts + protocol/README.md (ADR 0013).
 */
data class JunctionArmHint(
    /** left | right | through | back */
    val side: String,
    /** before | at | after */
    val at: String = "at",
    val highway: String? = null,
    /** True when this arm is the routed outbound. */
    val isOutbound: Boolean = false,
)

data class JunctionBuildInput(
    val maneuver: String,
    val drive: String? = null,
    val dual: Boolean = false,
    val approachHighway: String? = null,
    val turnHighway: String? = null,
    val arms: List<JunctionArmHint> = emptyList(),
    val roundaboutExits: Int = 0,
    val roundaboutExit: Int = 0,
)

object JunctionBuilder {

    /** Minimal IR — same defaults as protocol `synthesizeJunctionFromManeuver`. */
    fun fromManeuver(maneuver: String, drive: String? = null): JunctionMessage {
        val outbound = outboundFromManeuver(maneuver)
        return when (maneuver) {
            "arrive" -> JunctionMessage(kind = "arrive", outbound = "straight", through = false, drive = drive)
            "depart" -> JunctionMessage(kind = "depart", outbound = "straight", through = true, drive = drive)
            "u_turn" -> JunctionMessage(kind = "u_turn", outbound = "u_turn", through = false, drive = drive)
            "roundabout" -> JunctionMessage(
                kind = "roundabout",
                outbound = "right",
                through = false,
                drive = drive,
                exits = 4,
                exit = 2,
            )
            "slight_left" -> JunctionMessage(kind = "fork", outbound = "slight_left", through = false, drive = drive)
            "slight_right" -> JunctionMessage(kind = "fork", outbound = "slight_right", through = false, drive = drive)
            "straight" -> JunctionMessage(kind = "simple", outbound = "straight", through = true, drive = drive)
            "left" -> JunctionMessage(kind = "simple", outbound = "left", through = false, drive = drive)
            "right" -> JunctionMessage(kind = "simple", outbound = "right", through = false, drive = drive)
            else -> JunctionMessage(kind = "simple", outbound = outbound, through = outbound == "straight", drive = drive)
        }
    }

    /** Rich IR from Full Library topology (+ maneuver). */
    fun build(input: JunctionBuildInput): JunctionMessage {
        val maneuver = input.maneuver
        val outbound = outboundFromManeuver(maneuver)
        val drive = input.drive
        val sides = sidesFromArms(input.arms, outbound)

        val hasLeft = input.arms.any { it.side == "left" } ||
            outbound == "left" || outbound == "slight_left"
        val hasRight = input.arms.any { it.side == "right" } ||
            outbound == "right" || outbound == "slight_right"
        val hasThrough = input.arms.any { it.side == "through" } || outbound == "straight"

        when (maneuver) {
            "arrive" -> return JunctionMessage(
                kind = "arrive", outbound = "straight", through = false, drive = drive, sides = sides,
            )
            "depart" -> return JunctionMessage(
                kind = "depart", outbound = "straight", through = true, drive = drive, sides = sides,
            )
            "roundabout" -> {
                val exits = input.roundaboutExits.coerceIn(0, 6).let { if (it < 2) 4 else it }
                val exit = input.roundaboutExit.coerceIn(0, exits).let { if (it < 1) 2.coerceAtMost(exits) else it }
                return JunctionMessage(
                    kind = "roundabout",
                    outbound = outbound.takeIf { it != "straight" && it != "u_turn" } ?: "right",
                    through = false,
                    drive = drive,
                    sides = sides,
                    exits = exits,
                    exit = exit,
                )
            }
            "u_turn" -> return JunctionMessage(
                kind = "u_turn", outbound = "u_turn", through = false, drive = drive, sides = sides,
            )
        }

        // Dual upgrades even slight_keep when the corridor is divided (protocol rich path).
        if (input.dual) {
            val crossMedian = outbound == "left" || outbound == "right"
            return JunctionMessage(
                kind = "dual_carriageway",
                outbound = outbound,
                through = outbound == "straight" || hasThrough,
                drive = drive,
                sides = sides,
                crossMedian = crossMedian,
            )
        }

        rampKind(input, outbound)?.let { (kind, side, through) ->
            return JunctionMessage(
                kind = kind,
                outbound = outbound,
                through = through,
                drive = drive,
                sides = sides,
                side = side,
            )
        }

        if (maneuver == "slight_left" || maneuver == "slight_right" ||
            outbound == "slight_left" || outbound == "slight_right"
        ) {
            return JunctionMessage(
                kind = "fork",
                outbound = outbound,
                through = false,
                drive = drive,
                sides = sides,
            )
        }

        if (hasLeft && hasRight && hasThrough) {
            return JunctionMessage(
                kind = "crossroads",
                outbound = outbound,
                through = true,
                drive = drive,
                sides = sides,
            )
        }

        if ((hasLeft || hasRight) && !hasThrough && outbound != "straight") {
            return JunctionMessage(
                kind = "t_junction",
                outbound = outbound,
                through = false,
                drive = drive,
                sides = sides,
            )
        }

        return JunctionMessage(
            kind = "simple",
            outbound = outbound,
            through = hasThrough || outbound == "straight",
            drive = drive,
            sides = sides,
        )
    }

    fun outboundFromManeuver(maneuver: String): String = when (maneuver) {
        "left", "right", "slight_left", "slight_right", "straight", "u_turn" -> maneuver
        "arrive", "depart" -> "straight"
        "roundabout" -> "right"
        else -> "straight"
    }

    fun isLink(highway: String?): Boolean =
        highway != null && highway.endsWith("_link")

    fun isMajorHighway(highway: String?): Boolean {
        if (highway == null) return false
        val base = highway.removeSuffix("_link")
        return base == "motorway" || base == "trunk" || base == "primary" || base == "secondary"
    }

    private fun rampKind(
        input: JunctionBuildInput,
        outbound: String,
    ): Triple<String, String?, Boolean>? {
        val approach = input.approachHighway
        val turn = input.turnHighway
        if (isLink(turn) && isMajorHighway(approach) && !isLink(approach)) {
            return Triple("ramp_exit", outboundSide(outbound), true)
        }
        if (isLink(approach) && isMajorHighway(turn) && !isLink(turn)) {
            return Triple("ramp_enter", outboundSide(outbound), false)
        }
        return null
    }

    private fun outboundSide(outbound: String): String? = when (outbound) {
        "left", "slight_left" -> "left"
        "right", "slight_right" -> "right"
        else -> null
    }

    private fun sidesFromArms(arms: List<JunctionArmHint>, outbound: String): List<JunctionSideArm> {
        val outSide = when (outbound) {
            "left", "slight_left" -> "left"
            "right", "slight_right" -> "right"
            else -> null
        }
        val seen = mutableSetOf<String>()
        val out = mutableListOf<JunctionSideArm>()
        for (a in arms) {
            if (a.side != "left" && a.side != "right") continue
            if (a.isOutbound && outSide != null && a.side == outSide && a.at == "at") continue
            val key = "${a.side}:${a.at}"
            if (!seen.add(key)) continue
            out.add(JunctionSideArm(side = a.side, at = a.at, style = "dashed"))
        }
        return out
    }
}
