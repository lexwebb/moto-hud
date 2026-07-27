package com.motohud.companion

/**
 * Platform nav source that publishes [NavState] into [HudBus].
 * Implementations: OsmAnd AIDL (default), OsmAnd Full Library (embedded flavor), Maps scrape.
 */
interface NavEngine {
    /** Human-readable id for UI/status (e.g. "osmand-aidl", "osmand-embedded"). */
    val id: String

    fun start()
    fun stop()
}
