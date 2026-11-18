package com.playingwithclouds.veil.cams

/** What a rule asks for once its scene was seen long enough, or once its recording should end. */
sealed interface CamRuleAction {
    val ruleId: String
    val roomId: String

    /** A discreet notification. */
    data class Notify(override val ruleId: String, override val roomId: String) : CamRuleAction

    /** Start recording the room, [fromSecondsAgo] back in its buffer. */
    data class StartRecording(override val ruleId: String, override val roomId: String, val fromSecondsAgo: Int) : CamRuleAction

    /** Stop the recording this engine started. */
    data class StopRecording(override val ruleId: String, override val roomId: String) : CamRuleAction
}

/**
 * Decides when scene rules fire, from what the classifier saw in each room's frames (a few seconds
 * apart): a rule's scene must be seen without a gap longer than [maxGapMilliseconds] for its
 * [CamSceneRule.seenSeconds]; then it fires once and rests for its cooldown in that room. It also
 * ends the recordings it started: after their time, or once the scene was gone long enough. Only
 * one rule's recording runs per room. Pure: the caller grabs the frames and carries out the actions.
 */
class CamRuleEngine(private val maxGapMilliseconds: Long = DEFAULT_MAX_GAP_MILLISECONDS) {

    /** One rule in one room: since when its scene runs, when it was last seen, when it last fired. */
    private class Track {
        var firstSeen: Long? = null
        var lastSeen: Long? = null
        var lastFired: Long? = null
    }

    /** A recording this engine started in a room. */
    private class Recording(val ruleId: String, val startedAt: Long, var lastSeen: Long)

    private val tracks = mutableMapOf<String, Track>()
    private val recordings = mutableMapOf<String, Recording>()

    /** What [rules] ask for after [roomId]'s frame at [at] (epoch milliseconds) showed [seen] (tag names). */
    fun observe(rules: List<CamSceneRule>, roomId: String, at: Long, seen: Set<String>): List<CamRuleAction> {
        val actions = mutableListOf<CamRuleAction>()
        for (rule in rules.filter { candidate -> candidate.appliesTo(roomId) }) {
            val matched = rule.matches(seen)
            noteRecordingScene(rule, roomId, at, matched)
            if (fires(trackOf(rule.id, roomId), rule, at, matched)) {
                actions.addAll(fire(rule, roomId, at))
            }
        }
        actions.addAll(due(rules, roomId, at))
        return actions
    }

    /** Whether a recording this engine started runs in [roomId]. */
    fun isRecording(roomId: String): Boolean {
        return roomId in recordings
    }

    /** Forgets [roomId] (no longer watched); a recording of it is left to the backend, which ends it when the room goes offline. */
    fun forget(roomId: String) {
        recordings.remove(roomId)
        tracks.keys.removeAll { key -> key.endsWith("@$roomId") }
    }

    /** Updates the track of one rule; true when its scene ran long enough and it isn't resting. */
    private fun fires(track: Track, rule: CamSceneRule, at: Long, matched: Boolean): Boolean {
        val lastSeen = track.lastSeen
        if (!matched) {
            if (lastSeen != null && at - lastSeen > maxGapMilliseconds) {
                track.firstSeen = null
            }
            return false
        }
        if (lastSeen == null || at - lastSeen > maxGapMilliseconds || track.firstSeen == null) {
            track.firstSeen = at
        }
        track.lastSeen = at
        val firstSeen = track.firstSeen ?: at
        if (at - firstSeen < rule.seenSeconds * MILLISECONDS_PER_SECOND) {
            return false
        }
        val lastFired = track.lastFired
        if (lastFired != null && at - lastFired < rule.cooldownMinutes * MILLISECONDS_PER_MINUTE) {
            return false
        }
        track.lastFired = at
        return true
    }

    /** The actions of a rule that fired. */
    private fun fire(rule: CamSceneRule, roomId: String, at: Long): List<CamRuleAction> {
        val actions = mutableListOf<CamRuleAction>()
        if (rule.notify) {
            actions.add(CamRuleAction.Notify(rule.id, roomId))
        }
        if (rule.record && roomId !in recordings) {
            recordings[roomId] = Recording(rule.id, at, at)
            actions.add(CamRuleAction.StartRecording(rule.id, roomId, rule.seenSeconds + RECORD_LEAD_SECONDS))
        }
        return actions
    }

