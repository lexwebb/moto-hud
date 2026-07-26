package com.motohud.companion

import android.app.Notification
import android.service.notification.NotificationListenerService
import android.service.notification.StatusBarNotification
import android.util.Log

class NavNotificationListener : NotificationListenerService() {

    override fun onNotificationPosted(sbn: StatusBarNotification?) {
        sbn ?: return
        if (sbn.packageName != MAPS_PACKAGE && sbn.packageName != MAPS_PACKAGE_GO) return
        val nav = parse(sbn.notification) ?: return
        Log.d(TAG, "maps nav update: $nav")
        HudBus.publishNav(nav, NavSource.MAPS)
    }

    override fun onNotificationRemoved(sbn: StatusBarNotification?) {
        sbn ?: return
        if (sbn.packageName != MAPS_PACKAGE && sbn.packageName != MAPS_PACKAGE_GO) return
        HudBus.publishNav(
            NavState(active = false, instruction = "Navigation ended"),
            NavSource.MAPS,
        )
    }

    private fun parse(n: Notification): NavState? {
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
            etaMin = 0,
            maneuver = ManeuverParser.fromText(instruction),
        )
    }

    companion object {
        private const val TAG = "NavListener"
        const val MAPS_PACKAGE = "com.google.android.apps.maps"
        const val MAPS_PACKAGE_GO = "com.google.android.apps.mapslite"
    }
}
