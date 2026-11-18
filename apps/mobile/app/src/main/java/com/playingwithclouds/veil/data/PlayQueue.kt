package com.playingwithclouds.veil.data

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.MoveInCollectionMutation
import com.playingwithclouds.veil.graphql.RemoveFromCollectionMutation
import com.playingwithclouds.veil.graphql.type.CollectionKind
import com.playingwithclouds.veil.graphql.type.RecordKind
import com.playingwithclouds.veil.graphql.type.RecordRefInput
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlin.random.Random

/** The queue's videos as last read or changed here, in the queue's order; [collectionId] is empty until read. */
data class QueueContents(val collectionId: String = "", val videos: List<VideoSummary> = emptyList())

/**
 * The player's queue: the library's built-in queue collection (what "Play next" fills), with
 * shuffle and repeat kept on this phone. Nothing leaves the queue by playing it; the order moves
 * on from the video playing. Reorders and removals change the collection right away, so the
 * Library's Queue and the web see them too.
 */
object PlayQueue {

    private const val PREFERENCES = "queue"
    private const val KEY_SHUFFLE = "shuffle"
    private const val KEY_REPEAT = "repeat"

    private lateinit var preferences: SharedPreferences

    private val mutableContents = MutableStateFlow(QueueContents())
    private val mutableShuffle = MutableStateFlow(false)
    private val mutableRepeat = MutableStateFlow(RepeatMode.OFF)
    private val mutableShuffledIds = MutableStateFlow<List<String>>(emptyList())

    /** The queue's videos. */
    val contents: StateFlow<QueueContents> = mutableContents

    /** Whether the queue plays shuffled. */
    val shuffle: StateFlow<Boolean> = mutableShuffle

    /** What happens at the end. */
    val repeat: StateFlow<RepeatMode> = mutableRepeat

    /** The shuffled order while shuffling; kept in memory, a new one each time shuffle is turned on. */
    val shuffledIds: StateFlow<List<String>> = mutableShuffledIds

    /** Reads shuffle and repeat; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableShuffle.value = preferences.getBoolean(KEY_SHUFFLE, false)
        mutableRepeat.value = RepeatMode.entries.firstOrNull { mode -> mode.name == preferences.getString(KEY_REPEAT, null) } ?: RepeatMode.OFF
    }

    /** Reads the queue again. */
    suspend fun refresh() {
        val page = LibraryRepository.builtInCollection(CollectionKind.queue)
        val videos = page.entries.filterIsInstance<CollectionEntry.Video>().map { entry -> entry.video }
        mutableContents.value = QueueContents(page.id, videos)
        if (mutableShuffle.value && mutableShuffledIds.value.isEmpty()) {
            mutableShuffledIds.value = QueueOrder.shuffled(videos.map { video -> video.id }, null, Random.Default)
        }
    }

    /** The videos in the order they play, shuffled or not. */
    fun playOrder(): List<VideoSummary> {
        val videos = mutableContents.value.videos
        var shuffled: List<String>? = null
        if (mutableShuffle.value) {
            shuffled = mutableShuffledIds.value
        }
        val byId = videos.associateBy { video -> video.id }
        return QueueOrder.playOrder(videos.map { video -> video.id }, shuffled).mapNotNull { id -> byId[id] }
    }

    /** What plays when [currentId] ends; [currentId] itself on repeat one, null at the end. */
    fun next(currentId: String): String? {
        return QueueOrder.next(playOrder().map { video -> video.id }, currentId, mutableRepeat.value)
    }

    /** What the next button skips to from [currentId]; null when nothing comes after it. */
    fun skip(currentId: String): String? {
        return QueueOrder.skip(playOrder().map { video -> video.id }, currentId, mutableRepeat.value)
    }

    /** Turns shuffle on (a new order, [currentId] first) or off. */
    fun setShuffle(on: Boolean, currentId: String?) {
        preferences.edit { putBoolean(KEY_SHUFFLE, on) }
        mutableShuffle.value = on
        if (on) {
            mutableShuffledIds.value = QueueOrder.shuffled(mutableContents.value.videos.map { video -> video.id }, currentId, Random.Default)
        }
    }

    /** Moves the repeat button on: off → all → one. */
    fun cycleRepeat() {
        val next = mutableRepeat.value.next()
        preferences.edit { putString(KEY_REPEAT, next.name) }
        mutableRepeat.value = next
    }

    /** Takes a video out of the queue. */
    suspend fun remove(videoId: String) {
        val contents = mutableContents.value
        mutableContents.update { current -> current.copy(videos = current.videos.filter { video -> video.id != videoId }) }
        val mutation = RemoveFromCollectionMutation(contents.collectionId, listOf(RecordRefInput(RecordKind.video, videoId)))
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }

    /** Moves a video to [position] of the queue's own order (0 = first). */
    suspend fun move(videoId: String, position: Int) {
        val contents = mutableContents.value
        val from = contents.videos.indexOfFirst { video -> video.id == videoId }
        if (from < 0 || from == position) {
            return
        }
        val ids = QueueOrder.moved(contents.videos.map { video -> video.id }, from, position)
        val byId = contents.videos.associateBy { video -> video.id }
        mutableContents.value = contents.copy(videos = ids.mapNotNull { id -> byId[id] })
        val mutation = MoveInCollectionMutation(contents.collectionId, RecordRefInput(RecordKind.video, videoId), position)
        VeilApi.client.mutation(mutation).execute().dataOrThrow()
    }
}
