package com.motohud.companion

import android.app.Notification
import android.service.notification.NotificationListenerService
import android.service.notification.StatusBarNotification
import android.util.Log

class NavNotificationListener : NotificationListenerService() {

    override fun onNotificationPosted(sbn: StatusBarNotification?) {
        sbn ?: return
        when (sbn.packageName) {
            in OSMAND_PACKAGES -> handleOsmand(sbn.notification)
            MAPS_PACKAGE, MAPS_PACKAGE_GO -> handleMaps(sbn.notification)
        }
    }

    override fun onNotificationRemoved(sbn: StatusBarNotification?) {
        sbn ?: return
        when (sbn.packageName) {
            MAPS_PACKAGE, MAPS_PACKAGE_GO -> {
                HudBus.publishNav(
                    NavState(active = false, instruction = "Navigation ended"),
                    NavSource.MAPS,
                )
            }
            in OSMAND_PACKAGES -> {
                // OsmAnd AIDL owns active/inactive; notification remove is not authoritative.
            }
        }
    }

    private fun handleMaps(n: Notification) {
        val nav = parseCommon(n) ?: return
        Log.d(TAG, "maps nav update: $nav")
        HudBus.publishNav(nav, NavSource.MAPS)
    }

    private fun handleOsmand(n: Notification) {
        val parsed = parseCommon(n) ?: return
        // Soft enrichment while AIDL supplies typed turns.
        Log.d(TAG, "osmand notif enrich: $parsed")
        HudBus.publishNav(parsed, NavSource.OSMAND_ENRICH)
    }

    private fun parseCommon(n: Notification): NavState? {
        val extras = n.extras ?: return null
        val title = extras.getCharSequence(Notification.EXTRA_TITLE)?.toString().orEmpty()
        val text = extras.getCharSequence(Notification.EXTRA_TEXT)?.toString().orEmpty()
        val big = extras.getCharSequence(Notification.EXTRA_BIG_TEXT)?.toString().orEmpty()
        val sub = extras.getCharSequence(Notification.EXTRA_SUB_TEXT)?.toString().orEmpty()
        val lines = extras.getCharSequenceArray(Notification.EXTRA_TEXT_LINES)
            ?.joinToString(" ") { it?.toString().orEmpty() }
            .orEmpty()

        val blob = listOf(title, text, big, sub, lines).filter { it.isNotBlank() }.joinToString(" | ")
        if (blob.isBlank()) return null

        val distanceCandidate = listOf(title, text, sub, big).firstOrNull {
            it.contains(Regex("""\d+\s*(m|km)""", RegexOption.IGNORE_CASE))
        }.orEmpty()

        val instruction = when {
            text.isNotBlank() -> text
            big.isNotBlank() -> big
            else -> title
        }

        val road = when {
            title.isNotBlank() && title != instruction -> title
            else -> ""
        }

        return NavState(
            active = true,
            instruction = instruction,
            distanceM = ManeuverParser.parseDistanceMeters(distanceCandidate.ifBlank { blob }),
            distanceText = distanceCandidate.ifBlank { "" },
            road = road,
            etaMin = ManeuverParser.parseEtaMinutes(blob),
            maneuver = ManeuverParser.fromText(instruction),
        )
    }

    companion object {
        private const val TAG = "NavListener"
        const val MAPS_PACKAGE = "com.google.android.apps.maps"
        const val MAPS_PACKAGE_GO = "com.google.android.apps.mapslite"
        val OSMAND_PACKAGES = setOf("net.osmand", "net.osmand.plus")
    }
}
