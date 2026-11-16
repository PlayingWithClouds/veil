package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.snapping.SnapPosition
import androidx.compose.foundation.gestures.snapping.rememberSnapFlingBehavior
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.RecommendedVideo
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.AudioCard
import com.playingwithclouds.veil.ui.components.CamCard
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.GlassTopBarLayout
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.components.PerformerBubble
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.StoryCard
import com.playingwithclouds.veil.ui.components.TabTopBar
import com.playingwithclouds.veil.ui.components.TrackShown
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Card widths of the rows; the next card peeks in at the edge. */
private val VideoCardWidth = 260.dp
private val LiveCardWidth = 220.dp
private val StoryCardWidth = 130.dp
private val AudioCardWidth = 150.dp
private val PerformerWidth = 88.dp

/** Share of the screen a top pick takes, so the next one peeks in. */
private const val TOP_PICK_WIDTH_FRACTION = 0.86f

/** The list key of a More for you card, apart from the same video in another row. */
private fun forYouKey(videoId: String): String = "for-you:$videoId"

/**
 * The Home tab, after the web home: the slim top bar (menu, wordmark, search), then the top picks
 * as a row of large cards (the one in front plays its preview), Continue watching, Live now, More
 * for you as full-width cards, Continue reading, Stories, Audio and Performers. Empty rows hide;
 * pull down to load everything again. Recommendation cards log impressions on "home", Continue
 * watching on "continue".
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(navigator: AppNavigator) {
    val viewModel = viewModel { HomeViewModel() }
    val picks by viewModel.picks.collectAsStateWithLifecycle()
    val refreshing by viewModel.refreshing.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    ShowAmbientGlow((picks as? LoadState.Loaded)?.value?.top?.firstOrNull()?.video?.coverUrl)
    GlassTopBarLayout(bar = { TabTopBar(navigator) }) {
        PullToRefreshBox(isRefreshing = refreshing, onRefresh = viewModel::refresh, modifier = Modifier.fillMaxSize()) {
            LazyColumn(
                Modifier.fillMaxSize(),
                state = listState,
                contentPadding = PaddingValues(top = LocalTopBarInset.current, bottom = VeilSpacing.large + LocalFloatingBarInset.current),
            ) {
                item("top-picks") { TopPicks(picks, viewModel, navigator) }
                item("continue-watching") { ContinueWatchingRow(viewModel, navigator) }
                item("live") { LiveRow(viewModel, navigator) }
                forYou(picks, viewModel, navigator)
                item("more-rows") { MoreRows(viewModel, navigator) }
            }
            ForYouImpressions(listState, picks, viewModel)
        }
    }
}

/**
 * The top picks: a heading over a row of large cards that snaps to each, the one in front playing
 * its preview once the row rests. A placeholder while loading, the failure with Retry when they
 * didn't load, nothing when there are none.
 */
@Composable
private fun TopPicks(picks: LoadState<HomePicks>, viewModel: HomeViewModel, navigator: AppNavigator) {
    when (picks) {
        is LoadState.Loading -> TopPicksPlaceholder()
        is LoadState.Failed -> FailedMessage(picks.message, onRetry = viewModel::retryPicks)
        is LoadState.Loaded -> TopPicksRow(picks.value.top, viewModel, navigator)
    }
}

/** The loaded top picks; see [TopPicks]. */
@Composable
private fun TopPicksRow(top: List<RecommendedVideo>, viewModel: HomeViewModel, navigator: AppNavigator) {
    if (top.isEmpty()) {
        return
    }
    val rowState = rememberLazyListState()
    val inFront by remember { derivedStateOf { pickInFront(rowState) } }
    TrackShown(rowState, top.map { pick -> pick.video.id }) { position ->
        viewModel.recordShown(top[position].video.id, ImpressionSurface.HOME, position)
    }
    Column(Modifier.fillMaxWidth()) {
        SectionHeading("Top picks for you", modifier = Modifier.gutterPadding())
        LazyRow(
            Modifier.fillMaxWidth(),
            state = rowState,
            contentPadding = GutterRowPadding,
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.cardGap),
            flingBehavior = rememberSnapFlingBehavior(rowState, SnapPosition.Start),
        ) {
            itemsIndexed(top, key = { _, pick -> pick.video.id }) { position, pick ->
                VideoCard(
                    pick.video,
                    onClick = {
                        viewModel.recordClicked(pick.video.id, ImpressionSurface.HOME, position)
                        navigator.openVideo(pick.video.id)
                    },
                    modifier = Modifier.fillParentMaxWidth(TOP_PICK_WIDTH_FRACTION),
                    note = pick.reason,
                    previewing = position == inFront && !rowState.isScrollInProgress,
                )
            }
        }
    }
}

