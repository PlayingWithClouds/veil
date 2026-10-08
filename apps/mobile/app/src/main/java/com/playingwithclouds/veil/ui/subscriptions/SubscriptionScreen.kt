package com.playingwithclouds.veil.ui.subscriptions

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.AssistChip
import androidx.compose.material3.Button
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.FeedFilter
import com.playingwithclouds.veil.data.SubscriptionFeedEntry
import com.playingwithclouds.veil.data.SubscriptionRepository
import com.playingwithclouds.veil.data.SubscriptionSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.ConfirmDialog
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.util.intervalLabel
import com.playingwithclouds.veil.util.scheduleLabel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** Re-run intervals offered in the settings: hourly up to weekly. */
private val intervalHourOptions = listOf(1, 3, 6, 12, 24, 48, 168)

/** One subscription: its schedule settings and the scenes it found. */
class SubscriptionViewModel(private val subscriptionId: String) : ViewModel() {

    private val mutableSubscription = MutableStateFlow<LoadState<SubscriptionSummary>>(LoadState.Loading)
    private val mutableRunning = MutableStateFlow(false)

    /** The subscription. */
    val subscription: StateFlow<LoadState<SubscriptionSummary>> = mutableSubscription

    /** Whether a manual run is in progress. */
    val running: StateFlow<Boolean> = mutableRunning

    /** The scenes this subscription found, newest first. */
    val feed = PagedList<SubscriptionFeedEntry>(viewModelScope, PAGE_SIZE, { entry -> entry.scene.id }) { offset ->
        SubscriptionRepository.feed(FeedFilter(subscriptionId = subscriptionId), PAGE_SIZE, offset)
    }

    init {
        load()
        feed.loadMore()
    }

    /** Loads the subscription. */
    fun load() {
        viewModelScope.loadInto(mutableSubscription) {
            requireNotNull(SubscriptionRepository.subscriptions().firstOrNull { candidate -> candidate.id == subscriptionId }) { "Subscription not found" }
        }
    }

    /** Changes the re-run interval. */
    fun setInterval(hours: Int) = applyChange { SubscriptionRepository.update(subscriptionId, hours, null) }

    /** Pauses or resumes the schedule. */
    fun setEnabled(enabled: Boolean) = applyChange { SubscriptionRepository.update(subscriptionId, null, enabled) }

    /** Runs the subscription now; this can take a minute. */
    fun runNow() {
        viewModelScope.launch {
            mutableRunning.value = true
            try {
                mutableSubscription.value = LoadState.Loaded(SubscriptionRepository.runNow(subscriptionId))
                feed.refresh()
            } catch (error: Exception) {
                // The schedule label shows the failure on the next load.
            }
            mutableRunning.value = false
        }
    }

    /** Deletes the subscription and leaves the page. */
    fun unsubscribe(onDone: () -> Unit) {
        viewModelScope.launch {
            val result = runCatching { SubscriptionRepository.unsubscribe(subscriptionId) }
            if (result.isSuccess) {
                onDone()
            }
        }
    }

    /** Clears the new-scene count; runs after the screen closed, so it outlives the view model. */
    fun markSeen() {
        AppScope.launch { runCatching { SubscriptionRepository.markSeen(subscriptionId) } }
    }

    /** Applies a subscription change and shows the result. */
    private fun applyChange(change: suspend () -> SubscriptionSummary) {
        viewModelScope.launch {
            try {
                mutableSubscription.value = LoadState.Loaded(change())
            } catch (error: Exception) {
                // Keep the previous settings on screen.
            }
        }
    }

    companion object {
        private const val PAGE_SIZE = 30
    }
}

