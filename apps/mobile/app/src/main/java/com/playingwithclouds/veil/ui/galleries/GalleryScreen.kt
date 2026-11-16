package com.playingwithclouds.veil.ui.galleries

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.GalleryDetail
import com.playingwithclouds.veil.data.GalleryImage
import com.playingwithclouds.veil.data.GalleryRepository
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.GalleryCard
import com.playingwithclouds.veil.ui.components.ImageViewer
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RatingStars
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.components.Shelf
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.GutterRowPadding
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.theme.VeilType
import com.playingwithclouds.veil.util.formatReleaseDate
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Columns of the photo grid. */
private const val PHOTO_COLUMNS = 2

/** Related galleries under the photos. */
private const val RELATED_LIMIT = 24

/** Photos without a known size are laid out as this portrait shape. */
private const val FALLBACK_RATIO = 3f / 4f

private val RelatedWidth = 140.dp

/** One gallery, opened (its pages read for the images) and logged as viewed. */
class GalleryViewModel(private val galleryId: String) : ViewModel() {

    private val detailState = MutableStateFlow<LoadState<GalleryDetail>>(LoadState.Loading)
    private val relatedState = MutableStateFlow(emptyList<GallerySummary>())

    val detail: StateFlow<LoadState<GalleryDetail>> = detailState.asStateFlow()
    val related: StateFlow<List<GallerySummary>> = relatedState.asStateFlow()

    init {
        load()
        viewModelScope.launch { relatedState.value = runCatching { GalleryRepository.related(galleryId, RELATED_LIMIT) }.getOrDefault(emptyList()) }
    }

    /** Opens the gallery and logs the view. */
    fun load() {
        detailState.value = LoadState.Loading
        viewModelScope.launch {
            val opened = runCatching { requireNotNull(GalleryRepository.open(galleryId)) { "This gallery is gone." } }
            val gallery = opened.getOrElse { error ->
                detailState.value = LoadState.Failed(error.displayMessage())
                return@launch
            }
            detailState.value = LoadState.Loaded(gallery)
            AppServices.events.log(LoggedEvent(EventKind.view, RecordKind.gallery, galleryId))
        }
    }

    /** Flips the favourite mark. */
    fun toggleFavorite() {
        val gallery = loaded() ?: return
        val favorite = !gallery.favorite
        detailState.update { LoadState.Loaded(gallery.copy(favorite = favorite)) }
        viewModelScope.launch { runCatching { GalleryRepository.setFavorite(galleryId, favorite) } }
    }

    /** Rates the gallery 0..10; null clears it. */
    fun rate(rating: Double?) {
        val gallery = loaded() ?: return
        detailState.update { LoadState.Loaded(gallery.copy(rating = rating)) }
        viewModelScope.launch { runCatching { GalleryRepository.rate(galleryId, rating) } }
    }

    /** The gallery, once opened. */
    private fun loaded(): GalleryDetail? {
        return (detailState.value as? LoadState.Loaded)?.value
    }
}

/** A gallery's page: what it is and what to do with it, its photos in two columns, related galleries. */
@Composable
fun GalleryScreen(galleryId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "gallery-$galleryId") { GalleryViewModel(galleryId) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val related by viewModel.related.collectAsStateWithLifecycle()
    var viewing by remember { mutableStateOf<Int?>(null) }
    val title = (detail as? LoadState.Loaded)?.value?.summary?.title.orEmpty()
    ShowAmbientGlow((detail as? LoadState.Loaded)?.value?.summary?.coverUrl)
    Scaffold(topBar = { VeilTopBar(title, onBack = navigator::back) }) { padding ->
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { gallery ->
            LazyVerticalStaggeredGrid(
                StaggeredGridCells.Fixed(PHOTO_COLUMNS),
                contentPadding = PaddingValues(start = VeilSpacing.gutter, end = VeilSpacing.gutter, bottom = VeilSpacing.huge),
                horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small),
                verticalItemSpacing = VeilSpacing.small,
            ) {
                item(key = "header", span = StaggeredGridItemSpan.FullLine) { GalleryHeader(gallery, viewModel, navigator) }
                itemsIndexed(gallery.images, key = { _, image -> image.id }) { index, image ->
                    Photo(image, onClick = { viewing = index })
                }
                if (related.isNotEmpty()) {
                    item(key = "related", span = StaggeredGridItemSpan.FullLine) {
                        Shelf("Related", related, key = { item -> item.id }, itemWidth = RelatedWidth, modifier = Modifier.bleed()) { item, _ ->
                            GalleryCard(item, onClick = { navigator.openGallery(item.id) })
                        }
                    }
                }
            }
            val index = viewing
            if (index != null) {
                ImageViewer(gallery.images.map { image -> image.url }, initialIndex = index, onDismiss = { viewing = null })
            }
        }
    }
}

/** One photo at its own proportions; tapping it opens the viewer there. */
@Composable
private fun Photo(image: GalleryImage, onClick: () -> Unit) {
    var ratio = FALLBACK_RATIO
    val width = image.width
    val height = image.height
    if (width != null && height != null && width > 0 && height > 0) {
        ratio = width.toFloat() / height
    }
    RemoteImage(image.url, Modifier.fillMaxWidth().aspectRatio(ratio).clip(VeilShapes.card).pressClickable(onClick))
}

/** Who and when, the photo count, Favourite / the video / the rating, the description and tags. */
@Composable
private fun GalleryHeader(gallery: GalleryDetail, viewModel: GalleryViewModel, navigator: AppNavigator) {
    Column(Modifier.padding(top = VeilSpacing.medium, bottom = VeilSpacing.medium), verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
        val facts = listOfNotNull(gallery.studio?.name, "${gallery.summary.imageCount} photos", formatReleaseDate(gallery.date))
        Text(facts.joinToString(" · "), style = VeilType.cardMeta, color = VeilColors.contentMuted)
        LazyRow(Modifier.bleed(), contentPadding = GutterRowPadding, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            item {
                var icon = VeilIcons.Heart
                if (gallery.favorite) {
                    icon = VeilIcons.HeartFilled
                }
                Pill("Favourite", onClick = viewModel::toggleFavorite, selected = gallery.favorite, icon = icon)
            }
            val videoId = gallery.videoId
            if (videoId != null) {
                item { Pill("Watch the video", onClick = { navigator.openVideo(videoId) }, icon = VeilIcons.Play) }
            }
            for (performer in gallery.performers) {
                item(key = performer.id) { Pill(performer.name, onClick = { navigator.openPerformer(performer.id) }, icon = VeilIcons.Performer) }
            }
        }
        RatingStars(gallery.rating, onRate = viewModel::rate)
        val details = gallery.details
        if (!details.isNullOrBlank()) {
            Text(details, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        if (gallery.tags.isNotEmpty()) {
            FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                for (tag in gallery.tags) {
                    TagPill(tag.name, onClick = { navigator.openTag(tag) })
                }
            }
        }
    }
}
