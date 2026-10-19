package com.playingwithclouds.veil.ui.cams

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.common.C
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import com.playingwithclouds.veil.data.CamModel
import com.playingwithclouds.veil.data.CamRecording
import com.playingwithclouds.veil.data.CamRepository
import com.playingwithclouds.veil.data.EntityRef
import com.playingwithclouds.veil.ui.AppNavigator
import com.playingwithclouds.veil.ui.LoadState
import com.playingwithclouds.veil.ui.components.FailedMessage
import com.playingwithclouds.veil.ui.components.LoadingIndicator
import com.playingwithclouds.veil.ui.components.ReactionButton
import com.playingwithclouds.veil.ui.components.ReactionRowInset
import com.playingwithclouds.veil.ui.components.RemoteImage
import com.playingwithclouds.veil.ui.components.TagChips
import com.playingwithclouds.veil.ui.components.VeilTopBar
import com.playingwithclouds.veil.ui.components.pressClickable
import com.playingwithclouds.veil.ui.design.Pill
import com.playingwithclouds.veil.ui.design.PrimaryButton
import com.playingwithclouds.veil.ui.design.SectionHeading
import com.playingwithclouds.veil.ui.design.SheetAction
import com.playingwithclouds.veil.ui.design.Spinner
import com.playingwithclouds.veil.ui.design.VeilBottomSheet
import com.playingwithclouds.veil.ui.design.VeilCard
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilSnackbarHost
import com.playingwithclouds.veil.ui.design.VeilSwitch
import com.playingwithclouds.veil.ui.design.fadingEdges
import com.playingwithclouds.veil.ui.design.gutterPadding
import com.playingwithclouds.veil.ui.scene.PlayerExtras
import com.playingwithclouds.veil.ui.scene.ScenePlayerView
import com.playingwithclouds.veil.ui.scene.liveLagMilliseconds
import com.playingwithclouds.veil.ui.scene.playStream
import com.playingwithclouds.veil.ui.scene.rememberFullscreenState
import com.playingwithclouds.veil.ui.scene.rememberPictureInPicture
import com.playingwithclouds.veil.ui.scene.rememberPlayerOptions
import com.playingwithclouds.veil.ui.scene.rememberPlayerTracks
import com.playingwithclouds.veil.ui.scene.rememberScenePlayer
import com.playingwithclouds.veil.ui.scene.selectedQualityHeight
import com.playingwithclouds.veil.ui.scene.videoQualities
import com.playingwithclouds.veil.ui.scene.videoQualityLabel
import com.playingwithclouds.veil.ui.theme.VeilColors
import com.playingwithclouds.veil.ui.theme.VeilShapes
import com.playingwithclouds.veil.ui.theme.VeilSpacing
import com.playingwithclouds.veil.util.formatBytes
import com.playingwithclouds.veil.util.formatClock
import com.playingwithclouds.veil.util.formatRelativeTime
import com.playingwithclouds.veil.util.parseTimestamp
import java.time.Duration
import java.time.Instant
import kotlinx.coroutines.delay

private const val PLAYER_ASPECT_RATIO = 16f / 9f
private const val MILLISECONDS_PER_SECOND = 1000.0
private const val CLOCK_TICK_MILLISECONDS = 1000L

/**
 * A live room: the stream in the scene player's live mode (rewind within the room's buffer, Go
 * live), follow, record now or from the past, auto-record, then the room's finished recordings.
 */
