package com.playingwithclouds.veil.ui.previews

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.pager.PagerState
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.RecommendedScene
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.SceneQuickActionsSheet
import com.playingwithclouds.veil.ui.components.SceneQuickActionsViewModel
import com.playingwithclouds.veil.ui.components.ScenePreview
import com.playingwithclouds.veil.ui.components.TagFilterSheet
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.tagLabel

/** How many previews before the end the next page starts loading. */
private const val PREFETCH_DISTANCE = 4

/** Blur radius of the poster that fills the screen behind the clip. */
private val BackdropBlur = 28.dp

/** Width over height of the preview clips. */
private const val CLIP_RATIO = 16f / 9f

/**
 * The vertical preview feed: one preview clip per screen, swipe up for the next, tap to open the
 * scene. Only the clip on screen plays; the others show their poster.
 */
@Composable
fun PreviewFeedScreen(navigator: AppNavigator) {
    val viewModel = viewModel { PreviewFeedViewModel() }
    val state by viewModel.feed.state.collectAsStateWithLifecycle()
    val tagFilter by viewModel.tagFilter.collectAsStateWithLifecycle()
    var showFilter by remember { mutableStateOf(false) }
    val items = state.items
    val pagerState = rememberPagerState { items.size }

    LoadMoreNearEnd(pagerState, itemCount = items.size, isLoading = state.isLoading, endReached = state.endReached, onLoadMore = viewModel.feed::loadMore)
    LaunchedEffect(pagerState.currentPage, items.size) {
        val item = items.getOrNull(pagerState.currentPage)
        if (item != null) {
            viewModel.recordShown(item, pagerState.currentPage)
        }
    }

    Box(Modifier.fillMaxSize().background(VeilColors.canvas)) {
        if (items.isEmpty()) {
            EmptyFeed(state.isLoading || !state.hasLoaded, state.error, onRetry = viewModel.feed::refresh)
        }
        VerticalPager(pagerState, Modifier.fillMaxSize(), key = { page -> items[page].scene.id }) { page ->
            val item = items[page]
            PreviewPage(
                item = item,
                active = page == pagerState.currentPage,
                navigator = navigator,
                onOpen = {
                    viewModel.recordClicked(item, page)
                    navigator.openScene(item.scene.id)
                },
                onNotInterested = { viewModel.drop(item.scene.id) },
            )
        }
        Row(
            Modifier.fillMaxWidth().statusBarsPadding().padding(horizontal = VeilSpacing.medium, vertical = VeilSpacing.small),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            RoundIconButton(VeilIcons.Back, contentDescription = "Back", onClick = navigator::back)
            RoundIconButton(VeilIcons.Tags, contentDescription = "Hide tags", onClick = { showFilter = true }, tint = filterTint(tagFilter.exclude.isNotEmpty()))
        }
    }
    if (showFilter) {
        TagFilterSheet(tagFilter, allowInclude = false, onChange = viewModel::setTagFilter, onDismiss = { showFilter = false })
    }
}

/** The filter button's tint: accent while tags are hidden. */
private fun filterTint(active: Boolean): Color {
    if (active) {
        return VeilColors.accent
    }
    return VeilColors.content
}

/** Loads the next page when the user is within [PREFETCH_DISTANCE] previews of the end, or the loaded page left nothing to show. */
@Composable
private fun LoadMoreNearEnd(pagerState: PagerState, itemCount: Int, isLoading: Boolean, endReached: Boolean, onLoadMore: () -> Unit) {
    LaunchedEffect(pagerState.currentPage, itemCount, isLoading, endReached) {
        if (!isLoading && !endReached && pagerState.currentPage >= itemCount - PREFETCH_DISTANCE) {
            onLoadMore()
        }
    }
}

/** The spinner, failure or note shown while the pager has nothing. */
@Composable
private fun EmptyFeed(loading: Boolean, error: String?, onRetry: () -> Unit) {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        if (loading) {
            Spinner()
            return@Box
        }
        if (error != null) {
            FailedMessage(error, onRetry)
            return@Box
        }
        EmptyMessage("No previews yet. Search or browse to fill the library.")
    }
}

