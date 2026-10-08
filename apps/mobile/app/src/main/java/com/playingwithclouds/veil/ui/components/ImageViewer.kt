package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTransformGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties

private const val MAX_ZOOM = 5f

/** A full-screen pager over images with pinch-to-zoom; system back or the close button dismisses it. */
@Composable
fun ImageViewer(imagePaths: List<String>, initialIndex: Int, onDismiss: () -> Unit) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false, decorFitsSystemWindows = false),
    ) {
        val pagerState = rememberPagerState(initialPage = initialIndex) { imagePaths.size }
        Box(Modifier.fillMaxSize().background(Color.Black)) {
            HorizontalPager(pagerState, Modifier.fillMaxSize(), beyondViewportPageCount = 1) { page ->
                ZoomableImage(imagePaths[page], isCurrent = pagerState.currentPage == page)
            }
            IconButton(onClick = onDismiss, modifier = Modifier.align(Alignment.TopEnd).padding(8.dp)) {
                Icon(Icons.Filled.Close, contentDescription = "Close", tint = Color.White)
            }
            Text(
                "${pagerState.currentPage + 1} / ${imagePaths.size}",
                modifier = Modifier.align(Alignment.BottomCenter).padding(16.dp),
                color = Color.White,
            )
        }
    }
}

/** One image that can be pinched and dragged; it resets when the pager moves away from it. */
@Composable
private fun ZoomableImage(path: String, isCurrent: Boolean) {
    var scale by remember { mutableFloatStateOf(1f) }
    var offset by remember { mutableStateOf(Offset.Zero) }
    LaunchedEffect(isCurrent) {
        if (!isCurrent) {
            scale = 1f
            offset = Offset.Zero
        }
    }
    RemoteImage(
        path = path,
        contentScale = ContentScale.Fit,
        modifier = Modifier
            .fillMaxSize()
            .pointerInput(Unit) {
                detectTransformGestures { _, pan, zoom, _ ->
                    scale = (scale * zoom).coerceIn(1f, MAX_ZOOM)
                    offset = nextOffset(scale, offset, pan)
                }
            }
            .graphicsLayer(scaleX = scale, scaleY = scale, translationX = offset.x, translationY = offset.y),
    )
}

/** The image offset after a drag: panning only makes sense while zoomed in. */
private fun nextOffset(scale: Float, offset: Offset, pan: Offset): Offset {
    if (scale > 1f) {
        return offset + pan
    }
    return Offset.Zero
}
