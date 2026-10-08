package com.playingwithclouds.veil.ui.subscriptions

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.FeedFilter
import com.playingwithclouds.veil.data.SubscriptionFeedEntry
import com.playingwithclouds.veil.data.SubscriptionRepository
import com.playingwithclouds.veil.data.SubscriptionSummary
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.Badge
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SettingsMenuButton
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.LargeHeader
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** What the following feed currently shows. */
data class FeedOptions(val newOnly: Boolean = false, val unwatchedOnly: Boolean = false)

/** Every followed thing, and the scenes they found. */
class SubscriptionsViewModel : ViewModel() {

    private val mutableSubscriptions = MutableStateFlow<List<SubscriptionSummary>>(emptyList())
    private val mutableOptions = MutableStateFlow(FeedOptions())

    /** The subscriptions, for the strip on top. */
    val subscriptions: StateFlow<List<SubscriptionSummary>> = mutableSubscriptions

    /** The selected feed filters. */
    val options: StateFlow<FeedOptions> = mutableOptions

    /** The combined feed, newest find first. */
    val feed = PagedList<SubscriptionFeedEntry>(viewModelScope, PAGE_SIZE, { entry -> entry.scene.id }) { offset ->
        val filter = FeedFilter(newOnly = mutableOptions.value.newOnly, unwatchedOnly = mutableOptions.value.unwatchedOnly)
        SubscriptionRepository.feed(filter, PAGE_SIZE, offset)
    }

    init {
        feed.loadMore()
        loadSubscriptions()
    }

    /** Reloads the subscription strip (new-scene counts change as runs finish). */
    fun loadSubscriptions() {
        viewModelScope.launch {
            try {
                mutableSubscriptions.value = SubscriptionRepository.subscriptions()
            } catch (error: Exception) {
                // Keep the strip as it was.
            }
        }
    }

    /** Applies feed filters and reloads the feed. */
    fun setOptions(options: FeedOptions) {
        mutableOptions.value = options
        feed.refresh()
    }

    companion object {
        private const val PAGE_SIZE = 30
    }
}

/** The Following tab. */
@Composable
fun SubscriptionsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { SubscriptionsViewModel() }
    val subscriptions by viewModel.subscriptions.collectAsStateWithLifecycle()
    val options by viewModel.options.collectAsStateWithLifecycle()
    LaunchedEffect(Unit) { viewModel.loadSubscriptions() }

    Scaffold { padding ->
        PagedGrid(
            paged = viewModel.feed,
            keyOf = { entry -> entry.scene.id },
            emptyText = "Nothing found yet. Follow a studio, performer, tag or search to fill this feed.",
            modifier = Modifier.padding(padding),
            header = {
                fullWidthItem("header") {
                    LargeHeader("Following") { SettingsMenuButton(navigator) }
                }
                fullWidthItem("strip") { SubscriptionStrip(subscriptions, navigator) }
                fullWidthItem("filters") { FeedFilters(options, viewModel::setOptions) }
            },
        ) { _, entry ->
            SceneCard(entry.scene, onClick = { navigator.openScene(entry.scene.id) }, isNew = entry.isNew)
        }
    }
}

/** The followed things as a row of avatars with their new-scene counts. */
@Composable
private fun SubscriptionStrip(subscriptions: List<SubscriptionSummary>, navigator: AppNavigator) {
    if (subscriptions.isEmpty()) {
        EmptyMessage("You are not following anything yet.")
        return
    }
    LazyRow(contentPadding = PaddingValues(horizontal = 4.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        items(subscriptions, key = { subscription -> subscription.id }) { subscription ->
            Column(
                Modifier.width(76.dp).pressClickable { navigator.openSubscription(subscription.id) },
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Box {
                    Avatar(subscription.targetImageUrl, Modifier.size(64.dp))
                    if (subscription.newCount > 0) {
                        Badge(
                            subscription.newCount.toString(),
                            Modifier.align(Alignment.TopEnd),
                            background = VeilColors.accent,
                            foreground = VeilColors.canvas,
                        )
                    }
                }
                Text(
                    subscription.displayName,
                    style = MaterialTheme.typography.labelMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    textAlign = TextAlign.Center,
                )
            }
        }
    }
}

/** Pills narrowing the feed to new or unwatched scenes. */
@Composable
private fun FeedFilters(options: FeedOptions, onChange: (FeedOptions) -> Unit) {
    Row(Modifier.padding(horizontal = 4.dp, vertical = 8.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        Pill(
            "New only",
            selected = options.newOnly,
            onClick = { onChange(options.copy(newOnly = !options.newOnly)) },
        )
        Pill(
            "Unwatched",
            selected = options.unwatchedOnly,
            onClick = { onChange(options.copy(unwatchedOnly = !options.unwatchedOnly)) },
        )
    }
}