/** A subscription's page. */
@Composable
fun SubscriptionScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "subscription-$id") { SubscriptionViewModel(id) }
    val subscription by viewModel.subscription.collectAsStateWithLifecycle()
    val running by viewModel.running.collectAsStateWithLifecycle()
    var confirmingUnfollow by remember { mutableStateOf(false) }

    DisposableEffect(id) { onDispose { viewModel.markSeen() } }
    if (confirmingUnfollow) {
        ConfirmDialog(
            title = "Stop following?",
            message = "The subscription and its found scenes are removed.",
            confirmLabel = "Unfollow",
            onConfirm = {
                confirmingUnfollow = false
                viewModel.unsubscribe(navigator::back)
            },
            onDismiss = { confirmingUnfollow = false },
        )
    }

    Scaffold(
        topBar = {
            VeilTopBar(
                title = (subscription as? LoadState.Loaded)?.value?.displayName.orEmpty(),
                onBack = navigator::back,
                actions = {
                    IconButton(onClick = { confirmingUnfollow = true }) { Icon(Icons.Filled.Delete, contentDescription = "Unfollow") }
                },
            )
        },
    ) { padding ->
        LoadStateContent(subscription, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { current ->
            PagedGrid(
                paged = viewModel.feed,
                keyOf = { entry -> entry.scene.id },
                emptyText = "Nothing found yet.",
                modifier = Modifier.padding(padding),
                header = { fullWidthItem("settings") { SubscriptionSettings(current, running, viewModel, navigator) } },
            ) { _, entry ->
                SceneCard(entry.scene, onClick = { navigator.openScene(entry.scene.id) }, isNew = entry.isNew)
            }
        }
    }
}

/** Schedule, counts and controls of a subscription. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun SubscriptionSettings(
    subscription: SubscriptionSummary,
    running: Boolean,
    viewModel: SubscriptionViewModel,
    navigator: AppNavigator,
) {
    Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Avatar(subscription.targetImageUrl, Modifier.size(56.dp))
            Column {
                Text(subscription.kind.lowercase().replaceFirstChar { letter -> letter.uppercase() }, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Text("${subscription.totalCount} videos, ${subscription.newCount} new", style = MaterialTheme.typography.bodyMedium)
            }
        }
        Text(
            scheduleLabel(subscription.lastRunAt, subscription.nextRunAt, subscription.enabled),
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            style = MaterialTheme.typography.bodySmall,
        )
        val problem = subscription.lastError
        if (!problem.isNullOrEmpty()) {
            Text(problem, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
        }
        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Button(onClick = viewModel::runNow, enabled = !running) {
                Icon(Icons.Filled.PlayArrow, contentDescription = null, modifier = Modifier.size(18.dp))
                Text(runLabel(running))
            }
            IntervalPicker(subscription.intervalHours, viewModel::setInterval)
            val targetId = subscription.targetId
            if (targetId != null) {
                OutlinedButton(onClick = { openTarget(subscription, targetId, navigator) }) { Text("Open page") }
            }
        }
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Switch(checked = subscription.enabled, onCheckedChange = viewModel::setEnabled)
            Text("Run on schedule")
        }
    }
}

/** The label of the manual-run button. */
private fun runLabel(running: Boolean): String {
    if (running) {
        return "Running..."
    }
    return "Run now"
}

/** Opens the studio, performer or tag a subscription follows. */
private fun openTarget(subscription: SubscriptionSummary, targetId: String, navigator: AppNavigator) {
    when (subscription.kind) {
        "STUDIO" -> navigator.openStudio(targetId)
        "PERFORMER" -> navigator.openPerformer(targetId)
        "TAG" -> navigator.openTag(targetId)
    }
}

/** A chip that opens a menu of re-run intervals. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun IntervalPicker(currentHours: Int, onSelect: (Int) -> Unit) {
    var expanded by remember { mutableStateOf(false) }
    val options = (intervalHourOptions + currentHours).distinct().sorted()
    androidx.compose.foundation.layout.Box {
        AssistChip(onClick = { expanded = true }, label = { Text(intervalLabel(currentHours)) })
        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            for (hours in options) {
                DropdownMenuItem(
                    text = { Text(intervalLabel(hours)) },
                    onClick = {
                        expanded = false
                        onSelect(hours)
                    },
                )
            }
        }
    }
}
