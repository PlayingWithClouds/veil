package com.playingwithclouds.veil.ui.search

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.FilterOption
import com.playingwithclouds.veil.data.SearchFilter
import com.playingwithclouds.veil.data.SiteChoice
import com.playingwithclouds.veil.data.SiteStyle
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** Size of a site's logo in its chip. */
private val SiteIconSize = 16.dp

/** What a filter sheet offers: the sites (with their styles, video search only) and the tags to hide. */
data class FilterChoices(
    val siteChoices: List<SiteChoice> = emptyList(),
    val siteStyles: List<SiteStyle> = emptyList(),
    val tagCounts: List<Pair<TagSummary, Int>> = emptyList(),
    /** The POV tag the Camera group picks; the group is left out while the library has none. */
    val povTag: TagSummary? = null,
    /** The scene labels in the library with how many videos have each, most common first; "What's in it" is left out without them. */
    val sceneLabels: List<Pair<TagSummary, Int>> = emptyList(),
)

/**
 * The chip row under the search box, as the desktop's: Filters (with how many of its settings are
 * on), the scene labels asked for and the hidden tags (tap one to drop it), then the recent searches.
 */
@Composable
fun FilterChipRow(
    filter: SearchFilter,
    onOpenFilters: () -> Unit,
    onFilter: (SearchFilter) -> Unit,
    recent: List<String> = emptyList(),
    onRecent: (String) -> Unit = {},
) {
    var filtersLabel = "Filters"
    if (filter.panelCount > 0) {
        filtersLabel = "Filters · ${filter.panelCount}"
    }
    LazyRow(
        Modifier.fillMaxWidth().fadingEdges().padding(vertical = VeilSpacing.extraSmall),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item("filters") { FeedChip(filtersLabel, onClick = onOpenFilters, selected = filter.panelCount > 0, icon = VeilIcons.Tune) }
        items(filter.sceneLabels.entries.toList(), key = { entry -> "scene-${entry.key}" }) { entry ->
            FeedChip(entry.value, onClick = { onFilter(filter.copy(sceneLabels = filter.sceneLabels - entry.key)) }, selected = true, icon = VeilIcons.Close)
        }
        items(filter.hiddenTags.entries.toList(), key = { entry -> "hidden-${entry.key}" }) { entry ->
            FeedChip("Not ${entry.value}", onClick = { onFilter(filter.copy(hiddenTags = filter.hiddenTags - entry.key)) }, selected = true, icon = VeilIcons.Close)
        }
        items(recent, key = { text -> "recent-$text" }) { text ->
            FeedChip(text, onClick = { onRecent(text) }, icon = VeilIcons.Search)
        }
    }
}

/**
 * The filters, as the desktop's panel has them: length, quality, the user's rating, show only
 * unwatched or favourites, sort, sites (a whole style at once, or one by one with their logos)
 * and tags to hide. Every tap applies at once. For galleries ([forVideos] false) length, quality
 * and sort are left out, and "unwatched" reads "not viewed".
 */
@Composable
fun FilterSheet(
    filter: SearchFilter,
    choices: FilterChoices,
    onFilter: (SearchFilter) -> Unit,
    onDismiss: () -> Unit,
    forVideos: Boolean = true,
) {
    var unseenLabel = "Unwatched"
    if (!forVideos) {
        unseenLabel = "Not viewed"
    }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(
            Modifier.verticalScroll(rememberScrollState()).gutterPadding().padding(bottom = VeilSpacing.huge),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Text("Filters", Modifier.weight(1f), style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
                if (!filter.isEmpty) {
                    TextAction("Clear all", onClick = { onFilter(SearchFilter()) })
                }
            }
            if (forVideos) {
                SingleChoice("Length", SearchFilter.LENGTH_OPTIONS, filter.length) { value -> onFilter(filter.copy(length = value)) }
                SingleChoice("Quality", SearchFilter.RESOLUTION_OPTIONS, filter.minHeight) { value -> onFilter(filter.copy(minHeight = value)) }
            }
            SingleChoice("Your rating", SearchFilter.RATING_OPTIONS, filter.minRating) { value -> onFilter(filter.copy(minRating = value)) }
            FilterGroup("Show only") {
                FeedChip(unseenLabel, onClick = { onFilter(filter.copy(unwatched = !filter.unwatched)) }, selected = filter.unwatched)
                FeedChip("Favourites", onClick = { onFilter(filter.copy(favorite = !filter.favorite)) }, selected = filter.favorite)
            }
            if (forVideos && choices.sceneLabels.isNotEmpty()) {
                FilterGroup("What's in it") {
                    for ((label, count) in choices.sceneLabels) {
                        FeedChip("${label.name} · $count", onClick = { onFilter(filter.toggledSceneLabel(label)) }, selected = label.id in filter.sceneLabels)
                    }
                }
            }
            val povTag = choices.povTag
            if (forVideos && povTag != null) {
                FilterGroup("Camera") {
                    FeedChip("POV", onClick = { onFilter(filter.toggledRequiredTag(povTag)) }, selected = povTag.id in filter.requiredTags)
                }
            }
            if (forVideos) {
                SingleChoice("Sort by", SearchFilter.SORT_OPTIONS, filter.sort) { value -> onFilter(filter.copy(sort = value)) }
            }
            SitesGroup(filter, choices, onFilter)
            if (choices.tagCounts.isNotEmpty()) {
                FilterGroup("Hide tags") {
                    for ((tag, count) in choices.tagCounts) {
                        FeedChip("${tag.name} · $count", onClick = { onFilter(filter.copy(hiddenTags = filter.hiddenTags + (tag.id to tag.name))) }, icon = VeilIcons.Block)
                    }
                }
            }
        }
    }
}

/** A group of options of which one at most is on; tapping the one that's on turns it off. */
@Composable
private fun SingleChoice(title: String, options: List<FilterOption>, current: String?, onPick: (String?) -> Unit) {
    FilterGroup(title) {
        for (option in options) {
            FeedChip(option.label, onClick = {
                if (option.value == current) {
                    onPick(null)
                } else {
                    onPick(option.value)
                }
            }, selected = option.value == current)
        }
    }
}

/** The styles (each picks all its sites), then every site with its logo. */
@Composable
private fun SitesGroup(filter: SearchFilter, choices: FilterChoices, onFilter: (SearchFilter) -> Unit) {
    if (choices.siteChoices.isEmpty()) {
        return
    }
    FilterGroup("Sites") {
        for (style in choices.siteStyles) {
            FeedChip("${style.label} · ${style.sites.size}", onClick = { onFilter(filter.toggledStyle(style)) }, selected = filter.styleOn(style))
        }
    }
    FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        for (choice in choices.siteChoices) {
            FeedChip(
                choice.label,
                onClick = { onFilter(filter.toggledSite(choice.value)) },
                selected = choice.value in filter.sites,
                leading = { SiteIcon(choice.iconUrl) },
            )
        }
    }
}

/** A site's small logo in its chip; nothing when it has none. */
@Composable
private fun SiteIcon(iconUrl: String?) {
    if (iconUrl == null) {
        return
    }
    RemoteImage(iconUrl, Modifier.size(SiteIconSize).clip(VeilShapes.badge))
}

/** A titled group of chips that wrap. */
@Composable
private fun FilterGroup(title: String, chips: @Composable () -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        Text(title.uppercase(), style = MaterialTheme.typography.labelSmall, color = VeilColors.contentMuted)
        FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            chips()
        }
    }
}
