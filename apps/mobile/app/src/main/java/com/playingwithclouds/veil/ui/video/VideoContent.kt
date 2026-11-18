package com.playingwithclouds.veil.ui.video

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.produceState
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.media3.common.Player
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.ImpressionSurface
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.UploaderRepository
import com.playingwithclouds.veil.data.VideoDetail
import com.playingwithclouds.veil.data.VideoUploader
import kotlinx.coroutines.CancellationException
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.RatingStars
import com.playingwithclouds.veil.ui.components.TrackShown
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SkeletonBlock
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatReleaseDate

/** Up next asks for more this close to its end. */
private const val LOAD_MORE_DISTANCE = 5

/** Lines of the description before it's tapped open. */
private const val DETAILS_COLLAPSED_LINES = 3

/** Width of a person in the people row. */
private val PersonWidth = 72.dp

/** Size of a person's picture in the people row. */
private val PersonAvatarSize = 56.dp

/** The list key of an Up next card. */
private fun upNextKey(videoId: String): String = "up-next:$videoId"

/**
 * Everything under the player: title and facts, the actions, rating, sources, people, tags,
 * description and where it was found, then Up next, which grows while scrolled to its end.
 */
@Composable
fun VideoContent(state: VideoState, viewModel: VideoViewModel, player: Player, navigator: AppNavigator) {
    val listState = rememberLazyListState()
    val related = state.related
    LaunchedEffect(listState) {
        snapshotFlow {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisible >= listState.layoutInfo.totalItemsCount - LOAD_MORE_DISTANCE
        }.collect { nearEnd ->
            if (nearEnd) {
                viewModel.loadMoreRelated()
            }
        }
    }
    TrackShown(listState, related.map { video -> upNextKey(video.id) }) { position ->
        AppServices.impressions.recordShown(related[position].id, RecordKind.video, ImpressionSurface.RELATED, position)
    }
    LazyColumn(Modifier.fillMaxWidth(), state = listState, contentPadding = PaddingValues(bottom = VeilSpacing.huge)) {
        val detail = (state.detail as? LoadState.Loaded)?.value
        if (detail != null) {
            aboutVideo(detail, state, viewModel, player, navigator)
        } else {
            aboutVideoSkeleton()
        }
        upNext(state, navigator)
    }
}

/** The part about the video itself, above Up next. */
private fun LazyListScope.aboutVideo(detail: VideoDetail, state: VideoState, viewModel: VideoViewModel, player: Player, navigator: AppNavigator) {
    item("title") { TitleBlock(detail) }
    item("actions") { ActionsRow(state, viewModel, player) }
    item("rating") { RatingStars(state.rating, onRate = viewModel::rate, modifier = Modifier.gutterPadding().padding(top = VeilSpacing.small)) }
    if (detail.sources.size > 1) {
        item("sources") { SourcesRow(detail, state, viewModel) }
    }
    item("uploader") { UploaderChip(detail.id, navigator) }
    val people = listOfNotNull(detail.studio) + detail.performers
    if (people.isNotEmpty()) {
        item("people") {
            PeopleRow(detail, onOpen = { person ->
                if (person == detail.studio) {
                    navigator.openStudio(person.id)
                } else {
                    navigator.openPerformer(person.id)
                }
            })
        }
    }
    if (detail.tags.isNotEmpty()) {
        item("tags") { TagsBlock(detail, navigator) }
    }
    val details = detail.details
    if (!details.isNullOrBlank()) {
        item("details") { DetailsText(details) }
    }
    if (detail.sites.isNotEmpty()) {
        item("found-on") {
            Text(
                "Found on " + detail.sites.map { site -> site.plugin }.distinct().joinToString(", "),
                Modifier.gutterPadding().padding(top = VeilSpacing.medium),
                style = VeilType.cardMeta,
                color = VeilColors.contentFaint,
            )
        }
    }
}

/**
 * Stand-ins for the part about the video while it opens, under the same keys as the real rows: the
 * list stays at the top instead of keeping Up next (which often arrives first) in view.
 */
