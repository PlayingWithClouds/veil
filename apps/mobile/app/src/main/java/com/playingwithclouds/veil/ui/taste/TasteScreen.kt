package com.playingwithclouds.veil.ui.taste

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
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
import com.playingwithclouds.veil.data.DiscoveryRepository
import com.playingwithclouds.veil.data.TasteEntry
import com.playingwithclouds.veil.data.TasteProfile
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.tagLabel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlin.math.abs

/** Loads the taste profile the recommender ranks with. */
class TasteViewModel : ViewModel() {

    private val mutableProfile = MutableStateFlow<LoadState<TasteProfile>>(LoadState.Loading)

    /** The profile once loaded. */
    val profile: StateFlow<LoadState<TasteProfile>> = mutableProfile

    init {
        load()
    }

    /** Loads the profile again. */
    fun load() {
        viewModelScope.loadInto(mutableProfile) { DiscoveryRepository.tasteProfile(ENTRIES_PER_SIGN) }
    }

    companion object {
        private const val ENTRIES_PER_SIGN = 8
    }
}

/** The taste dashboard: top tags, performers, studios and sites, and what is avoided. */
@Composable
fun TasteScreen(navigator: AppNavigator) {
    val viewModel = viewModel { TasteViewModel() }
    val profile by viewModel.profile.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Your taste", onBack = navigator::back) }) { padding ->
        LoadStateContent(profile, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { loaded ->
            if (loaded.signalCount == 0) {
                EmptyMessage("Watch, like or save a few videos and your taste shows up here.", Modifier.padding(padding))
                return@LoadStateContent
            }
            TasteContent(loaded, navigator, Modifier.padding(padding))
        }
    }
}

/** The sections of the profile, each with its liked entries on top and its avoided ones below. */
@Composable
private fun TasteContent(profile: TasteProfile, navigator: AppNavigator, modifier: Modifier) {
    LazyColumn(modifier, contentPadding = PaddingValues(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.large)) {
        item { Summary(profile.signalCount) }
        taste("Top tags", profile.tags, onOpen = { entry -> navigator.openTag(entry.id) }, labelOf = { entry -> tagLabel(entry.name) })
        taste("Top performers", profile.performers, onOpen = { entry -> navigator.openPerformer(entry.id) })
        taste("Top studios", profile.studios, onOpen = { entry -> navigator.openStudio(entry.id) })
        taste("Top sites", profile.sites, onOpen = null)
    }
}

/** Adds a heading and rows for one kind: liked entries first, then a muted "Avoided" group. */
private fun LazyListScope.taste(
    title: String,
    entries: List<TasteEntry>,
    onOpen: ((TasteEntry) -> Unit)?,
    labelOf: (TasteEntry) -> String = { entry -> entry.name },
) {
    val liked = entries.filter { entry -> entry.affinity > 0 }
    val avoided = entries.filter { entry -> entry.affinity < 0 }
    if (liked.isEmpty() && avoided.isEmpty()) {
        return
    }
    item { SectionHeading(title) }
    items(liked.size) { index -> TasteRow(liked[index], labelOf(liked[index]), onOpen) }
    if (avoided.isNotEmpty()) {
        item {
            Text("Avoided", modifier = Modifier.padding(top = VeilSpacing.medium), style = MaterialTheme.typography.labelLarge, color = VeilColors.contentMuted)
        }
        items(avoided.size) { index -> TasteRow(avoided[index], labelOf(avoided[index]), onOpen) }
    }
}

/** How much the profile is built on. */
@Composable
private fun Summary(signalCount: Int) {
    Text(
        "Built from $signalCount signals: watches, likes, saves and what you skipped. Recent ones count more.",
        style = MaterialTheme.typography.bodyMedium,
        color = VeilColors.contentMuted,
    )
}

/** One entity: photo (people and studios), name and a bar as long as its affinity; red for avoided. */
@Composable
private fun TasteRow(entry: TasteEntry, label: String, onOpen: ((TasteEntry) -> Unit)?) {
    var rowModifier = Modifier.fillMaxWidth()
    if (onOpen != null) {
        rowModifier = rowModifier.pressClickable { onOpen(entry) }
    }
    Row(
        rowModifier.padding(vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (entry.imagePath != null) {
            Avatar(entry.imagePath, Modifier.size(36.dp))
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
            Text(label, style = MaterialTheme.typography.bodyLarge, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
            var barColor = VeilColors.accent
            if (entry.affinity < 0) {
                barColor = VeilColors.error
            }
            ProgressBar(progress = abs(entry.affinity).toFloat(), color = barColor)
        }
    }
}