@Composable
fun CamRoomScreen(camModelId: String, navigator: AppNavigator) {
    val viewModel = viewModel(key = "cam-$camModelId") { CamRoomViewModel(camModelId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val player = rememberScenePlayer()
    val fullscreen = rememberFullscreenState(player)
    val pictureInPicture = rememberPictureInPicture(player)
    val options = rememberPlayerOptions(player)
    val snackbarHostState = remember { SnackbarHostState() }
    var showingRecordMenu by remember { mutableStateOf(false) }
    PollWhileStarted(CAM_POLL_MILLISECONDS) { viewModel.refreshRecordings() }

    LaunchedEffect(state.streamGeneration) {
        val stream = state.stream ?: return@LaunchedEffect
        player.playStream(stream, C.TIME_UNSET)
    }
    LaunchedEffect(state.message) {
        val note = state.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(note)
        viewModel.messageShown()
    }
    LiveStreamReporting(player, viewModel)
    if (showingRecordMenu) {
        RecordSheet(player, onRecord = { quality, fromSecondsAgo ->
            showingRecordMenu = false
            viewModel.startRecording(quality, fromSecondsAgo)
        }, onDismiss = { showingRecordMenu = false })
    }

    val model = state.model
    if (model is LoadState.Failed) {
        Scaffold(topBar = { VeilTopBar("Cam", onBack = navigator::back) }) { padding ->
            FailedMessage(model.message, onRetry = viewModel::load, modifier = Modifier.padding(padding))
        }
        return
    }
    val room = (model as? LoadState.Loaded)?.value
    val immersive = fullscreen.isFullscreen || pictureInPicture.isActive
    Scaffold(
        topBar = {
            if (!immersive) {
                VeilTopBar(room?.displayName.orEmpty(), onBack = navigator::back)
            }
        },
        snackbarHost = { VeilSnackbarHost(snackbarHostState) },
    ) { padding ->
        Column(Modifier.padding(padding)) {
            var boxModifier = Modifier.fillMaxWidth().aspectRatio(PLAYER_ASPECT_RATIO)
            if (immersive) {
                boxModifier = Modifier.fillMaxSize()
            }
            Box(boxModifier.background(VeilColors.canvas), contentAlignment = Alignment.Center) {
                val stream = state.stream
                if (stream != null && !state.offline) {
                    ScenePlayerView(
                        player,
                        stream = stream,
                        title = room?.displayName.orEmpty(),
                        fullscreen = fullscreen,
                        pictureInPicture = pictureInPicture,
                        options = options,
                        extras = PlayerExtras(live = true, onEnterPictureInPicture = pictureInPicture::enterPictureInPicture),
                        related = emptyList(),
                        onOpenScene = {},
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    RoomPlaceholder(room, state, onRetry = viewModel::load)
                }
            }
            if (!immersive && room != null) {
                RoomContent(room, state, viewModel, navigator, onOpenRecordMenu = { showingRecordMenu = true })
            }
        }
    }
}

/** Tells the view model when the stream fails, plays or ends, so it can fetch a fresh URL or show the room offline. */
@Composable
private fun LiveStreamReporting(player: ExoPlayer, viewModel: CamRoomViewModel) {
    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onPlayerError(error: PlaybackException) {
                viewModel.onPlayerError()
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                if (isPlaying) {
                    viewModel.onPlaying()
                }
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                if (playbackState == Player.STATE_ENDED) {
                    viewModel.onStreamEnded()
                }
            }
        }
        player.addListener(listener)
        onDispose { player.removeListener(listener) }
    }
}

/** The room's snapshot behind a spinner while connecting, or the offline note with a retry. */
@Composable
private fun RoomPlaceholder(room: CamModel?, state: CamRoomState, onRetry: () -> Unit) {
    if (room == null) {
        LoadingIndicator()
        return
    }
    RemoteImage(room.thumbnailUrl, Modifier.fillMaxSize(), ContentScale.Crop)
    Box(Modifier.fillMaxSize().background(VeilColors.scrim), contentAlignment = Alignment.Center) {
        if (!state.offline) {
            Spinner()
            return@Box
        }
        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            Text("${room.displayName} is offline", color = VeilColors.content, style = MaterialTheme.typography.titleSmall)
            PrimaryButton("Try again", onClick = onRetry, icon = VeilIcons.Refresh)
        }
    }
}

