package com.playingwithclouds.veil.ui.collections

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
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
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Badge
import com.playingwithclouds.veil.ui.components.ConfirmDialog
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.ImageViewer
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.loadInto
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
                    if (loaded != null && loaded.collection.isUserCreated) {
                        IconButton(onClick = { renaming = true }) { Icon(Icons.Filled.Edit, contentDescription = "Rename") }
                        IconButton(onClick = { deleting = true }) { Icon(Icons.Filled.Delete, contentDescription = "Delete") }
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
@OptIn(ExperimentalFoundationApi::class)
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
        contentPadding = androidx.compose.foundation.layout.PaddingValues(12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
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
@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun MemberCard(member: CollectionMember, onClick: () -> Unit, onLongClick: () -> Unit) {
    Column(
        Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).combinedClickable(onClick = onClick, onLongClick = onLongClick),
    ) {
        Box(Modifier.fillMaxWidth().aspectRatio(4f / 3f).clip(RoundedCornerShape(12.dp))) {
            RemoteImage(member.posterPath, Modifier.matchParentSize())
            Badge(member.type.name.lowercase(), Modifier.align(Alignment.TopStart).padding(6.dp))
        }
        Text(
            member.title,
            modifier = Modifier.padding(4.dp),
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
    }
}
