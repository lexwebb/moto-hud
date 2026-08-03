package com.motohud.companion

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.provider.Settings
import android.util.Log
import android.view.View
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.appcompat.widget.SwitchCompat
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.lifecycleScope
import com.google.android.play.core.splitcompat.SplitCompat
import com.google.android.play.core.splitinstall.SplitInstallManagerFactory
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {

    override fun attachBaseContext(newBase: android.content.Context) {
        super.attachBaseContext(newBase)
        SplitCompat.installActivity(this)
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val status = findViewById<TextView>(R.id.statusText)
        val navText = findViewById<TextView>(R.id.navText)
        val mediaText = findViewById<TextView>(R.id.mediaText)
        val httpEnable = findViewById<SwitchCompat>(R.id.httpEnable)
        val httpUrl = findViewById<EditText>(R.id.httpUrl)
        val richBtn = findViewById<Button>(R.id.btnRichNav)
        val openOsmand = findViewById<Button>(R.id.btnOpenOsmand)

        httpEnable.isChecked = LinkPrefs.httpEnabled(this)
        httpUrl.setText(LinkPrefs.httpBaseUrl(this))
        refreshRichNavUi(richBtn, openOsmand)

        richBtn.setOnClickListener { onRichNavClicked(richBtn, openOsmand) }
        openOsmand.setOnClickListener { openEmbeddedOsmandMap() }

        findViewById<Button>(R.id.btnNotificationAccess).setOnClickListener {
            startActivity(Intent(Settings.ACTION_NOTIFICATION_LISTENER_SETTINGS))
        }
        findViewById<Button>(R.id.btnStart).setOnClickListener {
            LinkPrefs.setHttp(this, httpEnable.isChecked, httpUrl.text.toString())
            if (!ensurePermissions()) return@setOnClickListener
            ContextCompat.startForegroundService(this, Intent(this, HudForegroundService::class.java))
            Toast.makeText(this, R.string.start_hud, Toast.LENGTH_SHORT).show()
        }
        findViewById<Button>(R.id.btnStop).setOnClickListener {
            stopService(Intent(this, HudForegroundService::class.java))
            HudBus.setStatus("Stopped")
        }

        lifecycleScope.launch {
            HudBus.status.collect { status.text = it }
        }
        lifecycleScope.launch {
            HudBus.nav.collect { updateNavLabel(navText, it) }
        }
        lifecycleScope.launch {
            // engine= is derived from bind state; refresh when AIDL connects
            // even if nav is still idle.
            HudBus.osmandBound.collect { updateNavLabel(navText, HudBus.nav.value) }
        }
        lifecycleScope.launch {
            HudBus.media.collect {
                mediaText.text = listOf(it.title, it.artist).filter { s -> s.isNotBlank() }.joinToString(" — ")
                    .ifBlank { "No media" }
            }
        }
    }

    private fun updateNavLabel(navText: TextView, nav: NavState) {
        val engine = when {
            OsmandModule.isRichNavReady(this) -> "embedded"
            HudBus.isOsmandBound() -> "aidl"
            else -> "maps"
        }
        val src = when {
            HudBus.isOsmandBound() && nav.lanes.isNotEmpty() -> "OsmAnd+$engine·lanes"
            HudBus.isOsmandBound() -> "OsmAnd+$engine"
            else -> "Maps?"
        }
        navText.text = if (nav.active) {
            val then = nav.thenNext?.let { t -> " → then ${t.maneuver} ${t.distanceText}" }.orEmpty()
            val junc = nav.junction?.let { j -> " · j=${j.kind}/${j.outbound}" }.orEmpty()
            "[$src] ${nav.maneuver} ${nav.distanceText}$then$junc\n${nav.road.ifBlank { nav.instruction }}"
        } else {
            "Nav idle · engine=$engine"
        }
    }

    private fun refreshRichNavUi(richBtn: Button, openOsmand: Button) {
        when {
            OsmandModule.isRichNavReady(this) -> {
                richBtn.text = getString(R.string.rich_nav_ready)
                richBtn.isEnabled = false
                openOsmand.visibility = View.VISIBLE
            }
            OsmandModule.isInstalled(this) -> {
                richBtn.text = getString(R.string.rich_nav_restart)
                richBtn.isEnabled = true
                openOsmand.visibility = View.GONE
            }
            else -> {
                richBtn.text = getString(R.string.rich_nav_download)
                richBtn.isEnabled = true
                openOsmand.visibility = View.GONE
            }
        }
    }

    private fun onRichNavClicked(richBtn: Button, openOsmand: Button) {
        if (OsmandModule.isInstalled(this) && !OsmandModule.isRichNavReady(this)) {
            Toast.makeText(this, R.string.rich_nav_restart_hint, Toast.LENGTH_LONG).show()
            // Kill process so AppComponentFactory can pick OsmAnd Application.
            finishAffinity()
            Runtime.getRuntime().exit(0)
            return
        }
        richBtn.isEnabled = false
        lifecycleScope.launch {
            OsmandModule.requestInstall(this@MainActivity).collect { event ->
                when (event) {
                    is OsmandModule.InstallEvent.Progress ->
                        richBtn.text = getString(R.string.rich_nav_progress, event.percent)
                    OsmandModule.InstallEvent.Installing,
                    OsmandModule.InstallEvent.AlreadyInstalled,
                    -> richBtn.text = getString(R.string.rich_nav_installing)
                    OsmandModule.InstallEvent.Installed -> {
                        Toast.makeText(this@MainActivity, R.string.rich_nav_restart_hint, Toast.LENGTH_LONG).show()
                        refreshRichNavUi(richBtn, openOsmand)
                        richBtn.isEnabled = true
                    }
                    is OsmandModule.InstallEvent.Failed -> {
                        Toast.makeText(
                            this@MainActivity,
                            getString(R.string.rich_nav_failed, event.code),
                            Toast.LENGTH_LONG,
                        ).show()
                        refreshRichNavUi(richBtn, openOsmand)
                    }
                    OsmandModule.InstallEvent.Canceled -> refreshRichNavUi(richBtn, openOsmand)
                    is OsmandModule.InstallEvent.NeedsConfirmation -> {
                        SplitInstallManagerFactory.create(this@MainActivity)
                            .startConfirmationDialogForResult(event.state, this@MainActivity, REQ_CONFIRM)
                    }
                }
            }
        }
    }

    private fun openEmbeddedOsmandMap() {
        try {
            val cls = Class.forName("net.osmand.plus.activities.MapActivity")
            startActivity(Intent(this, cls))
        } catch (e: Exception) {
            Log.e(TAG, "MapActivity unavailable", e)
            Toast.makeText(this, R.string.rich_nav_map_unavailable, Toast.LENGTH_SHORT).show()
        }
    }

    /** @return true if all required permissions are already granted */
    private fun ensurePermissions(): Boolean {
        val needed = mutableListOf<String>()
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION)
            != PackageManager.PERMISSION_GRANTED
        ) {
            needed += Manifest.permission.ACCESS_FINE_LOCATION
        }
        if (Build.VERSION.SDK_INT >= 31) {
            listOf(
                Manifest.permission.BLUETOOTH_CONNECT,
                Manifest.permission.BLUETOOTH_SCAN,
            ).forEach {
                if (ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED) {
                    needed += it
                }
            }
        }
        if (Build.VERSION.SDK_INT >= 33) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED
            ) {
                needed += Manifest.permission.POST_NOTIFICATIONS
            }
        }
        if (needed.isNotEmpty()) {
            ActivityCompat.requestPermissions(this, needed.toTypedArray(), 1001)
            Toast.makeText(this, "Grant permissions, then tap Start again", Toast.LENGTH_LONG).show()
            return false
        }
        return true
    }

    companion object {
        private const val TAG = "MainActivity"
        private const val REQ_CONFIRM = 4401
    }
}
