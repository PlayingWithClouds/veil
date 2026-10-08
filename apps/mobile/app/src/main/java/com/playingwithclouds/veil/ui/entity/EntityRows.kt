package com.playingwithclouds.veil.ui.entity

import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.StudioSummary
import com.playingwithclouds.veil.ui.components.CollectionCard
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.PerformerCard
import com.playingwithclouds.veil.ui.components.StudioCard
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.theme.VeilSpacing

// The rows below sit inside a grid that pads with the gutter: headings take that padding, the
// scrolling rows bleed past it to the screen edges.

/** A titled horizontal row of performers; nothing when the list is empty. */
@Composable
fun PerformerRow(title: String, performers: List<PerformerSummary>, onOpen: (String) -> Unit) {
    if (performers.isEmpty()) {
        return
    }
    SectionHeading(title)
    LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        items(performers, key = { performer -> performer.id }) { performer ->
            PerformerCard(performer, onClick = { onOpen(performer.id) }, modifier = Modifier.width(112.dp))
        }
    }
}

/** A titled horizontal row of studios; nothing when the list is empty. */
@Composable
fun StudioRow(title: String, studios: List<StudioSummary>, onOpen: (String) -> Unit) {
    if (studios.isEmpty()) {
        return
    }
    SectionHeading(title)
    LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        items(studios, key = { studio -> studio.id }) { studio ->
            StudioCard(studio, onClick = { onOpen(studio.id) }, modifier = Modifier.width(112.dp))
        }
    }
}

/** A titled horizontal row of galleries; nothing when the list is empty. */
@Composable
fun GalleryRow(title: String, galleries: List<GallerySummary>, onOpen: (String) -> Unit) {
    if (galleries.isEmpty()) {
        return
    }
    SectionHeading(title)
    LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap)) {
        items(galleries, key = { gallery -> gallery.id }) { gallery ->
            GalleryCard(gallery, onClick = { onOpen(gallery.id) }, modifier = Modifier.width(140.dp))
        }
    }
}

/** A titled horizontal row of collections; nothing when the list is empty. */
@Composable
fun CollectionRow(title: String, collections: List<CollectionSummary>, onOpen: (String) -> Unit) {
    if (collections.isEmpty()) {
        return
    }
    SectionHeading(title)
    LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap)) {
        items(collections, key = { collection -> collection.id }) { collection ->
            CollectionCard(collection, onClick = { onOpen(collection.id) }, modifier = Modifier.width(220.dp))
        }
    }
}
