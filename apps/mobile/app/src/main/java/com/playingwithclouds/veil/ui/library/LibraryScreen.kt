package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Card
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.PrimaryTabRow
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.DownloadJob
import com.playingwithclouds.veil.data.JobQueue
import com.playingwithclouds.veil.data.MediaCard
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SectionTitle
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.util.formatBytes
import kotlinx.coroutines.delay

/** How often the download queue is reloaded while it is on screen. */
private const val QUEUE_POLL_MILLISECONDS = 3000L

/** The sections of the Library tab. */
enum class LibrarySection(val label: String) {
    DOWNLOADED("Downloaded"),
    QUEUE("Queue"),
    WATCHLIST("Watchlist"),
}

/** The Library tab: downloaded scenes, the download queue and the watchlist. */
@Composable
fun LibraryScreen(navigator: AppNavigator, onMenu: () -> Unit) {
    val viewModel = viewModel { LibraryViewModel() }
    var selected by rememberSaveable { mutableIntStateOf(0) }
    val section = LibrarySection.entries[selected]

    LaunchedEffect(section) {
        viewModel.loadSection(section)
        while (section == LibrarySection.QUEUE) {
            delay(QUEUE_POLL_MILLISECONDS)
            viewModel.reloadQueue()
        }
    }

    Scaffold(topBar = { VeilTopBar("Library", onMenu = onMenu) }) { padding ->
        Column(Modifier.padding(padding)) {
            PrimaryTabRow(selectedTabIndex = selected) {
                for (entry in LibrarySection.entries) {
                    Tab(selected = entry == section, onClick = { selected = entry.ordinal }, text = { Text(entry.label) })
                }
            }
            when (section) {
                LibrarySection.DOWNLOADED -> DownloadedSection(viewModel, navigator)
                LibrarySection.QUEUE -> QueueSection(viewModel, navigator)
                LibrarySection.WATCHLIST -> WatchlistSection(viewModel, navigator)
            }
        }
    }
}

/** Scenes whose download finished. */
@Composable
private fun DownloadedSection(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val downloaded by viewModel.downloaded.collectAsStateWithLifecycle()
    LoadStateContent(downloaded, onRetry = { viewModel.loadSection(LibrarySection.DOWNLOADED) }) { scenes ->
        SceneList(scenes, "Nothing downloaded yet. Queue a download from a video's page.", navigator)
    }
}

/** A scrollable grid of scene cards with an empty note. */
@Composable
private fun SceneList(scenes: List<SceneSummary>, emptyText: String, navigator: AppNavigator) {
    if (scenes.isEmpty()) {
        EmptyMessage(emptyText)
        return
    }
    LazyVerticalGrid(
        columns = GridCells.Adaptive(300.dp),
        contentPadding = PaddingValues(12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        items(scenes, key = { scene -> scene.id }) { scene ->
            SceneCard(scene, onClick = { navigator.openScene(scene.id) })
        }
    }
}

/** Saved-for-later scenes. */
@Composable
private fun WatchlistSection(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val watchlist by viewModel.watchlist.collectAsStateWithLifecycle()
    LoadStateContent(watchlist, onRetry = { viewModel.loadSection(LibrarySection.WATCHLIST) }) { cards ->
        if (cards.isEmpty()) {
            EmptyMessage("Your watchlist is empty.")
            return@LoadStateContent
        }
        LazyColumn(contentPadding = PaddingValues(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            items(cards, key = { card -> card.mediaId }) { card ->
                WatchlistRow(card, navigator, onRemove = { viewModel.removeFromWatchlist(card) })
            }
        }
    }
}

/** A watchlist entry with its poster, title and a remove button. */
@Composable
private fun WatchlistRow(card: MediaCard, navigator: AppNavigator, onRemove: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().clickable { navigator.openScene(card.mediaId) },
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(card.posterPath, Modifier.width(120.dp).aspectRatio(16f / 9f))
        Text(card.title, Modifier.weight(1f), maxLines = 3)
        IconButton(onClick = onRemove) { Icon(Icons.Filled.Delete, contentDescription = "Remove from watchlist") }
    }
}

/** The download queue and background jobs. */
@Composable
private fun QueueSection(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val queue by viewModel.queue.collectAsStateWithLifecycle()
    LoadStateContent(queue, onRetry = { viewModel.loadSection(LibrarySection.QUEUE) }) { jobs ->
        QueueList(jobs, viewModel, navigator)
    }
}

/** Downloads with progress, then a summary of background jobs. */
@Composable
private fun QueueList(jobs: JobQueue, viewModel: LibraryViewModel, navigator: AppNavigator) {
    if (jobs.downloads.isEmpty() && jobs.background.isEmpty()) {
        EmptyMessage("No downloads or background jobs.")
        return
    }
    LazyColumn(contentPadding = PaddingValues(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        items(jobs.downloads, key = { job -> job.id }) { job ->
            DownloadRow(job, navigator, onRetry = { viewModel.retry(job.id) }, onDelete = { viewModel.delete(job.id) })
        }
        if (jobs.background.isNotEmpty()) {
            item {
                Column {
                    SectionTitle("Background jobs")
                    val failed = jobs.background.count { job -> job.status == "failed" }
                    Text(
                        "${jobs.background.size} jobs, $failed failed",
                        Modifier.padding(horizontal = 16.dp),
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    TextButton(onClick = viewModel::clearBackgroundJobs) { Text("Clear background jobs") }
                }
            }
        }
    }
}

/** One download: title, progress bar, size and the retry/remove actions. */
@Composable
private fun DownloadRow(job: DownloadJob, navigator: AppNavigator, onRetry: () -> Unit, onDelete: () -> Unit) {
    Card(Modifier.fillMaxWidth()) {
        Row(Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(
                Modifier.weight(1f).clickable(enabled = job.scene != null) { job.scene?.let { scene -> navigator.openScene(scene.id) } },
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                Text(job.title, maxLines = 2, style = MaterialTheme.typography.bodyMedium)
                Text(statusLine(job), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                val progress = job.progress
                if (job.isActive && progress != null) {
                    LinearProgressIndicator(progress = { progress.toFloat().coerceIn(0f, 1f) }, modifier = Modifier.fillMaxWidth())
                }
            }
            if (job.isFailed) {
                IconButton(onClick = onRetry) { Icon(Icons.Filled.Refresh, contentDescription = "Retry") }
            }
            IconButton(onClick = onDelete) { Icon(Icons.Filled.Delete, contentDescription = "Remove") }
        }
    }
}

/** "downloading · 120 MB of 800 MB", or the error of a failed job. */
private fun statusLine(job: DownloadJob): String {
    val parts = mutableListOf(job.status)
    val received = job.bytesReceived
    val total = job.bytesTotal
    if (received != null && total != null && total > 0) {
        parts.add("${formatBytes(received)} of ${formatBytes(total)}")
    }
    val error = job.error
    if (job.isFailed && !error.isNullOrEmpty()) {
        parts.add(error)
    }
    return parts.joinToString(" · ")
}
