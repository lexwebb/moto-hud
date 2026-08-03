package com.motohud.companion

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow

enum class NavSource {
    /** Typed OsmAnd AIDL / Full Library (preferred when actively navigating). */
    OSMAND,
    /** Soft fields (road, ETA) scraped from OsmAnd notifications while AIDL owns turns. */
    OSMAND_ENRICH,
    /** Google Maps / Maps Go notification scrape (fallback). */
    MAPS,
}

object HudBus {
    private val _nav = MutableStateFlow(NavState())
    val nav: StateFlow<NavState> = _nav.asStateFlow()

    private val _media = MutableStateFlow(MediaState())
    val media: StateFlow<MediaState> = _media.asStateFlow()

    private val _cmds = MutableSharedFlow<String>(extraBufferCapacity = 8)
    val cmds: SharedFlow<String> = _cmds.asSharedFlow()

    private val _status = MutableStateFlow("Idle")
    val status: StateFlow<String> = _status.asStateFlow()

    private val _osmandBound = MutableStateFlow(false)
    val osmandBound: StateFlow<Boolean> = _osmandBound.asStateFlow()

    /** When true, Maps notification updates are ignored. */
    @Volatile
    private var osmandOwnsNav = false

    fun setOsmandBound(bound: Boolean) {
        _osmandBound.value = bound
        if (!bound) osmandOwnsNav = false
    }

    fun isOsmandBound(): Boolean = _osmandBound.value

    fun publishNav(n: NavState, source: NavSource = NavSource.MAPS) {
        when (source) {
            NavSource.OSMAND -> {
                osmandOwnsNav = n.active
                if (!n.active) {
                    _nav.value = n
                    return
                }
                val cur = _nav.value
                // Don't let AIDL's farther "speakable" turn clobber a closer
                // imminent turn we already took from the OsmAnd notification.
                if (cur.active && cur.distanceM > 0 && n.distanceM > cur.distanceM + 25) {
                    _nav.value = n.copy(
                        distanceM = cur.distanceM,
                        distanceText = cur.distanceText,
                        maneuver = cur.maneuver,
                        junction = cur.junction ?: n.junction,
                        instruction = cur.instruction.ifBlank { n.instruction },
                        road = n.road.ifBlank { cur.road },
                        etaMin = if (n.etaMin > 0) n.etaMin else cur.etaMin,
                        remainingM = if (n.remainingM > 0) n.remainingM else cur.remainingM,
                        // Promote AIDL's farther turn to then-next when missing.
                        thenNext = n.thenNext ?: ThenNext(
                            maneuver = n.maneuver,
                            distanceM = n.distanceM,
                            distanceText = n.distanceText.ifBlank {
                                ManeuverParser.formatDistanceMeters(n.distanceM)
                            },
                            instruction = n.instruction,
                            road = n.road,
                        ),
                    )
                } else {
                    _nav.value = n
                }
            }
            NavSource.OSMAND_ENRICH -> {
                if (!_osmandBound.value) return
                // If AIDL is bound but hasn't published an active turn yet (or
                // Connected Apps blocked registerForNavigationUpdates), let the
                // OsmAnd notification drive the HUD.
                if (!osmandOwnsNav) {
                    osmandOwnsNav = n.active
                    _nav.value = n.copy(
                        distanceText = if (n.distanceM > 0) {
                            ManeuverParser.formatDistanceMeters(n.distanceM)
                        } else {
                            n.distanceText
                        },
                        road = n.road.takeIf { it.isNotBlank() && !looksLikeTurnBanner(it) }.orEmpty(),
                    )
                    return
                }
                val cur = _nav.value
                if (!cur.active) return
                // OsmAnd AIDL uses getNextRouteDirectionInfo(..., toSpeak=true), which
                // can skip the imminent turn and jump to a farther speakable one
                // (e.g. roundabout at 733 m while the banner says right in 40 m).
                // When the notification reports a closer turn, prefer that for the HUD.
                val notifCloser =
                    n.distanceM > 0 && (cur.distanceM <= 0 || n.distanceM + 25 < cur.distanceM)
                val road = n.road.takeIf { it.isNotBlank() && !looksLikeTurnBanner(it) } ?: cur.road
                _nav.value = if (notifCloser) {
                    cur.copy(
                        distanceM = n.distanceM,
                        distanceText = ManeuverParser.formatDistanceMeters(n.distanceM),
                        maneuver = n.maneuver.takeIf { it != "unknown" } ?: cur.maneuver,
                        junction = n.junction ?: cur.junction,
                        road = road,
                        etaMin = if (n.etaMin > 0) n.etaMin else cur.etaMin,
                        remainingM = if (n.remainingM > 0) n.remainingM else cur.remainingM,
                        instruction = n.instruction.ifBlank { cur.instruction },
                    )
                } else {
                    cur.copy(
                        road = road,
                        etaMin = if (n.etaMin > 0) n.etaMin else cur.etaMin,
                        remainingM = if (n.remainingM > 0) n.remainingM else cur.remainingM,
                        instruction = n.instruction
                            .takeIf { it.isNotBlank() && !looksLikeTurnBanner(it) }
                            ?: cur.instruction,
                    )
                }
            }
            NavSource.MAPS -> {
                if (osmandOwnsNav) return
                _nav.value = n
            }
        }
    }

    fun publishMedia(m: MediaState) {
        _media.value = m
    }

    fun publishCmd(action: String) {
        _cmds.tryEmit(action)
    }

    fun setStatus(s: String) {
        _status.value = s
    }

    /** OsmAnd turn banners often look like "40 m • Turn right and go". */
    private fun looksLikeTurnBanner(s: String): Boolean =
        TURN_BANNER_PREFIX.containsMatchIn(s.trim())

    private val TURN_BANNER_PREFIX = Regex("""^\d+([.,]\d+)?\s*(m|km)\b""", RegexOption.IGNORE_CASE)
}
