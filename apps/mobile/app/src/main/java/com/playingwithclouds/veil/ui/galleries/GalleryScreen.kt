package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.GalleryDetail
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.CollectionPickerDialog
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.ImageViewer
import com.playingwithclouds.veil.ui.components.ReactionButton
import com.playingwithclouds.veil.ui.components.ReactionRowInset
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.dislikeIcon
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.components.likeIcon
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.face.FaceIdentifySheet
import com.playingwithclouds.veil.ui.face.FaceRequest
import com.playingwithclouds.veil.ui.scene.AlternateRow
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * A gallery's page: reactions and the set's video on top, then the images as a grid that opens
 * in a zoomable viewer, where "Who is this?" asks about the image on screen.
 */
@Composable
fun GalleryScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "gallery-$id") { GalleryViewModel(id) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var viewerIndex by remember { mutableStateOf<Int?>(null) }
    var faceRequest by remember { mutableStateOf<FaceRequest?>(null) }
    var showingCollections by remember { mutableStateOf(false) }
    var showingVideoSearch by remember { mutableStateOf(false) }
    val snackbarHostState = remember { SnackbarHostState() }
    val title = (state.gallery as? LoadState.Loaded)?.value?.title.orEmpty()

    LaunchedEffect(state.message) {
        val note = state.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    if (showingCollections) {
        val collections by viewModel.userCollections.collectAsStateWithLifecycle()
        CollectionPickerDialog(
            collections,
            state.collectionIds,
            onToggle = viewModel::setInCollection,
            onCreate = viewModel::createCollectionWithGallery,
            onDismiss = { showingCollections = false },
        )
    }
    if (showingVideoSearch) {
        VideoSearchSheet(
            viewModel,
            onOpenScene = { sceneId ->
                showingVideoSearch = false
                navigator.openScene(sceneId)
            },
            onDismiss = { showingVideoSearch = false },
        )
    }
    val request = faceRequest
    if (request != null) {
        FaceIdentifySheet(request, navigator, onDismiss = { faceRequest = null })
    }

    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }, snackbarHost = { VeilSnackbarHost(snackbarHostState) }) { padding ->
        LoadStateContent(state.gallery, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { detail ->
            val viewing = viewerIndex
            if (viewing != null) {
                ImageViewer(
                    detail.images.map { image -> image.filePath },
                    viewing,
                    onDismiss = { viewerIndex = null },
                    onIdentify = { index -> faceRequest = FaceRequest(detail.id, imageId = detail.images[index].id) },
                )
            }
            ImageGrid(
                detail,
                state,
                navigator,
                Modifier.padding(padding),
                actions = GalleryActions(
                    onReact = viewModel::react,
                    onIncrementOCount = viewModel::incrementOCount,
                    onDecrementOCount = viewModel::decrementOCount,
                    onToggleFavorite = viewModel::toggleFavorite,
                    onSave = {
                        viewModel.loadUserCollections()
                        showingCollections = true
                    },
                    onFindVideo = {
                        viewModel.findVideo()
                        showingVideoSearch = true
                    },
                    onIdentify = { faceRequest = FaceRequest(detail.id) },
                ),
                onOpenImage = { index -> viewerIndex = index },
            )
        }
    }
}

/** What the gallery's reaction row can do. */
private class GalleryActions(
    val onReact: (Verdict) -> Unit,
    val onIncrementOCount: () -> Unit,
    val onDecrementOCount: () -> Unit,
    val onToggleFavorite: () -> Unit,
    val onSave: () -> Unit,
    val onFindVideo: () -> Unit,
    val onIdentify: () -> Unit,
)

