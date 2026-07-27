package com.motohud.companion

import android.content.Context

/**
 * Prefers in-process OsmAnd Full Library when the on-demand module is installed
 * and the process was restarted onto [MotoHudOsmandApp]; otherwise AIDL.
 */
object NavEngines {
    fun create(context: Context): NavEngine {
        if (OsmandModule.isRichNavReady(context)) {
            OsmandModule.createEmbeddedEngine(context)?.let { return it }
        }
        return OsmandNavClient(context)
    }
}