    /** Keeps the scene's last sighting for the room's recording when it belongs to [rule]. */
    private fun noteRecordingScene(rule: CamSceneRule, roomId: String, at: Long, matched: Boolean) {
        val recording = recordings[roomId] ?: return
        if (recording.ruleId == rule.id && matched) {
            recording.lastSeen = at
        }
    }

    /** The stop of the room's recording when its time is up, its scene is gone long enough or its rule went away. */
    private fun due(rules: List<CamSceneRule>, roomId: String, at: Long): List<CamRuleAction> {
        val recording = recordings[roomId] ?: return emptyList()
        val rule = rules.firstOrNull { candidate -> candidate.id == recording.ruleId && candidate.enabled }
        if (rule == null || recordingOver(rule, recording, at)) {
            recordings.remove(roomId)
            return listOf(CamRuleAction.StopRecording(recording.ruleId, roomId))
        }
        return emptyList()
    }

    /** Whether a rule's recording has run its course. */
    private fun recordingOver(rule: CamSceneRule, recording: Recording, at: Long): Boolean {
        val running = at - recording.startedAt
        if (rule.recordStop == RecordStop.FIXED) {
            return running >= rule.recordMinutes * MILLISECONDS_PER_MINUTE
        }
        if (running >= CamSceneRule.MAX_RECORD_MINUTES * MILLISECONDS_PER_MINUTE) {
            return true
        }
        return at - recording.lastSeen >= rule.goneMinutes * MILLISECONDS_PER_MINUTE
    }

    /** The track of [ruleId] in [roomId], made on first use. */
    private fun trackOf(ruleId: String, roomId: String): Track {
        return tracks.getOrPut("$ruleId@$roomId") { Track() }
    }

    companion object {
        /** A frame comes every ~5 s; one missed or misread frame doesn't break a scene. */
        const val DEFAULT_MAX_GAP_MILLISECONDS = 12_000L

        /** A recording begins this much before the scene was first seen. */
        const val RECORD_LEAD_SECONDS = 15

        private const val MILLISECONDS_PER_SECOND = 1000L
        private const val MILLISECONDS_PER_MINUTE = 60_000L
    }
}

/** A followed or auto-recorded room as background watching weighs it. */
data class WatchCandidate(val roomId: String, val live: Boolean, val followed: Boolean, val autoRecorded: Boolean, val viewers: Int)

/** Which rooms background watching takes, and whether it may run at all. */
object CamWatchPlan {

    /** Rooms watched at once at most. */
    const val MAX_ROOMS = 3

    /** Below this battery share (and not charging) background watching stops. */
    const val MIN_BATTERY_PERCENT = 20

    /**
     * The live rooms some rule covers (one naming it, or an every-room rule for a followed or
     * auto-recorded one), those named by a rule first, then by viewers; at most [limit].
     */
    fun targets(rules: List<CamSceneRule>, candidates: List<WatchCandidate>, limit: Int = MAX_ROOMS): List<String> {
        val enabled = rules.filter { rule -> rule.enabled }
        val named = enabled.mapNotNull { rule -> rule.roomId }.toSet()
        val everyRoom = enabled.any { rule -> rule.roomId == null }
        val covered = candidates.filter { room ->
            room.live && (room.roomId in named || (everyRoom && (room.followed || room.autoRecorded)))
        }
        return covered
            .sortedWith(compareByDescending<WatchCandidate> { room -> room.roomId in named }.thenByDescending { room -> room.viewers })
            .take(limit)
            .map { room -> room.roomId }
    }

    /** Whether background watching may run: a rule, enough battery (or charging), and Wi-Fi when Wi-Fi only is on. */
    fun allowed(hasRules: Boolean, batteryPercent: Int, charging: Boolean, wifiOnly: Boolean, metered: Boolean): Boolean {
        if (!hasRules) {
            return false
        }
        if (!charging && batteryPercent < MIN_BATTERY_PERCENT) {
            return false
        }
        return !(wifiOnly && metered)
    }
}
