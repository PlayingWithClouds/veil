package com.playingwithclouds.veil.ui.studios

import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.EntityFilter
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.FollowKind
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.StudioDetail
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FollowButton
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.bleed
import com.playingwithclouds.veil.ui.entity.EntityHeader
import com.playingwithclouds.veil.ui.entity.EntitySceneFetch
import com.playingwithclouds.veil.ui.entity.EntitySceneFetchBar
import com.playingwithclouds.veil.ui.entity.GalleryRow
import com.playingwithclouds.veil.ui.entity.ProfileInfo
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** One studio's profile, galleries and scenes. */
class StudioViewModel(private val studioId: String) : ViewModel() {

    private val mutableDetail = MutableStateFlow<LoadState<StudioDetail>>(LoadState.Loading)
    private val mutableGalleries = MutableStateFlow<List<GallerySummary>>(emptyList())

    /** The profile. */
    val detail: StateFlow<LoadState<StudioDetail>> = mutableDetail

    /** The studio's galleries, first page. */
    val galleries: StateFlow<List<GallerySummary>> = mutableGalleries

    /** The studio's scenes, newest release first. */
    val scenes = PagedList(viewModelScope, PAGE_SIZE, { tagged -> tagged.scene.id }) { offset ->
        EntityRepository.scenes(EntityFilter(studioId = studioId), PAGE_SIZE, offset)
    }

    /** The search for the studio's videos on every site, started when the page opens. */
    val siteFetch = EntitySceneFetch(viewModelScope, studioId) {
        scenes.refresh()
        refreshProfile()
    }

    init {
        load()
        scenes.loadMore()
        siteFetch.start()
    }

    /** Reloads the profile in place (its video count), keeping the current one on failure. */
    private fun refreshProfile() {
        viewModelScope.launch {
            try {
                val studio = EntityRepository.studio(studioId) ?: return@launch
                mutableDetail.value = LoadState.Loaded(studio)
            } catch (error: Exception) {
                // The page keeps the profile it shows.
            }
        }
    }

    /** Loads the profile and galleries. */
    fun load() {
        viewModelScope.loadInto(mutableDetail) {
            requireNotNull(EntityRepository.studio(studioId)) { "Studio not found" }
        }
        viewModelScope.launch {
            try {
                mutableGalleries.value = EntityRepository.galleries(EntityFilter(studioId = studioId), GALLERY_ROW_SIZE, 0)
            } catch (error: Exception) {
                mutableGalleries.value = emptyList()
            }
        }
    }

    companion object {
        private const val PAGE_SIZE = 30
        private const val GALLERY_ROW_SIZE = 20
    }
}

/** A studio's page. */
@Composable
fun StudioScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "studio-$id") { StudioViewModel(id) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val galleries by viewModel.galleries.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Studio", onBack = navigator::back) }) { padding ->
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { studio ->
            PagedGrid(
                paged = viewModel.scenes,
                keyOf = { tagged -> tagged.scene.id },
                emptyText = "No videos for this studio yet.",
                modifier = Modifier.padding(padding),
                header = {
                    fullWidthItem("header") { StudioHeader(studio, navigator) }
                    fullWidthItem("galleries") { GalleryRow("Galleries", galleries, navigator::openGallery) }
                    fullWidthItem("site-fetch") { EntitySceneFetchBar(viewModel.siteFetch) }
                },
            ) { _, tagged ->
                SceneCard(tagged.scene, onClick = { navigator.openScene(tagged.scene.id) })
            }
        }
    }
}

/** Name, logo, actions, facts and tags of a studio. */
@Composable
private fun StudioHeader(studio: StudioDetail, navigator: AppNavigator) {
    EntityHeader(
        name = studio.name,
        imagePath = studio.imagePath,
        subtitle = "${studio.sceneCount} videos",
        actions = { FollowButton(FollowKind.STUDIO, studio.id, navigator) },
    )
    val parent = studio.parent
    if (parent != null) {
        // Pulled back by the action's own inset so its text lines up with the gutter.
        TextAction("Part of ${parent.name}", onClick = { navigator.openStudio(parent.id) }, modifier = Modifier.offset(x = -VeilSpacing.small))
    }
    ProfileInfo(listOf("Aliases" to studio.aliases.joinToString(", ").ifEmpty { null }, "Website" to studio.url), studio.details)
    TagChips(studio.tags, onClick = { tag -> navigator.openTag(tag.id) }, modifier = Modifier.bleed())
}
