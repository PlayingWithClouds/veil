package com.playingwithclouds.veil.privacy

import kotlin.math.abs

/**
 * Spots the phone being laid screen-down from accelerometer samples: gravity pulls along the
 * screen's back (z below -[FACE_DOWN_Z]) with the phone otherwise flat, held for [HOLD_MILLISECONDS].
 * Fires once, then waits for the phone to be lifted face-up again.
 */
class FaceDownDetector {

    private var faceDownSinceMilliseconds: Long? = null
    private var fired = false

    /** Feeds one sample (m/s² per axis); true exactly when the flip is confirmed. */
    fun onSample(x: Float, y: Float, z: Float, timestampMilliseconds: Long): Boolean {
        if (z > FACE_UP_Z) {
            fired = false
        }
        val flat = abs(x) < FLAT_TILT && abs(y) < FLAT_TILT
        if (z > -FACE_DOWN_Z || !flat) {
            faceDownSinceMilliseconds = null
            return false
        }
        val since = faceDownSinceMilliseconds
        if (since == null) {
            faceDownSinceMilliseconds = timestampMilliseconds
            return false
        }
        if (fired || timestampMilliseconds - since < HOLD_MILLISECONDS) {
            return false
        }
        fired = true
        return true
    }

    private companion object {
        const val FACE_DOWN_Z = 8.5f
        const val FACE_UP_Z = 3f
        const val FLAT_TILT = 3.5f
        const val HOLD_MILLISECONDS = 500L
    }
}

/** Recognises two taps in quick succession; the caller decides what counts as a tap. */
class DoubleTapDetector(private val windowMilliseconds: Long = DEFAULT_WINDOW_MILLISECONDS) {

    private var lastTapMilliseconds: Long? = null

    /** Registers a tap; true when it completes a double tap. */
    fun onTap(timestampMilliseconds: Long): Boolean {
        val last = lastTapMilliseconds
        if (last != null && timestampMilliseconds - last <= windowMilliseconds) {
            lastTapMilliseconds = null
            return true
        }
        lastTapMilliseconds = timestampMilliseconds
        return false
    }

    private companion object {
        const val DEFAULT_WINDOW_MILLISECONDS = 400L
    }
}
