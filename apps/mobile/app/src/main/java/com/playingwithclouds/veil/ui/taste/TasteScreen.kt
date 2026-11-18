package com.playingwithclouds.veil.ui.taste

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.SceneLabelTaste
import com.playingwithclouds.veil.data.SceneRepository
import com.playingwithclouds.veil.data.TagSummary
import com.playingwithclouds.veil.data.Taste
import com.playingwithclouds.veil.data.TasteEntry
import com.playingwithclouds.veil.data.TasteRepository
import com.playingwithclouds.veil.data.TasteSubject
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.FeedChip
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

private val EntryPictureSize = 32.dp

/** Answers on the taste page, as like event values. */
private const val YES = 1
private const val NOT_REALLY = -1
private const val TAKEN_BACK = 0

/** The user's taste and their answers on it; answers show at once and are sent in the background. */
class TasteViewModel : ViewModel() {

    private val tasteState = MutableStateFlow<LoadState<Taste>>(LoadState.Loading)
    val taste: StateFlow<LoadState<Taste>> = tasteState.asStateFlow()

    init {
        load()
    }

    /** Loads the taste profile. */
    fun load() {
        viewModelScope.loadInto(tasteState) { TasteRepository.taste() }
    }

    /** Answers about a scene label: yes makes it a liked scene, not really or taking it back removes it. */
    fun answerLabel(label: SceneLabelTaste, value: Int) {
        val confirmed = confirmedOf(value)
        update { taste -> taste.copy(sceneLabels = taste.sceneLabels.map { other -> if (other.tag.id == label.tag.id) other.copy(confirmed = confirmed) else other }) }
        SceneRepository.noteLabelAnswer(label.tag.name, value == YES)
        send(TasteSubject(RecordKind.tag, label.tag.id), value)
    }

    /** Answers about a performer, studio, site or tag of [dimension]. */
    fun answerEntry(dimension: TasteDimension, entry: TasteEntry, value: Int) {
        val confirmed = confirmedOf(value)
        update { taste -> replaceEntry(taste, dimension, entry.id, confirmed) }
        send(subjectOf(dimension, entry.id), value)
    }

    /** Applies a change to the loaded taste. */
    private fun update(change: (Taste) -> Taste) {
        val current = tasteState.value
        if (current is LoadState.Loaded) {
            tasteState.value = LoadState.Loaded(change(current.value))
        }
    }

    /** Sends an answer; when it fails, the page reloads to show what stands. */
    private fun send(subject: TasteSubject, value: Int) {
        viewModelScope.launch {
            try {
                TasteRepository.answer(subject, value)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                load()
            }
        }
    }
}

/** An answer value as the confirmed state: true yes, false not really, null taken back. */
private fun confirmedOf(value: Int): Boolean? {
    if (value == TAKEN_BACK) {
        return null
    }
    return value == YES
}

/** What an entry of [dimension] is, for the like event. */
private fun subjectOf(dimension: TasteDimension, id: String): TasteSubject {
    return when (dimension) {
        TasteDimension.PERFORMERS -> TasteSubject(RecordKind.performer, id)
        TasteDimension.STUDIOS -> TasteSubject(RecordKind.studio, id)
        TasteDimension.TAGS -> TasteSubject(RecordKind.tag, id)
        TasteDimension.SITES -> TasteSubject(null, id)
    }
}

/** The taste with one entry's answer changed. */
private fun replaceEntry(taste: Taste, dimension: TasteDimension, id: String, confirmed: Boolean?): Taste {
    val change = { entries: List<TasteEntry> -> entries.map { entry -> if (entry.id == id) entry.copy(confirmed = confirmed) else entry } }
    return when (dimension) {
        TasteDimension.PERFORMERS -> taste.copy(performers = change(taste.performers))
        TasteDimension.STUDIOS -> taste.copy(studios = change(taste.studios))
        TasteDimension.SITES -> taste.copy(sites = change(taste.sites))
        TasteDimension.TAGS -> taste.copy(tags = change(taste.tags))
    }
}

