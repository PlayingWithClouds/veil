package com.playingwithclouds.veil.ui.face

import android.graphics.BitmapFactory
import android.util.Base64
import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import com.playingwithclouds.veil.data.FaceCandidate
import com.playingwithclouds.veil.data.FaceIdentification
import com.playingwithclouds.veil.data.FaceRepository
import com.playingwithclouds.veil.data.FaceVideoMatch
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.Avatar
import com.playingwithclouds.veil.ui.components.EmptyMessage
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.SecondaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.TextAction
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * The picture to ask about: a frame of [mediaId] ([grabFrame] returns it as a base64 JPEG, or null
 * when it cannot be read), the gallery image [imageId], or else the media's own picture. [id]
 * tells requests apart, so asking again about the same media searches again.
 */
data class FaceRequest(
    val mediaId: String,
    val imageId: String? = null,
    val grabFrame: (suspend () -> String?)? = null,
    val id: Long = System.nanoTime(),
)

/** What the sheet shows besides the answer: who was credited and which video is opening. */
data class FaceSheetState(
    val result: LoadState<FaceIdentification> = LoadState.Loading,
    /** The frame that was searched, as a base64 JPEG. */
    val frameBase64: String? = null,
    /** Names credited to the media in this sheet. */
    val credited: Set<String> = emptySet(),
    val creditingName: String? = null,
    val openingVideoUrl: String? = null,
    val message: String? = null,
)

/** Runs one face search and the credits and video opens that follow it. */
class FaceIdentifyViewModel : ViewModel() {

    private val mutableState = MutableStateFlow(FaceSheetState())

    /** The sheet state. */
    val state: StateFlow<FaceSheetState> = mutableState

    /** Asks every face-recognition plugin about the request's picture. */
    fun identify(request: FaceRequest) {
        mutableState.value = FaceSheetState()
        viewModelScope.launch {
            try {
                val frame = request.grabFrame?.invoke()
                mutableState.update { current -> current.copy(frameBase64 = frame) }
                val result = FaceRepository.identify(request.mediaId, request.imageId, frame)
                mutableState.update { current -> current.copy(result = LoadState.Loaded(result)) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(result = LoadState.Failed(error.displayMessage())) }
            }
        }
    }

    /** Credits the candidate's performer to the media. */
    fun credit(mediaId: String, candidate: FaceCandidate) {
        mutableState.update { current -> current.copy(creditingName = candidate.name, message = null) }
        viewModelScope.launch {
            try {
                FaceRepository.creditPerformer(mediaId, candidate.name)
                mutableState.update { current -> current.copy(credited = current.credited + candidate.name, creditingName = null) }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(creditingName = null, message = error.displayMessage()) }
            }
        }
    }

    /** Stores a similar video as a scene and hands its id to [onOpened]. */
    fun openVideo(video: FaceVideoMatch, onOpened: (String) -> Unit) {
        mutableState.update { current -> current.copy(openingVideoUrl = video.url, message = null) }
        viewModelScope.launch {
            try {
                val sceneId = FaceRepository.openVideoAsScene(video)
                mutableState.update { current -> current.copy(openingVideoUrl = null) }
                onOpened(sceneId)
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                mutableState.update { current -> current.copy(openingVideoUrl = null, message = "Could not open that video") }
            }
        }
    }
}

/**
 * "Who is this?": performers who look like the person in the request's picture, best first, each
 * creditable to the media, then tube videos with a similar face (opened, never credited).
 * [onPickAnotherMoment] offers to close the sheet and search another frame.
 */
@Composable
fun FaceIdentifySheet(
    request: FaceRequest,
    navigator: AppNavigator,
    onDismiss: () -> Unit,
    onPickAnotherMoment: (() -> Unit)? = null,
) {
    val viewModel = viewModel(key = "face-${request.mediaId}") { FaceIdentifyViewModel() }
    val state by viewModel.state.collectAsStateWithLifecycle()
    LaunchedEffect(request.id) { viewModel.identify(request) }
    val frame = remember(state.frameBase64) { decodeFrame(state.frameBase64) }

    VeilBottomSheet(onDismissRequest = onDismiss) {
        SectionHeading("Who is this?", Modifier.gutterPadding())
        LazyColumn(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            item { SearchedPicture(frame, (state.result as? LoadState.Loaded)?.value?.pictureUrl) }
            if (onPickAnotherMoment != null) {
                item { TextAction("Pick another moment", onClick = onPickAnotherMoment, modifier = Modifier.gutterPadding()) }
            }
            val message = state.message
            if (message != null) {
                item { Text(message, Modifier.gutterPadding(), color = VeilColors.error, style = MaterialTheme.typography.bodySmall) }
            }
            when (val result = state.result) {
                is LoadState.Loading -> item { Searching() }
                is LoadState.Failed -> item { EmptyMessage(result.message) }
                is LoadState.Loaded -> answer(result.value, state, request.mediaId, viewModel, navigator, onDismiss)
            }
        }
    }
}

