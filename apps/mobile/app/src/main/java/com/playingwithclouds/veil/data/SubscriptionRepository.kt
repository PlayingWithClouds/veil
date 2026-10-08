package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.api.orAbsent
import com.playingwithclouds.veil.api.orAbsentIfEmpty
import com.playingwithclouds.veil.graphql.MarkSearchSubscriptionSeenMutation
import com.playingwithclouds.veil.graphql.RunSearchSubscriptionMutation
import com.playingwithclouds.veil.graphql.SearchSubscriptionsQuery
import com.playingwithclouds.veil.graphql.SubscribeSearchMutation
import com.playingwithclouds.veil.graphql.SubscriptionFeedQuery
import com.playingwithclouds.veil.graphql.UnsubscribeSearchMutation
import com.playingwithclouds.veil.graphql.UpdateSearchSubscriptionMutation
import com.playingwithclouds.veil.graphql.type.SubscriptionFeedFilter

/** Narrows the combined subscription feed; unset fields match everything. */
data class FeedFilter(val subscriptionId: String? = null, val newOnly: Boolean = false, val unwatchedOnly: Boolean = false)

/** Saved searches and followed studios, performers and tags. */
object SubscriptionRepository {

    /** Every subscription. */
    suspend fun subscriptions(): List<SubscriptionSummary> {
        val data = VeilApi.client.query(SearchSubscriptionsQuery()).execute().dataOrThrow()
        return data.searchSubscriptions.map { subscription -> subscription.subscriptionFields.toSummary() }
    }

    /** A page of the scenes the subscriptions found, newest find first. */
    suspend fun feed(filter: FeedFilter, limit: Int, offset: Int): List<SubscriptionFeedEntry> {
        val backendFilter = SubscriptionFeedFilter(
            subscriptionId = filter.subscriptionId.orAbsent(),
            kinds = Optional.Absent,
            newOnly = Optional.present(filter.newOnly),
            unwatchedOnly = Optional.present(filter.unwatchedOnly),
        )
        val query = SubscriptionFeedQuery(Optional.present(backendFilter), Optional.present(limit), Optional.present(offset))
        val data = VeilApi.client.query(query).execute().dataOrThrow()
        return data.subscriptionFeed.map { item ->
            SubscriptionFeedEntry(item.scene.sceneCardFields.toSummary(), item.foundAt, item.subscriptionId, item.isNew)
        }
    }

    /** Subscribes to a search query; an empty source list searches every plugin. */
    suspend fun subscribeSearch(query: String, sources: List<String>): SubscriptionSummary {
        val data = VeilApi.client.mutation(SubscribeSearchMutation(query, sources.orAbsentIfEmpty())).execute().dataOrThrow()
        return data.subscribeSearch.subscriptionFields.toSummary()
    }

    /** Changes a subscription's re-run interval and/or enabled flag. */
    suspend fun update(id: String, intervalHours: Int?, enabled: Boolean?): SubscriptionSummary {
        val mutation = UpdateSearchSubscriptionMutation(id, intervalHours.orAbsent(), enabled.orAbsent())
        val data = VeilApi.client.mutation(mutation).execute().dataOrThrow()
        return data.updateSearchSubscription.subscriptionFields.toSummary()
    }

    /** Deletes a subscription and its feed. */
    suspend fun unsubscribe(id: String) {
        VeilApi.client.mutation(UnsubscribeSearchMutation(id)).execute().dataOrThrow()
    }

    /** Runs a subscription now; returns when the run finishes, which can take a minute. */
    suspend fun runNow(id: String): SubscriptionSummary {
        val data = VeilApi.client.mutation(RunSearchSubscriptionMutation(id)).execute().dataOrThrow()
        return data.runSearchSubscription.subscriptionFields.toSummary()
    }

    /** Clears a subscription's new-scene count. */
    suspend fun markSeen(id: String): SubscriptionSummary {
        val data = VeilApi.client.mutation(MarkSearchSubscriptionSeenMutation(id)).execute().dataOrThrow()
        return data.markSearchSubscriptionSeen.subscriptionFields.toSummary()
    }
}