/** The reactions, tags and description, then the images. */
@Composable
private fun ImageGrid(
    detail: GalleryDetail,
    state: GalleryState,
    navigator: AppNavigator,
    modifier: Modifier,
    actions: GalleryActions,
    onOpenImage: (Int) -> Unit,
) {
    LazyVerticalGrid(
        columns = GridCells.Adaptive(110.dp),
        modifier = modifier,
        contentPadding = PaddingValues(VeilSpacing.gutter),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        fullWidthItem("reactions") { ReactionRow(detail, state, navigator, actions) }
        fullWidthItem("about") {
            if (!detail.details.isNullOrBlank()) {
                Text(
                    detail.details,
                    Modifier.padding(vertical = VeilSpacing.small),
                    style = MaterialTheme.typography.bodyMedium,
                    color = VeilColors.contentMuted,
                )
            }
            TagChips(detail.tags, onClick = { tag -> navigator.openTag(tag.id) }, modifier = Modifier.bleed())
        }
        if (detail.images.isEmpty()) {
            fullWidthItem("empty") { NoImages(state.fetchingImages) }
        }
        itemsIndexed(detail.images, key = { _, image -> image.id }) { index, image ->
            RemoteImage(
                image.filePath,
                Modifier.fillMaxWidth().aspectRatio(1f).pressClickable { onOpenImage(index) }.clip(VeilShapes.small),
            )
        }
    }
}

/** Like, dislike, O-counter, favorite, save, the set's video and "Who is this?", running edge to edge. */
@Composable
private fun ReactionRow(detail: GalleryDetail, state: GalleryState, navigator: AppNavigator, actions: GalleryActions) {
    LazyRow(
        Modifier.bleed(),
        contentPadding = PaddingValues(horizontal = ReactionRowInset),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.hairline),
    ) {
        item {
            ReactionButton(likeIcon(state.verdict == Verdict.UP), "Like", active = state.verdict == Verdict.UP) { actions.onReact(Verdict.UP) }
        }
        item {
            ReactionButton(dislikeIcon(state.verdict == Verdict.DOWN), "Dislike", active = state.verdict == Verdict.DOWN) { actions.onReact(Verdict.DOWN) }
        }
        item {
            ReactionButton(VeilIcons.HeartFilled, "O ${state.oCount}", onLongClick = actions.onDecrementOCount) { actions.onIncrementOCount() }
        }
        item { ReactionButton(favoriteIcon(state.isFavorite), "Favorite", active = state.isFavorite) { actions.onToggleFavorite() } }
        item { ReactionButton(VeilIcons.CollectionAdd, "Save") { actions.onSave() } }
        val video = detail.video
        if (video != null) {
            item { ReactionButton(VeilIcons.Play, "Watch video") { navigator.openScene(video.id) } }
        }
        item { ReactionButton(VeilIcons.FindVideo, "Find video") { actions.onFindVideo() } }
        item { ReactionButton(VeilIcons.Identify, "Who is this?") { actions.onIdentify() } }
    }
}

/** The favorite icon, filled while the gallery is a favorite. */
private fun favoriteIcon(favorite: Boolean): ImageVector {
    if (favorite) {
        return VeilIcons.HeartFilled
    }
    return VeilIcons.Heart
}

/** Shown while a gallery has no images: fetching them from the site, or that it failed. */
@Composable
private fun NoImages(fetching: Boolean) {
    if (fetching) {
        Spinner(Modifier.padding(VeilSpacing.extraLarge))
        return
    }
    EmptyMessage("This gallery has no images yet.")
}

/** Videos the set may come from, best match first; picking one stores it as the set's video and opens it. */
@Composable
private fun VideoSearchSheet(viewModel: GalleryViewModel, onOpenScene: (String) -> Unit, onDismiss: () -> Unit) {
    val search by viewModel.videoSearch.collectAsStateWithLifecycle()
    VeilBottomSheet(onDismissRequest = onDismiss) {
        SectionHeading("Find the video", Modifier.gutterPadding())
        if (search.isLoading) {
            Text("Searching every site for this set's video...", Modifier.gutterPadding(), color = VeilColors.contentMuted)
            Spinner(Modifier.padding(VeilSpacing.extraLarge).align(Alignment.CenterHorizontally))
            return@VeilBottomSheet
        }
        if (search.candidates.isEmpty()) {
            EmptyMessage("No matching videos found.")
            return@VeilBottomSheet
        }
        LazyColumn {
            items(search.candidates, key = { candidate -> candidate.sourceUrl }) { candidate ->
                AlternateRow(candidate, attaching = search.linkingUrl == candidate.sourceUrl) {
                    viewModel.linkVideo(candidate, onOpenScene)
                }
            }
        }
    }
}
