package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
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
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CollectionPage
import com.playingwithclouds.veil.data.CollectionTile
import kotlinx.coroutines.flow.StateFlow
import com.playingwithclouds.veil.data.DownloadItem
import com.playingwithclouds.veil.graphql.type.JobStatus
import androidx.compose.ui.graphics.Color
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.GlassTopBarLayout
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TabTopBar
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.fadingEdges
import androidx.compose.foundation.lazy.LazyRow
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatBytes
import kotlinx.coroutines.delay

/** Columns of the collections grid. */
private const val COLLECTION_COLUMNS = 2

/** How often running downloads are read again while the tab shows them. */
private const val DOWNLOAD_POLL_MILLISECONDS = 3000L

/** Size of a download's thumbnail in its row. */
private val DownloadThumbWidth = 128.dp

private const val LANDSCAPE_RATIO = 16f / 9f

/**
 * The Library tab, as the web's library page: the watchlist, the queue, the user's collections and
 * the downloads, one segment each.
 */
@Composable
fun LibraryScreen(navigator: AppNavigator) {
    val viewModel = viewModel { LibraryViewModel() }
    val tab by viewModel.tab.collectAsStateWithLifecycle()
    val message by viewModel.message.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(message) {
        val note = message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    Scaffold(snackbarHost = { VeilSnackbarHost(snackbarHostState) }) { padding ->
        GlassTopBarLayout(
            bar = {
                TabTopBar(navigator, "Library")
                // Six tabs don't fit a segmented control on a phone; they scroll as chips.
                LazyRow(
                    Modifier.fillMaxWidth().fadingEdges().padding(bottom = VeilSpacing.small),
                    contentPadding = GutterRowPadding,
                    horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
                ) {
                    items(LibraryTab.entries, key = { entry -> entry.name }) { entry ->
                        FeedChip(entry.label, onClick = { viewModel.select(entry) }, selected = entry == tab)
                    }
                }
            },
            modifier = Modifier.padding(padding),
        ) {
            when (tab) {
                LibraryTab.WATCHLIST -> EntriesTab(viewModel.watchlist, "Nothing on your watchlist. Add videos from their page.", viewModel, navigator)
                LibraryTab.LIKED -> EntriesTab(viewModel.liked, "No liked videos yet. Give one a thumbs up on its page.", viewModel, navigator)
                LibraryTab.FAVORITES -> EntriesTab(viewModel.favorites, "No favourites yet. Tap the heart on a video or gallery.", viewModel, navigator)
                LibraryTab.QUEUE -> EntriesTab(viewModel.queue, "Nothing queued up next.", viewModel, navigator)
                LibraryTab.COLLECTIONS -> CollectionsTab(viewModel, navigator)
                LibraryTab.DOWNLOADS -> DownloadsTab(viewModel, navigator)
                LibraryTab.BLOCKED -> BlockedTab(viewModel, navigator)
            }
        }
    }
}

/** The watchlist, the favourites or the queue; a long press takes an entry out. */
@Composable
private fun EntriesTab(flow: StateFlow<LoadState<CollectionPage>>, emptyText: String, viewModel: LibraryViewModel, navigator: AppNavigator) {
    val state by flow.collectAsStateWithLifecycle()
    LoadStateContent(state, onRetry = viewModel::reload) { page ->
        CollectionEntries(page, emptyText, navigator, onRemove = { entry -> viewModel.remove(page, entry) })
    }
}

/** The user's collections as tiles, with a button to start a new one. */
@Composable
private fun CollectionsTab(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val state by viewModel.collections.collectAsStateWithLifecycle()
    var creating by remember { mutableStateOf(false) }
    if (creating) {
        TextInputDialog(
            title = "New collection",
            label = "Name",
            confirmLabel = "Create",
            onConfirm = { name ->
                creating = false
                if (name.isNotBlank()) {
                    viewModel.createCollection(name)
                }
            },
            onDismiss = { creating = false },
        )
    }
    LoadStateContent(state, onRetry = viewModel::reload) { collections ->
        LazyVerticalGrid(
            GridCells.Fixed(COLLECTION_COLUMNS),
            contentPadding = PaddingValues(
                start = VeilSpacing.gutter,
                top = LocalTopBarInset.current,
                end = VeilSpacing.gutter,
                bottom = VeilSpacing.large + LocalFloatingBarInset.current,
            ),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        ) {
            item(span = { GridItemSpan(maxLineSpan) }) {
                Row { SecondaryButton("New collection", onClick = { creating = true }, icon = VeilIcons.CollectionAdd) }
            }
            items(collections, key = { collection -> collection.id }) { collection ->
                CollectionTileCard(collection, onClick = { navigator.openCollection(collection.id) })
            }
        }
    }
}

/** A collection as a tile: its first cover, its name and how much it holds. */
@Composable
private fun CollectionTileCard(collection: CollectionTile, onClick: () -> Unit) {
    Column(Modifier.fillMaxWidth().pressClickable(onClick)) {
        RemoteImage(collection.coverUrls.firstOrNull(), Modifier.fillMaxWidth().aspectRatio(LANDSCAPE_RATIO).clip(VeilShapes.card))
        Text(
            collection.name,
            Modifier.padding(top = VeilSpacing.small),
            style = VeilType.cardTitle,
            color = VeilColors.content,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text("${collection.itemCount} items", style = VeilType.cardMeta, color = VeilColors.contentMuted)
    }
}

/** The downloads, newest first, read again every few seconds while one is still going. */
@Composable
private fun DownloadsTab(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val state by viewModel.downloads.collectAsStateWithLifecycle()
    val loaded = (state as? LoadState.Loaded)?.value.orEmpty()
    val going = loaded.any { item -> item.status == JobStatus.queued || item.status == JobStatus.running }
    LaunchedEffect(going) {
        while (going) {
            delay(DOWNLOAD_POLL_MILLISECONDS)
            viewModel.refreshDownloads()
        }
    }
    LoadStateContent(state, onRetry = viewModel::reload) { downloads ->
        if (downloads.isEmpty()) {
            EmptyMessage("No downloads. Download a video from its page to keep it.", Modifier.padding(top = LocalTopBarInset.current))
            return@LoadStateContent
        }
        LazyColumn(contentPadding = PaddingValues(top = LocalTopBarInset.current, bottom = VeilSpacing.large + LocalFloatingBarInset.current)) {
            items(downloads, key = { item -> item.jobId }) { item ->
                DownloadRow(item, onOpen = { navigator.openVideo(item.video.id) }, onDelete = { viewModel.deleteDownload(item) })
            }
        }
    }
}

/** One download: the video's cover and title, how it's going, and a delete button. */
@Composable
private fun DownloadRow(item: DownloadItem, onOpen: () -> Unit, onDelete: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(onOpen).padding(start = VeilSpacing.gutter, end = VeilSpacing.small, top = VeilSpacing.small, bottom = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(item.video.coverUrl, Modifier.size(DownloadThumbWidth, DownloadThumbWidth / LANDSCAPE_RATIO).clip(VeilShapes.card))
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
            Text(item.video.title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
            Text(downloadStatus(item), style = MaterialTheme.typography.bodySmall, color = statusColor(item), maxLines = 1)
            if (item.status == JobStatus.running) {
                ProgressBar(item.progress.toFloat())
            }
        }
        IconTap(VeilIcons.Delete, contentDescription = "Delete download", onClick = onDelete)
    }
}

/** What a download is doing: "Queued", "42 %", "Done · 1.2 GB" or why it failed. */
private fun downloadStatus(item: DownloadItem): String {
    return when (item.status) {
        JobStatus.queued -> "Queued"
        JobStatus.running -> "${(item.progress * 100).toInt()} %"
        JobStatus.succeeded -> listOfNotNull("Done", item.sizeBytes?.let(::formatBytes)).joinToString(" · ")
        JobStatus.failed -> "Failed: ${item.error ?: "unknown error"}"
        else -> "Cancelled"
    }
}

/** Failures in the error colour, the rest muted. */
private fun statusColor(item: DownloadItem): Color {
    if (item.status == JobStatus.failed) {
        return VeilColors.error
    }
    return VeilColors.contentMuted
}
