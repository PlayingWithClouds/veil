package com.playingwithclouds.veil.ui

import com.playingwithclouds.veil.data.SearchSuggestion
import com.playingwithclouds.veil.ui.search.SearchState
import com.playingwithclouds.veil.ui.search.withSiteCompletions
import org.junit.Assert.assertEquals
import org.junit.Test

class SiteCompletionsTest {

    private fun suggestion(text: String, kind: String = "QUERY"): SearchSuggestion {
        return SearchSuggestion(kind, text, null, null, null)
    }

    @Test
    fun completionsFollowTheLibrarySuggestionsWithoutRepeats() {
        val state = SearchState(query = "ann", suggestions = listOf(suggestion("Anna", "PERFORMER")))
        val merged = state.withSiteCompletions("ann", listOf(suggestion("anna"), suggestion("annie")))
        assertEquals(listOf("Anna", "annie"), merged.suggestions.map { item -> item.text })
    }

    @Test
    fun lateCompletionsForOtherTextAreIgnored() {
        val state = SearchState(query = "anni", suggestions = emptyList())
        assertEquals(state, state.withSiteCompletions("ann", listOf(suggestion("annie"))))
        val submitted = SearchState(query = "ann", submitted = "ann")
        assertEquals(submitted, submitted.withSiteCompletions("ann", listOf(suggestion("annie"))))
    }
}
