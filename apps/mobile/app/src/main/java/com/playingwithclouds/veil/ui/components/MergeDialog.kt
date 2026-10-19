package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilMenu
import com.playingwithclouds.veil.ui.design.VeilMenuItem
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** The kinds of record that can be merged by hand. */
enum class MergeKind(val noun: String) {
    PERFORMER("performer"),
    STUDIO("studio"),
    TAG("tag"),
}

/** Most candidates shown while picking the record to keep. */
private const val CANDIDATE_LIMIT = 8

/** Pause after typing before the candidates are searched. */
private const val SEARCH_DEBOUNCE_MILLISECONDS = 200L

/** Fewest letters searched for. */
private const val MIN_SEARCH_LENGTH = 2

/** A ⋮ in a page's top bar with "Merge into…", which opens the [MergeDialog]. */
@Composable
fun MergeMenuButton(kind: MergeKind, recordId: String, recordName: String, onMerged: (String) -> Unit) {
    var menuOpen by remember { mutableStateOf(false) }
    var dialogOpen by remember { mutableStateOf(false) }
    Box {
        IconTap(VeilIcons.More, contentDescription = "More", onClick = { menuOpen = true }, tint = VeilColors.content)
        VeilMenu(expanded = menuOpen, onDismissRequest = { menuOpen = false }) {
            VeilMenuItem("Merge into...", icon = VeilIcons.Merge, onClick = {
                menuOpen = false
                dialogOpen = true
            })
        }
    }
    if (dialogOpen) {
        MergeDialog(kind, recordId, recordName, onMerged = onMerged, onDismiss = { dialogOpen = false })
    }
}

/**
 * Merges the record whose page is open into another one the user picks: search for the record to
 * keep, confirm, and everything of the open record (credits, scenes, history, site addresses)
 * moves over. Irreversible, so the confirm step says so. [onMerged] gets the kept record's id.
 */
@Composable
fun MergeDialog(kind: MergeKind, recordId: String, recordName: String, onMerged: (String) -> Unit, onDismiss: () -> Unit) {
    var search by remember { mutableStateOf("") }
    var candidates by remember { mutableStateOf<List<EntityRef>>(emptyList()) }
    var chosen by remember { mutableStateOf<EntityRef?>(null) }
    var merging by remember { mutableStateOf(false) }
    var failure by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    LaunchedEffect(search) {
        val text = search.trim()
        if (text.length < MIN_SEARCH_LENGTH) {
            candidates = emptyList()
            return@LaunchedEffect
        }
        delay(SEARCH_DEBOUNCE_MILLISECONDS)
        val found = runCatching { mergeCandidates(kind, text) }.getOrDefault(emptyList())
        candidates = found.filter { candidate -> candidate.id != recordId }
    }

    val survivor = chosen
    if (survivor != null) {
        val confirm = {
            merging = true
            failure = null
            scope.launch {
                try {
                    onMerged(EntityRepository.merge(survivor.id, recordId))
                } catch (error: CancellationException) {
                    throw error
                } catch (error: Exception) {
                    failure = error.displayMessage()
                }
                merging = false
            }
            Unit
        }
        MergeConfirmDialog(kind, recordName, survivor.name, merging, failure, onConfirm = confirm, onBack = { chosen = null })
        return
    }
    VeilDialog(
        title = "Merge into another ${kind.noun}",
        onDismissRequest = onDismiss,
        buttons = { SecondaryButton("Cancel", onClick = onDismiss) },
    ) {
        Text("Pick the ${kind.noun} to keep. $recordName will become part of it.", color = VeilColors.contentMuted, style = MaterialTheme.typography.bodyMedium)
        VeilTextField(search, { text -> search = text }, placeholder = "Search by name", leadingIcon = VeilIcons.Search, modifier = Modifier.fillMaxWidth())
        LazyColumn {
            items(candidates, key = { candidate -> candidate.id }) { candidate ->
                Text(
                    candidate.name,
                    Modifier.fillMaxWidth().pressClickable { chosen = candidate }.padding(vertical = VeilSpacing.medium),
                    style = MaterialTheme.typography.bodyLarge,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

/** The last step: says what moves where and that it cannot be undone. */
@Composable
private fun MergeConfirmDialog(
    kind: MergeKind,
    mergedName: String,
    survivorName: String,
    merging: Boolean,
    failure: String?,
    onConfirm: () -> Unit,
    onBack: () -> Unit,
) {
    VeilDialog(
        title = "Merge $mergedName?",
        onDismissRequest = onBack,
        buttons = {
            SecondaryButton("Back", onClick = onBack, enabled = !merging)
            PrimaryButton("Merge", onClick = onConfirm, enabled = !merging)
        },
    ) {
        Text(
            "Everything of $mergedName (videos, credits, history, site addresses) moves to the ${kind.noun} $survivorName, " +
                "and $mergedName is deleted. This cannot be undone.",
            style = MaterialTheme.typography.bodyMedium,
            color = VeilColors.contentMuted,
        )
        if (failure != null) {
            Text(failure, style = MaterialTheme.typography.bodySmall, color = VeilColors.error)
        }
    }
}

/** Records of [kind] whose name matches [search]. */
private suspend fun mergeCandidates(kind: MergeKind, search: String): List<EntityRef> {
    when (kind) {
        MergeKind.PERFORMER -> {
            return EntityRepository.performers(search, null, CANDIDATE_LIMIT, 0).map { performer -> EntityRef(performer.id, performer.name, performer.imagePath) }
        }
        MergeKind.STUDIO -> {
            return EntityRepository.studios(search, null, CANDIDATE_LIMIT, 0).map { studio -> EntityRef(studio.id, studio.name, studio.imagePath) }
        }
        MergeKind.TAG -> {
            return EntityRepository.tags(search, CANDIDATE_LIMIT, 0).map { tag -> EntityRef(tag.id, tag.name, null) }
        }
    }
}
