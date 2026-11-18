package com.playingwithclouds.veil.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.NoticeDot
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.update.UpdateChecker

/** The drawer's sections, in order, as the desktop sidebar has them. */
enum class DestinationGroup(val label: String) {
    WATCH("Watch"),
    LIBRARY("Library"),
    YOU("You"),
}

/**
 * A top-level page in the drawer, in menu order (apps/web/src/lib/navigation.ts). Those with a
 * [tab] are tabs of the bottom bar too.
 */
enum class Destination(val label: String, val icon: ImageVector, val group: DestinationGroup, val tab: Tab? = null) {
    HOME("Home", VeilIcons.Home, DestinationGroup.WATCH, Tab.HOME),
    CAMS("Cams", VeilIcons.Cams, DestinationGroup.WATCH, Tab.CAMS),
    SHORTS("Shorties", VeilIcons.Shorts, DestinationGroup.WATCH),
    GALLERIES("Galleries", VeilIcons.Galleries, DestinationGroup.WATCH),
    AUDIO("Audio", VeilIcons.Audio, DestinationGroup.WATCH),
    STORIES("Stories", VeilIcons.Stories, DestinationGroup.WATCH),
    CATEGORIES("Categories", VeilIcons.Categories, DestinationGroup.LIBRARY),
    PERFORMERS("Performers", VeilIcons.Performers, DestinationGroup.LIBRARY),
    STUDIOS("Studios", VeilIcons.Studios, DestinationGroup.LIBRARY),
    LIBRARY("Library", VeilIcons.Library, DestinationGroup.YOU, Tab.LIBRARY),
    TASTE("Your taste", VeilIcons.Heart, DestinationGroup.YOU),
    PLUGINS("Plugins", VeilIcons.Plugins, DestinationGroup.YOU),
}

/** Width of the drawer panel. */
private val DrawerWidth = 300.dp

/** Height of one drawer entry. */
private val EntryHeight = 48.dp

/** Size of an entry's icon. */
private val EntryIconSize = 22.dp

/**
 * The side menu: the wordmark, every top-level page in its section with the [current] one
 * highlighted, and Settings at the bottom (with a dot while an update waits). Choosing an entry
 * calls [onDestination] or [onSettings]; closing the drawer is the caller's.
 */
@Composable
fun MainDrawer(
    current: Destination?,
    onDestination: (Destination) -> Unit,
    onSettings: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val updateMarked by UpdateChecker.hasUnseenUpdate.collectAsStateWithLifecycle()
    Column(
        modifier
            .width(DrawerWidth)
            .fillMaxHeight()
            .background(VeilColors.elevated)
            .statusBarsPadding()
            .navigationBarsPadding()
            .padding(horizontal = VeilSpacing.medium),
    ) {
        Text(
            "Veil",
            Modifier.padding(start = VeilSpacing.medium, top = VeilSpacing.large, bottom = VeilSpacing.large),
            style = VeilType.wordmark,
            color = VeilColors.content,
        )
        Column(
            Modifier.weight(1f).verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            for (group in DestinationGroup.entries) {
                DrawerSection(group, current, onDestination)
            }
        }
        DrawerEntry("Settings", VeilIcons.Settings, selected = false, marked = updateMarked, onClick = onSettings)
        Spacer(Modifier.height(VeilSpacing.small))
    }
}

/** A section's small heading over its entries. */
@Composable
private fun DrawerSection(group: DestinationGroup, current: Destination?, onDestination: (Destination) -> Unit) {
    Column {
        Text(
            group.label.uppercase(),
            Modifier.padding(start = VeilSpacing.medium, bottom = VeilSpacing.extraSmall),
            style = MaterialTheme.typography.labelSmall,
            color = VeilColors.contentFaint,
        )
        for (destination in Destination.entries.filter { entry -> entry.group == group }) {
            DrawerEntry(
                destination.label,
                destination.icon,
                selected = destination == current,
                onClick = { onDestination(destination) },
            )
        }
    }
}

/** One entry: icon and label, on a soft fill when [selected]; [marked] adds a notice dot. */
@Composable
private fun DrawerEntry(label: String, icon: ImageVector, selected: Boolean, onClick: () -> Unit, marked: Boolean = false) {
    var entryModifier = Modifier.fillMaxWidth().height(EntryHeight).clip(VeilShapes.small)
    var tint = VeilColors.contentMuted
    var weight = FontWeight.Normal
    if (selected) {
        entryModifier = entryModifier.background(VeilColors.glassSelection)
        tint = VeilColors.content
        weight = FontWeight.SemiBold
    }
    Row(
        entryModifier.pressClickable(onClick).padding(horizontal = VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large),
    ) {
        Icon(icon, contentDescription = null, tint = tint, modifier = Modifier.size(EntryIconSize))
        Text(label, Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge, fontWeight = weight, color = tint)
        if (marked) {
            NoticeDot()
        }
    }
}
