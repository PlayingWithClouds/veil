package com.playingwithclouds.veil.data

import android.graphics.Bitmap
import android.media.MediaMetadataRetriever
import android.util.Base64
import java.io.ByteArrayOutputStream
import kotlin.math.abs
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

/** A thumbnail is only redone when the best moment moved further than this from the one it shows. */
private const val THUMBNAIL_MOVE_SECONDS = 30.0

/** Width a captured frame is scaled down to; listings never show it larger. */
private const val THUMBNAIL_WIDTH = 640

private const val JPEG_QUALITY = 82
private const val MICROSECONDS_PER_SECOND = 1_000_000.0

/** Whether the scene needs a new thumbnail: it has a best moment and none yet, or one far from it. */
fun needsNewThumbnail(bestMomentSeconds: Double?, thumbnailSeconds: Double?): Boolean {
    if (bestMomentSeconds == null) {
        return false
    }
    if (thumbnailSeconds == null) {
        return true
    }
    return abs(bestMomentSeconds - thumbnailSeconds) > THUMBNAIL_MOVE_SECONDS
}

/** Takes a frame out of a scene's stream for use as its thumbnail. */
object ThumbnailCapture {

    /** Width a frame for the face search is scaled down to: larger than a thumbnail, so faces stay sharp. */
    const val FACE_FRAME_WIDTH = 1280

    /**
     * The frame at [atSeconds] as a base64 JPEG at most [width] wide, or null when the device
     * cannot read it (most HLS streams, a stream that has gone away).
     */
    suspend fun jpegBase64(stream: PlayableStream, atSeconds: Double, width: Int = THUMBNAIL_WIDTH): String? {
        return withContext(Dispatchers.IO) {
            runCatching { grabFrame(stream, atSeconds, width) }.getOrNull()
        }
    }

    /** Reads the frame through [MediaMetadataRetriever], scaled to [width]. */
    private fun grabFrame(stream: PlayableStream, atSeconds: Double, width: Int): String? {
        val retriever = MediaMetadataRetriever()
        try {
            retriever.setDataSource(stream.url, stream.headers)
            val frame = retriever.getFrameAtTime((atSeconds * MICROSECONDS_PER_SECOND).toLong(), MediaMetadataRetriever.OPTION_CLOSEST_SYNC)
                ?: return null
            return encode(scaled(frame, width))
        } finally {
            retriever.release()
        }
    }

    /** The frame scaled down to [width], keeping its shape. */
    private fun scaled(frame: Bitmap, width: Int): Bitmap {
        if (frame.width <= width) {
            return frame
        }
        val height = frame.height * width / frame.width
        return Bitmap.createScaledBitmap(frame, width, height, true)
    }

    /** The bitmap as a base64 JPEG. */
    private fun encode(bitmap: Bitmap): String {
        val bytes = ByteArrayOutputStream()
        bitmap.compress(Bitmap.CompressFormat.JPEG, JPEG_QUALITY, bytes)
        return Base64.encodeToString(bytes.toByteArray(), Base64.NO_WRAP)
    }
}
