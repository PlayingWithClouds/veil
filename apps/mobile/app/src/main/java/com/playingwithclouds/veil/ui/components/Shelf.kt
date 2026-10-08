package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Width of a scene card in a shelf; the next card peeks in at the edge. */
private val ShelfCardWidth = 240.dp

/**
 * A headed, horizontally scrolling row of scene cards, placed in a grid that pads with the gutter:
 * the row runs to the screen edges while its first card lines up with the heading. The heading
 * opens the full list when [onSeeAll] is set.
 */
@Composable
fun SceneShelf(
    title: String,
    scenes: List<SceneSummary>,
    onOpen: (SceneSummary) -> Unit,
    modifier: Modifier = Modifier,
    onSeeAll: (() -> Unit)? = null,
    onLongPress: ((SceneSummary) -> Unit)? = null,
) {
    if (scenes.isEmpty()) {
        return
    }
    Column(modifier.fillMaxWidth()) {
        SectionHeading(title, onClick = onSeeAll)
        LazyRow(
            Modifier.bleed(),
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        ) {
            items(scenes, key = { scene -> scene.id }) { scene ->
                var longPress: (() -> Unit)? = null
                if (onLongPress != null) {
                    longPress = { onLongPress(scene) }
                }
                SceneCard(scene, onClick = { onOpen(scene) }, modifier = Modifier.width(ShelfCardWidth), onLongClick = longPress)
            }
        }
    }
}
