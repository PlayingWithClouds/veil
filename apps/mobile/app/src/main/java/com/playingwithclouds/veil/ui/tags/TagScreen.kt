package com.playingwithclouds.veil.ui.tags

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.CollectionSummary
import com.playingwithclouds.veil.data.EntityFilter
import com.playingwithclouds.veil.data.EntityRepository
import com.playingwithclouds.veil.data.FollowKind
import com.playingwithclouds.veil.data.GallerySummary
import com.playingwithclouds.veil.data.PerformerSummary
import com.playingwithclouds.veil.data.StudioSummary
import com.playingwithclouds.veil.data.TagDetail
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Badge
import com.playingwithclouds.veil.ui.components.FollowButton
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.SceneCard
import com.playingwithclouds.veil.ui.components.SectionTitle
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.fullWidthItem
import com.playingwithclouds.veil.ui.entity.CollectionRow
import com.playingwithclouds.veil.ui.entity.EntityHeader
import com.playingwithclouds.veil.ui.entity.GalleryRow
import com.playingwithclouds.veil.ui.entity.PerformerRow
import com.playingwithclouds.veil.ui.entity.ProfileInfo
import com.playingwithclouds.veil.ui.entity.StudioRow
import com.playingwithclouds.veil.ui.loadInto
import com.playingwithclouds.veil.ui.paging.PagedList
import com.playingwithclouds.veil.util.tagLabel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/** What else is filed under a tag besides its scenes. */
data class TagRelations(
    val performers: List<PerformerSummary> = emptyList(),
    val studios: List<StudioSummary> = emptyList(),
    val galleries: List<GallerySummary> = emptyList(),
    val collections: List<CollectionSummary> = emptyList(),
)

/** The hub for everything a tag touches. */
class TagViewModel(private val tagId: String) : ViewModel() {

    private val mutableDetail = MutableStateFlow<LoadState<TagDetail>>(LoadState.Loading)
    private val mutableRelations = MutableStateFlow(TagRelations())

    /** The tag. */
    val detail: StateFlow<LoadState<TagDetail>> = mutableDetail

    /** Performers, studios, galleries and collections that carry the tag. */
    val relations: StateFlow<TagRelations> = mutableRelations

    /** Scenes with the tag: direct matches first, then ones inherited from studio or performers. */
    val scenes = PagedList(viewModelScope, PAGE_SIZE, { tagged -> tagged.scene.id }) { offset ->
        EntityRepository.scenes(EntityFilter(tagId = tagId), PAGE_SIZE, offset)
    }

    init {
        load()
        scenes.loadMore()
    }

    /** Loads the tag and what is filed under it. */
    fun load() {
        viewModelScope.loadInto(mutableDetail) {
            requireNotNull(EntityRepository.tag(tagId)) { "Tag not found" }
        }
        viewModelScope.launch { loadRelations() }
    }

    /** Fetches the related rows; a failing row stays empty. */
    private suspend fun loadRelations() {
        mutableRelations.value = TagRelations(
            performers = quietly { EntityRepository.performers(null, tagId, ROW_SIZE, 0) },
            studios = quietly { EntityRepository.studios(null, tagId, ROW_SIZE, 0) },
            galleries = quietly { EntityRepository.galleries(EntityFilter(tagId = tagId), ROW_SIZE, 0) },
            collections = quietly { EntityRepository.collectionsForTag(tagId, ROW_SIZE) },
        )
    }

    /** Runs a fetch, returning an empty list when it fails. */
    private suspend fun <T> quietly(fetch: suspend () -> List<T>): List<T> {
        return try {
            fetch()
        } catch (error: Exception) {
            emptyList()
        }
    }

    companion object {
        private const val PAGE_SIZE = 30
        private const val ROW_SIZE = 20
    }
}

/** A tag's page. */
@Composable
fun TagScreen(id: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "tag-$id") { TagViewModel(id) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val relations by viewModel.relations.collectAsStateWithLifecycle()
    Scaffold(topBar = { VeilTopBar("Tag", onBack = navigator::back) }) { padding ->
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(padding)) { tag ->
            PagedGrid(
                paged = viewModel.scenes,
                keyOf = { tagged -> tagged.scene.id },
                emptyText = "No videos with this tag yet.",
                modifier = Modifier.padding(padding),
                header = {
                    fullWidthItem("header") { TagHeader(tag, relations, navigator) }
                    fullWidthItem("scenes-title") { SectionTitle("Videos") }
                },
            ) { _, tagged ->
                Box {
                    SceneCard(tagged.scene, onClick = { navigator.openScene(tagged.scene.id) })
                    if (tagged.inherited) {
                        Badge("via studio or performer", Modifier.align(Alignment.TopEnd).padding(6.dp))
                    }
                }
            }
        }
    }
}

/** Name, description, follow button and the related rows of a tag. */
@Composable
private fun TagHeader(tag: TagDetail, relations: TagRelations, navigator: AppNavigator) {
    EntityHeader(
        name = tagLabel(tag.name),
        imagePath = null,
        subtitle = "${tag.sceneCount} videos",
        showAvatar = false,
        actions = { FollowButton(FollowKind.TAG, tag.id, navigator) },
    )
    ProfileInfo(listOf("Also known as" to tag.aliases.joinToString(", ").ifEmpty { null }), tag.description)
    PerformerRow("Performers", relations.performers, navigator::openPerformer)
    StudioRow("Studios", relations.studios, navigator::openStudio)
    GalleryRow("Galleries", relations.galleries, navigator::openGallery)
    CollectionRow("Collections", relations.collections, navigator::openCollection)
}