/** One screen of the feed: backdrop, the clip (while [active]), details and the action rail. */
@Composable
private fun PreviewPage(
    item: RecommendedScene,
    active: Boolean,
    navigator: AppNavigator,
    onOpen: () -> Unit,
    onNotInterested: () -> Unit,
) {
    val scene = item.scene
    Box(Modifier.fillMaxSize().pressClickable(onOpen)) {
        RemoteImage(scene.posterPath, Modifier.fillMaxSize().blur(BackdropBlur))
        Box(Modifier.fillMaxSize().background(VeilColors.scrim))
        Clip(scene, active, Modifier.align(Alignment.Center))
        Box(
            Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color.Transparent, VeilColors.scrim))),
        ) {
            Row(
                Modifier.navigationBarsPadding().padding(start = VeilSpacing.gutter, end = VeilSpacing.medium, top = VeilSpacing.huge, bottom = VeilSpacing.large),
                verticalAlignment = Alignment.Bottom,
            ) {
                SceneDetails(scene, onOpen, navigator, Modifier.weight(1f))
                if (active) {
                    ActionRail(scene, onNotInterested)
                }
            }
        }
    }
}

/** The preview clip centered on the screen, with the runtime in the corner. */
@Composable
private fun Clip(scene: SceneSummary, active: Boolean, modifier: Modifier) {
    Box(modifier.fillMaxWidth().aspectRatio(CLIP_RATIO)) {
        RemoteImage(scene.posterPath, Modifier.matchParentSize(), contentScale = ContentScale.Fit)
        val previewVideo = scene.previewVideo
        if (active && previewVideo != null) {
            ScenePreview(previewVideo, Modifier.matchParentSize(), ContentScale.Fit)
        }
        val duration = scene.durationSeconds
        if (duration != null && duration > 0) {
            Badge(formatClock(duration.toDouble()), Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
        }
    }
}

/** Byline, title, a few tags and the watch button. */
@Composable
private fun SceneDetails(scene: SceneSummary, onOpen: () -> Unit, navigator: AppNavigator, modifier: Modifier) {
    Column(modifier, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        val byline = scene.byline
        if (byline != null) {
            Text(byline, style = MaterialTheme.typography.labelLarge, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
        }
        Text(scene.title, style = MaterialTheme.typography.titleMedium, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            for (tag in scene.tags.take(TAGS_SHOWN)) {
                TagPill(tagLabel(tag.name), onClick = { navigator.openTag(tag.id) })
            }
        }
        PrimaryButton("Watch", onClick = onOpen, icon = VeilIcons.Play)
    }
}

/** Like, watchlist and the quick actions sheet (hide a tag, performer or studio, not interested). */
@Composable
private fun ActionRail(scene: SceneSummary, onNotInterested: () -> Unit) {
    val actions = viewModel(key = "preview-actions-${scene.id}") { SceneQuickActionsViewModel(scene.id) }
    val actionState by actions.state.collectAsStateWithLifecycle()
    var showSheet by remember { mutableStateOf(false) }
    Column(Modifier.padding(start = VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        val liked = actionState.verdict == Verdict.UP
        var likeIcon = VeilIcons.Like
        if (liked) {
            likeIcon = VeilIcons.LikeFilled
        }
        RoundIconButton(likeIcon, contentDescription = "Like", onClick = { actions.react(Verdict.UP) }, tint = filterTint(liked))
        var watchlistIcon = VeilIcons.Bookmark
        if (actionState.onWatchlist) {
            watchlistIcon = VeilIcons.BookmarkFilled
        }
        RoundIconButton(watchlistIcon, contentDescription = "Watchlist", onClick = actions::toggleWatchlist, tint = filterTint(actionState.onWatchlist))
        RoundIconButton(VeilIcons.More, contentDescription = "More", onClick = { showSheet = true })
    }
    if (showSheet) {
        SceneQuickActionsSheet(scene, onDismiss = { showSheet = false }, onNotInterested = onNotInterested)
    }
}

/** Tags shown under a title. */
private const val TAGS_SHOWN = 3
