package com.playingwithclouds.veil.ui.media

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.AppServices
import com.playingwithclouds.veil.data.AudioDetail
import com.playingwithclouds.veil.data.LoggedEvent
import com.playingwithclouds.veil.data.MediaRepository
import com.playingwithclouds.veil.data.PlayableStream
import com.playingwithclouds.veil.data.VideoRepository
import com.playingwithclouds.veil.graphql.type.EventKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.LoadStateContent
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.DetailScreen
import com.playingwithclouds.veil.ui.components.LocalTopBarInset
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.RoundIconButton
import com.playingwithclouds.veil.ui.design.ShowAmbientGlow
import com.playingwithclouds.veil.ui.design.TagPill
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.displayMessage
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.ui.video.PlayerSeekBar
import com.playingwithclouds.veil.ui.video.WatchEvent
import com.playingwithclouds.veil.ui.video.WatchTracking
import com.playingwithclouds.veil.ui.video.playStream
import com.playingwithclouds.veil.ui.video.rememberVideoPlayer
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.util.UUID

/** How far the skip buttons jump. */
private const val SKIP_MILLISECONDS = 15_000L

/** How often the position on screen follows the player. */
private const val POSITION_TICK_MILLISECONDS = 500L

/** An audio listened this far counts as finished, so it starts over instead of resuming. */
private const val FINISHED_FRACTION = 0.95

private const val MILLISECONDS_PER_SECOND = 1000.0

private val PlayButtonSize = 64.dp

/** One audio: opened (its pages read for the streams), the first stream that resolves playing, listening logged. */
class AudioViewModel(private val audioId: String) : ViewModel() {

    private val detailState = MutableStateFlow<LoadState<AudioDetail>>(LoadState.Loading)
    private val streamState = MutableStateFlow<PlayableStream?>(null)
    private val sessionId = UUID.randomUUID().toString()

    val detail: StateFlow<LoadState<AudioDetail>> = detailState.asStateFlow()

    /** The stream to play, once one resolved. */
    val stream: StateFlow<PlayableStream?> = streamState.asStateFlow()

    init {
        load()
    }

    /** Opens the audio and resolves its first playable stream. */
    fun load() {
        detailState.value = LoadState.Loading
        viewModelScope.launch {
            try {
                val audio = requireNotNull(MediaRepository.openAudio(audioId)) { "This audio is gone." }
                detailState.value = LoadState.Loaded(audio)
                log(EventKind.view, emptyMap())
                streamState.value = audio.streamUrls.firstNotNullOfOrNull { url -> runCatching { VideoRepository.play(url) }.getOrNull() }
            } catch (error: CancellationException) {
                throw error
            } catch (error: Exception) {
                detailState.value = LoadState.Failed(error.displayMessage())
            }
        }
    }

    /** Where listening starts: the resume point, unless it was listened (nearly) to the end. */
    fun startPositionMilliseconds(): Long {
        val audio = (detailState.value as? LoadState.Loaded)?.value ?: return 0
        val progress = audio.progressSeconds ?: return 0
        val duration = audio.summary.durationSeconds
        if (duration != null && progress >= duration * FINISHED_FRACTION) {
            return 0
        }
        return (progress * MILLISECONDS_PER_SECOND).toLong()
    }

    /** Logs what the listening session reports, as the web logs audio. */
    fun onWatchEvent(event: WatchEvent) {
        when (event) {
            is WatchEvent.Progress -> log(EventKind.progress, mapOf("position" to event.positionSeconds))
            is WatchEvent.Span -> log(EventKind.heatSpan, mapOf("start" to event.startSeconds, "end" to event.endSeconds))
            is WatchEvent.Finish -> log(EventKind.finish, mapOf("position" to event.positionSeconds))
        }
    }

    /** Sends what the session still holds when the page closes. */
    override fun onCleared() {
        AppServices.events.flushNow()
    }

    /** Logs an event about this audio in this page's session. */
    private fun log(kind: EventKind, data: Map<String, Any>) {
        AppServices.events.log(LoggedEvent(kind, RecordKind.audio, audioId, data, sessionId))
    }
}

