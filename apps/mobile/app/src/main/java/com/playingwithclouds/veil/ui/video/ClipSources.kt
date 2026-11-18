package com.playingwithclouds.veil.ui.video

import androidx.media3.exoplayer.source.ClippingMediaSource
import androidx.media3.exoplayer.source.MediaSource
import com.playingwithclouds.veil.data.PlayableStream

private const val MICROSECONDS_PER_SECOND = 1_000_000.0

/**
 * A stretch of a stream as a media source: playback starts at [startSeconds] and ends at
 * [endSeconds] (seeking within the stream; nothing is cut), so a playlist of them jumps from
 * stretch to stretch on its own. Used by the compilation and the shorties.
 */
fun clippedMediaSource(stream: PlayableStream, startSeconds: Double, endSeconds: Double): MediaSource {
    return ClippingMediaSource.Builder(streamMediaSource(stream))
        .setStartPositionUs((startSeconds * MICROSECONDS_PER_SECOND).toLong())
        .setEndPositionUs((endSeconds * MICROSECONDS_PER_SECOND).toLong())
        .setEnableInitialDiscontinuity(false)
        .build()
}
