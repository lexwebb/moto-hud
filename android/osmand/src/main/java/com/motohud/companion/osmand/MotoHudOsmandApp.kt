package com.motohud.companion.osmand

import com.google.android.play.core.splitcompat.SplitCompat
import net.osmand.plus.OsmandApplication

/**
 * Application used when the on-demand `:osmand` feature module is installed.
 * Selected at process start by [com.motohud.companion.MotoHudAppComponentFactory].
 */
class MotoHudOsmandApp : OsmandApplication() {
    override fun onCreate() {
        SplitCompat.install(this)
        super.onCreate()
    }
}