/** Everything below the player: who, the actions, auto-record, tags and recordings. */
@Composable
private fun RoomContent(room: CamModel, state: CamRoomState, viewModel: CamRoomViewModel, navigator: AppNavigator, onOpenRecordMenu: () -> Unit) {
    LazyColumn(contentPadding = PaddingValues(bottom = VeilSpacing.large)) {
        item { RoomTitleBlock(room, state.recording) }
        item { RoomActions(room, state, viewModel, onOpenRecordMenu) }
        if (room.autoRecord) {
            item { AutoRecordQuality(room, state.autoRecordBusy, viewModel) }
        }
        item {
            TagChips(room.tags.map { tag -> EntityRef(tag, tag, null) }, onClick = { tag -> navigator.openCams(tag.name) })
        }
        if (state.pastRecordings.isNotEmpty()) {
            item { SectionHeading("Recordings", Modifier.gutterPadding()) }
            items(state.pastRecordings, key = { recording -> recording.id }) { recording ->
                PastRecordingRow(recording, onOpen = { sceneId -> navigator.openScene(sceneId) })
            }
        }
    }
}

/** Name, live state, viewers and site, the room's subject and what is being recorded. */
@Composable
private fun RoomTitleBlock(room: CamModel, recording: CamRecording?) {
    Column(Modifier.padding(VeilSpacing.gutter), verticalArrangement = Arrangement.spacedBy(VeilSpacing.extraSmall)) {
        Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
            Text(room.displayName, Modifier.weight(1f, fill = false), style = MaterialTheme.typography.headlineSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
            OnlineLabel(room.online)
        }
        Text(roomFacts(room), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        val subject = room.subject
        if (!subject.isNullOrBlank()) {
            Text(subject, style = MaterialTheme.typography.bodyMedium, color = VeilColors.contentMuted)
        }
        if (recording != null) {
            RecordingNote(recording)
        }
    }
}

/** "1.2k watching · chaturbate", or when it was last live while offline. */
private fun roomFacts(room: CamModel): String {
    val facts = mutableListOf<String>()
    if (room.online) {
        facts.add("${formatViewers(room.viewers)} watching")
    } else {
        val lastSeen = room.lastSeenOnlineAt
        if (lastSeen != null) {
            facts.add("Last live ${formatRelativeTime(lastSeen)}")
        }
    }
    facts.add(room.pluginName)
    return facts.joinToString(" · ")
}

/** "Recording at 720p: 12:34 · 120 MB", its clock ticking every second between polls. */
@Composable
private fun RecordingNote(recording: CamRecording) {
    val seconds = rememberRecordingSeconds(recording)
    var text = "Recording"
    val quality = recording.quality
    if (quality != null) {
        text += " at $quality"
    }
    text += ": ${formatClock(seconds)} · ${formatBytes(recording.sizeBytes)}"
    Row(horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(RecordingDotSize).clip(VeilShapes.capsule).background(VeilColors.error))
        Text(text, style = MaterialTheme.typography.bodySmall, color = VeilColors.content)
    }
}

private val RecordingDotSize = 8.dp

/** Seconds a running recording has been going, counted from its start every second. */
@Composable
private fun rememberRecordingSeconds(recording: CamRecording): Double {
    var now by remember { mutableLongStateOf(System.currentTimeMillis()) }
    LaunchedEffect(recording.id) {
        while (true) {
            now = System.currentTimeMillis()
            delay(CLOCK_TICK_MILLISECONDS)
        }
    }
    val started = parseTimestamp(recording.startedAt) ?: return recording.durationSeconds
    val elapsed = Duration.between(started, Instant.ofEpochMilli(now)).toMillis() / MILLISECONDS_PER_SECOND
    return maxOf(recording.durationSeconds, elapsed)
}

/** Follow, record (or stop) and auto-record. */
@Composable
private fun RoomActions(room: CamModel, state: CamRoomState, viewModel: CamRoomViewModel, onOpenRecordMenu: () -> Unit) {
    LazyRow(contentPadding = PaddingValues(horizontal = ReactionRowInset), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.hairline)) {
        item {
            var label = "Follow"
            if (room.followed) {
                label = "Following"
            }
            ReactionButton(VeilIcons.Follow, label, active = room.followed) { viewModel.toggleFollow() }
        }
        item { RecordButton(room, state, viewModel, onOpenRecordMenu) }
        item {
            ReactionButton(VeilIcons.Repeat, "Auto-record", active = room.autoRecord) {
                if (!state.autoRecordBusy) {
                    viewModel.setAutoRecord(!room.autoRecord)
                }
            }
        }
    }
}

