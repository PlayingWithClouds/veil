package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
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
import androidx.compose.ui.text.style.TextOverflow
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilCheck
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/**
 * The user's collections, those in [memberOf] ticked; a tap adds to or removes from one through
 * [onToggle], and a new one (holding the media) can be created on the spot through [onCreate].
 */
@Composable
fun CollectionPickerDialog(
    collections: List<CollectionSummary>,
    memberOf: Set<String>,
    onToggle: (CollectionSummary, Boolean) -> Unit,
    onCreate: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    var creating by remember { mutableStateOf(false) }

    if (creating) {
        TextInputDialog(
            title = "New collection",
            label = "Name",
            confirmLabel = "Create",
            onConfirm = { name ->
                creating = false
                onCreate(name)
            },
            onDismiss = { creating = false },
        )
        return
    }
    VeilDialog(
        title = "Add to collection",
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("New collection", onClick = { creating = true })
            PrimaryButton("Done", onClick = onDismiss)
        },
    ) {
        if (collections.isEmpty()) {
            Text("You have no collections yet.", color = VeilColors.contentMuted)
            return@VeilDialog
        }
        LazyColumn {
            items(collections, key = { collection -> collection.id }) { collection ->
                val included = collection.id in memberOf
                Row(
                    Modifier.fillMaxWidth().pressClickable { onToggle(collection, !included) }.padding(vertical = VeilSpacing.small),
                    horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    VeilCheck(checked = included)
                    Text(collection.name, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
            }
        }
    }
}
