package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Width of a scene card in a shelf; the next card peeks in at the edge. */
private val ShelfCardWidth = 240.dp

/**
 * A headed, horizontally scrolling row of scene cards. In a grid that pads with the gutter the row
 * bleeds to the screen edges while its first card lines up with the heading; in a list without
 * side padding set [edgeToEdge] so the heading takes the gutter itself instead. The heading opens
 * the full list when [onSeeAll] is set. With an [impressionSurface] the shelf logs which cards were
 * shown and opened (position = index in the row), with the candidate [sources] of the scenes.
 */
@Composable
fun SceneShelf(
    title: String,
    scenes: List<SceneSummary>,
    onOpen: (SceneSummary) -> Unit,
    modifier: Modifier = Modifier,
    onSeeAll: (() -> Unit)? = null,
    onLongPress: ((SceneSummary) -> Unit)? = null,
    edgeToEdge: Boolean = false,
    impressionSurface: String? = null,
    sources: Map<String, String> = emptyMap(),
) {
    if (scenes.isEmpty()) {
        return
    }
    val rowState = rememberLazyListState()
    if (impressionSurface != null) {
        // Several shelves share a surface, so one starting must not make the others log again.
        LogShownScenes(rowState, impressionSurface, scenes.map { scene -> scene.id }, sources, forgetShown = false)
    }
    var headingModifier = Modifier.gutterPadding()
    var rowModifier = Modifier.fillMaxWidth()
    if (!edgeToEdge) {
        headingModifier = Modifier
        rowModifier = Modifier.bleed()
    }
    Column(modifier.fillMaxWidth()) {
        SectionHeading(title, modifier = headingModifier, onClick = onSeeAll)
        LazyRow(
            rowModifier,
            state = rowState,
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        ) {
            itemsIndexed(scenes, key = { _, scene -> scene.id }) { position, scene ->
                var longPress: (() -> Unit)? = null
                if (onLongPress != null) {
                    longPress = { onLongPress(scene) }
                }
                val open = {
                    if (impressionSurface != null) {
                        logSceneClicked(impressionSurface, scene.id, position, sources[scene.id])
                    }
                    onOpen(scene)
                }
                SceneCard(scene, onClick = open, modifier = Modifier.width(ShelfCardWidth), onLongClick = longPress)
            }
        }
    }
}

/**
 * A headed, horizontally scrolling row of round photos with names (performers or studios) for a
 * list without side padding; tapping one calls [onOpen].
 */
@Composable
fun AvatarShelf(
    title: String,
    entities: List<EntityRef>,
    onOpen: (EntityRef) -> Unit,
    modifier: Modifier = Modifier,
) {
    if (entities.isEmpty()) {
        return
    }
    Column(modifier.fillMaxWidth()) {
        SectionHeading(title, modifier = Modifier.gutterPadding())
        LazyRow(
            Modifier.fillMaxWidth(),
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        ) {
            items(entities, key = { entity -> entity.id }) { entity ->
                AvatarCard(
                    name = entity.name,
                    imagePath = entity.imagePath,
                    subtitle = null,
                    onClick = { onOpen(entity) },
                    modifier = Modifier.width(AvatarCellWidth),
                )
            }
        }
    }
}
