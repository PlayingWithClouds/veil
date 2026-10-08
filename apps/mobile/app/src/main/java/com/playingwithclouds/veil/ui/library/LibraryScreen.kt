package com.playingwithclouds.veil.ui.library

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
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
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
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SegmentedControl
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.formatBytes
import kotlinx.coroutines.delay

/** How often the download queue is reloaded while it is on screen. */
private const val QUEUE_POLL_MILLISECONDS = 3000L

/** Padding of each section's list: the screen gutter at the sides. */
private val SectionPadding = PaddingValues(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.medium)

/** The sections of the Library tab. */
enum class LibrarySection(val label: String) {
    DOWNLOADED("Downloaded"),
    QUEUE("Queue"),
    WATCHLIST("Watchlist"),
}

/** Downloaded scenes, the download queue and the watchlist, opened on [initialSection] (a [LibrarySection] name). */
@Composable
fun LibraryScreen(initialSection: String, navigator: AppNavigator) {
    val viewModel = viewModel { LibraryViewModel() }
    var selected by rememberSaveable { mutableIntStateOf(sectionIndex(initialSection)) }
    val section = LibrarySection.entries[selected]

    LaunchedEffect(section) {
        viewModel.loadSection(section)
        while (section == LibrarySection.QUEUE) {
            delay(QUEUE_POLL_MILLISECONDS)
            viewModel.reloadQueue()
        }
    }

    Scaffold(topBar = { VeilTopBar("Downloads & watchlist", onBack = navigator::back) }) { padding ->
        Column(Modifier.padding(padding)) {
            SegmentedControl(
                labels = LibrarySection.entries.map { entry -> entry.label },
                selectedIndex = selected,
                onSelect = { index -> selected = index },
                modifier = Modifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
            )
            when (section) {
                LibrarySection.DOWNLOADED -> DownloadedSection(viewModel, navigator)
                LibrarySection.QUEUE -> QueueSection(viewModel, navigator)
                LibrarySection.WATCHLIST -> WatchlistSection(viewModel, navigator)
            }
        }
    }
}

/** The tab index of a section name; the first tab when unknown. */
private fun sectionIndex(name: String): Int {
    val section = LibrarySection.entries.firstOrNull { entry -> entry.name == name } ?: return 0
    return section.ordinal
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
        contentPadding = SectionPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
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
        LazyColumn(contentPadding = SectionPadding, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
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
        Modifier.fillMaxWidth().pressClickable { navigator.openScene(card.mediaId) },
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(card.posterPath, Modifier.width(128.dp).aspectRatio(16f / 9f).clip(VeilShapes.card))
        Text(card.title, Modifier.weight(1f), style = MaterialTheme.typography.bodyMedium, maxLines = 3)
        IconTap(VeilIcons.Delete, contentDescription = "Remove from watchlist", onClick = onRemove)
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
    LazyColumn(contentPadding = SectionPadding, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        items(jobs.downloads, key = { job -> job.id }) { job ->
            DownloadRow(job, navigator, onRetry = { viewModel.retry(job.id) }, onDelete = { viewModel.delete(job.id) })
        }
        if (jobs.background.isNotEmpty()) {
            item {
                Column {
                    SectionHeading("Background jobs")
                    val failed = jobs.background.count { job -> job.status == "failed" }
                    Text("${jobs.background.size} jobs, $failed failed", color = VeilColors.contentMuted)
                    SecondaryButton(
                        "Clear background jobs",
                        onClick = viewModel::clearBackgroundJobs,
                        modifier = Modifier.padding(vertical = VeilSpacing.medium),
                    )
                }
            }
        }
    }
}

/** One download: title, progress bar, size and the retry/remove actions. */
@Composable
private fun DownloadRow(job: DownloadJob, navigator: AppNavigator, onRetry: () -> Unit, onDelete: () -> Unit) {
    VeilCard {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Column(
                Modifier.weight(1f).pressClickable(enabled = job.scene != null) { job.scene?.let { scene -> navigator.openScene(scene.id) } },
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            ) {
                Text(job.title, maxLines = 2, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Medium)
                Text(statusLine(job), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
                val progress = job.progress
                if (job.isActive && progress != null) {
                    ProgressBar(progress.toFloat())
                }
            }
            if (job.isFailed) {
                IconTap(VeilIcons.Refresh, contentDescription = "Retry", onClick = onRetry)
            }
            IconTap(VeilIcons.Delete, contentDescription = "Remove", onClick = onDelete)
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
