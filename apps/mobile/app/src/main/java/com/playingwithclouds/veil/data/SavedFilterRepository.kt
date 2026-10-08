package com.playingwithclouds.veil.data

import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.CreateSavedFilterMutation
import com.playingwithclouds.veil.graphql.DeleteSavedFilterMutation
import com.playingwithclouds.veil.graphql.SavedFiltersQuery
import com.playingwithclouds.veil.graphql.fragment.SavedFilterFields
import com.playingwithclouds.veil.graphql.type.SceneFilterInput

/** A saved combination of tag, sites and runtime, shown as a chip on Home. */
data class FilterPreset(
    val id: String,
    val name: String,
    val tagId: String?,
    val sources: List<String>,
    val runtime: RuntimeWindow,
)

/** The filter presets stored on the backend. */
object SavedFilterRepository {

    /** All presets, newest first. */
    suspend fun list(): List<FilterPreset> {
        return VeilApi.client.query(SavedFiltersQuery()).execute().dataOrThrow().savedFilters.map { saved -> saved.savedFilterFields.toPreset() }
    }

    /** Stores a preset and returns it with its id. */
    suspend fun create(name: String, tagId: String?, sources: List<String>, runtime: RuntimeWindow): FilterPreset {
        val filter = SceneFilterInput(
            tagId = tagId.orAbsent(),
            minDuration = runtime.minSeconds.orAbsent(),
            maxDuration = runtime.maxSeconds.orAbsent(),
            sources = sources.orAbsentIfEmpty(),
        )
        val data = VeilApi.client.mutation(CreateSavedFilterMutation(name, filter)).execute().dataOrThrow()
        return data.createSavedFilter.savedFilterFields.toPreset()
    }

    /** Removes a preset. */
    suspend fun delete(presetId: String) {
        VeilApi.client.mutation(DeleteSavedFilterMutation(presetId)).execute().dataOrThrow()
    }
}

/** Maps the preset fragment to the model. */
fun SavedFilterFields.toPreset(): FilterPreset {
    return FilterPreset(
        id = id,
        name = name,
        tagId = filter.tagId,
        sources = filter.sources.orEmpty(),
        runtime = RuntimeWindow(filter.minDuration, filter.maxDuration),
    )
}
