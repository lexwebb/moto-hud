package com.motohud.companion

import android.content.Context
import androidx.multidex.MultiDexApplication
import com.google.android.play.core.splitcompat.SplitCompat

/** Base application when the on-demand OsmAnd module is not installed. */
class MotoHudApp : MultiDexApplication() {
    override fun attachBaseContext(base: Context) {
        super.attachBaseContext(base)
        SplitCompat.install(this)
    }

    override fun onCreate() {
        super.onCreate()
        SplitCompat.install(this)
    }
}
