package com.playingwithclouds.veil.cams

import android.content.Context
import android.graphics.Bitmap
import android.graphics.ImageFormat
import android.media.Image
import android.media.ImageReader
import android.os.Handler
import android.os.HandlerThread
import androidx.media3.common.C
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.VideoSize
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.ui.video.streamMediaSource
import kotlinx.coroutines.delay
import kotlinx.coroutines.withTimeoutOrNull
import kotlin.math.roundToInt

/**
 * Pictures from a live room without showing it: a muted player of its own on the lowest rendition
 * decodes into an [ImageReader]; every decoded frame is dropped at once unless [grab] asked for one,
 * which is turned into a small bitmap. Lives on the main thread (the player's); frames arrive on a
 * thread of their own. [onError] is told when the stream fails, so a fresh URL can be asked for.
 */
class LiveFrameSampler(context: Context, private val onError: () -> Unit) {

    private val thread = HandlerThread("cam-frames").also { started -> started.start() }
    private val handler = Handler(thread.looper)
    private val player = ExoPlayer.Builder(context.applicationContext).build()
    private var reader: ImageReader? = null

    @Volatile
    private var wanted = false

    @Volatile
    private var frame: Bitmap? = null

    init {
        player.volume = 0f
        player.trackSelectionParameters = player.trackSelectionParameters.buildUpon()
            .setForceLowestBitrate(true)
            .setMaxVideoSize(MAX_WIDTH, MAX_HEIGHT)
            .setTrackTypeDisabled(C.TRACK_TYPE_AUDIO, true)
            .build()
        player.addListener(object : Player.Listener {
            override fun onVideoSizeChanged(videoSize: VideoSize) {
                if (videoSize.width > 0 && videoSize.height > 0) {
                    attachReader(videoSize.width, videoSize.height)
                }
            }

            override fun onPlayerError(error: PlaybackException) {
                onError()
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                if (playbackState == Player.STATE_ENDED) {
                    onError()
                }
            }
        })
        attachReader(DEFAULT_WIDTH, DEFAULT_HEIGHT)
    }

    /** Plays [stream] from its live edge. */
    fun play(stream: PlayableStream) {
        player.setMediaSource(streamMediaSource(stream))
        player.prepare()
        player.playWhenReady = true
    }

    /** The next decoded frame, scaled down; null when none comes within [timeoutMilliseconds]. */
    suspend fun grab(timeoutMilliseconds: Long = GRAB_TIMEOUT_MILLISECONDS): Bitmap? {
        frame = null
        wanted = true
        return withTimeoutOrNull(timeoutMilliseconds) {
            while (frame == null) {
                delay(POLL_MILLISECONDS)
            }
            frame
        }
    }

    /** Stops the player and frees the reader and its thread. */
    fun release() {
        player.release()
        reader?.close()
        reader = null
        thread.quitSafely()
    }

    /** Points the player at a new reader of [width]×[height], closing the old one. */
    private fun attachReader(width: Int, height: Int) {
        val previous = reader
        if (previous != null && previous.width == width && previous.height == height) {
            return
        }
        val next = ImageReader.newInstance(width, height, ImageFormat.YUV_420_888, READER_IMAGES)
        next.setOnImageAvailableListener({ source -> onImage(source) }, handler)
        player.setVideoSurface(next.surface)
        reader = next
        previous?.close()
    }

    /** Takes the newest image off the reader, converting it only when one was asked for. */
    private fun onImage(source: ImageReader) {
        val image = runCatching { source.acquireLatestImage() }.getOrNull() ?: return
        try {
            if (wanted) {
                frame = toBitmap(image)
                wanted = false
            }
        } finally {
            image.close()
        }
    }

