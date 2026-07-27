package com.motohud.companion

import android.app.Application
import android.util.Log
import androidx.core.app.AppComponentFactory

/**
 * Picks OsmAnd's Application when the `:osmand` split is on the classpath,
 * otherwise [MotoHudApp]. Extends AndroidX factory so Startup still works.
 */
class MotoHudAppComponentFactory : AppComponentFactory() {
    override fun instantiateApplicationCompat(cl: ClassLoader, className: String): Application {
        val resolved = resolveApplicationClassName(cl)
        Log.i(TAG, "Application → $resolved")
        return super.instantiateApplicationCompat(cl, resolved)
    }

    companion object {
        private const val TAG = "AppComponentFactory"
        const val BASE_APP = "com.motohud.companion.MotoHudApp"
        const val OSMAND_APP = "com.motohud.companion.osmand.MotoHudOsmandApp"

        fun resolveApplicationClassName(cl: ClassLoader): String {
            return try {
                cl.loadClass(OSMAND_APP)
                OSMAND_APP
            } catch (_: ClassNotFoundException) {
                BASE_APP
            }
        }
    }
}