/** An audio's page: its cover with the controls under it, then the description and tags. */
@Composable
fun AudioScreen(audioId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "audio-$audioId") { AudioViewModel(audioId) }
    val detail by viewModel.detail.collectAsStateWithLifecycle()
    val stream by viewModel.stream.collectAsStateWithLifecycle()
    val player = rememberVideoPlayer()
    LaunchedEffect(stream) {
        val playable = stream ?: return@LaunchedEffect
        player.playStream(playable, viewModel.startPositionMilliseconds())
    }
    WatchTracking(player, onEvent = viewModel::onWatchEvent)
    val title = (detail as? LoadState.Loaded)?.value?.summary?.title.orEmpty()
    ShowAmbientGlow((detail as? LoadState.Loaded)?.value?.summary?.coverUrl)
    DetailScreen(title, onBack = navigator::back) {
        val barInset = LocalTopBarInset.current
        LoadStateContent(detail, onRetry = viewModel::load, modifier = Modifier.padding(top = barInset)) { audio ->
            Column(
                Modifier.verticalScroll(rememberScrollState()).gutterPadding().padding(top = VeilSpacing.large + barInset, bottom = VeilSpacing.huge),
                verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
            ) {
                RemoteImage(audio.summary.coverUrl, Modifier.fillMaxWidth().aspectRatio(1f).clip(VeilShapes.card))
                Text(audio.summary.title, style = MaterialTheme.typography.titleLarge, color = VeilColors.content)
                val voice = audio.summary.voice
                if (voice != null) {
                    Text(voice, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
                }
                if (stream == null) {
                    Text("No stream to play yet.", style = MaterialTheme.typography.bodySmall, color = VeilColors.contentFaint)
                } else {
                    AudioControls(player)
                }
                val details = audio.details
                if (!details.isNullOrBlank()) {
                    Text(details, style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
                }
                if (audio.tags.isNotEmpty()) {
                    FlowRow(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
                        for (tag in audio.tags) {
                            TagPill(tag.name, onClick = { navigator.openTag(tag) })
                        }
                    }
                }
            }
        }
    }
}

/** The seek bar, then back 15 s, play or pause, and ahead 15 s. */
@Composable
private fun AudioControls(player: ExoPlayer) {
    var position by remember { mutableLongStateOf(0L) }
    var buffered by remember { mutableLongStateOf(0L) }
    var duration by remember { mutableLongStateOf(0L) }
    var playing by remember { mutableStateOf(false) }
    var scrubbing by remember { mutableStateOf(false) }
    LaunchedEffect(player) {
        while (true) {
            if (!scrubbing) {
                position = player.currentPosition
            }
            buffered = player.bufferedPosition
            duration = maxOf(0L, player.duration)
            playing = player.isPlaying
            delay(POSITION_TICK_MILLISECONDS)
        }
    }
    Column(verticalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
        PlayerSeekBar(
            positionMilliseconds = position,
            bufferedMilliseconds = buffered,
            durationMilliseconds = duration,
            onSeek = { target ->
                position = target
                player.seekTo(target)
            },
            onScrubbingChange = { active -> scrubbing = active },
        )
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly, verticalAlignment = Alignment.CenterVertically) {
            IconTap(VeilIcons.StepBack, contentDescription = "Back 15 seconds", tint = VeilColors.content, onClick = {
                player.seekTo(maxOf(0L, player.currentPosition - SKIP_MILLISECONDS))
            })
            var icon = VeilIcons.Play
            var label = "Play"
            if (playing) {
                icon = VeilIcons.Pause
                label = "Pause"
            }
            RoundIconButton(icon, contentDescription = label, size = PlayButtonSize, onClick = {
                if (player.isPlaying) {
                    player.pause()
                } else {
                    player.play()
                }
            })
            IconTap(VeilIcons.StepForward, contentDescription = "Ahead 15 seconds", tint = VeilColors.content, onClick = {
                player.seekTo(player.currentPosition + SKIP_MILLISECONDS)
            })
        }
    }
}
