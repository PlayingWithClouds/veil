package com.playingwithclouds.veil.ui.media

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.Chapter
import com.playingwithclouds.veil.data.ChapterText
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.MediaRepository
import com.playingwithclouds.veil.data.ReadingPoint
import com.playingwithclouds.veil.data.StoryReading
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.launch

/** How long scrolling has to rest before the place is saved. */
private const val SAVE_DELAY_MILLISECONDS = 1500L

/** The chapter open in the reader, with what can follow it. */
data class ReaderPage(
    val chapter: ChapterText,
    val following: List<Chapter>,
    val previous: Chapter?,
    /** How far into the chapter reading stopped last time (0..1), to scroll back there. */
    val startFraction: Double,
)

/**
 * Reading one story a chapter at a time. The place (the chapter's position plus how far into it)
 * is saved as the story's progress, as the web's reader does.
 */
class StoryReaderViewModel(private val storyId: String, private val startPosition: Int) : ViewModel() {

    private val pageState = MutableStateFlow<LoadState<ReaderPage>>(LoadState.Loading)
    val page: StateFlow<LoadState<ReaderPage>> = pageState.asStateFlow()

    private var chapters = emptyList<Chapter>()
    private var lastSaved: Double? = null

    init {
        viewModelScope.launch {
            val story = runCatching { MediaRepository.openStory(storyId) }.getOrNull()
            chapters = story?.chapters.orEmpty()
            val point = StoryReading.readingPointOf(story?.progress)
            var fraction = 0.0
            if (point != null && point.position == startPosition) {
                fraction = point.fraction
            }
            open(startPosition, fraction)
        }
    }

    /** Opens the chapter at [position], reading it from the site the first time. */
    fun open(position: Int, startFraction: Double = 0.0) {
        pageState.value = LoadState.Loading
        viewModelScope.launch {
            try {
                val text = requireNotNull(MediaRepository.readChapter(storyId, position)) { "This chapter is gone." }
                // Reading a chapter of a branching story can add the ones its choices lead to.
                chapters = runCatching { MediaRepository.chapters(storyId) }.getOrDefault(chapters)
                val current = chapters.firstOrNull { chapter -> chapter.id == text.chapter.id } ?: text.chapter
                pageState.value = LoadState.Loaded(
                    ReaderPage(text, StoryReading.followingChapters(chapters, current), StoryReading.previousChapter(chapters, current), startFraction),
                )
                save(ReadingPoint(position, startFraction))
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                pageState.value = LoadState.Failed(error.displayMessage())
            }
        }
    }

    /** Saves the place as the story's progress, unless it's the one saved last. */
    fun save(point: ReadingPoint) {
        val progress = StoryReading.progressOf(point)
        if (progress == lastSaved) {
            return
        }
        lastSaved = progress
        AppServices.events.log(LoggedEvent(EventKind.progress, RecordKind.story, storyId, mapOf("position" to progress)))
    }

    /** Sends the saved place when the reader closes. */
    override fun onCleared() {
        AppServices.events.flushNow()
    }
}

/** The reader: the chapter's text, then the next chapter or the choices, and the one before. */
@Composable
fun StoryReaderScreen(storyId: String, position: Int, navigator: AppNavigator) {
    val viewModel = viewModel(key = "reader-$storyId") { StoryReaderViewModel(storyId, position) }
    val state by viewModel.page.collectAsStateWithLifecycle()
    val title = (state as? LoadState.Loaded)?.value?.chapter?.chapter?.let(::chapterTitle).orEmpty()
    DetailScreen(title, onBack = navigator::back) {
        LoadStateContent(state, onRetry = { viewModel.open(position) }, modifier = Modifier.padding(top = LocalTopBarInset.current)) { page ->
            ChapterReader(page, viewModel)
        }
    }
}

/** A chapter's own title, else its number. */
private fun chapterTitle(chapter: Chapter): String {
    return chapter.title ?: "Chapter ${chapter.position + 1}"
}

/** The paragraphs, scrolled back to where reading stopped; scrolling saves the place once it rests. */
@OptIn(FlowPreview::class)
@Composable
private fun ChapterReader(page: ReaderPage, viewModel: StoryReaderViewModel) {
    val paragraphs = StoryReading.paragraphsOf(page.chapter.text)
    val listState = rememberLazyListState()
    val position = page.chapter.chapter.position
    LaunchedEffect(page.chapter.chapter.id) {
        listState.scrollToItem((page.startFraction * paragraphs.size).toInt())
        snapshotFlow { fractionRead(listState, paragraphs.size) }
            .debounce(SAVE_DELAY_MILLISECONDS)
            .collect { fraction -> viewModel.save(ReadingPoint(position, fraction)) }
    }
    DisposableEffect(page.chapter.chapter.id) {
        onDispose { viewModel.save(ReadingPoint(position, fractionRead(listState, paragraphs.size))) }
    }
    LazyColumn(state = listState, contentPadding = PaddingValues(top = VeilSpacing.medium + LocalTopBarInset.current, bottom = VeilSpacing.huge)) {
        items(paragraphs.size) { index ->
            Text(
                paragraphs[index],
                Modifier.gutterPadding().padding(bottom = VeilSpacing.large),
                style = MaterialTheme.typography.bodyLarge,
                color = VeilColors.content,
            )
        }
        item("next") { ChapterEnd(page, onOpen = { chapter -> viewModel.open(chapter.position) }) }
    }
}

/** How far through the chapter's paragraphs the top of the screen is, 0..1. */
private fun fractionRead(listState: LazyListState, paragraphCount: Int): Double {
    if (paragraphCount <= 0) {
        return 0.0
    }
    return (listState.firstVisibleItemIndex.toDouble() / paragraphCount).coerceIn(0.0, 1.0)
}

/** What ends a chapter: the choices (or the next chapter), and the one before; "The end" without any. */
@Composable
private fun ChapterEnd(page: ReaderPage, onOpen: (Chapter) -> Unit) {
    Column(Modifier.gutterPadding().padding(top = VeilSpacing.large), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        if (page.following.isEmpty()) {
            Text("The end", style = MaterialTheme.typography.titleMedium, color = VeilColors.contentMuted)
        }
        for (next in page.following) {
            val label = next.choice ?: "Next: ${chapterTitle(next)}"
            PrimaryButton(label, onClick = { onOpen(next) }, modifier = Modifier.fillMaxWidth())
        }
        val previous = page.previous
        if (previous != null) {
            Row { SecondaryButton("Back to ${chapterTitle(previous)}", onClick = { onOpen(previous) }) }
        }
    }
}