/** The candidates, the similar videos and the plugins that failed. */
private fun LazyListScope.answer(
    result: FaceIdentification,
    state: FaceSheetState,
    mediaId: String,
    viewModel: FaceIdentifyViewModel,
    navigator: AppNavigator,
    onDismiss: () -> Unit,
) {
    if (result.candidates.isEmpty() && result.videos.isEmpty()) {
        item { EmptyMessage(noMatchText(result)) }
    }
    items(result.candidates, key = { candidate -> "candidate-${candidate.name}" }) { candidate ->
        CandidateRow(
            candidate,
            credited = candidate.name in state.credited,
            crediting = state.creditingName == candidate.name,
            onCredit = { viewModel.credit(mediaId, candidate) },
            onOpen = { performerId ->
                onDismiss()
                navigator.openPerformer(performerId)
            },
        )
    }
    if (result.videos.isNotEmpty()) {
        item { SectionHeading("Similar videos", Modifier.gutterPadding()) }
        items(result.videos, key = { video -> "video-${video.url}" }) { video ->
            SimilarVideoRow(video, opening = state.openingVideoUrl == video.url, viewModel, navigator, onDismiss)
        }
    }
    items(result.failures, key = { failure -> "failure-$failure" }) { failure ->
        Text(failure, Modifier.gutterPadding(), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
    }
}

/** Why nothing came back: no plugin answered, or none found a match. */
private fun noMatchText(result: FaceIdentification): String {
    if (result.sources.isEmpty()) {
        return "No face search plugin answered. Turn one on in Plugins."
    }
    return "Nobody matching found."
}

/** Height of the searched picture on top of the sheet. */
private val SearchedPictureHeight = 160.dp

/** The picture that was searched: the grabbed frame, else the stored picture the backend sent. */
@Composable
private fun SearchedPicture(frame: ImageBitmap?, pictureUrl: String?) {
    val modifier = Modifier.gutterPadding().height(SearchedPictureHeight).clip(VeilShapes.card)
    if (frame != null) {
        Image(frame, contentDescription = "Searched frame", modifier = modifier, contentScale = ContentScale.Fit)
        return
    }
    if (pictureUrl != null) {
        RemoteImage(pictureUrl, modifier.fillMaxWidth(), ContentScale.Fit, contentDescription = "Searched picture")
    }
}

/** The spinner while the services answer, which takes a while. */
@Composable
private fun Searching() {
    Column(
        Modifier.fillMaxWidth().padding(VeilSpacing.extraLarge),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
    ) {
        Spinner()
        Text("Asking the face search services...", color = VeilColors.contentMuted, style = MaterialTheme.typography.bodyMedium)
    }
}

/** A candidate: photo, name, similarity and sources, then Credit (or Open once in the library). */
@Composable
private fun CandidateRow(candidate: FaceCandidate, credited: Boolean, crediting: Boolean, onCredit: () -> Unit, onOpen: (String) -> Unit) {
    val performerId = candidate.performerId
    var rowModifier = Modifier.fillMaxWidth()
    if (performerId != null) {
        rowModifier = rowModifier.pressClickable { onOpen(performerId) }
    }
    Row(
        rowModifier.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Avatar(candidate.imageUrl, Modifier.size(52.dp))
        Column(Modifier.weight(1f)) {
            Text(candidate.name, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(candidateFacts(candidate), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted, maxLines = 1, overflow = TextOverflow.Ellipsis)
        }
        CreditAction(credited, crediting, onCredit)
    }
}

/** "87% · In library · facedigger". */
private fun candidateFacts(candidate: FaceCandidate): String {
    val facts = mutableListOf(formatSimilarity(candidate.similarity))
    if (candidate.performerId != null) {
        facts.add("In library")
    }
    facts.addAll(candidate.sources)
    return facts.joinToString(" · ")
}

/** Credit, a spinner while crediting, or a note once done. */
@Composable
private fun CreditAction(credited: Boolean, crediting: Boolean, onCredit: () -> Unit) {
    if (crediting) {
        Spinner(size = 22.dp)
        return
    }
    if (credited) {
        Text("Credited", style = MaterialTheme.typography.labelLarge, color = VeilColors.accent)
        return
    }
    SecondaryButton("Credit", onClick = onCredit)
}

/** A video with a similar face: opens as a scene when a plugin serves it, else in the browser. */
@Composable
private fun SimilarVideoRow(video: FaceVideoMatch, opening: Boolean, viewModel: FaceIdentifyViewModel, navigator: AppNavigator, onDismiss: () -> Unit) {
    val uriHandler = LocalUriHandler.current
    val open = {
        if (video.scenePlugin != null) {
            viewModel.openVideo(video) { sceneId ->
                onDismiss()
                navigator.openScene(sceneId)
            }
        } else {
            uriHandler.openUri(video.url)
        }
    }
    Row(
        Modifier.fillMaxWidth().pressClickable(enabled = !opening, onClick = open).padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.extraSmall),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(video.thumbnailUrl, Modifier.size(width = 120.dp, height = 68.dp).clip(VeilShapes.card))
        Column(Modifier.weight(1f)) {
            Text(video.title, style = MaterialTheme.typography.bodyMedium, maxLines = 2, overflow = TextOverflow.Ellipsis)
            Text(formatSimilarity(video.similarity), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
        if (opening) {
            Spinner(size = 22.dp)
        }
    }
}

private const val PERCENT = 100

/** A 0 to 1 similarity as "87%". */
private fun formatSimilarity(similarity: Double): String {
    return "${(similarity * PERCENT).toInt()}%"
}

/** The grabbed frame as an image to show, or null when there is none or it does not decode. */
private fun decodeFrame(frameBase64: String?): ImageBitmap? {
    if (frameBase64 == null) {
        return null
    }
    val bytes = runCatching { Base64.decode(frameBase64, Base64.NO_WRAP) }.getOrNull() ?: return null
    return BitmapFactory.decodeByteArray(bytes, 0, bytes.size)?.asImageBitmap()
}
