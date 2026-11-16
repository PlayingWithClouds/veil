package com.playingwithclouds.veil.ui.media

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.ContentsEntry
import com.playingwithclouds.veil.data.MediaRepository
import com.playingwithclouds.veil.data.StoryDetail
import com.playingwithclouds.veil.data.StoryReading
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

private val CoverWidth = 120.dp
private val CoverHeight = 180.dp

/** How far each branch level is indented in the contents. */
private val BranchIndent = 16.dp

/** One story, opened (its pages read for the chapters). */
class StoryViewModel(private val storyId: String) : ViewModel() {

    private val storyState = MutableStateFlow<LoadState<StoryDetail>>(LoadState.Loading)
    val story: StateFlow<LoadState<StoryDetail>> = storyState.asStateFlow()

    init {
        load()
    }

    /** Opens the story. */
    fun load() {
        viewModelScope.loadInto(storyState) { requireNotNull(MediaRepository.openStory(storyId)) { "This story is gone." } }
    }
}

/** A story's page: cover and facts, Read (or Continue), the description, tags and the chapters. */
@Composable
fun StoryScreen(storyId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "story-$storyId") { StoryViewModel(storyId) }
    val state by viewModel.story.collectAsStateWithLifecycle()
    val title = (state as? LoadState.Loaded)?.value?.summary?.title.orEmpty()
    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }) { padding ->
        LoadStateContent(state, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { story ->
            val contents = StoryReading.contentsOf(story.chapters)
            LazyColumn(contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
                item("header") { StoryHeader(story, navigator) }
                if (contents.isNotEmpty()) {
                    item("chapters-heading") { SectionHeading("Chapters", modifier = Modifier.gutterPadding()) }
                }
                items(contents, key = { entry -> entry.chapter.id }) { entry ->
                    ChapterRow(entry, onOpen = { navigator.openStoryReader(story.summary.id, entry.chapter.position) })
                }
            }
        }
    }
}

/** Cover, title, author and length, Read or Continue, the description and tags. */
@Composable
private fun StoryHeader(story: StoryDetail, navigator: AppNavigator) {
    val point = StoryReading.readingPointOf(story.progress)
    Column(Modifier.gutterPadding().padding(top = VeilSpacing.large), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.large)) {
            RemoteImage(story.summary.coverUrl, Modifier.size(CoverWidth, CoverHeight).clip(VeilShapes.card))
            Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
                Text(story.summary.title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
                val author = story.summary.author
                if (author != null) {
                    Text(author, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
                }
                Text(lengthLine(story), style = VeilType.cardMeta, color = VeilColors.contentFaint)
            }
        }
        if (story.chapters.isNotEmpty()) {
            var label = "Read"
            var position = story.chapters.first().position
            if (point != null) {
                label = "Continue"
                position = point.position
            }
            PrimaryButton(label, onClick = { navigator.openStoryReader(story.summary.id, position) }, icon = VeilIcons.Stories)
        }
        val details = story.details
        if (!details.isNullOrBlank()) {
            Text(details, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        if (story.tags.isNotEmpty()) {
            FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                for (tag in story.tags) {
                    TagPill(tag.name, onClick = { navigator.openTag(tag.id) })
                }
            }
        }
    }
}

/** "12 chapters · 40,000 words · 3 h read"; parts unknown left out. */
private fun lengthLine(story: StoryDetail): String {
    val parts = mutableListOf("${story.summary.chapterCount} chapters")
    val words = story.wordCount
    if (words != null && words > 0) {
        parts.add("%,d words".format(words))
        val minutes = StoryReading.minutesLeft(words, 0.0)
        parts.add(readTime(minutes))
    }
    return parts.joinToString(" · ")
}

/** "45 min read", "3 h read". */
internal fun readTime(minutes: Int): String {
    if (minutes < MINUTES_PER_HOUR) {
        return "$minutes min read"
    }
    return "${minutes / MINUTES_PER_HOUR} h read"
}

private const val MINUTES_PER_HOUR = 60

/** One chapter of the contents, indented by its branch depth; a choice shows what was chosen. */
@Composable
private fun ChapterRow(entry: ContentsEntry, onOpen: () -> Unit) {
    val chapter = entry.chapter
    val label = chapter.choice ?: chapter.title ?: "Chapter ${chapter.position + 1}"
    Text(
        label,
        Modifier
            .fillMaxWidth()
            .pressClickable(onOpen)
            .padding(start = VeilSpacing.gutter + BranchIndent * entry.depth, end = VeilSpacing.gutter, top = VeilSpacing.medium, bottom = VeilSpacing.medium),
        style = MaterialTheme.typography.bodyLarge,
        color = VeilColors.content,
        maxLines = 2,
        overflow = TextOverflow.Ellipsis,
    )
}
