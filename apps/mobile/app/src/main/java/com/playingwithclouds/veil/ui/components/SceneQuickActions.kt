package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.AppScope
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.SceneSummary
import com.playingwithclouds.veil.data.Verdict
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** The user's current stance on a scene, as the quick actions show it. */
data class QuickActionsState(val onWatchlist: Boolean = false, val verdict: Verdict? = null)

/**
 * Loads and changes a scene's watchlist and like state for the quick actions sheet. Writes run on
 * [AppScope], so closing the sheet right after a tap does not cancel them.
 */
class SceneQuickActionsViewModel(private val sceneId: String) : ViewModel() {

    private val mutableState = MutableStateFlow(QuickActionsState())

    /** What the sheet shows. */
    val state: StateFlow<QuickActionsState> = mutableState

    init {
        viewModelScope.launch {
            val onWatchlist = runCatching { SceneRepository.isOnWatchlist(sceneId) }.getOrDefault(false)
            val verdict = runCatching { SceneRepository.verdict(sceneId) }.getOrNull()
            mutableState.value = QuickActionsState(onWatchlist, verdict)
        }
    }

    /** Puts the scene on or takes it off the watchlist. */
    fun toggleWatchlist() {
        val target = !mutableState.value.onWatchlist
        mutableState.update { current -> current.copy(onWatchlist = target) }
        AppScope.launch { runCatching { SceneRepository.setOnWatchlist(sceneId, target) } }
    }

    /** Sets [verdict], or clears it when it is already the current one. */
    fun react(verdict: Verdict) {
        if (mutableState.value.verdict == verdict) {
            mutableState.update { current -> current.copy(verdict = null) }
            AppScope.launch { runCatching { SceneRepository.clearVerdict(sceneId) } }
            return
        }
        mutableState.update { current -> current.copy(verdict = verdict) }
        AppScope.launch { runCatching { SceneRepository.setVerdict(sceneId, verdict) } }
    }
}

/**
 * The sheet a long press on a scene card opens: watchlist, like, and "not interested" (a dislike,
 * which the recommender treats as a negative signal). [onNotInterested] lets the caller drop the
 * card from its list.
 */
@Composable
fun SceneQuickActionsSheet(scene: SceneSummary, onDismiss: () -> Unit, onNotInterested: () -> Unit = {}) {
    val viewModel = viewModel(key = "quick-actions-${scene.id}") { SceneQuickActionsViewModel(scene.id) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Text(
            scene.title,
            modifier = Modifier.padding(start = VeilSpacing.extraLarge, end = VeilSpacing.extraLarge, bottom = VeilSpacing.small),
            style = MaterialTheme.typography.titleMedium,
            color = VeilColors.content,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        var watchlistIcon = VeilIcons.Bookmark
        var watchlistLabel = "Add to watchlist"
        if (state.onWatchlist) {
            watchlistIcon = VeilIcons.BookmarkFilled
            watchlistLabel = "On watchlist"
        }
        SheetAction(watchlistLabel, watchlistIcon, onClick = viewModel::toggleWatchlist, active = state.onWatchlist)
        val liked = state.verdict == Verdict.UP
        var likeIcon = VeilIcons.Like
        if (liked) {
            likeIcon = VeilIcons.LikeFilled
        }
        SheetAction("Like", likeIcon, onClick = { viewModel.react(Verdict.UP) }, active = liked)
        SheetAction(
            "Not interested",
            VeilIcons.Dislike,
            onClick = {
                if (state.verdict != Verdict.DOWN) {
                    viewModel.react(Verdict.DOWN)
                }
                onNotInterested()
                onDismiss()
            },
        )
    }
}
