package com.playingwithclouds.veil.data

import com.apollographql.apollo.api.Optional
import com.playingwithclouds.veil.api.VeilApi
import com.playingwithclouds.veil.api.dataOrThrow
import com.playingwithclouds.veil.graphql.EditPerformerMutation
import com.playingwithclouds.veil.graphql.PerformerAttributeValuesQuery
import com.playingwithclouds.veil.graphql.PerformerForEditQuery
import com.playingwithclouds.veil.graphql.type.PerformerAttributesInput
import com.playingwithclouds.veil.graphql.type.PerformerEdit
import java.text.Collator
import java.time.LocalDate
import java.util.Locale

/** How an attribute is typed into the editor. */
enum class AttributeFieldType { Suggest, Date, Number, Measurements }

/** One attribute input of the performer editor, in form order (the web's `ATTRIBUTE_FIELDS`). */
enum class PerformerAttribute(val key: String, val label: String, val type: AttributeFieldType, val presets: List<String> = emptyList()) {
    Gender("gender", "Gender", AttributeFieldType.Suggest, PerformerPresets.GENDERS),
    Birthdate("birthdate", "Born", AttributeFieldType.Date),
    DeathDate("deathDate", "Died", AttributeFieldType.Date),
    Country("country", "Country", AttributeFieldType.Suggest, PerformerPresets.countryNames()),
    Ethnicity("ethnicity", "Ethnicity", AttributeFieldType.Suggest, PerformerPresets.ETHNICITIES),
    HairColor("hairColor", "Hair", AttributeFieldType.Suggest, PerformerPresets.HAIR_COLORS),
    EyeColor("eyeColor", "Eyes", AttributeFieldType.Suggest, PerformerPresets.EYE_COLORS),
    HeightCm("heightCm", "Height (cm)", AttributeFieldType.Number),
    WeightKg("weightKg", "Weight (kg)", AttributeFieldType.Number),
    Measurements("measurements", "Measurements", AttributeFieldType.Measurements),
    FakeTits("fakeTits", "Enhanced", AttributeFieldType.Suggest, PerformerPresets.ENHANCEMENTS),
    Tattoos("tattoos", "Tattoos", AttributeFieldType.Suggest, PerformerPresets.BODY_ART),
    Piercings("piercings", "Piercings", AttributeFieldType.Suggest, PerformerPresets.BODY_ART),
    CareerStart("careerStart", "Active since", AttributeFieldType.Number),
    CareerEnd("careerEnd", "Active until", AttributeFieldType.Number),
}

/** A performer's body and career facts as stored or saved; null where nothing is known. */
data class PerformerAttributeValues(
    val gender: String? = null,
    val birthdate: String? = null,
    val deathDate: String? = null,
    val country: String? = null,
    val ethnicity: String? = null,
    val eyeColor: String? = null,
    val hairColor: String? = null,
    val heightCm: Int? = null,
    val weightKg: Int? = null,
    val measurements: String? = null,
    val fakeTits: String? = null,
    val tattoos: String? = null,
    val piercings: String? = null,
    val careerStart: Int? = null,
    val careerEnd: Int? = null,
) {
    /** One attribute as text; empty where nothing is known. */
    fun textOf(attribute: PerformerAttribute): String {
        return when (attribute) {
            PerformerAttribute.Gender -> gender.orEmpty()
            PerformerAttribute.Birthdate -> birthdate.orEmpty()
            PerformerAttribute.DeathDate -> deathDate.orEmpty()
            PerformerAttribute.Country -> country.orEmpty()
            PerformerAttribute.Ethnicity -> ethnicity.orEmpty()
            PerformerAttribute.HairColor -> hairColor.orEmpty()
            PerformerAttribute.EyeColor -> eyeColor.orEmpty()
            PerformerAttribute.HeightCm -> heightCm?.toString().orEmpty()
            PerformerAttribute.WeightKg -> weightKg?.toString().orEmpty()
            PerformerAttribute.Measurements -> measurements.orEmpty()
            PerformerAttribute.FakeTits -> fakeTits.orEmpty()
            PerformerAttribute.Tattoos -> tattoos.orEmpty()
            PerformerAttribute.Piercings -> piercings.orEmpty()
            PerformerAttribute.CareerStart -> careerStart?.toString().orEmpty()
            PerformerAttribute.CareerEnd -> careerEnd?.toString().orEmpty()
        }
    }

    /** These values with empty text as null, so a cleared attribute is cleared. */
    fun emptyAsNull(): PerformerAttributeValues {
        return fromTexts(PerformerAttribute.entries.associateWith { attribute -> textOf(attribute) })
    }

    companion object {
        /** Attributes from their texts: empty text is null, and a number that isn't one is null too. */
        fun fromTexts(texts: Map<PerformerAttribute, String>): PerformerAttributeValues {
            fun text(attribute: PerformerAttribute): String? = texts[attribute].orEmpty().ifEmpty { null }
            fun number(attribute: PerformerAttribute): Int? = texts[attribute].orEmpty().trim().toIntOrNull()
            return PerformerAttributeValues(
                gender = text(PerformerAttribute.Gender),
                birthdate = text(PerformerAttribute.Birthdate),
                deathDate = text(PerformerAttribute.DeathDate),
                country = text(PerformerAttribute.Country),
                ethnicity = text(PerformerAttribute.Ethnicity),
                eyeColor = text(PerformerAttribute.EyeColor),
                hairColor = text(PerformerAttribute.HairColor),
                heightCm = number(PerformerAttribute.HeightCm),
                weightKg = number(PerformerAttribute.WeightKg),
                measurements = text(PerformerAttribute.Measurements),
                fakeTits = text(PerformerAttribute.FakeTits),
                tattoos = text(PerformerAttribute.Tattoos),
                piercings = text(PerformerAttribute.Piercings),
                careerStart = number(PerformerAttribute.CareerStart),
                careerEnd = number(PerformerAttribute.CareerEnd),
            )
        }
    }
}

