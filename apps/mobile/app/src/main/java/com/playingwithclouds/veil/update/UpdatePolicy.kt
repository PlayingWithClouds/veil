package com.playingwithclouds.veil.update

/** When to look for an update and whether a release is newer than the installed app. */
object UpdatePolicy {

    /** The least time between two automatic checks. */
    const val CHECK_INTERVAL_MILLISECONDS = 24L * 60 * 60 * 1000

    private val versionPattern = Regex("""\d+\.\d+\.\d+""")

    /**
     * The version code a release tag builds into, with the formula of `veilVersionCode` in
     * app/build.gradle.kts: `v1.2.3` → 1_002_003. Null when the tag is not `major.minor.patch`.
     */
    fun versionCodeOf(tag: String): Int? {
        val versionName = tag.trim().removePrefix("v").substringBefore("-")
        if (!versionPattern.matches(versionName)) {
            return null
        }
        val parts = versionName.split(".").map { part -> part.toIntOrNull() ?: return null }
        return parts[0] * 1_000_000 + parts[1] * 1_000 + parts[2]
    }

    /** Whether the release tagged [tag] is newer than the installed [installedVersionCode]; false for unreadable tags. */
    fun isNewer(tag: String, installedVersionCode: Int): Boolean {
        val releaseVersionCode = versionCodeOf(tag) ?: return false
        return releaseVersionCode > installedVersionCode
    }

    /** Whether an automatic check is due, given when the last one succeeded (0 = never). */
    fun isCheckDue(lastCheckMillis: Long, nowMillis: Long): Boolean {
        if (lastCheckMillis <= 0) {
            return true
        }
        // A clock set back past the last check would otherwise block checks until it catches up.
        if (nowMillis < lastCheckMillis) {
            return true
        }
        return nowMillis - lastCheckMillis >= CHECK_INTERVAL_MILLISECONDS
    }
}