/** Stop while recording, else Record, which opens the quality and start menu; only while live. */
@Composable
private fun RecordButton(room: CamModel, state: CamRoomState, viewModel: CamRoomViewModel, onOpenRecordMenu: () -> Unit) {
    if (state.recording != null) {
        ReactionButton(VeilIcons.Stop, "Stop", active = true) {
            if (!state.recordBusy) {
                viewModel.stopRecording()
            }
        }
        return
    }
    ReactionButton(VeilIcons.Record, "Record") {
        if (room.online && !state.offline && !state.recordBusy) {
            onOpenRecordMenu()
        }
    }
}

/** The quality auto-record captures, while it is on. */
@Composable
private fun AutoRecordQuality(room: CamModel, busy: Boolean, viewModel: CamRoomViewModel) {
    VeilCard(Modifier.gutterPadding()) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium)) {
            Text("Record whenever ${room.displayName} is live", Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge)
            VeilSwitch(checked = true, onCheckedChange = { viewModel.setAutoRecord(false) })
        }
        LazyRow(Modifier.fadingEdges(), horizontalArrangement = Arrangement.spacedBy(VeilSpacing.small)) {
            items(AUTO_RECORD_QUALITIES) { quality ->
                Pill(
                    recordQualityLabel(quality),
                    onClick = {
                        if (!busy) {
                            viewModel.setAutoRecord(true, quality)
                        }
                    },
                    selected = room.recordQuality == quality,
                )
            }
        }
    }
}

/** A finished recording: its snapshot, when, how long and how big; opens its scene. */
@Composable
private fun PastRecordingRow(recording: CamRecording, onOpen: (String) -> Unit) {
    val sceneId = recording.sceneId ?: return
    Row(
        Modifier.fillMaxWidth().pressClickable { onOpen(sceneId) }.padding(horizontal = VeilSpacing.gutter, vertical = VeilSpacing.small),
        horizontalArrangement = Arrangement.spacedBy(VeilSpacing.medium),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RemoteImage(recording.thumbnailUrl, Modifier.size(width = 120.dp, height = 68.dp).clip(VeilShapes.card))
        Column(Modifier.weight(1f)) {
            Text("Live ${formatRelativeTime(recording.startedAt)}", style = MaterialTheme.typography.bodyMedium)
            val facts = listOfNotNull(formatClock(recording.durationSeconds), formatBytes(recording.sizeBytes), recording.quality)
            Text(facts.joinToString(" · "), style = MaterialTheme.typography.bodySmall, color = VeilColors.contentMuted)
        }
    }
}

/**
 * The record menu: start in the past when the room's buffer holds it (from the playhead when
 * rewound, one or three minutes back), or now at the source quality or one of the stream's
 * heights; the height being watched is ticked.
 */
@Composable
private fun RecordSheet(player: ExoPlayer, onRecord: (quality: String, fromSecondsAgo: Int?) -> Unit, onDismiss: () -> Unit) {
    val tracks = rememberPlayerTracks(player)
    val watching = recordQualityFor(selectedQualityHeight(player, tracks))
    val bufferedSeconds = maxOf(0L, player.duration) / MILLISECONDS_PER_SECOND
    val lagSeconds = liveLagMilliseconds(player.currentPosition, player.duration) / MILLISECONDS_PER_SECOND
    val starts = recordStarts(bufferedSeconds, lagSeconds)
    VeilBottomSheet(onDismissRequest = onDismiss) {
        if (starts.isNotEmpty()) {
            SectionHeading("Start in the past", Modifier.gutterPadding())
            for (start in starts) {
                SheetAction(start.label, VeilIcons.History, onClick = { onRecord(watching, start.secondsAgo) })
            }
        }
        SectionHeading("Record at", Modifier.gutterPadding())
        SheetAction("Source (best)", VeilIcons.Record, onClick = { onRecord(CamRepository.SOURCE_QUALITY, null) }, active = watching == CamRepository.SOURCE_QUALITY)
        for (quality in videoQualities(tracks)) {
            val height = quality.height.toString()
            SheetAction(videoQualityLabel(quality.height), VeilIcons.Record, onClick = { onRecord(height, null) }, active = watching == height)
        }
    }
}
