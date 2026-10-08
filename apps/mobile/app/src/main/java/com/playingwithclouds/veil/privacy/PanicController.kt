package com.playingwithclouds.veil.privacy

import android.app.Activity
import android.content.Context
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.os.SystemClock
import android.view.MotionEvent

/**
 * The panic action and its two triggers: laying the phone face-down (accelerometer, only while
 * the app is in front) and a two-finger double tap (read from the activity's touch stream, so it
 * never competes with single-finger gestures). The action pauses playback, locks the app when a
 * lock is set, and sends the app to the background, which lands on the launcher.
 */
class PanicController(private val activity: Activity) : SensorEventListener {

    private val sensorManager = activity.getSystemService(Context.SENSOR_SERVICE) as SensorManager
    private val faceDownDetector = FaceDownDetector()
    private val doubleTapDetector = DoubleTapDetector()
    private var listening = false

    /** Starts or stops the accelerometer to match whether the flip trigger is on. */
    fun updateSensor(flipEnabled: Boolean) {
        if (flipEnabled && !listening) {
            val accelerometer = sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER) ?: return
            sensorManager.registerListener(this, accelerometer, SensorManager.SENSOR_DELAY_NORMAL)
            listening = true
        }
        if (!flipEnabled) {
            stopSensor()
        }
    }

    /** Stops listening to the accelerometer. */
    fun stopSensor() {
        if (listening) {
            sensorManager.unregisterListener(this)
            listening = false
        }
    }

    /** Watches touches for the two-finger double tap; never consumes them. */
    fun onTouch(event: MotionEvent) {
        val isSecondFinger = event.actionMasked == MotionEvent.ACTION_POINTER_DOWN && event.pointerCount == 2
        if (!PrivacyPreferences.panicTap.value || !isSecondFinger) {
            return
        }
        if (doubleTapDetector.onTap(event.eventTime)) {
            trigger()
        }
    }

    /** Runs the panic action. */
    fun trigger() {
        PrivacyEvents.requestPause()
        AppLock.lockNow()
        activity.moveTaskToBack(true)
    }

    /** Feeds accelerometer samples to the face-down detector. */
    override fun onSensorChanged(event: SensorEvent) {
        val timestampMilliseconds = SystemClock.elapsedRealtime()
        if (faceDownDetector.onSample(event.values[0], event.values[1], event.values[2], timestampMilliseconds)) {
            trigger()
        }
    }

    /** Required by the listener interface; accuracy does not matter here. */
    override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) = Unit
}
