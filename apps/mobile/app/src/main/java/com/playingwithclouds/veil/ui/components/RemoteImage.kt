package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import coil3.compose.AsyncImage
import com.playingwithclouds.veil.api.ServerSettings
import com.playingwithclouds.veil.ui.theme.VeilColors

/**
 * An image from the backend or a source site. Remote URLs go through the backend's image cache
 * (hotlink-protected CDNs refuse the app directly); loopback blob URLs are rebased onto the
 * server in use.
 */
@Composable
fun RemoteImage(
    path: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
    contentDescription: String? = null,
) {
    val url = ServerSettings.imageUrl(path)
    if (url == null) {
        Box(modifier.background(VeilColors.surfaceHigh))
        return
    }
    AsyncImage(
        model = url,
        contentDescription = contentDescription,
        modifier = modifier.background(VeilColors.surfaceHigh),
        contentScale = contentScale,
    )
}
