package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.AudioHistoryQuery
import com.playingwithclouds.veil.graphql.AudioListingQuery
import com.playingwithclouds.veil.graphql.OpenAudioMutation
import com.playingwithclouds.veil.graphql.OpenStoryMutation
import com.playingwithclouds.veil.graphql.ReadChapterMutation
import com.playingwithclouds.veil.graphql.StoryChaptersQuery
import com.playingwithclouds.veil.graphql.StoryListingQuery
import com.playingwithclouds.veil.graphql.fragment.AudioCardFields
import com.playingwithclouds.veil.graphql.fragment.ChapterFields
import com.playingwithclouds.veil.graphql.fragment.StoryCardFields

/** A story's page: its card, description, chapters and where reading stopped ([progress], see [StoryReading]). */
data class StoryDetail(
    val summary: StorySummary,
    val details: String?,
    val wordCount: Int?,
    val tags: List<TagSummary>,
    val chapters: List<Chapter>,
    val progress: Double?,
)

/** A chapter with its text. */
data class ChapterText(val chapter: Chapter, val text: String)

/** An audio's page: its card, description, tags, the streams to play and the resume point in seconds. */
data class AudioDetail(
    val summary: AudioSummary,
    val details: String?,
    val tags: List<TagSummary>,
    val streamUrls: List<String>,
    val progressSeconds: Double?,
)

/** Stories and audio (`apps/web/src/lib/media.ts`). */
object MediaRepository {

    /** Stories, most recently added first. */
    suspend fun stories(limit: Int, offset: Int): List<StorySummary> {
        val data = VeilApi.client.query(StoryListingQuery(Optional.present(limit), Optional.present(offset))).execute().dataOrThrow()
        return data.stories.map { story -> story.storyCardFields.toStory() }
    }

    /** Audio, most recently added first. */
    suspend fun audios(limit: Int, offset: Int): List<AudioSummary> {
        val data = VeilApi.client.query(AudioListingQuery(Optional.present(limit), Optional.present(offset))).execute().dataOrThrow()
        return data.audios.map { audio -> audio.audioCardFields.toAudio() }
    }

    /** Audio stopped partway, most recently listened first. History holds finished ones too, so twice as many are asked for. */
    suspend fun continueListening(limit: Int): List<AudioSummary> {
        val data = VeilApi.client.query(AudioHistoryQuery(Optional.present(limit * 2))).execute().dataOrThrow()
        val started = data.history.filter { state -> state.progress != null }
        return started.mapNotNull { state -> state.record?.onAudio?.audioCardFields?.toAudio() }.take(limit)
    }

    /** Opens a story, reading its pages not read yet (chapters, credits); null when it's gone. */
    suspend fun openStory(id: String): StoryDetail? {
        val story = VeilApi.client.mutation(OpenStoryMutation(id)).execute().dataOrThrow().fetchMedia?.onStory ?: return null
        return StoryDetail(
            summary = story.storyCardFields.toStory(),
            details = story.details,
            wordCount = story.wordCount,
            tags = story.tags.map { tag -> TagSummary(tag.id, tag.name) },
            chapters = story.chapters.map { chapter -> chapter.chapterFields.toChapter() },
            progress = story.userState.progress,
        )
    }

    /** A story's chapters again: reading a chapter of a branching story can add the ones its choices lead to. */
    suspend fun chapters(storyId: String): List<Chapter> {
        val data = VeilApi.client.query(StoryChaptersQuery(storyId)).execute().dataOrThrow()
        return data.story?.chapters.orEmpty().map { chapter -> chapter.chapterFields.toChapter() }
    }

    /** A chapter with its text, read from the site the first time; null when there's none at that position. */
    suspend fun readChapter(storyId: String, position: Int): ChapterText? {
        val chapter = VeilApi.client.mutation(ReadChapterMutation(storyId, position)).execute().dataOrThrow().readChapter ?: return null
        return ChapterText(chapter.chapterFields.toChapter(), chapter.text)
    }

    /** Opens an audio, reading its pages not read yet (credits, streams); null when it's gone. */
    suspend fun openAudio(id: String): AudioDetail? {
        val audio = VeilApi.client.mutation(OpenAudioMutation(id)).execute().dataOrThrow().fetchMedia?.onAudio ?: return null
        return AudioDetail(
            summary = audio.audioCardFields.toAudio(),
            details = audio.details,
            tags = audio.tags.map { tag -> TagSummary(tag.id, tag.name) },
            streamUrls = audio.streams.map { stream -> stream.url },
            progressSeconds = audio.userState.progress,
        )
    }

    /** A chapter from its fragment. */
    private fun ChapterFields.toChapter(): Chapter {
        return Chapter(id, position, title, wordCount, parent, choice)
    }
}

/** A story card from its fragment; the first performer is the author. */
internal fun StoryCardFields.toStory(): StorySummary {
    return StorySummary(id, title, cover?.url, performers.firstOrNull()?.name, chapterCount)
}

/** An audio card from its fragment; the first performer is the voice. */
internal fun AudioCardFields.toAudio(): AudioSummary {
    return AudioSummary(id, title, cover?.url, performers.firstOrNull()?.name, durationSeconds)
}
