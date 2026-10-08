package com.playingwithclouds.veil.ui.scene

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.rememberUpdatedState
import androidx.media3.common.Player
import com.playingwithclouds.veil.data.HeatReport
import com.playingwithclouds.veil.data.HeatSpan
import kotlin.math.abs

/** Playback shorter than this is buffering noise, not watching. */
private const val MIN_SPAN_SECONDS = 1.0

/** Jumps shorter than this are nudges, not scrubs. */
private const val MIN_SCRUB_SECONDS = 3.0

/** How close to the loop's A point a jump has to land to count as the loop restarting. */
private const val LOOP_RESTART_TOLERANCE_SECONDS = 0.5

private const val MILLISECONDS_PER_SECOND = 1000.0

/**
 * What one viewing session did on a scene: which stretches played and where the viewer jumped to.
 * The backend adds many sessions up into the most-replayed graph.
 */
class HeatSession {
    private val spans = mutableListOf<HeatSpan>()
    private val scrubs = mutableListOf<Double>()
    private var openSpanStartSeconds: Double? = null

    /** Playback began at [positionSeconds]. */
    fun playbackStarted(positionSeconds: Double) {
        openSpanStartSeconds = positionSeconds
    }

    /** Playback stopped at [positionSeconds], ending the stretch it began. */
    fun playbackStopped(positionSeconds: Double) {
        closeSpan(positionSeconds)
    }

    /**
     * The viewer jumped from one position to another, [playing] or not. A long enough jump counts as a
     * scrub to [toSeconds], unless it is the A–B loop starting over at [loopStartSeconds].
     */
    fun jumped(fromSeconds: Double, toSeconds: Double, playing: Boolean, loopStartSeconds: Double?) {
        closeSpan(fromSeconds)
        if (playing) {
            openSpanStartSeconds = toSeconds
        }
        if (abs(toSeconds - fromSeconds) < MIN_SCRUB_SECONDS) {
            return
        }
        if (loopStartSeconds != null && abs(toSeconds - loopStartSeconds) <= LOOP_RESTART_TOLERANCE_SECONDS) {
            return
        }
        scrubs.add(toSeconds)
    }

    /** Hands over what was collected since the last call, or null when there is nothing worth sending. */
    fun drain(durationSeconds: Double): HeatReport? {
        if (durationSeconds <= 0 || (spans.isEmpty() && scrubs.isEmpty())) {
            return null
        }
        val report = HeatReport(durationSeconds, spans.toList(), scrubs.toList())
        spans.clear()
        scrubs.clear()
        return report
    }

    /** Ends the open stretch at [positionSeconds], keeping it when it was long enough to count. */
    private fun closeSpan(positionSeconds: Double) {
        val start = openSpanStartSeconds ?: return
        openSpanStartSeconds = null
        if (positionSeconds - start >= MIN_SPAN_SECONDS) {
            spans.add(HeatSpan(start, positionSeconds))
        }
    }
}

/**
 * Reports what the viewer watches and jumps to as replay data: whenever playback pauses and when
 * the page closes. [loopStartMilliseconds] tells loop restarts apart from scrubs.
 */
@Composable
fun PlaybackHeatReporting(
    player: Player,
    loopStartMilliseconds: () -> Long?,
    onReport: (HeatReport) -> Unit,
) {
    val currentLoopStart = rememberUpdatedState(loopStartMilliseconds)
    val currentOnReport = rememberUpdatedState(onReport)
    DisposableEffect(player) {
        val session = HeatSession()
        val flush = {
            val duration = player.duration
            if (duration > 0) {
                session.drain(duration / MILLISECONDS_PER_SECOND)?.let(currentOnReport.value)
            }
        }
        val listener = object : Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                val position = player.currentPosition / MILLISECONDS_PER_SECOND
                if (isPlaying) {
                    session.playbackStarted(position)
                    return
                }
                session.playbackStopped(position)
                flush()
            }

            override fun onPositionDiscontinuity(oldPosition: Player.PositionInfo, newPosition: Player.PositionInfo, reason: Int) {
                if (reason != Player.DISCONTINUITY_REASON_SEEK) {
                    return
                }
                val loopStart = currentLoopStart.value()?.let { milliseconds -> milliseconds / MILLISECONDS_PER_SECOND }
                session.jumped(
                    oldPosition.positionMs / MILLISECONDS_PER_SECOND,
                    newPosition.positionMs / MILLISECONDS_PER_SECOND,
                    player.isPlaying,
                    loopStart,
                )
            }
        }
        player.addListener(listener)
        onDispose {
            player.removeListener(listener)
            if (player.isPlaying) {
                session.playbackStopped(player.currentPosition / MILLISECONDS_PER_SECOND)
            }
            flush()
        }
    }
}
