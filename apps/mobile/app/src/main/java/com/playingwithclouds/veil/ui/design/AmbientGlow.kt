package com.playingwithclouds.veil.ui.design

import android.os.Build
import androidx.compose.animation.Crossfade
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.BlurredEdgeTreatment
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.ColorFilter
import androidx.compose.ui.graphics.ColorMatrix
import androidx.compose.ui.graphics.CompositingStrategy
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import coil3.compose.AsyncImage
import coil3.request.ImageRequest
import com.playingwithclouds.veil.api.ServerSettings

/** How opaque the glow is at the top of an unscrolled page, as on the web. */
private const val GLOW_OPACITY = 0.38f

/** How much more saturated the glow is than its picture. */
private const val GLOW_SATURATION = 1.5f

/** Share of the screen's height the glow covers. */
private const val GLOW_HEIGHT_FRACTION = 0.85f

/** The picture is decoded this small: the blur hides the detail, and it costs next to nothing. */
private const val GLOW_DECODE_PIXELS = 64

private const val GLOW_FADE_MILLISECONDS = 700

private val GlowBlurRadius = 60.dp

/**
 * The picture each page features, for the shared glow (the web's `ambientGlow`). Pages light it
 * with [ShowAmbientGlow]; the latest page still on screen wins, so a page opened over another
 * takes over and hands back when it closes.
 */
@Stable
class AmbientGlowState {
    private val shown = mutableStateListOf<Pair<Any, String>>()

    /** The picture the glow shows now; null for none. */
    val url: String?
        get() = shown.lastOrNull()?.second

    /** Lights the glow with [url] for [owner], over what was shown before. */
    internal fun show(owner: Any, url: String) {
        shown.removeAll { entry -> entry.first == owner }
        shown.add(owner to url)
    }

    /** Takes back what [owner] showed. */
    internal fun release(owner: Any) {
        shown.removeAll { entry -> entry.first == owner }
    }
}

/** The shell's glow, which pages light. */
val LocalAmbientGlow = staticCompositionLocalOf { AmbientGlowState() }

/** Lights the shared glow with a page's featured picture while the page is on screen; null lights nothing. */
@Composable
fun ShowAmbientGlow(url: String?) {
    val state = LocalAmbientGlow.current
    DisposableEffect(state, url) {
        val owner = Any()
        if (url != null) {
            state.show(owner, url)
        }
        onDispose { state.release(owner) }
    }
}

/**
 * The glow itself, as the web draws it: the featured picture blurred into soft colour behind the
 * top of the page, fading out toward the bottom and the sides; a new picture fades in over the old
 * one, and [visibility] (1 at the top of a page, 0 scrolled away) fades it as the page scrolls.
 * Needs Android 12's blur; older phones show no glow rather than a sharp picture.
 */
@Composable
fun AmbientGlow(state: AmbientGlowState, visibility: () -> Float, modifier: Modifier = Modifier) {
    if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
        return
    }
    val context = LocalContext.current
    Box(
        modifier
            .fillMaxWidth()
            .fillMaxHeight(GLOW_HEIGHT_FRACTION)
            .graphicsLayer {
                alpha = GLOW_OPACITY * visibility().coerceIn(0f, 1f)
                compositingStrategy = CompositingStrategy.Offscreen
            }
            .drawWithContent {
                drawContent()
                val mask = Brush.radialGradient(
                    0.2f to Color.Black,
                    1f to Color.Transparent,
                    center = Offset(size.width / 2, size.height * 0.3f),
                    radius = size.width * 0.9f,
                )
                drawRect(mask, blendMode = BlendMode.DstIn)
            },
    ) {
        Crossfade(state.url, animationSpec = tween(GLOW_FADE_MILLISECONDS), label = "glow") { url ->
            if (url == null) {
                return@Crossfade
            }
            val request = ImageRequest.Builder(context).data(ServerSettings.imageUrl(url)).size(GLOW_DECODE_PIXELS).build()
            AsyncImage(
                model = request,
                contentDescription = null,
                contentScale = ContentScale.Crop,
                colorFilter = ColorFilter.colorMatrix(ColorMatrix().apply { setToSaturation(GLOW_SATURATION) }),
                modifier = Modifier.fillMaxSize().blur(GlowBlurRadius, BlurredEdgeTreatment.Unbounded),
            )
        }
    }
}