/** The pick most in front: the first one, or the next once the first is scrolled more than halfway out. */
private fun pickInFront(rowState: LazyListState): Int {
    val first = rowState.layoutInfo.visibleItemsInfo.firstOrNull() ?: return 0
    if (rowState.firstVisibleItemScrollOffset > first.size / 2) {
        return first.index + 1
    }
    return first.index
}

/** The shape of the top picks while they load: a heading's room over one large card. */
@Composable
private fun TopPicksPlaceholder() {
    Column(Modifier.fillMaxWidth().padding(top = VeilSpacing.huge).gutterPadding()) {
        Column(
            Modifier
                .fillMaxWidth(TOP_PICK_WIDTH_FRACTION)
                .aspectRatio(16f / 9f)
                .clip(VeilShapes.card)
                .background(VeilColors.surface),
        ) {}
    }
}

/** Videos stopped partway, logged as shown on "continue". */
@Composable
private fun ContinueWatchingRow(viewModel: HomeViewModel, navigator: AppNavigator) {
    val videos by viewModel.continueWatching.collectAsStateWithLifecycle()
    val rowState = rememberLazyListState()
    TrackShown(rowState, videos.map { video -> video.id }) { position ->
        viewModel.recordShown(videos[position].id, ImpressionSurface.CONTINUE, position)
    }
    Shelf("Continue watching", videos, key = { video -> video.id }, itemWidth = VideoCardWidth, state = rowState) { video, position ->
        VideoCard(video, onClick = {
            viewModel.recordClicked(video.id, ImpressionSurface.CONTINUE, position)
            navigator.openVideo(video.id)
        })
    }
}

/** Rooms being recorded, then rooms live now. */
@Composable
private fun LiveRow(viewModel: HomeViewModel, navigator: AppNavigator) {
    val rooms by viewModel.live.collectAsStateWithLifecycle()
    Shelf("Live now", rooms, key = { room -> room.id }, itemWidth = LiveCardWidth) { room, _ ->
        CamCard(room, onClick = { navigator.openCam(room.id) })
    }
}

/** More for you: a heading, then the recommendations after the top picks as full-width cards. */
private fun LazyListScope.forYou(picks: LoadState<HomePicks>, viewModel: HomeViewModel, navigator: AppNavigator) {
    if (picks !is LoadState.Loaded) {
        return
    }
    val forYou = picks.value.forYou
    if (forYou.isEmpty()) {
        return
    }
    val firstPosition = picks.value.top.size
    item("for-you-heading") { SectionHeading("More for you", modifier = Modifier.gutterPadding()) }
    itemsIndexed(forYou, key = { _, pick -> forYouKey(pick.video.id) }) { index, pick ->
        VideoCard(
            pick.video,
            onClick = {
                viewModel.recordClicked(pick.video.id, ImpressionSurface.HOME, firstPosition + index)
                navigator.openVideo(pick.video.id)
            },
            note = pick.reason,
            edgeToEdge = true,
        )
    }
}

/** Logs the More for you cards seen in the page list; positions continue after the top picks. */
@Composable
private fun ForYouImpressions(listState: LazyListState, picks: LoadState<HomePicks>, viewModel: HomeViewModel) {
    if (picks !is LoadState.Loaded) {
        return
    }
    val forYou = picks.value.forYou
    val firstPosition = picks.value.top.size
    TrackShown(listState, forYou.map { pick -> forYouKey(pick.video.id) }) { index ->
        viewModel.recordShown(forYou[index].video.id, ImpressionSurface.HOME, firstPosition + index)
    }
}

/** The rows after More for you: Continue reading, Stories, Audio, Performers. */
@Composable
private fun MoreRows(viewModel: HomeViewModel, navigator: AppNavigator) {
    val continueReading by viewModel.continueReading.collectAsStateWithLifecycle()
    val stories by viewModel.stories.collectAsStateWithLifecycle()
    val audios by viewModel.audios.collectAsStateWithLifecycle()
    val performers by viewModel.performers.collectAsStateWithLifecycle()
    Column {
        Shelf("Continue reading", continueReading, key = { story -> story.id }, itemWidth = StoryCardWidth) { story, _ ->
            StoryCard(story, onClick = { navigator.openStory(story.id) })
        }
        Shelf("Stories", stories, key = { story -> story.id }, itemWidth = StoryCardWidth) { story, _ ->
            StoryCard(story, onClick = { navigator.openStory(story.id) })
        }
        Shelf("Audio", audios, key = { audio -> audio.id }, itemWidth = AudioCardWidth) { audio, _ ->
            AudioCard(audio, onClick = { navigator.openAudio(audio.id) })
        }
        Shelf("Performers", performers, key = { performer -> performer.id }, itemWidth = PerformerWidth) { performer, _ ->
            PerformerBubble(performer, onClick = { navigator.openPerformer(performer.id) })
        }
    }
}
