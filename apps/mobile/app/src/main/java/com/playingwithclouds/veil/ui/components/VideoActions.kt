package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.CollectionTile
import com.playingwithclouds.veil.data.HiddenRecords
import com.playingwithclouds.veil.data.LibraryRepository
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.VideoRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.launch

/** Credits offered to block in the sheet; more performers are blocked from their page, as on the web. */
private const val SHEET_PERFORMERS = 3

/**
 * A video card's long-press sheet, the web's card menu (`recordActions.ts`): play next, add to a
 * collection, favourite, not interested, and blocking the video, its first performers or its
 * studio. Not interested and the blocks take the card (and every card crediting what's blocked)
 * off the screen at once.
 */
@Composable
fun VideoActionsSheet(video: VideoSummary, onDismiss: () -> Unit) {
    var pickingCollection by remember { mutableStateOf(false) }
    var favorite by remember { mutableStateOf(video.favorite) }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Text(video.title, Modifier.gutterPadding().padding(bottom = VeilSpacing.small), style = MaterialTheme.typography.titleSmall, color = VeilColors.contentMuted, maxLines = 2)
        if (pickingCollection) {
            CollectionChoices(onPick = { collection ->
                inBackground { VideoRepository.addToCollection(collection.id, video.id) }
                onDismiss()
            })
            return@VeilBottomSheet
        }
        SheetAction("Play next", VeilIcons.Queue, onClick = {
            inBackground { VideoRepository.playNext(video.id) }
            onDismiss()
        })
        SheetAction("Add to collection…", VeilIcons.CollectionAdd, onClick = { pickingCollection = true })
        FavoriteAction(favorite, onClick = {
            favorite = !favorite
            val next = favorite
            inBackground { VideoRepository.setFavorite(video.id, next) }
        })
        SheetAction("Not interested", VeilIcons.NotInterested, onClick = {
            notInterested(video)
            onDismiss()
        })
        SheetAction("Block this video", VeilIcons.Block, onClick = {
            block(RecordKind.video, video.id)
            onDismiss()
        })
        for (performer in video.performers.take(SHEET_PERFORMERS)) {
            SheetAction("Block ${performer.name}", VeilIcons.Performer, onClick = {
                block(RecordKind.performer, performer.id)
                onDismiss()
            })
        }
        val studio = video.studio
        if (studio != null) {
            SheetAction("Block ${studio.name}", VeilIcons.Studios, onClick = {
                block(RecordKind.studio, studio.id)
                onDismiss()
            })
        }
    }
}

/** "Add to favourites", or "Remove from favourites" with a filled heart. */
@Composable
private fun FavoriteAction(favorite: Boolean, onClick: () -> Unit) {
    if (favorite) {
        SheetAction("Remove from favourites", VeilIcons.HeartFilled, onClick = onClick, active = true)
        return
    }
    SheetAction("Add to favourites", VeilIcons.Heart, onClick = onClick)
}

/** The user's collections to add the video to, loaded when the step opens. */
@Composable
private fun CollectionChoices(onPick: (CollectionTile) -> Unit) {
    var collections by remember { mutableStateOf<List<CollectionTile>?>(null) }
    LaunchedEffect(Unit) {
        collections = runCatching { LibraryRepository.collections() }.getOrDefault(emptyList())
    }
    val loaded = collections
    if (loaded == null) {
        Spinner(Modifier.gutterPadding().padding(vertical = VeilSpacing.large))
        return
    }
    if (loaded.isEmpty()) {
        EmptyMessage("No collections yet. Start one in the Library.")
        return
    }
    for (collection in loaded) {
        SheetAction(collection.name, VeilIcons.Collections, onClick = { onPick(collection) })
    }
}

/** Not interested: the recommendations learn it as a thumbs down (through the event log, so incognito keeps it), and the card leaves. */
private fun notInterested(video: VideoSummary) {
    AppServices.events.log(LoggedEvent(EventKind.like, RecordKind.video, video.id, mapOf("value" to -1)), flush = true)
    HiddenRecords.hide(RecordKind.video, video.id)
}

/** Blocks a record; it and every card crediting it leave the screen at once. */
private fun block(kind: RecordKind, id: String) {
    HiddenRecords.hide(kind, id)
    inBackground { VideoRepository.block(kind, id) }
}

/** Runs a write past the sheet and the card, which may leave the screen before it's done. */
private fun inBackground(write: suspend () -> Unit) {
    AppScope.launch { runCatching { write() } }
}
