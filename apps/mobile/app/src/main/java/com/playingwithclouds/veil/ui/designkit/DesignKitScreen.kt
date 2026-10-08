package com.playingwithclouds.veil.ui.designkit

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.design.Badge
import com.playingwithclouds.veil.ui.design.LoadingBar
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.PillRow
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.ProgressBar
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SegmentedControl
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilCheck
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing

/** One named value of the design system shown in the catalog. */
private data class Sample<T>(val name: String, val value: T)

private val colorSamples = listOf(
    Sample("canvas", VeilColors.canvas),
    Sample("elevated", VeilColors.elevated),
    Sample("surface", VeilColors.surface),
    Sample("surfaceHigh", VeilColors.surfaceHigh),
    Sample("content", VeilColors.content),
    Sample("contentMuted", VeilColors.contentMuted),
    Sample("contentFaint", VeilColors.contentFaint),
    Sample("accent", VeilColors.accent),
    Sample("accentSoft", VeilColors.accentSoft),
    Sample("info", VeilColors.info),
    Sample("success", VeilColors.success),
    Sample("warning", VeilColors.warning),
    Sample("error", VeilColors.error),
)

private val spacingSamples = listOf(
    Sample("hairline", VeilSpacing.hairline),
    Sample("extraSmall", VeilSpacing.extraSmall),
    Sample("small", VeilSpacing.small),
    Sample("medium", VeilSpacing.medium),
    Sample("large · gutter", VeilSpacing.large),
    Sample("extraLarge", VeilSpacing.extraLarge),
    Sample("huge", VeilSpacing.huge),
)

private val shapeSamples = listOf<Sample<Shape>>(
    Sample("badge", VeilShapes.badge),
    Sample("small", VeilShapes.small),
    Sample("card", VeilShapes.card),
    Sample("panel", VeilShapes.panel),
    Sample("sheet", VeilShapes.sheet),
    Sample("capsule", VeilShapes.capsule),
)

/**
 * A catalog of the design system: tokens and every `ui/design` component in its states, to check
 * changes to the kit on a device. Opened from Settings.
 */
@Composable
fun DesignKitScreen(navigator: AppNavigator) {
    val typeSamples = listOf(
        Sample("headlineLarge", MaterialTheme.typography.headlineLarge),
        Sample("titleLarge", MaterialTheme.typography.titleLarge),
        Sample("titleMedium", MaterialTheme.typography.titleMedium),
        Sample("bodyLarge", MaterialTheme.typography.bodyLarge),
        Sample("bodyMedium", MaterialTheme.typography.bodyMedium),
        Sample("bodySmall", MaterialTheme.typography.bodySmall),
        Sample("labelLarge", MaterialTheme.typography.labelLarge),
        Sample("labelMedium", MaterialTheme.typography.labelMedium),
        Sample("labelSmall", MaterialTheme.typography.labelSmall),
    )
    Scaffold(topBar = { VeilTopBar("Design kit", onBack = navigator::back) }) { padding ->
        LazyColumn(Modifier.padding(padding), contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
            item { SectionHeading("Colours", Modifier.gutterPadding()) }
            item { ColorSwatches() }
            item { SectionHeading("Type", Modifier.gutterPadding()) }
            for (sample in typeSamples) {
                item { TypeSample(sample) }
            }
            item { SectionHeading("Spacing", Modifier.gutterPadding()) }
            for (sample in spacingSamples) {
                item { SpacingSample(sample) }
            }
            item { SectionHeading("Shapes", Modifier.gutterPadding()) }
            item { ShapeSamples() }
            item { SectionHeading("Buttons", Modifier.gutterPadding()) }
            item { ButtonSamples() }
            item { SectionHeading("Pills", Modifier.gutterPadding()) }
            item { PillSamples() }
            item { SectionHeading("Controls", Modifier.gutterPadding()) }
            item { ControlSamples() }
            item { SectionHeading("Surfaces (heading with link)", Modifier.gutterPadding(), onClick = {}) }
            item { SurfaceSamples() }
        }
    }
}

/** Every colour token as a labelled swatch. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun ColorSwatches() {
    FlowRow(
        Modifier.gutterPadding(),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
    ) {
        for (sample in colorSamples) {
            Column(Modifier.width(96.dp), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(48.dp)
                        .clip(VeilShapes.small)
                        .background(sample.value)
                        .border(1.dp, VeilColors.glassEdge, VeilShapes.small),
                )
                Text(sample.name, style = MaterialTheme.typography.labelSmall, color = VeilColors.contentMuted)
            }
        }
    }
}

/** One text style, rendered in itself. */
@Composable
private fun TypeSample(sample: Sample<TextStyle>) {
    Text(
        sample.name,
        Modifier.gutterPadding().padding(vertical = VeilSpacing.extraSmall),
        style = sample.value,
        color = VeilColors.content,
    )
}

