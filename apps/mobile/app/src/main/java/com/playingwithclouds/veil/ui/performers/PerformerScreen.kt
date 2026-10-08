package com.playingwithclouds.veil.ui.performers

import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.EntityFilter
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.FollowKind
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerDetail
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FollowButton
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.entity.EntityHeader
import com.playingwithclouds.veil.ui.entity.GalleryRow
import com.playingwithclouds.veil.ui.entity.ProfileInfo
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.paging.PagedList
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** One performer's profile, galleries and scenes. */
class PerformerViewModel(private val performerId: String) : ViewModel() {

    private val mutableDetail = MutableStateFlow<LoadState<PerformerDetail>>(LoadState.Loading)
    private val mutableGalleries = MutableStateFlow<List<GallerySummary>>(emptyList())

    /** The profile. */
    val detail: StateFlow<LoadState<PerformerDetail>> = mutableDetail

    /** The performer's galleries, first page. */
    val galleries: StateFlow<List<GallerySummary>> = mutableGalleries

    /** The performer's scenes, newest release first. */
    val scenes = PagedList(viewModelScope, PAGE_SIZE, { tagged -> tagged.scene.id }) { offset ->
        EntityRepository.scenes(EntityFilter(performerId = performerId), PAGE_SIZE, offset)
    }

    init {
        load()
        scenes.loadMore()
    }

    /** Loads the profile and galleries. */
    fun load() {
        viewModelScope.loadInto(mutableDetail) {
            requireNotNull(EntityRepository.performer(performerId)) { "Performer not found" }
        }
        viewModelScope.launch {
            try {
                mutableGalleries.value = EntityRepository.galleries(EntityFilter(performerId = performerId), GALLERY_ROW_SIZE, 0)
            } catch (error: Exception) {
                mutableGalleries.value = emptyList()
            }
        }
    }

    /** Toggles the favorite flag. */
    fun toggleFavorite() {
        val current = (mutableDetail.value as? LoadState.Loaded)?.value ?: return
        val favorite = !current.favorite
        mutableDetail.update { LoadState.Loaded(current.copy(favorite = favorite)) }
        viewModelScope.launch {
            try {
                EntityRepository.setPerformerFavorite(performerId, favorite)
            } catch (error: Exception) {
                mutableDetail.update { LoadState.Loaded(current) }
            }
        }
    }

    /** Asks the backend to fill in the performer's photo and bio from enrich plugins. */
    fun enrich() {
        viewModelScope.launch {
            try {
                EntityRepository.enrichPerformer(performerId)
            } catch (error: Exception) {
                // Best-effort: the job simply is not queued.
            }
        }
    }

    companion object {
        private const val PAGE_SIZE = 30
        private const val GALLERY_ROW_SIZE = 20
    }
}

/** A performer's page. */
@Composable
fun PerformerScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "performer-$id") { PerformerViewModel(id) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val galleries by viewModel.galleries.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Performer", onBack = navigator::back) }) { padding ->
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { performer ->
            PagedGrid(
                paged = viewModel.scenes,
                keyOf = { tagged -> tagged.scene.id },
                emptyText = "No videos for this performer yet.",
                modifier = Modifier.padding(padding),
                header = {
                    fullWidthItem("header") { PerformerHeader(performer, viewModel, navigator) }
                    fullWidthItem("galleries") { GalleryRow("Galleries", galleries, navigator::openGallery) }
                },
            ) { _, tagged ->
                SceneCard(tagged.scene, onClick = { navigator.openScene(tagged.scene.id) })
            }
        }
    }
}

/** Name, photo, actions, facts and tags of a performer. */
@Composable
private fun PerformerHeader(performer: PerformerDetail, viewModel: PerformerViewModel, navigator: AppNavigator) {
    EntityHeader(
        name = performer.name,
        imagePath = performer.imagePath,
        subtitle = "${performer.sceneCount} videos",
        actions = {
            FollowButton(FollowKind.PERFORMER, performer.id, navigator)
            FavoriteButton(performer.favorite, viewModel::toggleFavorite)
            RoundIconButton(VeilIcons.Refresh, contentDescription = "Fetch details", onClick = viewModel::enrich, size = 44.dp)
        },
    )
    ProfileInfo(performerFacts(performer), performer.details)
    TagChips(performer.tags, onClick = { tag -> navigator.openTag(tag.id) })
}

/** The round heart that toggles the favorite flag, filled red while set. */
@Composable
private fun FavoriteButton(favorite: Boolean, onToggle: () -> Unit) {
    if (favorite) {
        RoundIconButton(VeilIcons.HeartFilled, contentDescription = "Remove favorite", onClick = onToggle, size = 44.dp, tint = VeilColors.error)
        return
    }
    RoundIconButton(VeilIcons.Heart, contentDescription = "Favorite", onClick = onToggle, size = 44.dp)
}

/** The facts worth listing, in reading order. */
private fun performerFacts(performer: PerformerDetail): List<Pair<String, String?>> {
    return listOf(
        "Aliases" to performer.aliases.joinToString(", ").ifEmpty { null },
        "Born" to formatBirthdate(performer.birthdate),
        "Country" to performer.country,
        "Ethnicity" to performer.ethnicity,
        "Eyes" to performer.eyeColor,
        "Hair" to performer.hairColor,
        "Height" to performer.heightCm?.let { centimeters -> "$centimeters cm" },
        "Weight" to performer.weightKg?.let { kilograms -> "$kilograms kg" },
        "Measurements" to performer.measurements,
        "Career" to performer.careerLength,
    )
}

/** The birth date as stored; relative wording would be wrong for decades-old dates. */
private fun formatBirthdate(birthdate: String?): String? {
    if (birthdate == null) {
        return null
    }
    return birthdate.take(10)
}