/**
 * Your taste: the scenes you're into with why the app thinks so, the performers, studios and sites
 * you like, what you avoid, and a few labels it isn't sure about yet. Every entry takes a Yes or a
 * Not really (a like on it, which the recommendations take as a verdict), shows the answer and can
 * be taken back. A scene said yes to is a liked scene: it plays in compilations, leads the
 * shorties and picks card previews; its play button starts a compilation of only that scene.
 */
@Composable
fun TasteScreen(navigator: AppNavigator) {
    val viewModel = viewModel { TasteViewModel() }
    val state by viewModel.taste.collectAsStateWithLifecycle()
    DetailScreen("Your taste", onBack = navigator::back) {
        val barInset = LocalTopBarInset.current
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(top = barInset)) { taste ->
            val sections = TasteSections.of(taste)
            LazyColumn(
                contentPadding = PaddingValues(top = barInset, bottom = VeilSpacing.huge),
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.small),
            ) {
                item("signals") {
                    Text(
                        "From ${taste.signalCount} things you watched, liked and skipped. Tell it where it's wrong.",
                        Modifier.gutterPadding().padding(top = VeilSpacing.medium),
                        style = MaterialTheme.typography.bodySmall,
                        color = VeilColors.contentMuted,
                    )
                }
                intoSection(sections.into, viewModel, navigator)
                for (dimension in listOf(TasteDimension.PERFORMERS, TasteDimension.STUDIOS, TasteDimension.SITES)) {
                    likedSection(dimension, sections.liked[dimension].orEmpty(), viewModel, navigator)
                }
                avoidedSection(sections.avoided, viewModel)
                questionSection(sections.questions, viewModel)
            }
        }
    }
}

/** "What you're into": the labels said yes to or sure enough, with their evidence. */
private fun LazyListScope.intoSection(labels: List<SceneLabelTaste>, viewModel: TasteViewModel, navigator: AppNavigator) {
    if (labels.isEmpty()) {
        return
    }
    item("into-heading") { SectionHeading("What you're into", modifier = Modifier.gutterPadding()) }
    items(labels, key = { label -> "into-${label.tag.id}" }) { label ->
        LabelCard(label, onAnswer = { value -> viewModel.answerLabel(label, value) }, onPlay = { navigator.openCompilationOf(setOf(label.tag.name)) })
    }
}

/** A dimension's liked entries, each with Yes and Not really. */
private fun LazyListScope.likedSection(dimension: TasteDimension, entries: List<TasteEntry>, viewModel: TasteViewModel, navigator: AppNavigator) {
    if (entries.isEmpty()) {
        return
    }
    item("heading-${dimension.name}") { SectionHeading(dimension.title, modifier = Modifier.gutterPadding()) }
    items(entries, key = { entry -> "${dimension.name}-${entry.id}" }) { entry ->
        EntryRow(entry, onOpen = openerOf(dimension, entry, navigator), onAnswer = { value -> viewModel.answerEntry(dimension, entry, value) })
    }
}

/** "Things you avoid": labels and entries said no to or disliked; each can be changed. */
private fun LazyListScope.avoidedSection(items: List<AvoidedItem>, viewModel: TasteViewModel) {
    if (items.isEmpty()) {
        return
    }
    item("avoid-heading") { SectionHeading("Things you avoid", modifier = Modifier.gutterPadding()) }
    items(items, key = { item -> "avoid-${item.key}" }) { item ->
        when (item) {
            is AvoidedItem.Label -> LabelCard(item.label, onAnswer = { value -> viewModel.answerLabel(item.label, value) }, onPlay = null)
            is AvoidedItem.Entry -> EntryRow(item.entry, onOpen = null, onAnswer = { value -> viewModel.answerEntry(item.dimension, item.entry, value) })
        }
    }
}

/** "Not sure yet": labels with little evidence, asked as quick questions. */
private fun LazyListScope.questionSection(labels: List<SceneLabelTaste>, viewModel: TasteViewModel) {
    if (labels.isEmpty()) {
        return
    }
    item("questions-heading") { SectionHeading("Not sure yet", modifier = Modifier.gutterPadding()) }
    items(labels, key = { label -> "question-${label.tag.id}" }) { label ->
        QuestionRow(label, onAnswer = { value -> viewModel.answerLabel(label, value) })
    }
}

