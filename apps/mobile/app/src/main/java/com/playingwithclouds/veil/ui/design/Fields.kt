package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Corner radius of text fields. */
val FieldShape = RoundedCornerShape(16.dp)

/**
 * A filled dark text field: an optional static label above, the input in a rounded box with an
 * optional leading icon and trailing content, and an optional note below. No underline, no
 * floating label.
 */
@Composable
fun VeilTextField(
    value: String,
    onValueChange: (String) -> Unit,
    modifier: Modifier = Modifier,
    label: String? = null,
    placeholder: String? = null,
    supportingText: String? = null,
    leadingIcon: ImageVector? = null,
    trailing: (@Composable () -> Unit)? = null,
    singleLine: Boolean = true,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default,
    keyboardActions: KeyboardActions = KeyboardActions.Default,
) {
    Column(modifier, verticalArrangement = Arrangement.spacedBy(6.dp)) {
        if (label != null) {
            Text(label, style = MaterialTheme.typography.labelMedium, color = VeilColors.contentMuted, modifier = Modifier.padding(start = 4.dp))
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = singleLine,
            textStyle = MaterialTheme.typography.bodyLarge.copy(color = VeilColors.content),
            cursorBrush = SolidColor(VeilColors.accent),
            keyboardOptions = keyboardOptions,
            keyboardActions = keyboardActions,
            modifier = Modifier.fillMaxWidth(),
            decorationBox = { innerTextField ->
                FieldBox(value.isEmpty(), placeholder, leadingIcon, trailing, innerTextField)
            },
        )
        if (supportingText != null) {
            Text(supportingText, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentFaint, modifier = Modifier.padding(start = 4.dp))
        }
    }
}

/** The rounded box around a field's input: icon, input or placeholder, trailing content. */
@Composable
private fun FieldBox(
    isEmpty: Boolean,
    placeholder: String?,
    leadingIcon: ImageVector?,
    trailing: (@Composable () -> Unit)?,
    innerTextField: @Composable () -> Unit,
) {
    var startPadding = 16.dp
    if (leadingIcon != null) {
        startPadding = 12.dp
    }
    var endPadding = 16.dp
    if (trailing != null) {
        endPadding = 4.dp
    }
    Row(
        Modifier
            .fillMaxWidth()
            .heightIn(min = 50.dp)
            .clip(FieldShape)
            .background(VeilColors.surfaceHigh)
            .padding(start = startPadding, end = endPadding),
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (leadingIcon != null) {
            Icon(leadingIcon, contentDescription = null, tint = VeilColors.contentMuted, modifier = Modifier.size(20.dp))
        }
        Box(Modifier.weight(1f).padding(vertical = 14.dp), contentAlignment = Alignment.CenterStart) {
            if (isEmpty && placeholder != null) {
                Text(placeholder, style = MaterialTheme.typography.bodyLarge, color = VeilColors.contentFaint, maxLines = 1)
            }
            innerTextField()
        }
        if (trailing != null) {
            trailing()
        }
    }
}
