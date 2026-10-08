package com.playingwithclouds.veil.ui.collections

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CollectionContents
import com.playingwithclouds.veil.data.CollectionMember
import com.playingwithclouds.veil.data.CollectionRepository
import com.playingwithclouds.veil.data.MemberType
import com.playingwithclouds.veil.data.PlaybackQueue
import com.playingwithclouds.veil.data.QueueEntry
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.ConfirmDialog
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.ImageViewer
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** One collection with its members. */
class CollectionViewModel(private val collectionId: String) : ViewModel() {

    private val mutableContents = MutableStateFlow<LoadState<CollectionContents>>(LoadState.Loading)

    /** The collection and its members in manual order. */
    val contents: StateFlow<LoadState<CollectionContents>> = mutableContents

    init {
        load()
    }

    /** Loads the collection. */
    fun load() {
        viewModelScope.loadInto(mutableContents) {
            requireNotNull(CollectionRepository.contents(collectionId)) { "Collection not found" }
        }
    }

    /** Renames the collection. */
    fun rename(name: String) {
        viewModelScope.launch {
            runCatching { CollectionRepository.rename(collectionId, name) }
            load()
        }
    }

    /** Deletes the collection and leaves the page. */
    fun delete(onDeleted: () -> Unit) {
        viewModelScope.launch {
            val result = runCatching { CollectionRepository.delete(collectionId) }
            if (result.isSuccess) {
                onDeleted()
            }
        }
    }

    /** Takes a member out of the collection. */
    fun removeMember(member: CollectionMember) {
        viewModelScope.launch {
            runCatching { CollectionRepository.remove(collectionId, member.mediaId) }
            load()
        }
    }
}

/** A collection's page: its members, with rename and delete for the user's own. */
@Composable
fun CollectionScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "collection-$id") { CollectionViewModel(id) }
    val contents by viewModel.contents.collectAsStateWithLifecycle()
    var renaming by remember { mutableStateOf(false) }
    var deleting by remember { mutableStateOf(false) }
    val loaded = (contents as? LoadState.Loaded)?.value

    if (renaming && loaded != null) {
        TextInputDialog(
            title = "Rename collection",
            label = "Name",
            confirmLabel = "Save",
            initialValue = loaded.collection.name,
            onConfirm = { name ->
                renaming = false
                viewModel.rename(name)
            },
            onDismiss = { renaming = false },
        )
    }
    if (deleting) {
        ConfirmDialog(
            title = "Delete collection?",
            message = "The collection is removed; its members stay in the library.",
            confirmLabel = "Delete",
            onConfirm = {
                deleting = false
                viewModel.delete(navigator::back)
            },
            onDismiss = { deleting = false },
        )
    }

    Scaffold(
        topBar = {
            VeilTopBar(
                title = loaded?.collection?.name.orEmpty(),
                onBack = navigator::back,
                actions = {
                    if (loaded != null && loaded.members.any { member -> member.type == MemberType.SCENE }) {
                        RoundIconButton(
                            VeilIcons.Play,
                            contentDescription = "Play all",
                            onClick = { playCollection(loaded, shuffled = false, navigator) },
                        )
                        RoundIconButton(
                            VeilIcons.Random,
                            contentDescription = "Shuffle play",
                            onClick = { playCollection(loaded, shuffled = true, navigator) },
                        )
                    }
                    if (loaded != null && loaded.collection.isUserCreated) {
                        RoundIconButton(VeilIcons.Edit, contentDescription = "Rename", onClick = { renaming = true })
                        RoundIconButton(VeilIcons.Delete, contentDescription = "Delete", onClick = { deleting = true })
                    }
                },
            )
        },
    ) { padding ->
        LoadStateContent(contents, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { collection ->
            MemberGrid(collection, viewModel, navigator, Modifier.padding(padding))
        }
    }
}

/** The members as cards; long-pressing one of the user's own collections asks to remove it. */
@Composable
private fun MemberGrid(
    collection: CollectionContents,
    viewModel: CollectionViewModel,
    navigator: AppNavigator,
    modifier: Modifier,
) {
    var viewingImage by remember { mutableStateOf<Int?>(null) }
    var removing by remember { mutableStateOf<CollectionMember?>(null) }
    val imagePaths = collection.members.filter { member -> member.type == MemberType.IMAGE }.map { member -> member.posterPath.orEmpty() }

    val viewing = viewingImage
    if (viewing != null) {
        ImageViewer(imagePaths, viewing, onDismiss = { viewingImage = null })
    }
    val pendingRemoval = removing
    if (pendingRemoval != null) {
        ConfirmDialog(
            title = "Remove from collection?",
            message = pendingRemoval.title,
            confirmLabel = "Remove",
            onConfirm = {
                removing = null
                viewModel.removeMember(pendingRemoval)
            },
            onDismiss = { removing = null },
        )
    }

    if (collection.members.isEmpty()) {
        EmptyMessage("This collection is empty.", modifier)
        return
    }
    LazyVerticalGrid(
        columns = GridCells.Adaptive(160.dp),
        modifier = modifier,
        contentPadding = PaddingValues(VeilSpacing.gutter),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
    ) {
        items(collection.members, key = { member -> "${member.type}:${member.mediaId}" }) { member ->
            MemberCard(
                member = member,
                onClick = {
                    if (member.type == MemberType.IMAGE) {
                        viewingImage = imagePaths.indexOf(member.posterPath.orEmpty())
                        return@MemberCard
                    }
                    openMember(member, navigator)
                },
                onLongClick = {
                    if (collection.collection.isUserCreated) {
                        removing = member
                    }
                },
            )
        }
    }
}

/** Plays the collection's scenes one after the other through the queue, in order or shuffled. */
private fun playCollection(collection: CollectionContents, shuffled: Boolean, navigator: AppNavigator) {
    val scenes = collection.members
        .filter { member -> member.type == MemberType.SCENE }
        .map { member -> QueueEntry(member.mediaId, member.title, member.posterPath) }
    val first = PlaybackQueue.startPlaylist(scenes, shuffled) ?: return
    navigator.openScene(first.sceneId)
}

/** Opens the page of a member. */
private fun openMember(member: CollectionMember, navigator: AppNavigator) {
    when (member.type) {
        MemberType.SCENE -> navigator.openScene(member.mediaId)
        MemberType.GALLERY -> navigator.openGallery(member.mediaId)
        MemberType.PERFORMER -> navigator.openPerformer(member.mediaId)
        MemberType.STUDIO -> navigator.openStudio(member.mediaId)
        MemberType.IMAGE -> Unit
    }
}

/** A member's poster and title, labeled with its type. */
@Composable
private fun MemberCard(member: CollectionMember, onClick: () -> Unit, onLongClick: () -> Unit) {
    Column(
        Modifier.fillMaxWidth().clip(VeilShapes.card).pressClickable(enabled = true, onLongClick = onLongClick, onClick = onClick),
    ) {
        Box(Modifier.fillMaxWidth().aspectRatio(4f / 3f).clip(VeilShapes.card)) {
            RemoteImage(member.posterPath, Modifier.matchParentSize())
            Badge(member.type.name.lowercase(), Modifier.align(Alignment.TopStart).padding(VeilSpacing.small))
        }
        Text(
            member.title,
            modifier = Modifier.padding(horizontal = VeilSpacing.extraSmall, vertical = VeilSpacing.small),
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
    }
}
