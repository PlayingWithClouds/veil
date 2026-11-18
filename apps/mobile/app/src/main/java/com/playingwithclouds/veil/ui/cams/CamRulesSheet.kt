package com.playingwithclouds.veil.ui.cams

import android.Manifest
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
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
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.cams.CamRuleStore
import com.playingwithclouds.veil.cams.CamSceneRule
import com.playingwithclouds.veil.cams.RecordStop
import com.playingwithclouds.veil.scenes.SceneLabels
import com.playingwithclouds.veil.scenes.SceneModelStore
import com.playingwithclouds.veil.scenes.ready
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** How long a scene must be seen, in seconds. */
private val SEEN_CHOICES = listOf(5, 10, 20, 30)

/** How long a rule's recording runs, in minutes. */
private val RECORD_CHOICES = listOf(5, 10, 30)

/** How long the scenes must be gone before a waiting recording ends, in minutes. */
private val GONE_CHOICES = listOf(2, 5)

/** How long a rule rests in a room after firing, in minutes. */
private val COOLDOWN_CHOICES = listOf(5, 15, 30, 60)

/**
 * The scene rules of one room ([roomId], with the every-room ones) or, without a room, all of
 * them (Settings → Cams): each with its switch, Edit and Delete, and New rule, which opens the
 * editor in the same sheet. New rules here are for [roomId], or for every room.
 */
@Composable
fun CamRulesSheet(roomId: String?, roomName: String?, onDismiss: () -> Unit) {
    val rules by CamRuleStore.rules.collectAsState()
    var editing by remember { mutableStateOf<CamSceneRule?>(null) }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(
            Modifier.verticalScroll(rememberScrollState()).gutterPadding().padding(bottom = VeilSpacing.huge),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            val rule = editing
            if (rule != null) {
                RuleEditor(rule, roomId, roomName, onDone = { editing = null })
                return@Column
            }
            var title = "Scene rules"
            if (roomName != null) {
                title = "Scene rules · $roomName"
            }
            Text(title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            val shown = rules.filter { candidate -> roomId == null || candidate.roomId == null || candidate.roomId == roomId }
            if (shown.isEmpty()) {
                Text(
                    "When the scene model sees a scene you pick for a while, get a discreet notification or start recording.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = VeilColors.contentMuted,
                )
            }
            for (existing in shown) {
                RuleRow(existing, onEdit = { editing = existing })
            }
            PrimaryButton("New rule", onClick = { editing = CamSceneRule(roomId = roomId, roomName = roomName) }, icon = VeilIcons.Plus)
        }
    }
}

/** One rule: what it watches for and does, its switch, Edit and Delete. */
@Composable
private fun RuleRow(rule: CamSceneRule, onEdit: () -> Unit) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(VeilSpacing.hairline)) {
            Text(rule.tags.sorted().joinToString(", "), style = MaterialTheme.typography.bodyLarge, color = VeilColors.content)
            Text(ruleSummary(rule), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
            TextAction("Edit", onClick = onEdit)
        }
        VeilSwitch(checked = rule.enabled, onCheckedChange = { on -> CamRuleStore.save(rule.copy(enabled = on)) })
        IconTap(VeilIcons.Delete, contentDescription = "Delete the rule", onClick = { CamRuleStore.delete(rule.id) })
    }
}

/**
 * The editor: the scenes (chips grouped as the model's heads), how long they must be seen, what
 * happens (notify, record and for how long), the rest after firing, and for a room's rule whether
 * it's for that room or every room. Saving asks for the notification permission when needed.
 */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun RuleEditor(start: CamSceneRule, roomId: String?, roomName: String?, onDone: () -> Unit) {
    val modelState by SceneModelStore.state.collectAsState()
    val groups = remember(modelState) { SceneLabels.pickerGroups(modelState.ready?.manifest) }
    var rule by remember { mutableStateOf(start) }
    val askPermission = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { _ -> onDone() }

    /** Saves the rule, asking for the notification permission first when it notifies. */
    fun save() {
        CamRuleStore.save(rule)
        if (rule.notify && Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            askPermission.launch(Manifest.permission.POST_NOTIFICATIONS)
            return
        }
        onDone()
    }

    Text("Scene rule", style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
    if (roomId != null && roomName != null) {
        SectionHeading("For")
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            FeedChip(roomName, onClick = { rule = rule.copy(roomId = roomId, roomName = roomName) }, selected = rule.roomId != null)
            FeedChip("Every room", onClick = { rule = rule.copy(roomId = null, roomName = null) }, selected = rule.roomId == null)
        }
    }
    SectionHeading("Seen for")
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        for (seconds in SEEN_CHOICES) {
            FeedChip("$seconds s", onClick = { rule = rule.copy(seenSeconds = seconds) }, selected = rule.seenSeconds == seconds)
        }
    }
    SectionHeading("Then")
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        FeedChip("Notify", onClick = { rule = rule.copy(notify = !rule.notify) }, selected = rule.notify, icon = VeilIcons.Follow)
        FeedChip("Record", onClick = { rule = rule.copy(record = !rule.record) }, selected = rule.record, icon = VeilIcons.Record)
    }
    if (rule.record) {
        RecordChoices(rule, onChange = { changed -> rule = changed })
    }
    SectionHeading("Then rest for")
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        for (minutes in COOLDOWN_CHOICES) {
            FeedChip("$minutes min", onClick = { rule = rule.copy(cooldownMinutes = minutes) }, selected = rule.cooldownMinutes == minutes)
        }
    }
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        PrimaryButton("Save", onClick = ::save, enabled = rule.tags.isNotEmpty() && (rule.notify || rule.record))
        SecondaryButton("Cancel", onClick = onDone)
    }
    for (group in groups) {
        SectionHeading(group.title)
        FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            for (tag in group.tags) {
                FeedChip(tag, onClick = { rule = rule.copy(tags = toggled(rule.tags, tag)) }, selected = tag in rule.tags)
            }
        }
    }
}

/** How long the rule records: a fixed time, or until its scenes are gone for a while. */
@Composable
private fun RecordChoices(rule: CamSceneRule, onChange: (CamSceneRule) -> Unit) {
    SectionHeading("Record")
    FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        for (minutes in RECORD_CHOICES) {
            val selected = rule.recordStop == RecordStop.FIXED && rule.recordMinutes == minutes
            FeedChip("$minutes min", onClick = { onChange(rule.copy(recordStop = RecordStop.FIXED, recordMinutes = minutes)) }, selected = selected)
        }
        for (minutes in GONE_CHOICES) {
            val selected = rule.recordStop == RecordStop.UNTIL_GONE && rule.goneMinutes == minutes
            FeedChip("Until gone $minutes min", onClick = { onChange(rule.copy(recordStop = RecordStop.UNTIL_GONE, goneMinutes = minutes)) }, selected = selected)
        }
    }
}

/** [tags] with [tag] added, or taken out when it's there. */
private fun toggled(tags: Set<String>, tag: String): Set<String> {
    if (tag in tags) {
        return tags - tag
    }
    return tags + tag
}
