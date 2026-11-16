package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.CategoryTilesQuery
import com.playingwithclouds.veil.graphql.FilteredVideosQuery
import com.playingwithclouds.veil.graphql.type.FilterInput

/** A category as a tile shows it: how many videos it holds and a picture. */
data class CategoryTile(val id: String, val name: String, val count: Int, val pictureUrl: String?)

/** A top-level category and the categories inside it, in their order. */
data class CategoryGroup(val parent: CategoryTile, val children: List<CategoryTile>)

/** What a list of videos is narrowed to: a category or a tag. */
sealed interface VideoScope {
    data class Category(val id: String) : VideoScope
    data class Tag(val id: String) : VideoScope
}

/** Our categories and the videos in a category or with a tag (`apps/web/src/lib/categories.ts`). */
object CategoryRepository {

    /** Every category grouped under its top-level one, in their order. */
    suspend fun groups(): List<CategoryGroup> {
        val tiles = VeilApi.client.query(CategoryTilesQuery()).execute().dataOrThrow().categoryTiles
        val parents = tiles.filter { tile -> tile.category.parent == null }
        return parents.map { parent ->
            val children = tiles.filter { tile -> tile.category.parent?.id == parent.category.id }
            CategoryGroup(parent.toTile(), children.map { child -> child.toTile() })
        }
    }

    /** Newest videos in a category or with a tag. */
    suspend fun videos(scope: VideoScope, limit: Int, offset: Int): List<VideoSummary> {
        val query = FilteredVideosQuery(Optional.present(filterOf(scope)), Optional.present(limit), Optional.present(offset))
        return VeilApi.client.query(query).execute().dataOrThrow().videos.map { video -> video.videoCardFields.toSummary() }
    }

    /** The filter of a category's or tag's videos. */
    private fun filterOf(scope: VideoScope): FilterInput {
        return when (scope) {
            is VideoScope.Category -> FilterInput(categoryIds = Optional.present(listOf(scope.id)))
            is VideoScope.Tag -> FilterInput(includeTagIds = Optional.present(listOf(scope.id)))
        }
    }

    /** A tile from the query's row. */
    private fun CategoryTilesQuery.CategoryTile.toTile(): CategoryTile {
        return CategoryTile(category.id, category.name, count, preview?.url)
    }
}
