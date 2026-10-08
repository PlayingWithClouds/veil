package com.playingwithclouds.veil

import android.os.Bundle
import android.os.SystemClock
import android.view.MotionEvent
import android.view.WindowManager
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.fragment.app.FragmentActivity
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.lifecycleScope
import androidx.lifecycle.repeatOnLifecycle
import com.playingwithclouds.veil.privacy.AppLock
import com.playingwithclouds.veil.privacy.PanicController
import com.playingwithclouds.veil.privacy.PrivacyPreferences
import com.playingwithclouds.veil.ui.VeilApp
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch

/** The single activity; every screen is Compose. A FragmentActivity so the biometric prompt can attach. */
class MainActivity : FragmentActivity() {

    private lateinit var panicController: PanicController

    /** Draws edge to edge, wires the privacy hooks and hands the window to the Compose app. */
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        panicController = PanicController(this)
        followSecureFlag()
        followPanicSensor()
        setContent { VeilApp() }
    }

    /** Applies the recents and screenshot protection whenever the option or the app lock changes. */
    private fun followSecureFlag() {
        lifecycleScope.launch {
            combine(PrivacyPreferences.hideInRecents, PrivacyPreferences.lockEnabled) { hide, lock -> hide || lock }
                .collect { secure -> setSecure(secure) }
        }
    }

    /** Sets or clears FLAG_SECURE: blank recents card, black screenshots and recordings. */
    private fun setSecure(secure: Boolean) {
        if (secure) {
            window.setFlags(WindowManager.LayoutParams.FLAG_SECURE, WindowManager.LayoutParams.FLAG_SECURE)
            return
        }
        window.clearFlags(WindowManager.LayoutParams.FLAG_SECURE)
    }

    /** Listens to the accelerometer for the face-down trigger while the app is in front. */
    private fun followPanicSensor() {
        lifecycleScope.launch {
            repeatOnLifecycle(Lifecycle.State.STARTED) {
                try {
                    PrivacyPreferences.panicFlip.collect { enabled -> panicController.updateSensor(enabled) }
                } finally {
                    panicController.stopSensor()
                }
            }
        }
    }

    /** Locks again when the app was in the background long enough. */
    override fun onStart() {
        super.onStart()
        AppLock.onForegrounded(SystemClock.elapsedRealtime())
    }

    /** Notes when the app left the foreground, for the lock delay. */
    override fun onStop() {
        super.onStop()
        AppLock.onBackgrounded(SystemClock.elapsedRealtime())
    }

    /** Lets the panic trigger watch touches, then handles them as usual. */
    override fun dispatchTouchEvent(event: MotionEvent): Boolean {
        panicController.onTouch(event)
        return super.dispatchTouchEvent(event)
    }
}
