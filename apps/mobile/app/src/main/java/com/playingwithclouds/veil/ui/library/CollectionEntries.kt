package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.CollectionEntry
import com.playingwithclouds.veil.data.CollectionPage
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType

/** Size of a gallery's cover in a collection's list. */
private val GalleryThumbWidth = 72.dp
private val GalleryThumbHeight = 96.dp

/**
 * A collection's videos as full-width cards and its galleries as rows, in its order. A long press
 * offers to take one out with [onRemove]; [emptyText] shows while it holds nothing.
 */
@Composable
fun CollectionEntries(page: CollectionPage, emptyText: String, navigator: AppNavigator, onRemove: (CollectionEntry) -> Unit) {
    var removing by remember { mutableStateOf<CollectionEntry?>(null) }
    if (page.entries.isEmpty()) {
        EmptyMessage(emptyText, Modifier.padding(top = LocalTopBarInset.current))
        return
    }
    LazyColumn(
        contentPadding = PaddingValues(top = VeilSpacing.small + LocalTopBarInset.current, bottom = VeilSpacing.large + LocalFloatingBarInset.current),
    ) {
        items(page.entries, key = { entry -> entry.id }) { entry ->
            when (entry) {
                is CollectionEntry.Video -> VideoCard(
                    entry.video,
                    onClick = { navigator.openVideo(entry.video.id) },
                    edgeToEdge = true,
                    onLongClick = { removing = entry },
                )
                is CollectionEntry.Gallery -> GalleryRow(
                    entry.gallery,
                    onClick = { navigator.openGallery(entry.gallery.id) },
                    onLongClick = { removing = entry },
                )
            }
        }
    }
    val entry = removing
    if (entry != null) {
        VeilBottomSheet(onDismissRequest = { removing = null }) {
            SheetAction("Remove from ${page.name}", VeilIcons.Delete, onClick = {
                removing = null
                onRemove(entry)
            })
        }
    }
}

/** A gallery in a list: its cover beside the title, who's in it and how many photos. */
@Composable
private fun GalleryRow(gallery: GallerySummary, onClick: () -> Unit, onLongClick: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick)
            .padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(gallery.coverUrl, Modifier.size(GalleryThumbWidth, GalleryThumbHeight).clip(VeilShapes.card))
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.hairline)) {
            Text(gallery.title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val facts = listOfNotNull(gallery.byline, "${gallery.imageCount} photos")
            Text(facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted, maxLines = 1)
        }
    }
}