private fun LazyListScope.aboutVideoSkeleton() {
    item("title") {
        Column(Modifier.gutterPadding().padding(top = VeilSpacing.medium), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            SkeletonBlock(Modifier.fillMaxWidth().height(SkeletonTitleHeight))
            SkeletonBlock(Modifier.fillMaxWidth(SKELETON_SHORT_LINE).height(SkeletonLineHeight))
        }
    }
    item("actions") {
        Row(Modifier.gutterPadding().padding(top = VeilSpacing.medium), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            repeat(SKELETON_PILLS) {
                SkeletonBlock(Modifier.width(SkeletonPillWidth).height(SkeletonPillHeight), VeilShapes.capsule)
            }
        }
    }
    item("tags") {
        Row(Modifier.gutterPadding().padding(top = VeilSpacing.large), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            repeat(SKELETON_PILLS) {
                SkeletonBlock(Modifier.width(SkeletonTagWidth).height(SkeletonTagHeight), VeilShapes.capsule)
            }
        }
    }
    item("details") {
        Column(Modifier.gutterPadding().padding(top = VeilSpacing.large), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            repeat(DETAILS_COLLAPSED_LINES) {
                SkeletonBlock(Modifier.fillMaxWidth().height(SkeletonLineHeight))
            }
        }
    }
}

/** Pills in each skeleton row. */
private const val SKELETON_PILLS = 4

/** Share of the width the short skeleton line (the facts under the title) takes. */
private const val SKELETON_SHORT_LINE = 0.55f

private val SkeletonTitleHeight = 20.dp
private val SkeletonLineHeight = 12.dp
private val SkeletonPillWidth = 84.dp
private val SkeletonPillHeight = 36.dp
private val SkeletonTagWidth = 64.dp
private val SkeletonTagHeight = 28.dp

/** Up next: a heading, the related videos as full-width cards, a spinner while more load. */
private fun LazyListScope.upNext(state: VideoState, navigator: AppNavigator) {
    if (state.related.isEmpty() && state.relatedDone) {
        return
    }
    item("up-next-heading") { SectionHeading("Up next", modifier = Modifier.gutterPadding()) }
    itemsIndexed(state.related, key = { _, video -> upNextKey(video.id) }) { position, video ->
        VideoCard(
            video,
            onClick = {
                AppServices.impressions.recordClicked(video.id, RecordKind.video, ImpressionSurface.RELATED, position)
                navigator.openVideo(video.id)
            },
            edgeToEdge = true,
        )
    }
    if (state.relatedLoading) {
        item("up-next-loading") {
            Row(Modifier.fillMaxWidth().padding(VeilSpacing.large), horizontalArrangement = Arrangement.Center) { Spinner() }
        }
    }
}

/** The title, then who it's from, when, how long and in what quality. */
@Composable
private fun TitleBlock(detail: VideoDetail) {
    val facts = listOfNotNull(
        detail.summary.byline,
        formatReleaseDate(detail.summary.date),
        detail.summary.durationSeconds?.let(::formatClock),
        qualityLabel(detail.height),
    )
    Column(Modifier.gutterPadding().padding(top = VeilSpacing.medium), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
        Text(detail.title, style = MaterialTheme.typography.titleMedium, color = VeilColors.content)
        if (facts.isNotEmpty()) {
            Text(facts.joinToString(" · "), style = VeilType.cardMeta, color = VeilColors.contentMuted)
        }
    }
}

/** "4K", "1080p" from a video's height; null when unknown. */
private fun qualityLabel(height: Int?): String? {
    if (height == null || height <= 0) {
        return null
    }
    if (height >= 2160) {
        return "4K"
    }
    return "${height}p"
}

/** The actions as a sideways row: like, dislike, favourite, O, watchlist, download. */
@Composable
private fun ActionsRow(state: VideoState, viewModel: VideoViewModel, player: Player) {
    var likeIcon = VeilIcons.Like
    var dislikeIcon = VeilIcons.Dislike
    var favoriteIcon = VeilIcons.Heart
    if (state.verdict == Verdict.LIKED) {
        likeIcon = VeilIcons.LikeFilled
    }
    if (state.verdict == Verdict.DISLIKED) {
        dislikeIcon = VeilIcons.DislikeFilled
    }
    if (state.favorite) {
        favoriteIcon = VeilIcons.HeartFilled
    }
    var watchlistLabel = "Watchlist"
    if (state.onWatchlist) {
        watchlistLabel = "On watchlist"
    }
    var downloadLabel = "Download"
    if (state.downloading) {
        downloadLabel = "Downloading"
    }
    LazyRow(
        Modifier.fillMaxWidth().padding(top = VeilSpacing.medium),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        item { Pill("Like", onClick = viewModel::like, selected = state.verdict == Verdict.LIKED, icon = likeIcon) }
        item { Pill("Dislike", onClick = viewModel::dislike, selected = state.verdict == Verdict.DISLIKED, icon = dislikeIcon) }
        item { Pill("Favourite", onClick = viewModel::toggleFavorite, selected = state.favorite, icon = favoriteIcon) }
        item { Pill(state.oCount.toString(), onClick = { viewModel.logO(player.currentPosition / MILLISECONDS_PER_SECOND) }, icon = VeilIcons.O) }
        item {
            Pill(watchlistLabel, onClick = {
                if (!state.onWatchlist) {
                    viewModel.saveForLater()
                }
            }, selected = state.onWatchlist, icon = VeilIcons.Bookmark)
        }
        item {
            Pill(downloadLabel, onClick = {
                if (!state.downloading) {
                    viewModel.download()
                }
            }, icon = VeilIcons.Download)
        }
    }
}

