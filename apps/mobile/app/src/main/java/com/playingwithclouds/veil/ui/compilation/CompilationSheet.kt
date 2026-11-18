package com.playingwithclouds.veil.ui.compilation

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
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.data.PlayQueue
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.scenes.CompilationOptions
import com.playingwithclouds.veil.scenes.CompilationOrder
import com.playingwithclouds.veil.scenes.SceneLabels
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SegmentedControl
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * What a compilation over the queue (the video playing, then the queue) plays: the liked scenes or
 * labels picked for this one (chips grouped as the model's heads, the liked ones picked to start),
 * queue order or shuffled, and the shortest clip. [onStart] gets the choice; the liked scenes
 * themselves don't change.
 */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun CompilationSheet(onDismiss: () -> Unit, onStart: (CompilationOptions) -> Unit) {
    val modelState by SceneModelStore.state.collectAsState()
    val liked by SceneRepository.likedTags.collectAsState()
    LaunchedEffect(Unit) { SceneRepository.ensureLikedTags() }
    val groups = remember(modelState) { SceneLabels.pickerGroups(modelState.ready?.manifest) }
    var picked by remember(liked) { mutableStateOf(liked) }
    var order by remember { mutableStateOf(CompilationOrder.QUEUE) }
    var minClipSeconds by remember { mutableStateOf(CompilationOptions.MIN_CLIP_CHOICES.first()) }

    /** Flips one label for this compilation. */
    fun toggle(tag: String) {
        if (tag in picked) {
            picked = picked - tag
            return
        }
        picked = picked + tag
    }

    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(
            Modifier.verticalScroll(rememberScrollState()).gutterPadding().padding(bottom = VeilSpacing.huge),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            Text("Compilation of the queue", style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
                FeedChip("Liked scenes", onClick = { picked = liked }, selected = picked == liked && liked.isNotEmpty())
                FeedChip("None", onClick = { picked = emptySet() })
            }
            SegmentedControl(CompilationOrder.entries.map { entry -> entry.label }, order.ordinal, onSelect = { index -> order = CompilationOrder.entries[index] })
            SectionHeading("Shortest clip")
            Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                for (seconds in CompilationOptions.MIN_CLIP_CHOICES) {
                    FeedChip("${seconds.toInt()} s", onClick = { minClipSeconds = seconds }, selected = minClipSeconds == seconds)
                }
            }
            PrimaryButton(
                "Play",
                onClick = { onStart(CompilationOptions(tagsFor(picked, liked), order, minClipSeconds)) },
                modifier = Modifier.fillMaxWidth(),
                enabled = picked.isNotEmpty(),
            )
            for (group in groups) {
                SectionHeading(group.title)
                FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                    for (tag in group.tags) {
                        FeedChip(tag, onClick = { toggle(tag) }, selected = tag in picked)
                    }
                }
            }
        }
    }
}

/**
 * Home's way into a compilation of the queue: one chip, shown with a scene model and something
 * queued, opening [CompilationSheet].
 */
@Composable
fun QueueCompilationButton(navigator: AppNavigator, modifier: Modifier = Modifier) {
    val modelState by SceneModelStore.state.collectAsState()
    val queue by PlayQueue.contents.collectAsState()
    var choosing by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) { runCatching { PlayQueue.refresh() } }
    if (modelState.ready == null || queue.videos.isEmpty()) {
        return
    }
    if (choosing) {
        CompilationSheet(
            onDismiss = { choosing = false },
            onStart = { options ->
                choosing = false
                navigator.openQueueCompilation("", options)
            },
        )
    }
    Row(modifier) {
        FeedChip("Compilation of your queue · ${queue.videos.size}", onClick = { choosing = true }, icon = VeilIcons.Scenes)
    }
}

/** The tags a compilation is asked for: none (the liked scenes, read again when it starts) when [picked] is the liked set. */
private fun tagsFor(picked: Set<String>, liked: Set<String>): Set<String> {
    if (picked == liked) {
        return emptySet()
    }
    return picked
}
