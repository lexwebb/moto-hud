package com.motohud.companion

import android.content.Context
import android.util.Log
import com.google.android.play.core.splitinstall.SplitInstallManagerFactory
import com.google.android.play.core.splitinstall.SplitInstallRequest
import com.google.android.play.core.splitinstall.model.SplitInstallSessionStatus
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow

/** Play Feature Delivery helper for the on-demand `:osmand` module. */
object OsmandModule {
    const val NAME = "osmand"
    private const val TAG = "OsmandModule"
    private const val ENGINE_CLASS = "com.motohud.companion.osmand.OsmandEmbeddedNavEngine"

    fun isInstalled(context: Context): Boolean {
        return try {
            context.classLoader.loadClass(MotoHudAppComponentFactory.OSMAND_APP)
            true
        } catch (_: ClassNotFoundException) {
            val mgr = SplitInstallManagerFactory.create(context.applicationContext)
            NAME in mgr.installedModules
        }
    }

    fun isRichNavReady(context: Context): Boolean {
        return isInstalled(context) &&
            context.applicationContext.javaClass.name == MotoHudAppComponentFactory.OSMAND_APP
    }

    /** Reflection — base module must not compile against the feature. */
    fun createEmbeddedEngine(context: Context): NavEngine? {
        return try {
            val cls = Class.forName(ENGINE_CLASS)
            cls.getConstructor(Context::class.java).newInstance(context) as NavEngine
        } catch (e: Exception) {
            Log.w(TAG, "Embedded engine unavailable", e)
            null
        }
    }

    fun requestInstall(context: Context): Flow<InstallEvent> = callbackFlow {
        val mgr = SplitInstallManagerFactory.create(context.applicationContext)
        if (NAME in mgr.installedModules) {
            trySend(InstallEvent.AlreadyInstalled)
            close()
            return@callbackFlow
        }

        val listener = com.google.android.play.core.splitinstall.SplitInstallStateUpdatedListener { state ->
            when (state.status()) {
                SplitInstallSessionStatus.PENDING,
                SplitInstallSessionStatus.DOWNLOADING,
                -> {
                    val total = state.totalBytesToDownload().coerceAtLeast(1L)
                    val pct = ((state.bytesDownloaded() * 100) / total).toInt()
                    trySend(InstallEvent.Progress(pct))
                }
                SplitInstallSessionStatus.INSTALLING -> trySend(InstallEvent.Installing)
                SplitInstallSessionStatus.INSTALLED -> {
                    trySend(InstallEvent.Installed)
                    close()
                }
                SplitInstallSessionStatus.FAILED -> {
                    trySend(InstallEvent.Failed(state.errorCode()))
                    close()
                }
                SplitInstallSessionStatus.CANCELED -> {
                    trySend(InstallEvent.Canceled)
                    close()
                }
                SplitInstallSessionStatus.REQUIRES_USER_CONFIRMATION -> {
                    trySend(InstallEvent.NeedsConfirmation(state))
                }
                else -> Unit
            }
        }
        mgr.registerListener(listener)
        val req = SplitInstallRequest.newBuilder().addModule(NAME).build()
        mgr.startInstall(req)
            .addOnFailureListener { e ->
                Log.e(TAG, "startInstall failed", e)
                trySend(InstallEvent.Failed(-1))
                close()
            }
        awaitClose { mgr.unregisterListener(listener) }
    }

    sealed class InstallEvent {
        data object AlreadyInstalled : InstallEvent()
        data class Progress(val percent: Int) : InstallEvent()
        data object Installing : InstallEvent()
        data object Installed : InstallEvent()
        data object Canceled : InstallEvent()
        data class Failed(val code: Int) : InstallEvent()
        data class NeedsConfirmation(
            val state: com.google.android.play.core.splitinstall.SplitInstallSessionState,
        ) : InstallEvent()
    }
}
