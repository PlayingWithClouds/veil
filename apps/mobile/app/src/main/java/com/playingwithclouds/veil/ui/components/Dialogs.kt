package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.VeilDialog
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Asks for one line of text. */
@Composable
fun TextInputDialog(
    title: String,
    label: String,
    confirmLabel: String,
    initialValue: String = "",
    onConfirm: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    var text by rememberSaveable { mutableStateOf(initialValue) }
    VeilDialog(
        title = title,
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("Cancel", onClick = onDismiss)
            PrimaryButton(confirmLabel, onClick = { onConfirm(text.trim()) }, enabled = text.isNotBlank())
        },
    ) {
        VeilTextField(
            value = text,
            onValueChange = { newText -> text = newText },
            placeholder = label,
            modifier = Modifier.fillMaxWidth(),
        )
    }
}

/** Asks to confirm a destructive action. */
@Composable
fun ConfirmDialog(
    title: String,
    message: String,
    confirmLabel: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    VeilDialog(
        title = title,
        onDismissRequest = onDismiss,
        buttons = {
            SecondaryButton("Cancel", onClick = onDismiss)
            PrimaryButton(confirmLabel, onClick = onConfirm)
        },
    ) {
        Text(message, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
    }
}
