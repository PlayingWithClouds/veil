package com.playingwithclouds.veil.ui.design

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.playingwithclouds.veil.ui.theme.VeilColors

/** Corner radius of cards. */
val CardShape = RoundedCornerShape(20.dp)

/** Corner radius of dialogs and sheets. */
private val SheetCorner = 28.dp

/** A flat rounded panel on the dark surface: no elevation, no tint. */
@Composable
fun VeilCard(
    modifier: Modifier = Modifier,
    verticalArrangement: Arrangement.Vertical = Arrangement.spacedBy(8.dp),
    content: @Composable ColumnScope.() -> Unit,
) {
    Column(
        modifier.fillMaxWidth().clip(CardShape).background(VeilColors.surface).padding(16.dp),
        verticalArrangement = verticalArrangement,
        content = content,
    )
}

/**
 * A centered rounded dialog: bold title, then [content] (scrolls itself if long), then the
 * buttons right-aligned, typically a [SecondaryButton] and a [PrimaryButton].
 */
@Composable
fun VeilDialog(
    title: String,
    onDismissRequest: () -> Unit,
    buttons: @Composable RowScope.() -> Unit,
    content: @Composable ColumnScope.() -> Unit,
) {
    Dialog(onDismissRequest = onDismissRequest) {
        Column(
            Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(SheetCorner))
                .background(VeilColors.surfaceHigh)
                .padding(24.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Text(title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            Column(Modifier.weight(1f, fill = false), verticalArrangement = Arrangement.spacedBy(12.dp), content = content)
            Row(
                Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.End),
                verticalAlignment = Alignment.CenterVertically,
                content = buttons,
            )
        }
    }
}

/** A bottom sheet with large top corners, a slim handle and the dark surface. */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VeilBottomSheet(onDismissRequest: () -> Unit, content: @Composable ColumnScope.() -> Unit) {
    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
        sheetState = rememberModalBottomSheetState(),
        shape = RoundedCornerShape(topStart = SheetCorner, topEnd = SheetCorner),
        containerColor = VeilColors.surface,
        contentColor = VeilColors.content,
        tonalElevation = 0.dp,
        scrimColor = VeilColors.canvas.copy(alpha = 0.6f),
        dragHandle = { SheetHandle() },
    ) {
        Column(Modifier.navigationBarsPadding(), content = content)
    }
}

/** The grabber at the top of a sheet. */
@Composable
private fun SheetHandle() {
    Box(Modifier.padding(top = 10.dp, bottom = 14.dp).size(width = 40.dp, height = 5.dp).clip(CircleShape).background(VeilColors.contentFaint))
}

/** A popup menu on the raised surface with large corners. */
@Composable
fun VeilMenu(expanded: Boolean, onDismissRequest: () -> Unit, content: @Composable ColumnScope.() -> Unit) {
    DropdownMenu(
        expanded = expanded,
        onDismissRequest = onDismissRequest,
        shape = RoundedCornerShape(18.dp),
        containerColor = VeilColors.surfaceHigh,
        tonalElevation = 0.dp,
        shadowElevation = 12.dp,
        content = content,
    )
}

/** One entry of a [VeilMenu], with an optional leading icon and a check mark when [selected]. */
@Composable
fun VeilMenuItem(label: String, onClick: () -> Unit, icon: ImageVector? = null, selected: Boolean = false) {
    var leading: (@Composable () -> Unit)? = null
    if (icon != null) {
        leading = { Icon(icon, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(20.dp)) }
    }
    var trailing: (@Composable () -> Unit)? = null
    if (selected) {
        trailing = { Icon(VeilIcons.Check, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(18.dp)) }
    }
    DropdownMenuItem(
        text = { Text(label, style = MaterialTheme.typography.bodyLarge, color = VeilColors.content) },
        leadingIcon = leading,
        trailingIcon = trailing,
        onClick = onClick,
    )
}

/** Snackbar messages as a dark capsule floating at the bottom. */
@Composable
fun VeilSnackbarHost(hostState: SnackbarHostState, modifier: Modifier = Modifier) {
    SnackbarHost(hostState, modifier) { data ->
        Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
            Text(
                data.visuals.message,
                modifier = Modifier.clip(CircleShape).background(VeilColors.surfaceHigh).padding(horizontal = 20.dp, vertical = 14.dp),
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.content,
            )
        }
    }
}
