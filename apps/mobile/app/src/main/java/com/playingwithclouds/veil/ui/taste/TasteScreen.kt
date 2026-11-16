package com.playingwithclouds.veil.ui.taste

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.data.Taste
import com.playingwithclouds.veil.data.TasteEntry
import com.playingwithclouds.veil.data.TasteRepository
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlin.math.abs

/** Entries shown per section. */
private const val ENTRY_LIMIT = 12

private val EntryPictureSize = 32.dp
private val WeightBarWidth = 72.dp

/** What the recommendations think the user likes and dislikes. */
class TasteViewModel : ViewModel() {

    private val tasteState = MutableStateFlow<LoadState<Taste>>(LoadState.Loading)
    val taste: StateFlow<LoadState<Taste>> = tasteState.asStateFlow()

    init {
        load()
    }

    /** Loads the taste profile. */
    fun load() {
        viewModelScope.loadInto(tasteState) { TasteRepository.taste() }
    }
}

/**
 * Taste, as the web shows it: tags, performers, studios and sites the recommendations think the
 * user likes (accent bars) and dislikes (muted bars), from what they watched, liked and skipped.
 */
@Composable
fun TasteScreen(navigator: AppNavigator) {
    val viewModel = viewModel { TasteViewModel() }
    val state by viewModel.taste.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Taste", onBack = navigator::back) }) { padding ->
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { taste ->
            LazyColumn(contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
                item("signals") {
                    Text(
                        "From ${taste.signalCount} things you watched, liked and skipped.",
                        Modifier.gutterPadding().padding(top = VeilSpacing.medium),
                        style = MaterialTheme.typography.bodySmall,
                        color = VeilColors.contentMuted,
                    )
                }
                tasteSection("Tags", taste.tags, onOpen = { entry -> navigator.openTag(TagSummary(entry.id, entry.name)) })
                tasteSection("Performers", taste.performers, onOpen = { entry -> navigator.openPerformer(entry.id) })
                tasteSection("Studios", taste.studios, onOpen = { entry -> navigator.openStudio(entry.id) })
                tasteSection("Sites", taste.sites, onOpen = null)
            }
        }
    }
}

/** One dimension of the taste: its heading, then its strongest entries, liked first. */
private fun LazyListScope.tasteSection(title: String, entries: List<TasteEntry>, onOpen: ((TasteEntry) -> Unit)?) {
    if (entries.isEmpty()) {
        return
    }
    item("heading-$title") { SectionHeading(title, modifier = Modifier.gutterPadding()) }
    items(entries.take(ENTRY_LIMIT), key = { entry -> "$title-${entry.id}" }) { entry ->
        TasteRow(entry, onOpen)
    }
}

/** An entry: picture, name, and a bar as long as how strongly it's liked or disliked. */
@Composable
private fun TasteRow(entry: TasteEntry, onOpen: ((TasteEntry) -> Unit)?) {
    var rowModifier = Modifier.fillMaxWidth()
    if (onOpen != null) {
        rowModifier = rowModifier.pressClickable { onOpen(entry) }
    }
    var barColor = VeilColors.accent
    var label = "Liked"
    if (entry.weight < 0) {
        barColor = VeilColors.contentFaint
        label = "Disliked"
    }
    Row(
        rowModifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Avatar(entry.imageUrl, Modifier.size(EntryPictureSize))
        Column(Modifier.weight(1f)) {
            Text(entry.name, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(label, style = VeilType.cardMeta, color = VeilColors.contentMuted)
        }
        ProgressBar(abs(entry.weight).toFloat(), Modifier.width(WeightBarWidth), color = barColor)
    }
}
