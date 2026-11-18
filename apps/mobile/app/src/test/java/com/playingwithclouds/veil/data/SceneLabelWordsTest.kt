package com.playingwithclouds.veil.data

import org.junit.Assert.assertEquals
import org.junit.Test

class SceneLabelWordsTest {

    private val labels = listOf(
        TagSummary("1", "Blowjob"),
        TagSummary("2", "POV"),
        TagSummary("3", "Doggystyle"),
        TagSummary("4", "Reverse Cowgirl"),
        TagSummary("5", "Cowgirl"),
    )

    /** The names of the labels found. */
    private fun names(match: LabelWords): List<String> = match.labels.map { label -> label.name }

    @Test
    fun labelOnlySearchesLeaveNoText() {
        val match = SceneLabelWords.match("pov blowjob", labels)
        assertEquals(listOf("POV", "Blowjob"), names(match))
        assertEquals("", match.text)
    }

    @Test
    fun otherWordsStayAsText() {
        val match = SceneLabelWords.match("Mia Doggy-Style hotel", labels)
        assertEquals(listOf("Doggystyle"), names(match))
        assertEquals("Mia hotel", match.text)
    }

    @Test
    fun longerNamesWinAndPluralsAndShortFormsCount() {
        assertEquals(listOf("Reverse Cowgirl"), names(SceneLabelWords.match("reverse cowgirl", labels)))
        assertEquals(listOf("Blowjob", "Doggystyle"), names(SceneLabelWords.match("blowjobs doggy bj", labels)))
        assertEquals("reverse", SceneLabelWords.match("reverse", labels).text)
    }

    @Test
    fun withoutLabelsEverythingIsText() {
        val match = SceneLabelWords.match("pov blowjob", emptyList())
        assertEquals(emptyList<TagSummary>(), match.labels)
        assertEquals("pov blowjob", match.text)
    }
}
