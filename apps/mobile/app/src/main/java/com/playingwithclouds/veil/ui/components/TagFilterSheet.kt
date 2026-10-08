package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.ui.text.style.TextOverflow
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.TagFilter
import com.playingwithclouds.veil.data.TagFilterState
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.tagLabel
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay

/** Tags offered per search of the sheet. */
private const val TAG_RESULT_LIMIT = 12

/** Pause after typing before the tag search runs. */
private const val TAG_SEARCH_DEBOUNCE_MILLISECONDS = 250L

/**
 * A sheet to require and forbid tags. The chosen tags show on top (tap one to drop it); below,
 * a search over all tags with Include and Exclude per row. With [allowInclude] off (feeds the
 * backend can only subtract from) only Exclude is offered.
 */
@Composable
fun TagFilterSheet(filter: TagFilter, allowInclude: Boolean, onChange: (TagFilter) -> Unit, onDismiss: () -> Unit) {
    var query by remember { mutableStateOf("") }
    var results by remember { mutableStateOf<List<EntityRef>>(emptyList()) }
    LaunchedEffect(query) {
        delay(TAG_SEARCH_DEBOUNCE_MILLISECONDS)
        results = searchTags(query)
    }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        TagFilterBody(allowInclude, filter, results, query, onQueryChange = { text -> query = text }, onChange = onChange)
    }
}

/** The sheet body: chosen tags, search box and result rows, scrollable. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun TagFilterBody(
    allowInclude: Boolean,
    filter: TagFilter,
    results: List<EntityRef>,
    query: String,
    onQueryChange: (String) -> Unit,
    onChange: (TagFilter) -> Unit,
) {
    Column(Modifier.verticalScroll(rememberScrollState())) {
        SectionHeading("Tags")
        if (filter.isEmpty) {
            Text(
                "Pick tags to require or hide.",
                modifier = Modifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.contentMuted,
            )
        }
        FlowRow(
            Modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter),
            horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        ) {
            for (tag in filter.include) {
                Pill(tagLabel(tag.name), onClick = { onChange(filter.with(tag, TagFilterState.NEUTRAL)) }, selected = true, icon = VeilIcons.Check)
            }
            for (tag in filter.exclude) {
                Pill(tagLabel(tag.name), onClick = { onChange(filter.with(tag, TagFilterState.NEUTRAL)) }, icon = VeilIcons.Block)
            }
        }
        SearchField(query, onQueryChange, placeholder = "Search tags")
        for (tag in results) {
            TagFilterRow(tag, filter.stateOf(tag), allowInclude) { state -> onChange(filter.with(tag, state)) }
        }
    }
}

/** One tag with its Include and Exclude pills; tapping the active pill clears the tag. */
@Composable
private fun TagFilterRow(tag: EntityRef, state: TagFilterState, allowInclude: Boolean, onState: (TagFilterState) -> Unit) {
    Row(
        Modifier.fillMaxWidth().padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            tagLabel(tag.name),
            modifier = Modifier.weight(1f),
            style = MaterialTheme.typography.bodyLarge,
            color = VeilColors.content,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        if (allowInclude) {
            Pill("Include", onClick = { onState(toggled(state, TagFilterState.INCLUDED)) }, selected = state == TagFilterState.INCLUDED)
        }
        Pill("Exclude", onClick = { onState(toggled(state, TagFilterState.EXCLUDED)) }, selected = state == TagFilterState.EXCLUDED, icon = VeilIcons.Block)
    }
}

/** [target], or neutral when the tag is in that state already. */
private fun toggled(current: TagFilterState, target: TagFilterState): TagFilterState {
    if (current == target) {
        return TagFilterState.NEUTRAL
    }
    return target
}

/** Tags matching the text as references; empty when the search fails. */
private suspend fun searchTags(text: String): List<EntityRef> {
    try {
        val search = text.trim().ifEmpty { null }
        return EntityRepository.tags(search, TAG_RESULT_LIMIT, 0).map { tag -> EntityRef(tag.id, tag.name, null) }
    } catch (error: CancellationException) {
        throw error
    } catch (error: Exception) {
        return emptyList()
    }
}
