package com.playingwithclouds.veil.ui.scene

import android.app.PictureInPictureParams
import android.os.Build
import android.util.Rational
import androidx.activity.ComponentActivity
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.app.PictureInPictureModeChangedInfo
import androidx.core.util.Consumer
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.media3.common.Player
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.util.findActivity

/** How much of the video the picture-in-picture window shows. */
enum class PictureInPicturePrivacy(val label: String, val summary: String) {
    SHOW("Show video", "Shown"),
    BLUR("Blur video", "Blurred"),
    HIDE("Hide video", "Hidden"),
}

/** Whether the page is in the picture-in-picture window, and how to move it there. */
@Stable
class PictureInPictureState {
    var isActive by mutableStateOf(false)
        internal set
    internal var enter: () -> Unit = {}

    /** Shrinks the page into the picture-in-picture window. */
    fun enterPictureInPicture() {
        enter()
    }
}

/**
 * Keeps the video playing in a small window when the app goes to the background (automatically
 * from Android 12, on leaving the app before that) or when the button asks for it. Closing the
 * window pauses the video.
 */
@Composable
fun rememberPictureInPicture(player: Player): PictureInPictureState {
    val activity = LocalContext.current.findActivity() as? ComponentActivity
    val lifecycle = LocalLifecycleOwner.current.lifecycle
    val state = remember { PictureInPictureState() }
    if (activity == null) {
        return state
    }
    val paramsFor = { autoEnter: Boolean -> pictureInPictureParams(player, autoEnter) }
    state.enter = { activity.enterPictureInPictureMode(paramsFor(false)) }

    DisposableEffect(player, activity) {
        val playerListener = object : Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                activity.setPictureInPictureParams(paramsFor(isPlaying))
            }

            override fun onVideoSizeChanged(videoSize: androidx.media3.common.VideoSize) {
                activity.setPictureInPictureParams(paramsFor(player.isPlaying))
            }
        }
        player.addListener(playerListener)
        activity.setPictureInPictureParams(paramsFor(player.isPlaying))
        onDispose {
            player.removeListener(playerListener)
            activity.setPictureInPictureParams(PictureInPictureParams.Builder().setAutoEnterEnabled(false).build())
        }
    }
    DisposableEffect(player, activity) {
        val leaveHint = Runnable {
            if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S && player.isPlaying) {
                activity.enterPictureInPictureMode(paramsFor(false))
            }
        }
        val modeChanged = Consumer<PictureInPictureModeChangedInfo> { info ->
            state.isActive = info.isInPictureInPictureMode
            if (!info.isInPictureInPictureMode && !lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED)) {
                player.pause()
            }
        }
        val stopObserver = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_STOP && state.isActive) {
                player.pause()
            }
        }
        activity.addOnUserLeaveHintListener(leaveHint)
        activity.addOnPictureInPictureModeChangedListener(modeChanged)
        lifecycle.addObserver(stopObserver)
        onDispose {
            activity.removeOnUserLeaveHintListener(leaveHint)
            activity.removeOnPictureInPictureModeChangedListener(modeChanged)
            lifecycle.removeObserver(stopObserver)
        }
    }
    return state
}

/** The window parameters for the video on screen: its shape, and whether leaving the app starts the window. */
private fun pictureInPictureParams(player: Player, autoEnter: Boolean): PictureInPictureParams {
    val size = player.videoSize
    val shape = pictureInPictureAspect(size.width, size.height)
    val builder = PictureInPictureParams.Builder().setAspectRatio(Rational(shape.first, shape.second))
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
        builder.setAutoEnterEnabled(autoEnter)
    }
    return builder.build()
}

/** What replaces the video in the picture-in-picture window while the viewer asked to hide it. */
@Composable
fun PictureInPictureCover(modifier: Modifier = Modifier) {
    Box(modifier.fillMaxSize().background(VeilColors.canvas), contentAlignment = Alignment.Center) {
        Icon(VeilIcons.Play, contentDescription = null, tint = VeilColors.contentMuted, modifier = Modifier.size(32.dp))
    }
}
