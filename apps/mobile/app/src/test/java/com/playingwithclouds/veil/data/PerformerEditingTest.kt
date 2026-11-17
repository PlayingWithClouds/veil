package com.playingwithclouds.veil.data

import java.time.LocalDate
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class PerformerEditingTest {

    private val today = LocalDate.of(2026, 11, 18)

    private val stored = EditablePerformer(
        id = "performer-1",
        name = "Jane Doe",
        aliases = listOf("Jane", "J. Doe"),
        details = "Bio",
        links = listOf("https://example.com/jane"),
        imageUri = "blob://abc",
        imageUrl = "http://localhost/blob/abc",
        attributes = PerformerAttributeValues(
            gender = "female",
            birthdate = "1995-04-12",
            country = "USA",
            hairColor = "blonde",
            heightCm = 165,
            measurements = "34C-24-34",
            careerStart = 2015,
        ),
    )

    private fun formWith(change: (PerformerForm) -> PerformerForm): PerformerForm {
        return change(PerformerEditing.formOf(stored))
    }

    private fun PerformerForm.withText(attribute: PerformerAttribute, text: String): PerformerForm {
        return copy(attributes = attributes + (attribute to text))
    }

    @Test
    fun openingAndSavingChangesNothing() {
        val changes = PerformerEditing.changes(stored, PerformerEditing.formOf(stored))
        assertTrue(changes.isEmpty())
    }

    @Test
    fun theFormShowsTextAttributesCapitalisedAndMeasurementsSplit() {
        val form = PerformerEditing.formOf(stored)
        assertEquals("Female", form.attributes[PerformerAttribute.Gender])
        assertEquals("Blonde", form.attributes[PerformerAttribute.HairColor])
        assertEquals("165", form.attributes[PerformerAttribute.HeightCm])
        assertEquals(MeasurementParts("34C", "24", "34"), form.measurementParts)
        assertEquals("Jane, J. Doe", form.aliases)
    }

    @Test
    fun aTextAttributeChangedOnlyInCaseKeepsTheStoredSpelling() {
        val form = formWith { it.withText(PerformerAttribute.HairColor, "BLONDE") }
        assertEquals("blonde", PerformerEditing.formAttributes(stored.attributes, form).hairColor)
        assertTrue(PerformerEditing.changes(stored, form).isEmpty())
    }

    @Test
    fun aChangedTextAttributeIsSavedCapitalisedWithEveryAttribute() {
        val form = formWith { it.withText(PerformerAttribute.HairColor, " red ") }
        val attributes = PerformerEditing.changes(stored, form).attributes
        assertEquals(stored.attributes.copy(hairColor = "Red"), attributes)
    }

    @Test
    fun anUntouchedCountryStaysAsTheSiteSpelledItButATypedOneIsNormalised() {
        val untouched = PerformerEditing.formAttributes(stored.attributes, PerformerEditing.formOf(stored))
        assertEquals("USA", untouched.country)
        val typed = formWith { it.withText(PerformerAttribute.Country, "uk") }
        assertEquals("United Kingdom", PerformerEditing.formAttributes(stored.attributes, typed).country)
    }

    @Test
    fun clearedAttributesBecomeNull() {
        val form = formWith { it.withText(PerformerAttribute.HeightCm, "").withText(PerformerAttribute.Birthdate, "  ") }
        val attributes = PerformerEditing.changes(stored, form).attributes
        assertNull(attributes?.heightCm)
        assertNull(attributes?.birthdate)
        assertEquals("USA", attributes?.country)
    }

    @Test
    fun measurementsAreJoinedFromTheirPartsAndKeepTheStoredSpellingWhenUnchanged() {
        val unchanged = formWith { it.copy(measurementParts = MeasurementParts(" 34C", "24", "34 ")) }
        assertTrue(PerformerEditing.changes(stored, unchanged).isEmpty())
        val changed = formWith { it.copy(measurementParts = MeasurementParts("36D", "26", "36")) }
        assertEquals("36D-26-36", PerformerEditing.changes(stored, changed).attributes?.measurements)
        val cleared = formWith { it.copy(measurementParts = MeasurementParts("", "", "")) }
        assertNull(PerformerEditing.changes(stored, cleared).attributes?.measurements)
    }

    @Test
    fun measurementsThatDoNotSplitAreEditedAsText() {
        val odd = stored.copy(attributes = stored.attributes.copy(measurements = "34C 24 34"))
        val form = PerformerEditing.formOf(odd)
        assertNull(form.measurementParts)
        assertTrue(PerformerEditing.changes(odd, form).isEmpty())
        val edited = form.withText(PerformerAttribute.Measurements, "36D 26 36")
        assertEquals("36D 26 36", PerformerEditing.changes(odd, edited).attributes?.measurements)
    }

    @Test
    fun onlyTheChangedPlainFieldsAreSaved() {
        val form = formWith { it.copy(name = " Jane Roe ", aliases = "Jane, , J. Doe, Janie", links = "", details = "Bio  ") }
        val changes = PerformerEditing.changes(stored, form)
        assertEquals("Jane Roe", changes.name)
        assertEquals(listOf("Jane", "J. Doe", "Janie"), changes.aliases)
        assertEquals(emptyList<String>(), changes.links)
        assertNull(changes.details)
        assertNull(changes.imageUrl)
        assertNull(changes.attributes)
    }

    @Test
    fun anEmptyPictureUrlRemovesThePicture() {
        val changes = PerformerEditing.changes(stored, formWith { it.copy(imageUrl = " ") })
        assertEquals("", changes.imageUrl)
    }

    @Test
    fun theAgeHintFollowsTheBirthAndDeathDates() {
        val form = PerformerEditing.formOf(stored)
        assertEquals("31 years old", PerformerEditing.hint(PerformerAttribute.Birthdate, form, today))
        val dead = form.withText(PerformerAttribute.DeathDate, "2020-04-11")
        assertEquals("died aged 24", PerformerEditing.hint(PerformerAttribute.Birthdate, dead, today))
        val unreadable = form.withText(PerformerAttribute.Birthdate, "soon")
        assertNull(PerformerEditing.hint(PerformerAttribute.Birthdate, unreadable, today))
        assertNull(PerformerEditing.hint(PerformerAttribute.Gender, form, today))
    }

    @Test
    fun theYearsActiveHintRunsToTheEndYearOrNow() {
        val form = PerformerEditing.formOf(stored)
        assertEquals("11 years active", PerformerEditing.hint(PerformerAttribute.CareerStart, form, today))
        val ended = form.withText(PerformerAttribute.CareerEnd, "2020")
        assertEquals("5 years active", PerformerEditing.hint(PerformerAttribute.CareerStart, ended, today))
        val unset = form.withText(PerformerAttribute.CareerStart, "")
        assertNull(PerformerEditing.hint(PerformerAttribute.CareerStart, unset, today))
    }

    @Test
    fun suggestionsPutPresetsFirstAndDropDuplicatesIgnoringCase() {
        val merged = PerformerPresets.mergeSuggestions(listOf("Blonde", "Red"), listOf("red", "auburn", " ", "BLONDE", "teal"))
        assertEquals(listOf("Blonde", "Red", "Auburn", "Teal"), merged)
    }

    @Test
    fun matchingSuggestionsPutPrefixesFirstAndHideAnExactMatch() {
        val options = listOf("Brown", "Auburn", "Brunette", "Black")
        assertEquals(listOf("Brown", "Brunette"), PerformerPresets.matchingSuggestions(options, "br", 10))
        assertEquals(listOf("Auburn", "Brunette"), PerformerPresets.matchingSuggestions(options, "u", 10))
        assertEquals(emptyList<String>(), PerformerPresets.matchingSuggestions(options, "black", 10))
        assertEquals(listOf("Brown", "Auburn"), PerformerPresets.matchingSuggestions(options, "", 2))
    }

    @Test
    fun countriesAreNamedInEnglishAndNormalisedFromShortSpellings() {
        val names = PerformerPresets.countryNames()
        assertTrue(names.contains("Japan"))
        assertEquals("United States", PerformerPresets.normalizeCountry(" USA "))
        assertEquals("Japan", PerformerPresets.normalizeCountry("Japan"))
    }

    @Test
    fun measurementsSplitOnlyIntoThreeParts() {
        assertEquals(MeasurementParts("", "", ""), PerformerPresets.splitMeasurements(null))
        assertEquals(MeasurementParts("34C", "24", "34"), PerformerPresets.splitMeasurements("34C - 24 - 34"))
        assertNull(PerformerPresets.splitMeasurements("34-24"))
        assertEquals("", PerformerPresets.joinMeasurements(MeasurementParts(" ", "", "")))
    }
}
