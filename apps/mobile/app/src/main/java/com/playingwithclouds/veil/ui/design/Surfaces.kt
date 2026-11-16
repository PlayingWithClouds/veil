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
import android.os.Build
import android.view.WindowManager
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.window.DialogWindowProvider
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** A flat rounded panel on the dark surface: no elevation, no tint. */
@Composable
fun VeilCard(
    modifier: Modifier = Modifier,
    verticalArrangement: Arrangement.Vertical = Arrangement.spacedBy(VeilSpacing.small),
    content: @Composable ColumnScope.() -> Unit,
) {
    Column(
        modifier.fillMaxWidth().clip(VeilShapes.panel).background(VeilColors.surface).padding(VeilSpacing.large),
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
                .clip(VeilShapes.sheet)
                .background(VeilColors.surfaceHigh)
                .padding(VeilSpacing.extraLarge),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            Text(title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            Column(Modifier.weight(1f, fill = false), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium), content = content)
            Row(
                Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small, Alignment.End),
                verticalAlignment = Alignment.CenterVertically,
                content = buttons,
            )
        }
    }
}

/** Top corners of a bottom sheet, matching [VeilShapes.sheet]. */
private val SheetTopShape = RoundedCornerShape(topStart = 28.dp, topEnd = 28.dp)

/**
 * A bottom sheet with large top corners and a slim handle, as frosted glass: where the phone can
 * blur across windows (Android 12+), the page behind is blurred and the sheet is a dark tint over
 * it; elsewhere the sheet is the near-opaque glass fill.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VeilBottomSheet(onDismissRequest: () -> Unit, content: @Composable ColumnScope.() -> Unit) {
    val blurs = crossWindowBlurAvailable()
    var container = VeilColors.glassFallback
    if (blurs) {
        container = VeilColors.glassTint
    }
    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
        sheetState = rememberModalBottomSheetState(),
        shape = SheetTopShape,
        containerColor = container,
        contentColor = VeilColors.content,
        tonalElevation = 0.dp,
        scrimColor = VeilColors.scrim,
        dragHandle = { SheetHandle() },
    ) {
        if (blurs) {
            BlurBehindWindow()
        }
        Column(Modifier.navigationBarsPadding(), content = content)
    }
}

/** How strongly a sheet blurs the page behind it. */
private val SheetBlurRadius = 24.dp

/** Whether this phone blurs behind windows: Android 12+ with cross-window blur on (not in battery saver, say). */
@Composable
private fun crossWindowBlurAvailable(): Boolean {
    if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
        return false
    }
    val windowManager = LocalContext.current.getSystemService(WindowManager::class.java) ?: return false
    return windowManager.isCrossWindowBlurEnabled
}

/** Blurs what's behind the window this is composed in (a sheet's or a dialog's). */
@Composable
private fun BlurBehindWindow() {
    val view = LocalView.current
    val radius = with(LocalDensity.current) { SheetBlurRadius.roundToPx() }
    DisposableEffect(view) {
        val window = ((view as? DialogWindowProvider) ?: (view.parent as? DialogWindowProvider))?.window
        if (window != null && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            window.addFlags(WindowManager.LayoutParams.FLAG_BLUR_BEHIND)
            window.attributes = window.attributes.also { attributes -> attributes.blurBehindRadius = radius }
        }
        onDispose {}
    }
}

/** The grabber at the top of a sheet. */
@Composable
private fun SheetHandle() {
    Box(
        Modifier
            .padding(top = VeilSpacing.medium, bottom = VeilSpacing.medium)
            .size(width = 40.dp, height = 5.dp)
            .clip(VeilShapes.capsule)
            .background(VeilColors.contentFaint),
    )
}

/** A popup menu on the raised surface with large corners. */
@Composable
fun VeilMenu(expanded: Boolean, onDismissRequest: () -> Unit, content: @Composable ColumnScope.() -> Unit) {
    DropdownMenu(
        expanded = expanded,
        onDismissRequest = onDismissRequest,
        shape = VeilShapes.panel,
        containerColor = VeilColors.surfaceHigh,
        tonalElevation = 0.dp,
        shadowElevation = 12.dp,
        content = content,
    )
}

/**
 * One entry of a [VeilMenu], with an optional leading icon, a check mark when [selected] and a
 * [NoticeDot] when [marked].
 */
@Composable
fun VeilMenuItem(label: String, onClick: () -> Unit, icon: ImageVector? = null, selected: Boolean = false, marked: Boolean = false) {
    var leading: (@Composable () -> Unit)? = null
    if (icon != null) {
        leading = { Icon(icon, contentDescription = null, tint = VeilColors.content, modifier = Modifier.size(20.dp)) }
    }
    var trailing: (@Composable () -> Unit)? = null
    if (marked) {
        trailing = { NoticeDot() }
    }
    if (selected) {
        trailing = { Icon(VeilIcons.Check, contentDescription = null, tint = VeilColors.accent, modifier = Modifier.size(18.dp)) }
    }
    DropdownMenuItem(
        text = { Text(label, style = MaterialTheme.typography.bodyLarge, color = VeilColors.content) },
        leadingIcon = leading,
        trailingIcon = trailing,
        onClick = onClick,
    )
}

/** One tappable row of an action sheet: icon, label, and an accent tint while [active]. */
@Composable
fun SheetAction(label: String, icon: ImageVector, onClick: () -> Unit, active: Boolean = false) {
    var tint = VeilColors.content
    if (active) {
        tint = VeilColors.accent
    }
    Row(
        Modifier
            .fillMaxWidth()
            .pressClickable(onClick)
            .padding(horizontal = VeilSpacing.extraLarge, vertical = VeilSpacing.medium),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, contentDescription = null, tint = tint, modifier = Modifier.size(22.dp))
        Text(label, style = MaterialTheme.typography.bodyLarge, color = tint)
    }
}

/** Snackbar messages as a dark capsule floating at the bottom. */
@Composable
fun VeilSnackbarHost(hostState: SnackbarHostState, modifier: Modifier = Modifier) {
    SnackbarHost(hostState, modifier) { data ->
        Box(Modifier.fillMaxWidth().padding(VeilSpacing.large), contentAlignment = Alignment.Center) {
            Text(
                data.visuals.message,
                modifier = Modifier
                    .clip(VeilShapes.capsule)
                    .background(VeilColors.surfaceHigh)
                    .padding(horizontal = VeilSpacing.large, vertical = VeilSpacing.medium),
                style = MaterialTheme.typography.bodyMedium,
                color = VeilColors.content,
            )
        }
    }
}
