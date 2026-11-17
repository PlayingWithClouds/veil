package com.playingwithclouds.veil.ui.people

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.data.AttributeFieldType
import com.playingwithclouds.veil.data.EditablePerformer
import com.playingwithclouds.veil.data.MeasurementParts
import com.playingwithclouds.veil.data.PerformerAttribute
import com.playingwithclouds.veil.data.PerformerEditing
import com.playingwithclouds.veil.data.PerformerForm
import com.playingwithclouds.veil.data.PerformerPresets
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilTextField
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import java.time.LocalDate
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.launch

/** Width of the picture preview; it is as tall as a portrait. */
private val PicturePreviewWidth = 96.dp
private const val PICTURE_RATIO = 2f / 3f

/** How many matching suggestions show under a field at once. */
private const val SUGGESTION_LIMIT = 12

private const val LOCK_NOTE = "Fields you change here stay as you set them; the sites no longer update them."

/** The attribute inputs in form order; two on a row where both are short. */
private val AttributeRows = listOf(
    listOf(PerformerAttribute.Gender),
    listOf(PerformerAttribute.Birthdate, PerformerAttribute.DeathDate),
    listOf(PerformerAttribute.Country),
    listOf(PerformerAttribute.Ethnicity),
    listOf(PerformerAttribute.HairColor),
    listOf(PerformerAttribute.EyeColor),
    listOf(PerformerAttribute.HeightCm, PerformerAttribute.WeightKg),
    listOf(PerformerAttribute.Measurements),
    listOf(PerformerAttribute.FakeTits),
    listOf(PerformerAttribute.Tattoos),
    listOf(PerformerAttribute.Piercings),
    listOf(PerformerAttribute.CareerStart, PerformerAttribute.CareerEnd),
)

/**
 * What the sheet shows and does: the performer as stored, the form over it, the suggestions per
 * attribute (presets first, then the library's own values as they arrive) and the save.
 */
@Stable
private class PerformerEditState(private val performerId: String, private val scope: CoroutineScope) {
    var performer by mutableStateOf<LoadState<EditablePerformer>>(LoadState.Loading)
        private set
    var form by mutableStateOf<PerformerForm?>(null)
        private set
    var suggestions by mutableStateOf(presetSuggestions())
        private set
    var saving by mutableStateOf(false)
        private set
    var error by mutableStateOf<String?>(null)
        private set

    init {
        load()
    }

    /** Reads the performer, starts the form from them and asks the library for its values. */
    fun load() {
        performer = LoadState.Loading
        scope.launch {
            try {
                val stored = requireNotNull(PerformerEditing.load(performerId)) { "They're gone from the library." }
                form = PerformerEditing.formOf(stored)
                performer = LoadState.Loaded(stored)
                loadLibraryValues()
            } catch (failure: CancellationException) {
                throw failure
            } catch (failure: Exception) {
                performer = LoadState.Failed(failure.displayMessage())
            }
        }
    }

    /** Applies an edit to the form. */
    fun change(transform: (PerformerForm) -> PerformerForm) {
        val current = form ?: return
        form = transform(current)
    }

    /** Saves what changed; [onFinished] gets whether anything was saved. Nothing changed saves nothing. */
    fun save(onFinished: (saved: Boolean) -> Unit) {
        val stored = (performer as? LoadState.Loaded)?.value ?: return
        val current = form ?: return
        val changes = PerformerEditing.changes(stored, current)
        if (changes.isEmpty()) {
            onFinished(false)
            return
        }
        saving = true
        error = null
        scope.launch {
            try {
                PerformerEditing.save(stored.id, changes)
                saving = false
                onFinished(true)
            } catch (failure: CancellationException) {
                throw failure
            } catch (failure: Exception) {
                error = failure.displayMessage()
                saving = false
            }
        }
    }

    /** Fetches the library's values for each suggested attribute, each merged in as it arrives. */
    private fun loadLibraryValues() {
        for (attribute in PerformerAttribute.entries) {
            if (attribute.type == AttributeFieldType.Suggest) {
                scope.launch { loadLibraryValues(attribute) }
            }
        }
    }

    /** One attribute's library values behind its presets; the presets alone when they can't be read. */
    private suspend fun loadLibraryValues(attribute: PerformerAttribute) {
        val values = try {
            PerformerEditing.libraryValues(attribute)
        } catch (failure: CancellationException) {
            throw failure
        } catch (failure: Exception) {
            return
        }
        suggestions = suggestions + (attribute to PerformerPresets.mergeSuggestions(attribute.presets, values))
    }

