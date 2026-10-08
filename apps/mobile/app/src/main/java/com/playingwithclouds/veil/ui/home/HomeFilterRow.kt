package com.playingwithclouds.veil.ui.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.FilterPreset
import com.playingwithclouds.veil.data.SearchSite
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TextInputDialog
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilCheck
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.tagLabel

/** Tallest the site list in the filter sheet grows before it scrolls. */
private val SiteListMaxHeight = 320.dp

/**
 * The chips under the Home top bar: the site picker, "All", saved presets, runtime windows and the
 * tags most present in the feed. Selected chips turn white.
 */
@Composable
fun HomeFilterRow(viewModel: HomeViewModel, modifier: Modifier = Modifier) {
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val presets by viewModel.presets.collectAsStateWithLifecycle()
    val tagChips by viewModel.tagChips.collectAsStateWithLifecycle()
    val sites by viewModel.sites.collectAsStateWithLifecycle()
    var sheetOpen by remember { mutableStateOf(false) }

    LazyRow(
        modifier.fillMaxWidth().padding(vertical = VeilSpacing.small),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item(key = "sites") {
            FeedChip(
                siteChipLabel(filter.sources, sites),
                onClick = { sheetOpen = true },
                selected = filter.sources.isNotEmpty(),
                icon = VeilIcons.Tune,
            )
        }
        item(key = "all") {
            FeedChip("All", onClick = { viewModel.applyFilter(HomeFilter()) }, selected = !filter.isActive)
        }
        items(presets, key = { preset -> "preset-${preset.id}" }) { preset ->
            FeedChip(
                preset.name,
                onClick = { viewModel.applyFilter(togglePreset(filter, preset)) },
                selected = filter.matches(preset),
            )
        }
        items(RuntimeChip.entries, key = { chip -> chip.name }) { chip ->
            FeedChip(chip.label, onClick = { viewModel.applyFilter(filter.toggleRuntime(chip)) }, selected = filter.runtime == chip.window)
        }
        items(tagChips, key = { tag -> "tag-${tag.id}" }) { tag ->
            FeedChip(tagLabel(tag.name), onClick = { viewModel.applyFilter(filter.toggleTag(tag.id)) }, selected = filter.tagId == tag.id)
        }
    }

    if (sheetOpen) {
        HomeFilterSheet(viewModel, onDismiss = { sheetOpen = false })
    }
}

/** The filter of picking a preset chip, or none when it is already the active one. */
private fun togglePreset(filter: HomeFilter, preset: FilterPreset): HomeFilter {
    if (filter.matches(preset)) {
        return HomeFilter()
    }
    return HomeFilter.of(preset)
}

/** "Sites" while none is picked, the site's name for one, "3 sites" for more. */
private fun siteChipLabel(picked: Set<String>, sites: List<SearchSite>): String {
    if (picked.isEmpty()) {
        return "Sites"
    }
    if (picked.size > 1) {
        return "${picked.size} sites"
    }
    val name = picked.first()
    val site = sites.firstOrNull { candidate -> candidate.name == name }
    if (site == null) {
        return name
    }
    return site.label
}

/** The sheet behind the site chip: pick sites, save the filter as a chip, delete saved chips. */
@Composable
private fun HomeFilterSheet(viewModel: HomeViewModel, onDismiss: () -> Unit) {
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val presets by viewModel.presets.collectAsStateWithLifecycle()
    val sites by viewModel.sites.collectAsStateWithLifecycle()
    val tagChips by viewModel.tagChips.collectAsStateWithLifecycle()
    var naming by remember { mutableStateOf(false) }

    VeilBottomSheet(onDismissRequest = onDismiss) {
        SectionHeading("Sites", modifier = Modifier.gutterPadding())
        LazyColumn(Modifier.heightIn(max = SiteListMaxHeight)) {
            items(sites, key = { site -> site.name }) { site ->
                SiteRow(site, picked = filter.sources.contains(site.name), onClick = { viewModel.toggleSource(site.name) })
            }
        }
        if (filter.isActive && presets.none { preset -> filter.matches(preset) }) {
            SheetAction("Save this filter as a chip", VeilIcons.Plus, onClick = { naming = true })
        }
        if (presets.isNotEmpty()) {
            SectionHeading("Saved chips", modifier = Modifier.gutterPadding())
            for (preset in presets) {
                PresetRow(preset, onDelete = { viewModel.deletePreset(preset) })
            }
        }
    }

    if (naming) {
        val tagName = tagChips.firstOrNull { tag -> tag.id == filter.tagId }?.name
        TextInputDialog(
            title = "Save filter",
            label = "Name",
            confirmLabel = "Save",
            initialValue = suggestedPresetName(filter, tagName?.let { name -> tagLabel(name) }),
            onConfirm = { name ->
                viewModel.saveCurrentAsPreset(name)
                naming = false
            },
            onDismiss = { naming = false },
        )
    }
}

/** One site of the picker: its icon, name and a check. */
@Composable
private fun SiteRow(site: SearchSite, picked: Boolean, onClick: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .pressClickable(onClick)
            .padding(horizontal = VeilSpacing.extraLarge, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(site.iconUrl, Modifier.size(24.dp).clip(VeilShapes.badge), contentScale = ContentScale.Fit)
        Text(site.label, Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge, color = VeilColors.content, maxLines = 1)
        VeilCheck(picked)
    }
}

/** A saved chip in the sheet's list, with a delete button. */
@Composable
private fun PresetRow(preset: FilterPreset, onDelete: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().padding(start = VeilSpacing.extraLarge, end = VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(preset.name, Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge, color = VeilColors.content, maxLines = 1)
        IconTap(VeilIcons.Delete, contentDescription = "Delete ${preset.name}", onClick = onDelete)
    }
}
