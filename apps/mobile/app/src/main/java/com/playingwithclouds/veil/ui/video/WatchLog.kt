package com.playingwithclouds.veil.ui.video

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.media3.common.Player
import kotlinx.coroutines.delay

/** Progress is logged after this much actual playing. */
private const val PROGRESS_EVERY_SECONDS = 15.0

/** Stretches shorter than this are a seek landing, not watching. */
private const val MIN_SPAN_SECONDS = 1.0

/** A larger jump between two ticks is a seek, not playing time. */
private const val MAX_TICK_SECONDS = 2.0

/** How often the playing position is sampled. */
private const val TICK_MILLISECONDS = 1000L

private const val MILLISECONDS_PER_SECOND = 1000.0

/** What a playback session reports to the event log. */
sealed interface WatchEvent {

    /** The resume point. */
    data class Progress(val positionSeconds: Double) : WatchEvent

    /** A stretch that actually played, for the most-replayed data. */
    data class Span(val startSeconds: Double, val endSeconds: Double) : WatchEvent

    /** Played to the end. */
    data class Finish(val positionSeconds: Double) : WatchEvent
}

/**
 * One playback session, as the web's `watchLog.ts` logs it: progress every
 * [PROGRESS_EVERY_SECONDS] of playing and on pause, a span per stretch that played (closed by a
 * pause, a seek, the end or leaving) and finish at the end. Each call returns what to log.
 */
class WatchSession {
    private var spanStartSeconds: Double? = null
    private var lastPositionSeconds = 0.0
    private var playedSinceProgress = 0.0

    /** Playback started (or went on after buffering) at [positionSeconds]. */
    fun playing(positionSeconds: Double) {
        if (spanStartSeconds == null) {
            spanStartSeconds = positionSeconds
        }
        lastPositionSeconds = positionSeconds
    }

    /** The playing position, sampled about once a second; counts playing time toward progress. */
    fun tick(positionSeconds: Double): List<WatchEvent> {
        val step = positionSeconds - lastPositionSeconds
        lastPositionSeconds = positionSeconds
        if (step <= 0 || step > MAX_TICK_SECONDS) {
            return emptyList()
        }
        playedSinceProgress += step
        if (playedSinceProgress < PROGRESS_EVERY_SECONDS) {
            return emptyList()
        }
        return listOf(progress())
    }

    /** Playback paused at [positionSeconds]: the stretch ends and the resume point is logged. */
    fun paused(positionSeconds: Double): List<WatchEvent> {
        lastPositionSeconds = positionSeconds
        return closeSpan() + progress()
    }

    /** The viewer jumped from [fromSeconds] to [toSeconds]; the stretch ends at the jump. */
    fun seeked(fromSeconds: Double, toSeconds: Double, isPlaying: Boolean): List<WatchEvent> {
        lastPositionSeconds = fromSeconds
        val closed = closeSpan()
        lastPositionSeconds = toSeconds
        if (isPlaying) {
            spanStartSeconds = toSeconds
        }
        return closed
    }

    /** Played to the end at [durationSeconds]. */
    fun ended(durationSeconds: Double): List<WatchEvent> {
        lastPositionSeconds = durationSeconds
        return closeSpan() + WatchEvent.Finish(durationSeconds)
    }

    /** The page closes: what's still open, and the resume point if anything played since the last. */
    fun closed(): List<WatchEvent> {
        val closed = closeSpan()
        if (playedSinceProgress > 0) {
            return closed + progress()
        }
        return closed
    }

    /** The resume point at the last position. */
    private fun progress(): WatchEvent {
        playedSinceProgress = 0.0
        return WatchEvent.Progress(lastPositionSeconds)
    }

    /** Ends the open stretch at the last position, kept when it's long enough to count. */
    private fun closeSpan(): List<WatchEvent> {
        val start = spanStartSeconds ?: return emptyList()
        spanStartSeconds = null
        if (lastPositionSeconds - start < MIN_SPAN_SECONDS) {
            return emptyList()
        }
        return listOf(WatchEvent.Span(start, lastPositionSeconds))
    }
}

/** Feeds the player's playback into a [WatchSession] while composed and hands what it reports to [onEvent]. */
@Composable
fun WatchTracking(player: Player, onEvent: (WatchEvent) -> Unit) {
    val currentOnEvent by rememberUpdatedState(onEvent)
    val session = remember(player) { WatchSession() }
    val report = { events: List<WatchEvent> -> events.forEach { event -> currentOnEvent(event) } }
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                val position = player.currentPosition / MILLISECONDS_PER_SECOND
                if (isPlaying) {
                    session.playing(position)
                    return
                }
                if (player.playbackState != Player.STATE_ENDED) {
                    report(session.paused(position))
                }
            }

            override fun onPositionDiscontinuity(oldPosition: Player.PositionInfo, newPosition: Player.PositionInfo, reason: Int) {
                if (reason != Player.DISCONTINUITY_REASON_SEEK) {
                    return
                }
                report(session.seeked(oldPosition.positionMs / MILLISECONDS_PER_SECOND, newPosition.positionMs / MILLISECONDS_PER_SECOND, player.isPlaying))
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                if (playbackState == Player.STATE_ENDED && player.duration > 0) {
                    report(session.ended(player.duration / MILLISECONDS_PER_SECOND))
                }
            }
        }
        player.addListener(listener)
        onDispose {
            player.removeListener(listener)
            report(session.closed())
        }
    }
    LaunchedEffect(player) {
        while (true) {
            delay(TICK_MILLISECONDS)
            if (player.isPlaying) {
                report(session.tick(player.currentPosition / MILLISECONDS_PER_SECOND))
            }
        }
    }
}
