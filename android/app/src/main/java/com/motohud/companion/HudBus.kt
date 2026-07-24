package com.motohud.companion

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow

object HudBus {
    private val _nav = MutableStateFlow(NavState())
    val nav: StateFlow<NavState> = _nav.asStateFlow()

    private val _media = MutableStateFlow(MediaState())
    val media: StateFlow<MediaState> = _media.asStateFlow()

    private val _cmds = MutableSharedFlow<String>(extraBufferCapacity = 8)
    val cmds: SharedFlow<String> = _cmds.asSharedFlow()

    private val _status = MutableStateFlow("Idle")
    val status: StateFlow<String> = _status.asStateFlow()

    fun publishNav(n: NavState) {
        _nav.value = n
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
