package com.playingwithclouds.veil.ui.library

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.data.BlockedRecord
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LocalFloatingBarInset
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType

/** Width of a blocked video's cover in its row; performers and studios show a round picture as tall. */
private val CoverWidth = 112.dp
private val PictureSize = 63.dp

private const val LANDSCAPE_RATIO = 16f / 9f

/**
 * What's blocked, by kind: videos, performers, studios. Tapping one opens it; Unblock lets it back
 * into every list and takes it off this one at once.
 */
@Composable
fun BlockedTab(viewModel: LibraryViewModel, navigator: AppNavigator) {
    val state by viewModel.blocked.collectAsStateWithLifecycle()
    LoadStateContent(state, onRetry = viewModel::reload) { records ->
        if (records.isEmpty()) {
            EmptyMessage("Nothing blocked. Block a video, performer or studio from its menu.", Modifier.padding(top = LocalTopBarInset.current))
            return@LoadStateContent
        }
        LazyColumn(contentPadding = PaddingValues(top = LocalTopBarInset.current, bottom = VeilSpacing.large + LocalFloatingBarInset.current)) {
            for (kind in listOf(RecordKind.video, RecordKind.performer, RecordKind.studio)) {
                val ofKind = records.filter { record -> record.kind == kind }
                if (ofKind.isEmpty()) {
                    continue
                }
                item("heading-${kind.rawValue}") { SectionHeading(headingOf(kind), modifier = Modifier.gutterPadding()) }
                items(ofKind, key = { record -> "${record.kind.rawValue}-${record.id}" }) { record ->
                    BlockedRow(record, onOpen = { open(record, navigator) }, onUnblock = { viewModel.unblock(record) })
                }
            }
        }
    }
}

/** One blocked record: its cover or picture, its name, and Unblock. */
@Composable
private fun BlockedRow(record: BlockedRecord, onOpen: () -> Unit, onUnblock: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().pressClickable(onOpen).padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (record.kind == RecordKind.video) {
            RemoteImage(record.imageUrl, Modifier.size(CoverWidth, CoverWidth / LANDSCAPE_RATIO).clip(VeilShapes.card), ContentScale.Crop)
        } else {
            Avatar(record.imageUrl, Modifier.size(PictureSize))
        }
        Column(Modifier.weight(1f)) {
            Text(record.name, style = VeilType.cardTitle, color = VeilColors.content, maxLines = 2, overflow = TextOverflow.Ellipsis)
            val byline = record.video?.byline
            if (byline != null) {
                Text(byline, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted, maxLines = 1)
            }
        }
        SecondaryButton("Unblock", onClick = onUnblock)
    }
}

/** The section heading of a kind. */
private fun headingOf(kind: RecordKind): String {
    if (kind == RecordKind.performer) {
        return "Performers"
    }
    if (kind == RecordKind.studio) {
        return "Studios"
    }
    return "Videos"
}

/** Opens a blocked record's page. */
private fun open(record: BlockedRecord, navigator: AppNavigator) {
    if (record.kind == RecordKind.performer) {
        navigator.openPerformer(record.id)
        return
    }
    if (record.kind == RecordKind.studio) {
        navigator.openStudio(record.id)
        return
    }
    navigator.openVideo(record.id)
}