private const val MILLISECONDS_PER_SECOND = 1000.0

/** The video's sources as pills, the playing one selected; picking one plays it from here. */
@Composable
private fun SourcesRow(detail: VideoDetail, state: VideoState, viewModel: VideoViewModel) {
    Column(Modifier.padding(top = VeilSpacing.small)) {
        SectionHeading("Sources", modifier = Modifier.gutterPadding())
        LazyRow(contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            items(detail.sources, key = { source -> source.id }) { source ->
                val label = listOfNotNull(source.provider, source.quality).joinToString(" · ")
                Pill(label, onClick = { viewModel.selectSource(source) }, selected = source.id == state.active?.sourceId)
            }
        }
    }
}

/** The studio and performers as round pictures with names; tapping one opens their page. */
@Composable
private fun PeopleRow(detail: VideoDetail, onOpen: (PerformerSummary) -> Unit) {
    val people = listOfNotNull(detail.studio) + detail.performers
    LazyRow(
        Modifier.fillMaxWidth().padding(top = VeilSpacing.large),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
    ) {
        items(people, key = { person -> person.id }) { person ->
            Column(
                Modifier.width(PersonWidth).pressClickable { onOpen(person) },
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall),
            ) {
                Avatar(person.imageUrl, Modifier.size(PersonAvatarSize))
                Text(
                    person.name,
                    style = VeilType.cardMeta,
                    color = VeilColors.content,
                    textAlign = TextAlign.Center,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

/**
 * Who uploaded the video to the site it came from, read once the page has opened (the sites name
 * them on the pages it reads). A chip that opens their videos when their channel page is known,
 * else a plain line; nothing when no site names an uploader.
 */
@Composable
private fun UploaderChip(videoId: String, navigator: AppNavigator) {
    val uploader by produceState<VideoUploader?>(null, videoId) {
        value = try {
            UploaderRepository.ofVideo(videoId)
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            null
        }
    }
    val shown = uploader ?: return
    val url = shown.url
    if (url == null) {
        Text(
            "Uploaded by ${shown.name}",
            Modifier.gutterPadding().padding(top = VeilSpacing.large),
            style = VeilType.cardMeta,
            color = VeilColors.contentMuted,
        )
        return
    }
    Row(Modifier.gutterPadding().padding(top = VeilSpacing.large), verticalAlignment = Alignment.CenterVertically) {
        Pill(shown.name, onClick = { navigator.openUploader(shown.plugin, url, shown.name) }, icon = VeilIcons.Performer)
    }
}

/** The tags as one row of pills that scrolls sideways, so they never take more than a line; tapping one opens the tag. */
@Composable
private fun TagsBlock(detail: VideoDetail, navigator: AppNavigator) {
    LazyRow(
        Modifier.fillMaxWidth().fadingEdges().padding(top = VeilSpacing.large),
        contentPadding = GutterRowPadding,
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
    ) {
        items(detail.tags, key = { tag -> tag.id }) { tag ->
            TagPill(tag.name, onClick = { navigator.openTag(tag) })
        }
    }
}

/** The description, a few lines until tapped. */
@Composable
private fun DetailsText(details: String) {
    var expanded by rememberSaveable { mutableStateOf(false) }
    var maxLines = DETAILS_COLLAPSED_LINES
    if (expanded) {
        maxLines = Int.MAX_VALUE
    }
    Text(
        details,
        Modifier.gutterPadding().padding(top = VeilSpacing.large).pressClickable { expanded = !expanded },
        style = MaterialTheme.typography.bodySmall,
        color = VeilColors.contentMuted,
        maxLines = maxLines,
        overflow = TextOverflow.Ellipsis,
    )
}