/** A performer as stored, for editing. [imageUri] is the stored picture URI, [imageUrl] where to load it from. */
data class EditablePerformer(
    val id: String,
    val name: String,
    val aliases: List<String>,
    val details: String?,
    val links: List<String>,
    val imageUri: String?,
    val imageUrl: String?,
    val attributes: PerformerAttributeValues,
)

/** Bust (with cup), waist and hips, as the editor's three inputs hold them. */
data class MeasurementParts(val bust: String, val waist: String, val hips: String)

/**
 * What the editor's inputs hold, all as typed. [attributes] holds every attribute's text;
 * [measurementParts] is null when the stored measurements aren't "bust-waist-hips", which are then
 * edited as the text in [attributes].
 */
data class PerformerForm(
    val name: String,
    val aliases: String,
    val imageUrl: String,
    val details: String,
    val links: String,
    val attributes: Map<PerformerAttribute, String>,
    val measurementParts: MeasurementParts?,
)

/** The fields of a performer the user changed; null ones are left as stored. */
data class PerformerChanges(
    val name: String? = null,
    val aliases: List<String>? = null,
    val imageUrl: String? = null,
    val details: String? = null,
    val links: List<String>? = null,
    val attributes: PerformerAttributeValues? = null,
) {
    /** Whether nothing changed. */
    fun isEmpty(): Boolean {
        return name == null && aliases == null && imageUrl == null && details == null && links == null && attributes == null
    }
}

/**
 * The common values the performer editor suggests, before what the library already holds.
 * Suggestions only: any value can be typed. Port of `apps/web/src/lib/performerPresets.ts`.
 */
object PerformerPresets {
    val GENDERS = listOf("Female", "Male", "Trans woman", "Trans man", "Non-binary")
    val ETHNICITIES = listOf("Caucasian", "Latina", "Asian", "Black", "Middle Eastern", "Indian", "Mixed")
    val HAIR_COLORS = listOf("Blonde", "Brunette", "Black", "Red", "Auburn", "Brown", "Gray", "Dyed")
    val EYE_COLORS = listOf("Blue", "Green", "Brown", "Hazel", "Gray", "Amber")
    val ENHANCEMENTS = listOf("Natural", "Enhanced")
    val BODY_ART = listOf("None", "Some")

    // ISO 3166 region codes, as the web lists them.
    private val REGION_CODES = (
        "AD AE AF AG AL AM AO AR AT AU AZ BA BB BD BE BF BG BH BI BJ BN BO BR BS BT BW BY BZ CA CD CF CG CH CI CL CM CN CO CR CU CV CY " +
            "CZ DE DJ DK DM DO DZ EC EE EG ER ES ET FI FJ FR GA GB GD GE GH GM GN GQ GR GT GW GY HK HN HR HT HU ID IE IL IN IQ IR IS IT " +
            "JM JO JP KE KG KH KM KN KP KR KW KZ LA LB LC LI LK LR LS LT LU LV LY MA MC MD ME MG MK ML MM MN MO MR MT MU MV MW MX MY MZ " +
            "NA NE NG NI NL NO NP NZ OM PA PE PG PH PK PL PR PS PT PY QA RO RS RU RW SA SB SC SD SE SG SI SK SL SM SN SO SR SS ST SV SY " +
            "SZ TD TG TH TJ TL TM TN TO TR TT TW TZ UA UG US UY UZ VA VC VE VN VU WS YE ZA ZM ZW"
        ).split(" ")