/** Where tapping an entry goes: a performer's or studio's page; sites and tags open nothing here. */
private fun openerOf(dimension: TasteDimension, entry: TasteEntry, navigator: AppNavigator): (() -> Unit)? {
    return when (dimension) {
        TasteDimension.PERFORMERS -> { { navigator.openPerformer(entry.id) } }
        TasteDimension.STUDIOS -> { { navigator.openStudio(entry.id) } }
        TasteDimension.TAGS -> { { navigator.openTag(TagSummary(entry.id, entry.name)) } }
        TasteDimension.SITES -> null
    }
}

/** A scene label: its name, play button, evidence, then the answer buttons. */
@Composable
private fun LabelCard(label: SceneLabelTaste, onAnswer: (Int) -> Unit, onPlay: (() -> Unit)?) {
    VeilCard(Modifier.gutterPadding()) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f)) {
                Text(label.tag.name, style = VeilType.cardTitle, color = VeilColors.content)
                Text(labelMeta(label), style = VeilType.cardMeta, color = VeilColors.contentMuted)
            }
            if (onPlay != null) {
                IconTap(VeilIcons.Play, contentDescription = "Play ${label.tag.name} scenes", onClick = onPlay, tint = VeilColors.content)
            }
        }
        for (sentence in label.evidence) {
            Text(sentence, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        AnswerButtons(label.confirmed, onAnswer)
    }
}

/** "Said yes · 12 videos", or how sure the app is. */
private fun labelMeta(label: SceneLabelTaste): String {
    var answer = "A guess"
    if (label.certainty >= SURE_CERTAINTY) {
        answer = "From what you watch"
    }
    if (label.confirmed == true) {
        answer = "You said yes"
    }
    if (label.confirmed == false) {
        answer = "You said not really"
    }
    return "$answer · ${label.videoCount} videos"
}

/** A performer, studio, site or tag: picture, name, then the answer buttons. */
@Composable
private fun EntryRow(entry: TasteEntry, onOpen: (() -> Unit)?, onAnswer: (Int) -> Unit) {
    var nameModifier = Modifier.fillMaxWidth()
    if (onOpen != null) {
        nameModifier = nameModifier.pressClickable(onClick = onOpen)
    }
    Column(Modifier.gutterPadding().padding(vertical = VeilSpacing.extraSmall), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        Row(nameModifier, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium), verticalAlignment = Alignment.CenterVertically) {
            Avatar(entry.imageUrl, Modifier.size(EntryPictureSize))
            Text(entry.name, Modifier.weight(1f), style = VeilType.cardTitle, color = VeilColors.content, maxLines = 1, overflow = TextOverflow.Ellipsis)
        }
        AnswerButtons(entry.confirmed, onAnswer)
    }
}

/** A label asked about: "Into blowjob?" with Yes and Not really. */
@Composable
private fun QuestionRow(label: SceneLabelTaste, onAnswer: (Int) -> Unit) {
    Column(Modifier.gutterPadding().padding(vertical = VeilSpacing.extraSmall), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        Text("Into ${label.tag.name}?", style = VeilType.cardTitle, color = VeilColors.content)
        Text("In ${label.videoCount} videos", style = VeilType.cardMeta, color = VeilColors.contentMuted)
        AnswerButtons(label.confirmed, onAnswer)
    }
}

/** Yes and Not really; the chosen one is lit, with Undo to take it back. */
@Composable
private fun AnswerButtons(confirmed: Boolean?, onAnswer: (Int) -> Unit) {
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
        FeedChip("Yes", onClick = { onAnswer(YES) }, selected = confirmed == true, icon = VeilIcons.Like)
        FeedChip("Not really", onClick = { onAnswer(NOT_REALLY) }, selected = confirmed == false, icon = VeilIcons.Dislike)
        if (confirmed != null) {
            TextAction("Undo", onClick = { onAnswer(TAKEN_BACK) }, color = VeilColors.contentMuted)
        }
    }
}
