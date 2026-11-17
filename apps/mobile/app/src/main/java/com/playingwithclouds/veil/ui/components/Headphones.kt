package com.playingwithclouds.veil.ui.components

import android.content.Context
import android.media.AudioDeviceCallback
import android.media.AudioDeviceInfo
import android.media.AudioManager
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext

/** Output devices that only the wearer hears. */
private val HEADPHONE_TYPES = setOf(
    AudioDeviceInfo.TYPE_WIRED_HEADPHONES,
    AudioDeviceInfo.TYPE_WIRED_HEADSET,
    AudioDeviceInfo.TYPE_BLUETOOTH_A2DP,
    AudioDeviceInfo.TYPE_USB_HEADSET,
    AudioDeviceInfo.TYPE_BLE_HEADSET,
)

/**
 * Whether headphones are connected (wired, USB or Bluetooth), kept current while composed. Previews
 * play their sound only then, so nothing is heard through the speaker by surprise.
 */
@Composable
fun rememberHeadphonesConnected(): Boolean {
    val audioManager = LocalContext.current.getSystemService(Context.AUDIO_SERVICE) as AudioManager
    var connected by remember { mutableStateOf(headphonesIn(audioManager)) }
    DisposableEffect(audioManager) {
        val callback = object : AudioDeviceCallback() {
            override fun onAudioDevicesAdded(addedDevices: Array<out AudioDeviceInfo>) {
                connected = headphonesIn(audioManager)
            }

            override fun onAudioDevicesRemoved(removedDevices: Array<out AudioDeviceInfo>) {
                connected = headphonesIn(audioManager)
            }
        }
        audioManager.registerAudioDeviceCallback(callback, null)
        onDispose { audioManager.unregisterAudioDeviceCallback(callback) }
    }
    return connected
}

/** Whether any output device now is a pair of headphones. */
private fun headphonesIn(audioManager: AudioManager): Boolean {
    return audioManager.getDevices(AudioManager.GET_DEVICES_OUTPUTS).any { device -> device.type in HEADPHONE_TYPES }
}