    // Short spellings sites use for a country, by the name the list gives it.
    private val COUNTRY_SPELLINGS = mapOf(
        "usa" to "United States",
        "us" to "United States",
        "u.s.a." to "United States",
        "united states of america" to "United States",
        "uk" to "United Kingdom",
        "great britain" to "United Kingdom",
        "england" to "United Kingdom",
    )

    /** Every country by its English name, alphabetically. */
    fun countryNames(): List<String> {
        val collator = Collator.getInstance(Locale.ENGLISH)
        return REGION_CODES
            .map { code -> Locale.Builder().setRegion(code).build().getDisplayCountry(Locale.ENGLISH).ifEmpty { code } }
            .sortedWith(collator)
    }

    /** A typed country in the list's spelling: "USA" becomes "United States". */
    fun normalizeCountry(country: String): String {
        val trimmed = country.trim()
        return COUNTRY_SPELLINGS[trimmed.lowercase()] ?: trimmed
    }

    /** A value with its first letter capitalised: "female" as "Female". */
    fun capitalize(value: String): String {
        return value.replaceFirstChar { first -> first.uppercaseChar() }
    }

    /**
     * The presets first, then the library's values they don't already cover, ignoring case; all
     * capitalised, since sites spell them either way.
     */
    fun mergeSuggestions(presets: List<String>, libraryValues: List<String>): List<String> {
        val suggestions = mutableListOf<String>()
        val known = mutableSetOf<String>()
        for (value in presets + libraryValues) {
            val key = value.trim().lowercase()
            if (key.isEmpty() || !known.add(key)) {
                continue
            }
            suggestions.add(capitalize(value.trim()))
        }
        return suggestions
    }

    /**
     * The suggestions that fit what is typed, at most [limit]: those starting with it first, then
     * those containing it; none for a value that is already exactly one of them.
     */
    fun matchingSuggestions(suggestions: List<String>, typed: String, limit: Int): List<String> {
        val needle = typed.trim().lowercase()
        if (needle.isEmpty()) {
            return suggestions.take(limit)
        }
        if (suggestions.any { suggestion -> suggestion.lowercase() == needle }) {
            return emptyList()
        }
        val (starting, others) = suggestions.partition { suggestion -> suggestion.lowercase().startsWith(needle) }
        val containing = others.filter { suggestion -> suggestion.lowercase().contains(needle) }
        return (starting + containing).take(limit)
    }

    /** Splits "34C-24-34" into its parts; null when it isn't in that form. */
    fun splitMeasurements(measurements: String?): MeasurementParts? {
        if (measurements.isNullOrEmpty()) {
            return MeasurementParts("", "", "")
        }
        val parts = measurements.split("-").map { part -> part.trim() }
        if (parts.size != MEASUREMENT_PARTS) {
            return null
        }
        return MeasurementParts(parts[0], parts[1], parts[2])
    }

    /** "34C-24-34" from its parts; empty when every part is. */
    fun joinMeasurements(measurements: MeasurementParts): String {
        val parts = listOf(measurements.bust, measurements.waist, measurements.hips).map { part -> part.trim() }
        if (parts.all { part -> part.isEmpty() }) {
            return ""
        }
        return parts.joinToString("-")
    }

    /** Whole years between two dates, or from a date until [today]; null when a date can't be read. */
    fun yearsBetween(from: String, until: String?, today: LocalDate): Int? {
        val start = parseDate(from) ?: return null
        var end = today
        if (!until.isNullOrBlank()) {
            end = parseDate(until) ?: return null
        }
        return java.time.Period.between(start, end).years
    }

    /** A date like 1995-04-12 (a longer timestamp is cut to its day); null when it can't be read. */
    private fun parseDate(text: String): LocalDate? {
        return runCatching { LocalDate.parse(text.trim().take(ISO_DATE_LENGTH)) }.getOrNull()
    }

    private const val MEASUREMENT_PARTS = 3
    private const val ISO_DATE_LENGTH = 10
}

