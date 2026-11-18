package com.playingwithclouds.veil.ui.uploader

import androidx.compose.runtime.Composable
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.UploaderRepository
import com.playingwithclouds.veil.data.VideoSummary
import com.playingwithclouds.veil.data.VideoUploader
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.PagedGrid
import com.playingwithclouds.veil.ui.components.VideoCard
import com.playingwithclouds.veil.ui.paging.PagedList

/** Videos per page read from the uploader's channel page. */
private const val VIDEO_PAGE = 24

/** The videos on an uploader's channel page on their site, read page by page. */
class UploaderViewModel(uploader: VideoUploader) : ViewModel() {

    // A site's page can hold fewer videos than asked while more follow, so only an empty page ends the list.
    val videos = PagedList<VideoSummary>(viewModelScope, VIDEO_PAGE, { video -> video.id }, isLastPage = { page -> page.isEmpty() }) { offset ->
        UploaderRepository.videos(uploader, VIDEO_PAGE, offset)
    }

    init {
        videos.loadMore()
    }
}

/** One uploader of a tube site: their videos, fetched from their channel page there. */
@Composable
fun UploaderScreen(uploader: VideoUploader, navigator: AppNavigator) {
    val viewModel = viewModel(key = "uploader-${uploader.plugin}-${uploader.url}") { UploaderViewModel(uploader) }
    DetailScreen(uploader.name, onBack = navigator::back) {
        PagedGrid(
            paged = viewModel.videos,
            keyOf = { video -> video.id },
            emptyText = "No videos found on their ${uploader.plugin} page.",
        ) { _, video ->
            VideoCard(video, onClick = { navigator.openVideo(video.id) })
        }
    }
}