    /** Each suggested attribute's presets, before the library's values arrive. */
    private fun presetSuggestions(): Map<PerformerAttribute, List<String>> {
        return PerformerAttribute.entries
            .filter { attribute -> attribute.type == AttributeFieldType.Suggest }
            .associateWith { attribute -> PerformerPresets.mergeSuggestions(attribute.presets, emptyList()) }
    }
}

/**
 * Editing a performer in a bottom sheet, as the web's PerformerEditor: picture URL, name, aliases,
 * about, every attribute and the links. It loads the performer itself. Only fields that changed
 * are saved, since a saved field is locked against the sites. [onSaved] runs after a save that
 * changed something; [onDismiss] after that too, and on Cancel or a save with nothing changed.
 */
@Composable
fun PerformerEditSheet(performerId: String, onSaved: () -> Unit, onDismiss: () -> Unit) {
    val scope = rememberCoroutineScope()
    val state = remember(performerId) { PerformerEditState(performerId, scope) }
    VeilBottomSheet(onDismissRequest = onDismiss) {
        Column(
            Modifier
                .imePadding()
                .verticalScroll(rememberScrollState())
                .padding(start = VeilSpacing.gutter, end = VeilSpacing.gutter, bottom = VeilSpacing.large),
            verticalArrangement = Arrangement.spacedBy(VeilSpacing.large),
        ) {
            Text("Edit performer", style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
            val loaded = state.performer
            val form = state.form
            if (loaded is LoadState.Loaded && form != null) {
                EditForm(loaded.value, form, state, onSaved, onDismiss)
            } else if (loaded is LoadState.Failed) {
                LoadFailure(loaded.message, onRetry = state::load, onCancel = onDismiss)
            } else {
                Spinner(Modifier.align(Alignment.CenterHorizontally))
            }
        }
    }
}

/** The sheet when the performer could not be read: why, and to try again or give up. */
@Composable
private fun LoadFailure(message: String, onRetry: () -> Unit, onCancel: () -> Unit) {
    Text(message, style = MaterialTheme.typography.bodyMedium, color = VeilColors.error)
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small, Alignment.End)) {
        TextAction("Try again", onRetry)
        SecondaryButton("Cancel", onCancel)
    }
}

/** Every input of the form, then the note, the error and the buttons. */
@Composable
private fun EditForm(stored: EditablePerformer, form: PerformerForm, state: PerformerEditState, onSaved: () -> Unit, onDismiss: () -> Unit) {
    PictureAndName(stored, form, state)
    VeilTextField(
        form.aliases,
        { text -> state.change { it.copy(aliases = text) } },
        Modifier.fillMaxWidth(),
        label = "Also known as (comma-separated)",
        keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Words),
    )
    VeilTextField(
        form.details,
        { text -> state.change { it.copy(details = text) } },
        Modifier.fillMaxWidth(),
        label = "About",
        singleLine = false,
        keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Sentences),
    )
    val today = LocalDate.now()
    for (row in AttributeRows) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            for (attribute in row) {
                val hint = PerformerEditing.hint(attribute, form, today)
                AttributeInput(attribute, form, state, hint, Modifier.weight(1f))
            }
        }
    }
    VeilTextField(
        form.links,
        { text -> state.change { it.copy(links = text) } },
        Modifier.fillMaxWidth(),
        label = "Links (one per line)",
        singleLine = false,
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Uri),
    )
    Text(LOCK_NOTE, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentFaint)
    val error = state.error
    if (error != null) {
        Text(error, style = MaterialTheme.typography.bodyMedium, color = VeilColors.error)
    }
    SheetButtons(form, state, onSaved, onDismiss)
}

/** Cancel and Save, right-aligned; Save needs a name and waits while saving. */
@Composable
private fun SheetButtons(form: PerformerForm, state: PerformerEditState, onSaved: () -> Unit, onDismiss: () -> Unit) {
    var saveLabel = "Save"
    if (state.saving) {
        saveLabel = "Saving…"
    }
    Row(
        Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small, Alignment.End),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        SecondaryButton("Cancel", onDismiss)
        PrimaryButton(
            saveLabel,
            onClick = {
                state.save { saved ->
                    if (saved) {
                        onSaved()
                    }
                    onDismiss()
                }
            },
            enabled = form.name.isNotBlank() && !state.saving,
        )
    }
}