/**
 * Editing a performer (`apps/web/src/lib/components/PerformerEditor.svelte`, `performers.ts`):
 * reading one for the form, the library's values per attribute, and saving only what changed,
 * since a saved field is locked against the sites. The form and diff logic are plain functions.
 */
object PerformerEditing {

    /** The performer as stored, for the form; null when gone. */
    suspend fun load(id: String): EditablePerformer? {
        val performer = VeilApi.client.query(PerformerForEditQuery(id)).execute().dataOrThrow().performer ?: return null
        val stored = performer.attributes
        return EditablePerformer(
            id = performer.id,
            name = performer.name,
            aliases = performer.aliases,
            details = performer.details,
            links = performer.links,
            imageUri = performer.image?.uri,
            imageUrl = performer.image?.url,
            attributes = PerformerAttributeValues(
                gender = stored.gender,
                birthdate = stored.birthdate,
                deathDate = stored.deathDate,
                country = stored.country,
                ethnicity = stored.ethnicity,
                eyeColor = stored.eyeColor,
                hairColor = stored.hairColor,
                heightCm = stored.heightCm,
                weightKg = stored.weightKg,
                measurements = stored.measurements,
                fakeTits = stored.fakeTits,
                tattoos = stored.tattoos,
                piercings = stored.piercings,
                careerStart = stored.careerStart,
                careerEnd = stored.careerEnd,
            ),
        )
    }

    /** The values a text attribute has across the library's performers, most common first. */
    suspend fun libraryValues(attribute: PerformerAttribute): List<String> {
        return VeilApi.client.query(PerformerAttributeValuesQuery(attribute.key)).execute().dataOrThrow().performerAttributeValues
    }

    /** Saves the user's [changes]; every field they set is locked against sites. */
    suspend fun save(id: String, changes: PerformerChanges) {
        val edit = PerformerEdit(
            name = Optional.presentIfNotNull(changes.name),
            aliases = Optional.presentIfNotNull(changes.aliases),
            details = Optional.presentIfNotNull(changes.details),
            links = Optional.presentIfNotNull(changes.links),
            attributes = Optional.presentIfNotNull(changes.attributes?.toInput()),
            imageUrl = Optional.presentIfNotNull(changes.imageUrl),
        )
        VeilApi.client.mutation(EditPerformerMutation(id, edit)).execute().dataOrThrow()
    }

    /** The attributes as the GraphQL input; the ones that are null are cleared. */
    private fun PerformerAttributeValues.toInput(): PerformerAttributesInput {
        return PerformerAttributesInput(
            gender = Optional.presentIfNotNull(gender),
            birthdate = Optional.presentIfNotNull(birthdate),
            deathDate = Optional.presentIfNotNull(deathDate),
            country = Optional.presentIfNotNull(country),
            ethnicity = Optional.presentIfNotNull(ethnicity),
            eyeColor = Optional.presentIfNotNull(eyeColor),
            hairColor = Optional.presentIfNotNull(hairColor),
            heightCm = Optional.presentIfNotNull(heightCm),
            weightKg = Optional.presentIfNotNull(weightKg),
            measurements = Optional.presentIfNotNull(measurements),
            fakeTits = Optional.presentIfNotNull(fakeTits),
            tattoos = Optional.presentIfNotNull(tattoos),
            piercings = Optional.presentIfNotNull(piercings),
            careerStart = Optional.presentIfNotNull(careerStart),
            careerEnd = Optional.presentIfNotNull(careerEnd),
        )
    }

    /** The form as it opens: the stored performer, text attributes capitalised, measurements split. */
    fun formOf(performer: EditablePerformer): PerformerForm {
        val attributes = PerformerAttribute.entries.associateWith { attribute -> shownText(performer.attributes, attribute) }
        return PerformerForm(
            name = performer.name,
            aliases = performer.aliases.joinToString(", "),
            imageUrl = performer.imageUri.orEmpty(),
            details = performer.details.orEmpty(),
            links = performer.links.joinToString("\n"),
            attributes = attributes,
            measurementParts = PerformerPresets.splitMeasurements(performer.attributes.measurements),
        )
    }

    /** A note under an input: the age by Born, the years active by Active since; null for the rest. */
    fun hint(attribute: PerformerAttribute, form: PerformerForm, today: LocalDate): String? {
        if (attribute == PerformerAttribute.Birthdate) {
            return ageHint(form.attributes[attribute].orEmpty(), form.attributes[PerformerAttribute.DeathDate].orEmpty(), today)
        }
        if (attribute == PerformerAttribute.CareerStart) {
            val start = form.attributes[attribute].orEmpty().trim().toIntOrNull()
            val end = form.attributes[PerformerAttribute.CareerEnd].orEmpty().trim().toIntOrNull()
            return activeHint(start, end, today.year)
        }
        return null
    }

