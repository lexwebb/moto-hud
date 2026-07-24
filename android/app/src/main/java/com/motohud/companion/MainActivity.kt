package com.motohud.companion

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.provider.Settings
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val status = findViewById<TextView>(R.id.statusText)
        val navText = findViewById<TextView>(R.id.navText)
        val mediaText = findViewById<TextView>(R.id.mediaText)

        findViewById<Button>(R.id.btnNotificationAccess).setOnClickListener {
            startActivity(Intent(Settings.ACTION_NOTIFICATION_LISTENER_SETTINGS))
        }
        findViewById<Button>(R.id.btnStart).setOnClickListener {
            ensurePermissions()
            ContextCompat.startForegroundService(this, Intent(this, HudForegroundService::class.java))
        }
        findViewById<Button>(R.id.btnStop).setOnClickListener {
            stopService(Intent(this, HudForegroundService::class.java))
            HudBus.setStatus("Stopped")
        }

        lifecycleScope.launch {
            HudBus.status.collect { status.text = it }
        }
        lifecycleScope.launch {
            HudBus.nav.collect {
                navText.text = if (it.active) {
                    "${it.maneuver} ${it.distanceText}\n${it.instruction}"
                } else {
                    "Nav idle"
                }
            }
        }
        lifecycleScope.launch {
            HudBus.media.collect {
                mediaText.text = listOf(it.title, it.artist).filter { s -> s.isNotBlank() }.joinToString(" — ")
                    .ifBlank { "No media" }
            }
        }
    }

    private fun ensurePermissions() {
        val needed = mutableListOf<String>()
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
        }
    }
}
