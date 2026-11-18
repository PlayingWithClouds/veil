package com.playingwithclouds.veil.habits

/** Watching at least this long makes leaving the app right after it count as finishing. */
private const val MIN_WATCHED_SECONDS = 180.0

/** Leaving the app this long after the video last played still counts: a few minutes to clean up. */
private const val MAX_LEAVE_DELAY_MILLISECONDS = 5 * 60 * 1000L

/** One video counts at most once in this long, however often the app is left. */
private const val GUESS_COOLDOWN_MILLISECONDS = 6 * 60 * 60 * 1000L

/** A progress step longer than the wall-clock time it took (plus this slack) was a seek, not watching. */
private const val SEEK_SLACK_SECONDS = 2.0

private const val MILLISECONDS_PER_SECOND = 1000.0

/** A finish the app guessed: the video and where it was when it last played. */
data class GuessedFinish(val videoId: String, val positionSeconds: Double)

/**
 * Guesses an O nobody tapped, since hardly anyone does: the app was left (closed, switched away
 * or the screen turned off) within a few minutes of a video playing, after watching it for at
 * least a few minutes. Fed the watch log's progress reports; [now] is the wall clock in
 * milliseconds.
 */
class FinishGuesser(private val now: () -> Long) {

    private var videoId: String? = null
    private var watchedSeconds = 0.0
    private var lastPositionSeconds = 0.0
    private var lastPlayedAt = 0L
    private val guessedAt = mutableMapOf<String, Long>()

    /** The player of [id] reported its position: counts the time watched since the last report. */
    fun progress(id: String, positionSeconds: Double) {
        val at = now()
        if (id != videoId) {
            videoId = id
            watchedSeconds = 0.0
            lastPositionSeconds = positionSeconds
            lastPlayedAt = at
            return
        }
        val step = positionSeconds - lastPositionSeconds
        val elapsedSeconds = (at - lastPlayedAt) / MILLISECONDS_PER_SECOND
        if (step > 0 && step <= elapsedSeconds + SEEK_SLACK_SECONDS) {
            watchedSeconds += step
        }
        lastPositionSeconds = positionSeconds
        lastPlayedAt = at
    }

    /** The app was left: the finish this suggests, if any; the watch is then forgotten. */
    fun appLeft(): GuessedFinish? {
        val id = videoId ?: return null
        val at = now()
        val guess = guessFor(id, at)
        videoId = null
        watchedSeconds = 0.0
        if (guess != null) {
            guessedAt[id] = at
        }
        return guess
    }

    /** The finish for [id] when it was watched long enough, played recently and wasn't guessed lately. */
    private fun guessFor(id: String, at: Long): GuessedFinish? {
        if (watchedSeconds < MIN_WATCHED_SECONDS) {
            return null
        }
        if (at - lastPlayedAt > MAX_LEAVE_DELAY_MILLISECONDS) {
            return null
        }
        val previous = guessedAt[id]
        if (previous != null && at - previous < GUESS_COOLDOWN_MILLISECONDS) {
            return null
        }
        return GuessedFinish(id, lastPositionSeconds)
    }
}