    /** "29 years old", or "died aged 61" with a death date; null without a readable birth date. */
    private fun ageHint(birthdate: String, deathDate: String, today: LocalDate): String? {
        if (birthdate.isBlank()) {
            return null
        }
        val age = PerformerPresets.yearsBetween(birthdate, deathDate, today) ?: return null
        if (deathDate.isNotBlank()) {
            return "died aged $age"
        }
        return "$age years old"
    }

    /** "6 years active" from the start year until the end year or [currentYear]; null without a start. */
    private fun activeHint(start: Int?, end: Int?, currentYear: Int): String? {
        if (start == null || start == 0) {
            return null
        }
        var until = currentYear
        if (end != null && end != 0) {
            until = end
        }
        return "${until - start} years active"
    }

    /** The fields of [form] that differ from the [original]; nothing when it was left alone. */
    fun changes(original: EditablePerformer, form: PerformerForm): PerformerChanges {
        val aliases = splitList(form.aliases, ',')
        val links = splitList(form.links, '\n')
        val attributes = formAttributes(original.attributes, form)
        return PerformerChanges(
            name = form.name.trim().takeIf { name -> name != original.name },
            aliases = aliases.takeIf { it != original.aliases },
            imageUrl = form.imageUrl.trim().takeIf { url -> url != original.imageUri.orEmpty() },
            details = form.details.trim().takeIf { text -> text != original.details.orEmpty() },
            links = links.takeIf { it != original.links },
            attributes = attributes.takeIf { it != original.attributes.emptyAsNull() },
        )
    }

    /**
     * The attributes as the form holds them: measurements joined from their parts, and a country
     * the user changed in the list's spelling ("USA" as "United States"). An untouched country
     * stays as the site spelled it, so opening and saving locks nothing.
     */
    fun formAttributes(original: PerformerAttributeValues, form: PerformerForm): PerformerAttributeValues {
        val texts = form.attributes.toMutableMap()
        val parts = form.measurementParts
        if (parts != null) {
            texts[PerformerAttribute.Measurements] = joinedMeasurements(original.measurements, parts)
        }
        val country = texts[PerformerAttribute.Country].orEmpty()
        if (country.isNotEmpty() && country != original.country) {
            texts[PerformerAttribute.Country] = PerformerPresets.normalizeCountry(country)
        }
        for (attribute in PerformerAttribute.entries) {
            texts[attribute] = savedText(attribute, original, texts[attribute].orEmpty())
        }
        return PerformerAttributeValues.fromTexts(texts)
    }

    /** A text attribute as the form shows it: capitalised. */
    private fun shownText(values: PerformerAttributeValues, attribute: PerformerAttribute): String {
        val text = values.textOf(attribute)
        if (attribute.type == AttributeFieldType.Suggest) {
            return PerformerPresets.capitalize(text)
        }
        return text
    }

    /**
     * An attribute's text as it's saved: a suggested one capitalised, but in the stored spelling
     * when only the case differs, so showing it capitalised locks nothing; a date trimmed.
     */
    private fun savedText(attribute: PerformerAttribute, original: PerformerAttributeValues, value: String): String {
        if (attribute.type == AttributeFieldType.Date) {
            return value.trim()
        }
        if (attribute.type != AttributeFieldType.Suggest || value.isEmpty()) {
            return value
        }
        val stored = original.textOf(attribute)
        if (stored.isNotEmpty() && stored.lowercase() == value.trim().lowercase()) {
            return stored
        }
        return PerformerPresets.capitalize(value.trim())
    }

    /** The measurements from their parts; the stored spelling when the parts didn't change. */
    private fun joinedMeasurements(stored: String?, parts: MeasurementParts): String {
        val joined = PerformerPresets.joinMeasurements(parts)
        val storedParts = PerformerPresets.splitMeasurements(stored)
        if (storedParts != null && PerformerPresets.joinMeasurements(storedParts) == joined) {
            return stored.orEmpty()
        }
        return joined
    }

    /** Splits a list typed into one field, dropping blanks. */
    private fun splitList(text: String, separator: Char): List<String> {
        return text.split(separator).map { item -> item.trim() }.filter { item -> item.isNotEmpty() }
    }
}
