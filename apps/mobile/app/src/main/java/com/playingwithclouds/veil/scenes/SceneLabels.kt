package com.playingwithclouds.veil.scenes

/**
 * The one place taxonomy labels become tag names. The names follow how the sites tag videos
 * ("Doggystyle", "Facial"), so a segment's tag is the same tag a site gave and taste learned from
 * either counts for both. Labels missing here (a newer taxonomy) fall back to the manifest's
 * title, else their name in words; labels in [skipped] never become segments.
 */
object SceneLabels {

    /** The heads whose labels become segments and cover tags: what happens, not how it's filmed. */
    val segmentHeads = listOf("oral_intensity", "act", "position", "fetish", "climax", "camera")

    /** Tag names by `head.label`. */
    private val tagNames = mapOf(
        "oral_intensity.regular" to "Blowjob",
        "oral_intensity.deepthroat" to "Deepthroat",
        "oral_intensity.facefuck" to "Face Fucking",
        "act.blowjob" to "Blowjob",
        "act.pussy_licking" to "Pussy Licking",
        "act.rimming" to "Rimming",
        "act.ball_licking" to "Ball Licking",
        "act.sixty_nine" to "69",
        "act.facesitting" to "Facesitting",
        "act.kissing" to "Kissing",
        "act.breast_play" to "Breast Play",
        "act.handjob" to "Handjob",
        "act.fingering" to "Fingering",
        "act.anal_fingering" to "Anal Fingering",
        "act.fisting" to "Fisting",
        "act.masturbation" to "Masturbation",
        "act.squirting" to "Squirting",
        "act.vaginal" to "Vaginal Sex",
        "act.anal" to "Anal",
        "act.double_pen" to "Double Penetration",
        "act.spitroast" to "Spitroast",
        "act.strapon" to "Strapon",
        "act.toy_insertion" to "Dildo",
        "act.toy_external" to "Vibrator",
        "act.titjob" to "Titjob",
        "act.footjob" to "Footjob",
        "act.tribbing" to "Tribbing",
        "act.grinding" to "Dry Humping",
        "act.undressing" to "Undressing",
        "act.stripping" to "Striptease",
        "position.missionary" to "Missionary",
        "position.mating_press" to "Mating Press",
        "position.doggy" to "Doggystyle",
        "position.prone_bone" to "Prone Bone",
        "position.cowgirl" to "Cowgirl",
        "position.reverse_cowgirl" to "Reverse Cowgirl",
        "position.spooning" to "Spooning",
        "position.standing" to "Standing Sex",
        "position.standing_carry" to "Standing Carry",
        "position.sitting" to "Lap Sitting",
        "position.edge_of_bed" to "Edge of Bed",
        "position.piledriver" to "Piledriver",
        "fetish.impact_play" to "Impact Play",
        "fetish.choking" to "Choking",
        "fetish.gagging" to "Gagging",
        "fetish.puking" to "Puking",
        "fetish.hair_pulling" to "Hair Pulling",
        "fetish.gag_ball" to "Ball Gag",
        "fetish.collar_leash" to "Collar and Leash",
        "fetish.wax_play" to "Wax Play",
        "fetish.feet" to "Feet",
        "fetish.watersports" to "Pissing",
        "fetish.latex_leather" to "Latex",
        "fetish.costume_roleplay" to "Cosplay",
        "fetish.oil" to "Oiled",
        "fetish.machine" to "Fucking Machine",
        "fetish.glory_hole" to "Gloryhole",
        "climax.cumshot_face" to "Facial",
        "climax.cumshot_mouth" to "Cum in Mouth",
        "climax.cumshot_body" to "Cumshot",
        "climax.creampie" to "Creampie",
        "climax.swallow" to "Swallow",
        "climax.cum_play" to "Cum Play",
        "climax.orgasm" to "Orgasm",
        "camera.pov" to "POV",
    )

    /** Labels that say too little on their own to be a segment or a liked scene. */
    private val skipped = setOf(
        "position.kneeling",
        "position.lying_oral",
        "position.side_by_side",
        "position.other",
        "camera.overhead",
        "camera.low_angle",
        "camera.selfie",
        "camera.mirror",
        "camera.webcam",
        "camera.handheld",
        "camera.split_screen",
    )

    /** The group a liked-scenes picker shows each head under. */
    private val groupTitles = mapOf(
        "oral_intensity" to "Oral",
        "act" to "Acts",
        "position" to "Positions",
        "fetish" to "Kinks",
        "climax" to "Finish",
        "camera" to "Camera",
    )

    /** The tag name of [label] of [head]; null when the label never becomes a segment. */
    fun tagName(head: String, label: ModelLabel): String? {
        val key = "$head.${label.name}"
        if (head !in segmentHeads || key in skipped) {
            return null
        }
        val known = tagNames[key]
        if (known != null) {
            return known
        }
        if (label.title.isNotBlank()) {
            return words(label.title)
        }
        return words(label.name)
    }

    /** `cum_play` → "Cum Play". */
    fun words(name: String): String {
        return name.split('_', ' ').filter { part -> part.isNotEmpty() }.joinToString(" ") { part ->
            part.replaceFirstChar { first -> first.uppercaseChar() }
        }
    }

    /** One group of the liked-scenes picker: a title and its tag names in taxonomy order. */
    data class Group(val title: String, val tags: List<String>)

    /**
     * What the liked-scenes picker offers: with a model, its heads' labels; without one, the
     * built-in names (the taxonomy the app was written against), so the choice can be made first.
     */
    fun pickerGroups(manifest: SceneModelManifest?): List<Group> {
        if (manifest == null) {
            return builtInGroups()
        }
        return segmentHeads.mapNotNull { headName ->
            val head = manifest.head(headName) ?: return@mapNotNull null
            val tags = head.labels.filter { label -> !label.deprecated }.mapNotNull { label -> tagName(headName, label) }.distinct()
            Group(groupTitles[headName] ?: words(headName), tags)
        }.filter { group -> group.tags.isNotEmpty() }
    }

    /** The picker's groups from the built-in names. */
    private fun builtInGroups(): List<Group> {
        return segmentHeads.map { headName ->
            val tags = tagNames.filterKeys { key -> key.startsWith("$headName.") && key != "act.blowjob" }.values.distinct()
            Group(groupTitles[headName] ?: words(headName), tags)
        }
    }
}