    /** A YUV image as an ARGB bitmap, scaled to [OUTPUT_HEIGHT]. */
    private fun toBitmap(image: Image): Bitmap {
        val planes = image.planes
        val pixels = YuvFrames.toArgb(
            YuvPlanes(
                y = bytes(planes[0].buffer),
                u = bytes(planes[1].buffer),
                v = bytes(planes[2].buffer),
                width = image.width,
                height = image.height,
                yRowStride = planes[0].rowStride,
                uvRowStride = planes[1].rowStride,
                uvPixelStride = planes[1].pixelStride,
            ),
        )
        val full = Bitmap.createBitmap(pixels, image.width, image.height, Bitmap.Config.ARGB_8888)
        val width = (image.width * OUTPUT_HEIGHT / image.height.toFloat()).roundToInt().coerceAtLeast(1)
        val scaled = Bitmap.createScaledBitmap(full, width, OUTPUT_HEIGHT, true)
        if (scaled !== full) {
            full.recycle()
        }
        return scaled
    }

    /** A plane's bytes. */
    private fun bytes(buffer: java.nio.ByteBuffer): ByteArray {
        val out = ByteArray(buffer.remaining())
        buffer.get(out)
        return out
    }

    private companion object {
        const val DEFAULT_WIDTH = 426
        const val DEFAULT_HEIGHT = 240
        const val MAX_WIDTH = 854
        const val MAX_HEIGHT = 480
        const val OUTPUT_HEIGHT = 288
        const val READER_IMAGES = 3
        const val GRAB_TIMEOUT_MILLISECONDS = 8_000L
        const val POLL_MILLISECONDS = 100L
    }
}

/** One YUV 4:2:0 picture's planes as an [Image] hands them out. */
class YuvPlanes(
    val y: ByteArray,
    val u: ByteArray,
    val v: ByteArray,
    val width: Int,
    val height: Int,
    val yRowStride: Int,
    val uvRowStride: Int,
    val uvPixelStride: Int,
)

/** YUV 4:2:0 (any strides, planar or interleaved chroma) → ARGB pixels, BT.601 limited range. */
object YuvFrames {

    /** The picture's pixels, row by row. */
    fun toArgb(planes: YuvPlanes): IntArray {
        val out = IntArray(planes.width * planes.height)
        for (row in 0 until planes.height) {
            convertRow(planes, row, out)
        }
        return out
    }

    /** Converts one row into [out]. */
    private fun convertRow(planes: YuvPlanes, row: Int, out: IntArray) {
        val chromaRow = (row / 2) * planes.uvRowStride
        for (column in 0 until planes.width) {
            val luma = (planes.y[row * planes.yRowStride + column].toInt() and BYTE) - LUMA_OFFSET
            val chroma = chromaRow + (column / 2) * planes.uvPixelStride
            val u = (planes.u[chroma].toInt() and BYTE) - CHROMA_OFFSET
            val v = (planes.v[chroma].toInt() and BYTE) - CHROMA_OFFSET
            out[row * planes.width + column] = argb(luma, u, v)
        }
    }

    /** One pixel from its luma and centred chroma. */
    private fun argb(luma: Int, u: Int, v: Int): Int {
        val scaled = LUMA_SCALE * maxOf(0, luma)
        val red = clamp((scaled + RED_FROM_V * v) shr SHIFT)
        val green = clamp((scaled - GREEN_FROM_U * u - GREEN_FROM_V * v) shr SHIFT)
        val blue = clamp((scaled + BLUE_FROM_U * u) shr SHIFT)
        return OPAQUE or (red shl 16) or (green shl 8) or blue
    }

    /** A channel cut to 0..255. */
    private fun clamp(value: Int): Int {
        return value.coerceIn(0, BYTE)
    }

    private const val BYTE = 0xFF
    private const val OPAQUE = -0x1000000
    private const val LUMA_OFFSET = 16
    private const val CHROMA_OFFSET = 128

    // BT.601 coefficients in 10-bit fixed point.
    private const val SHIFT = 10
    private const val LUMA_SCALE = 1192
    private const val RED_FROM_V = 1634
    private const val GREEN_FROM_U = 401
    private const val GREEN_FROM_V = 833
    private const val BLUE_FROM_U = 2066
}
