package com.playingwithclouds.veil.ui.shorts

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.pager.PagerState
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.compose.ContentFrame
import androidx.media3.ui.compose.SURFACE_TYPE_TEXTURE_VIEW
import com.playingwithclouds.veil.data.ShortItem
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.OverlayIconButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.ui.video.clippedMediaSource
import com.playingwithclouds.veil.ui.video.rememberVideoPlayer

private const val MILLISECONDS_PER_SECOND = 1000.0

/** Shorts next to the one on screen whose streams are resolved ahead. */
private const val PREPARE_NEIGHBOURS = 2

/**
 * Shorties: a vertical swipe feed of the best bits (issue #103), one short a page, its stretch
 * looping. Tap a short to open the whole video at that moment. Only the short on screen holds the
 * player; the next ones' streams resolve ahead. Back sits over the feed, as on the video page.
 */
@Composable
fun ShortsScreen(navigator: AppNavigator) {
    val viewModel = viewModel { ShortsViewModel() }
    val state by viewModel.state.collectAsStateWithLifecycle()
    Box(Modifier.fillMaxSize().background(VeilColors.canvas)) {
        when (val shorts = state.shorts) {
            is LoadState.Loading -> LoadingIndicator()
            is LoadState.Failed -> FailedMessage(shorts.message, onRetry = viewModel::load)
            is LoadState.Loaded -> {
                if (shorts.value.isEmpty()) {
                    EmptyMessage("No shorts yet. Save clips, mark scenes or pick the scenes you like.", Modifier.align(Alignment.Center))
                } else {
                    ShortsPager(shorts.value, state, viewModel, navigator)
                }
            }
        }
        OverlayIconButton(VeilIcons.Back, contentDescription = "Back", onClick = navigator::back, modifier = Modifier.statusBarsPadding().padding(VeilSpacing.medium))
    }
}

/** The pager: the short on screen plays, its neighbours resolve, a short swiped away early logs a skip. */
@Composable
private fun ShortsPager(shorts: List<ShortItem>, state: ShortsState, viewModel: ShortsViewModel, navigator: AppNavigator) {
    val pagerState = rememberPagerState { shorts.size }
    val player = rememberVideoPlayer()
    PlayCurrent(pagerState, shorts, state, player, viewModel)
    VerticalPager(pagerState, Modifier.fillMaxSize(), key = { index -> shorts[index].key }) { index ->
        val short = shorts[index]
        val current = index == pagerState.currentPage
        ShortPage(short, current, state.streams.containsKey(short.key), player) {
            player.pause()
            navigator.openVideoAt(short.video.id, short.startSeconds + player.currentPosition / MILLISECONDS_PER_SECOND)
        }
    }
}

/**
 * Loads the current short's stretch into the player (looping) once its stream is there, resolves
 * the neighbours' streams, and reports the short that was left with how long it played.
 */
@Composable
private fun PlayCurrent(pagerState: PagerState, shorts: List<ShortItem>, state: ShortsState, player: ExoPlayer, viewModel: ShortsViewModel) {
    LaunchedEffect(pagerState, shorts) {
        var previous: ShortItem? = null
        snapshotFlow { pagerState.currentPage }.collect { page ->
            val left = previous
            if (left != null) {
                viewModel.left(left, player.currentPosition / MILLISECONDS_PER_SECOND)
            }
            previous = shorts[page]
            for (index in page - 1..page + PREPARE_NEIGHBOURS) {
                shorts.getOrNull(index)?.let(viewModel::prepare)
            }
        }
    }
    val short = shorts.getOrNull(pagerState.currentPage)
    val stream = short?.let { item -> state.streams[item.key] }
    LaunchedEffect(short?.key, stream) {
        player.stop()
        player.clearMediaItems()
        if (short == null || stream == null) {
            return@LaunchedEffect
        }
        player.setMediaSource(clippedMediaSource(stream, short.startSeconds, short.endSeconds))
        player.repeatMode = Player.REPEAT_MODE_ONE
        player.prepare()
        player.playWhenReady = true
    }
}

/** One short: the cover, the playing stretch over it when on screen, and its title and why it's there. */
@Composable
private fun ShortPage(short: ShortItem, current: Boolean, resolved: Boolean, player: ExoPlayer, onOpen: () -> Unit) {
    Box(Modifier.fillMaxSize().pressClickable(onOpen), contentAlignment = Alignment.Center) {
        RemoteImage(short.video.coverUrl, Modifier.fillMaxWidth(), ContentScale.Fit)
        if (current) {
            ContentFrame(player = player, modifier = Modifier.fillMaxSize(), surfaceType = SURFACE_TYPE_TEXTURE_VIEW, contentScale = ContentScale.Fit)
        }
        if (current && !resolved) {
            Spinner()
        }
        Column(
            Modifier.align(Alignment.BottomStart).fillMaxWidth().background(VeilColors.playerDim).navigationBarsPadding().padding(VeilSpacing.gutter),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
        ) {
            Text(short.video.title, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2)
            val detail = short.detail
            if (detail != null) {
                Text(detail, style = VeilType.cardMeta, color = VeilColors.contentMuted, maxLines = 1)
            }
        }
    }
}
