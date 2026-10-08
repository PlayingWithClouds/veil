package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.design.VeilIcons

/** A headed, horizontally scrolling row of scene cards; the heading opens the full list when [onSeeAll] is set. */
@Composable
fun SceneShelf(
    title: String,
    scenes: List<SceneSummary>,
    onOpen: (SceneSummary) -> Unit,
    modifier: Modifier = Modifier,
    onSeeAll: (() -> Unit)? = null,
) {
    if (scenes.isEmpty()) {
        return
    }
    Column(modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        ShelfHeading(title, onSeeAll)
        LazyRow(contentPadding = PaddingValues(horizontal = 4.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            items(scenes, key = { scene -> scene.id }) { scene ->
                SceneCard(scene, onClick = { onOpen(scene) }, modifier = Modifier.width(240.dp))
            }
        }
    }
}

/** A section heading, with a chevron when it leads somewhere. */
@Composable
fun ShelfHeading(title: String, onClick: (() -> Unit)? = null) {
    var rowModifier = Modifier.padding(horizontal = 4.dp, vertical = 4.dp)
    if (onClick != null) {
        rowModifier = Modifier.clickable(onClick = onClick).then(rowModifier)
    }
    Row(rowModifier, verticalAlignment = Alignment.CenterVertically) {
        Text(title, style = MaterialTheme.typography.titleLarge)
        if (onClick != null) {
            Icon(VeilIcons.Chevron, contentDescription = null, modifier = Modifier.padding(start = 4.dp).size(18.dp))
        }
    }
}