/** The picture preview with its URL (empty removes it), then the name. */
@Composable
private fun PictureAndName(stored: EditablePerformer, form: PerformerForm, state: PerformerEditState) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large)) {
        RemoteImage(
            previewPath(stored, form.imageUrl),
            Modifier.width(PicturePreviewWidth).aspectRatio(PICTURE_RATIO).clip(VeilShapes.card),
        )
        VeilTextField(
            form.imageUrl,
            { text -> state.change { it.copy(imageUrl = text) } },
            Modifier.weight(1f),
            label = "Picture URL (empty removes it)",
            singleLine = false,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Uri),
        )
    }
    VeilTextField(
        form.name,
        { text -> state.change { it.copy(name = text) } },
        Modifier.fillMaxWidth(),
        label = "Name",
        keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Words),
    )
}

/** What to load the preview from: the stored picture's URL while its URI is untouched, else what is typed. */
private fun previewPath(stored: EditablePerformer, typed: String): String? {
    val trimmed = typed.trim()
    if (trimmed.isEmpty()) {
        return null
    }
    if (trimmed == stored.imageUri) {
        return stored.imageUrl
    }
    return trimmed
}

/** The input an attribute is edited with, by its type, with a note under it when it has one. */
@Composable
private fun AttributeInput(attribute: PerformerAttribute, form: PerformerForm, state: PerformerEditState, hint: String?, modifier: Modifier) {
    val text = form.attributes[attribute].orEmpty()
    val onTextChange = { changed: String -> state.change { it.copy(attributes = it.attributes + (attribute to changed)) } }
    when (attribute.type) {
        AttributeFieldType.Suggest -> SuggestField(attribute.label, text, state.suggestions[attribute].orEmpty(), onTextChange, modifier)
        AttributeFieldType.Date -> VeilTextField(
            text,
            onTextChange,
            modifier,
            label = attribute.label,
            placeholder = "YYYY-MM-DD",
            supportingText = hint,
        )
        AttributeFieldType.Number -> VeilTextField(
            text,
            { changed -> onTextChange(changed.filter { character -> character.isDigit() }) },
            modifier,
            label = attribute.label,
            supportingText = hint,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
        )
        AttributeFieldType.Measurements -> MeasurementsInput(attribute.label, form, text, state, onTextChange, modifier)
    }
}

/** Bust, waist and hips as three inputs; one text input when the stored measurements don't split. */
@Composable
private fun MeasurementsInput(
    label: String,
    form: PerformerForm,
    text: String,
    state: PerformerEditState,
    onTextChange: (String) -> Unit,
    modifier: Modifier,
) {
    val parts = form.measurementParts
    if (parts == null) {
        VeilTextField(text, onTextChange, modifier, label = label)
        return
    }
    val onPartsChange = { changed: MeasurementParts -> state.change { it.copy(measurementParts = changed) } }
    Column(modifier, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        Text(label, style = MaterialTheme.typography.labelMedium, color = VeilColors.contentMuted, modifier = Modifier.padding(start = VeilSpacing.extraSmall))
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            VeilTextField(parts.bust, { changed -> onPartsChange(parts.copy(bust = changed)) }, Modifier.weight(1f), placeholder = "Bust")
            VeilTextField(
                parts.waist,
                { changed -> onPartsChange(parts.copy(waist = changed)) },
                Modifier.weight(1f),
                placeholder = "Waist",
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
            )
            VeilTextField(
                parts.hips,
                { changed -> onPartsChange(parts.copy(hips = changed)) },
                Modifier.weight(1f),
                placeholder = "Hips",
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
            )
        }
    }
}

/**
 * A text field that suggests values while it has focus: a row of chips with the [suggestions]
 * that fit what is typed. Tapping one fills the field; anything else can still be typed.
 */
@Composable
private fun SuggestField(label: String, value: String, suggestions: List<String>, onValueChange: (String) -> Unit, modifier: Modifier) {
    var focused by remember { mutableStateOf(false) }
    Column(modifier.onFocusChanged { focusState -> focused = focusState.hasFocus }, verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        VeilTextField(
            value,
            onValueChange,
            Modifier.fillMaxWidth(),
            label = label,
            keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Sentences),
        )
        if (focused) {
            SuggestionChips(PerformerPresets.matchingSuggestions(suggestions, value, SUGGESTION_LIMIT), onValueChange)
        }
    }
}

/** The matching suggestions as a sideways row of chips; nothing when none match. */
@Composable
private fun SuggestionChips(matches: List<String>, onPick: (String) -> Unit) {
    if (matches.isEmpty()) {
        return
    }
    LazyRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        items(matches, key = { match -> match }) { match ->
            FeedChip(match, onClick = { onPick(match) })
        }
    }
}
