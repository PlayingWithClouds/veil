package com.playingwithclouds.veil.ui.scenes

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.scenes.SceneLabels
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.launch

/**
 * The acts, positions and kinks the user likes, grouped as the model's heads (Oral, Acts,
 * Positions, …), each a chip that toggles; saved to the backend as each chip changes, so the web
 * and the shorts feed use them too.
 */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun LikedScenesSheet(onDismiss: () -> Unit) {
    val scope = rememberCoroutineScope()
    val modelState by SceneModelStore.state.collectAsState()
    val saved by SceneRepository.likedTags.collectAsState()
    var liked by remember { mutableStateOf(saved) }
    val groups = remember(modelState) { SceneLabels.pickerGroups(modelState.ready?.manifest) }

    /** Flips one tag and saves the choice. */
    fun toggle(tag: String) {
        var next = liked + tag
        if (tag in liked) {
            next = liked - tag
        }
        liked = next
        scope.launch { runCatching { SceneRepository.setLikedTags(next) } }
    }

    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(
            Modifier.verticalScroll(rememberScrollState()).gutterPadding().padding(bottom = VeilSpacing.huge),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Text("Scenes you like", Modifier.weight(1f), style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
                TextAction("Done", onClick = onDismiss)
            }
            for (group in groups) {
                SectionHeading(group.title)
                FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                    for (tag in group.tags) {
                        FeedChip(tag, onClick = { toggle(tag) }, selected = tag in liked)
                    }
                }
            }
        }
    }
}