/** One spacing step as a bar of its width. */
@Composable
private fun SpacingSample(sample: Sample<Dp>) {
    Row(
        Modifier.gutterPadding().padding(vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(Modifier.width(sample.value).height(16.dp).background(VeilColors.accent))
        Text("${sample.name} · ${sample.value.value.toInt()} dp", style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
    }
}

/** Every corner shape on a sample tile. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun ShapeSamples() {
    FlowRow(
        Modifier.gutterPadding(),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
    ) {
        for (sample in shapeSamples) {
            Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                Box(Modifier.size(72.dp).clip(sample.value).background(VeilColors.surfaceHigh))
                Text(sample.name, style = MaterialTheme.typography.labelSmall, color = VeilColors.contentMuted)
            }
        }
    }
}

/** The button family, enabled and disabled. */
@Composable
private fun ButtonSamples() {
    Column(Modifier.gutterPadding(), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            PrimaryButton("Primary", onClick = {})
            SecondaryButton("Secondary", onClick = {}, icon = VeilIcons.Plus)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            PrimaryButton("Disabled", onClick = {}, enabled = false)
            TextAction("Text action", onClick = {})
        }
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            RoundIconButton(VeilIcons.Refresh, contentDescription = "Refresh", onClick = {})
            RoundIconButton(VeilIcons.Settings, contentDescription = "Settings", onClick = {}, enabled = false)
        }
    }
}

/** Choice pills, tag pills and the segmented control. */
@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun PillSamples() {
    var selectedPill by remember { mutableStateOf("Newest") }
    var segment by remember { mutableIntStateOf(0) }
    Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        PillRow(
            listOf("Newest", "Popular", "Longest", "Random", "Shortest"),
            labelOf = { label -> label },
            onClick = { label -> selectedPill = label },
            isSelected = { label -> label == selectedPill },
        )
        FlowRow(Modifier.gutterPadding(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            Pill("With icon", onClick = {}, icon = VeilIcons.Follow)
            TagPill("tag pill", onClick = {})
            TagPill("POV", onClick = {})
            FeedChip("Feed chip", onClick = {}, icon = VeilIcons.Tune)
            FeedChip("Selected", onClick = {}, selected = true)
        }
        SegmentedControl(listOf("Scenes", "Galleries", "Images"), segment, onSelect = { index -> segment = index }, modifier = Modifier.gutterPadding())
    }
}

/** Toggles, progress and activity indicators, and a text field. */
@Composable
private fun ControlSamples() {
    var switched by remember { mutableStateOf(true) }
    var checked by remember { mutableStateOf(true) }
    var text by remember { mutableStateOf("") }
    Column(Modifier.gutterPadding(), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large), verticalAlignment = Alignment.CenterVertically) {
            VeilSwitch(switched, onCheckedChange = { value -> switched = value })
            VeilSwitch(!switched, onCheckedChange = { value -> switched = !value })
            VeilCheck(checked)
            TextAction("Toggle check", onClick = { checked = !checked })
            Spinner(size = 24.dp)
        }
        ProgressBar(progress = 0.4f)
        LoadingBar()
        VeilTextField(text, onValueChange = { value -> text = value }, label = "Label", placeholder = "Placeholder", supportingText = "Supporting text", leadingIcon = VeilIcons.Search)
    }
}

/** A panel, an image badge and sheet rows. */
@Composable
private fun SurfaceSamples() {
    Column(Modifier.gutterPadding(), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        VeilCard {
            Text("VeilCard", style = MaterialTheme.typography.titleMedium, color = VeilColors.content)
            Text("A flat panel on the surface colour.", style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
        }
        Box(Modifier.fillMaxWidth().height(96.dp).clip(VeilShapes.card).background(VeilColors.surfaceHigh)) {
            Badge("12:34", Modifier.align(Alignment.BottomEnd).padding(VeilSpacing.small))
            Badge("NEW", Modifier.align(Alignment.TopStart).padding(VeilSpacing.small), VeilColors.accent, VeilColors.onAccent)
        }
        Column(Modifier.clip(VeilShapes.panel).background(VeilColors.surface)) {
            SheetAction("Sheet action", VeilIcons.Bookmark, onClick = {})
            SheetAction("Active sheet action", VeilIcons.LikeFilled, onClick = {}, active = true)
        }
    }
}
