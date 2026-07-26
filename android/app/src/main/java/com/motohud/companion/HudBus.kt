package com.motohud.companion

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow

enum class NavSource {
    /** Typed OsmAnd AIDL (preferred when actively navigating). */
    OSMAND,
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

    @Volatile
    private var osmandBound = false

    /** When true, Maps notification updates are ignored. */
    @Volatile
    private var osmandOwnsNav = false

    fun setOsmandBound(bound: Boolean) {
        osmandBound = bound
        if (!bound) osmandOwnsNav = false
    }

    fun isOsmandBound(): Boolean = osmandBound

    fun publishNav(n: NavState, source: NavSource = NavSource.MAPS) {
        when (source) {
            NavSource.OSMAND -> {
                osmandOwnsNav = n.active
                _nav.value = n
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
}
