package com.playingwithclouds.veil.cams

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import org.json.JSONArray
import org.json.JSONObject
import java.util.UUID

/** When a rule's recording stops: after a fixed time, or once its scenes are gone for a while. */
enum class RecordStop { FIXED, UNTIL_GONE }

/**
 * A rule on a cam room's scenes: when one of [tags] (the scene model's tag names, "Deepthroat") is
 * seen for [seenSeconds], notify and/or record. [roomId] null is every room: the followed and
 * auto-recorded ones in the background, and whichever room is open in the app. A recording runs
 * [recordMinutes] ([RecordStop.FIXED]) or until the scenes are gone for [goneMinutes]
 * ([RecordStop.UNTIL_GONE], at most [MAX_RECORD_MINUTES]); it begins [seenSeconds] and a little
 * more back in the room's buffer, so the start of the scene is in it. After firing, the rule
 * rests [cooldownMinutes] for that room.
 */
data class CamSceneRule(
    val id: String = UUID.randomUUID().toString(),
    val roomId: String? = null,
    val roomName: String? = null,
    val tags: Set<String> = emptySet(),
    val seenSeconds: Int = 10,
    val notify: Boolean = true,
    val record: Boolean = false,
    val recordStop: RecordStop = RecordStop.FIXED,
    val recordMinutes: Int = 10,
    val goneMinutes: Int = 2,
    val cooldownMinutes: Int = 15,
    val enabled: Boolean = true,
) {

    /** Whether the rule looks at [roomId]. */
    fun appliesTo(roomId: String): Boolean {
        return enabled && (this.roomId == null || this.roomId == roomId)
    }

    /** Whether [seen] (tag names) holds one of the rule's scenes; names compare without case. */
    fun matches(seen: Set<String>): Boolean {
        val wanted = tags.map { tag -> tag.lowercase() }.toSet()
        return seen.any { tag -> tag.lowercase() in wanted }
    }

    /** The rule as stored. */
    fun toJson(): JSONObject {
        return JSONObject()
            .put("id", id)
            .put("roomId", roomId ?: JSONObject.NULL)
            .put("roomName", roomName ?: JSONObject.NULL)
            .put("tags", JSONArray(tags.sorted()))
            .put("seenSeconds", seenSeconds)
            .put("notify", notify)
            .put("record", record)
            .put("recordStop", recordStop.name)
            .put("recordMinutes", recordMinutes)
            .put("goneMinutes", goneMinutes)
            .put("cooldownMinutes", cooldownMinutes)
            .put("enabled", enabled)
    }

    companion object {
        /** The longest a recording waiting for its scenes to end runs. */
        const val MAX_RECORD_MINUTES = 120

        /** A stored rule; defaults fill what's missing. */
        fun fromJson(json: JSONObject): CamSceneRule {
            val tags = json.optJSONArray("tags") ?: JSONArray()
            return CamSceneRule(
                id = json.getString("id"),
                roomId = optionalString(json, "roomId"),
                roomName = optionalString(json, "roomName"),
                tags = (0 until tags.length()).map { index -> tags.getString(index) }.toSet(),
                seenSeconds = json.optInt("seenSeconds", 10),
                notify = json.optBoolean("notify", true),
                record = json.optBoolean("record", false),
                recordStop = RecordStop.entries.firstOrNull { stop -> stop.name == json.optString("recordStop") } ?: RecordStop.FIXED,
                recordMinutes = json.optInt("recordMinutes", 10),
                goneMinutes = json.optInt("goneMinutes", 2),
                cooldownMinutes = json.optInt("cooldownMinutes", 15),
                enabled = json.optBoolean("enabled", true),
            )
        }

        /** A string field, null when absent or JSON null. */
        private fun optionalString(json: JSONObject, name: String): String? {
            if (!json.has(name) || json.isNull(name)) {
                return null
            }
            return json.getString(name)
        }

        /** Rules from their stored JSON array; unreadable entries are left out. */
        fun listFromJson(text: String): List<CamSceneRule> {
            val array = runCatching { JSONArray(text) }.getOrNull() ?: return emptyList()
            return (0 until array.length()).mapNotNull { index -> runCatching { fromJson(array.getJSONObject(index)) }.getOrNull() }
        }

        /** Rules as a JSON array. */
        fun listToJson(rules: List<CamSceneRule>): String {
            return JSONArray(rules.map { rule -> rule.toJson() }).toString()
        }
    }
}

/**
 * The scene rules, kept on this phone (the classifier runs here, so the backend's automations
 * couldn't see a scene anyway), and whether rooms are watched in the background.
 */
object CamRuleStore {

    private const val PREFERENCES = "camRules"
    private const val KEY_RULES = "rules"
    private const val KEY_BACKGROUND = "background"
    private const val KEY_BLOCKED = "backgroundBlocked"

    private lateinit var preferences: SharedPreferences

    private val mutableRules = MutableStateFlow<List<CamSceneRule>>(emptyList())
    private val mutableBackground = MutableStateFlow(true)
    private val mutableBlocked = MutableStateFlow(false)

    /** Every rule, in the order they were made. */
    val rules: StateFlow<List<CamSceneRule>> = mutableRules

    /** Whether followed and auto-recorded rooms are watched while the app is closed; on by default. */
    val background: StateFlow<Boolean> = mutableBackground

    /** Android refused the last background start (battery optimisation); Settings offers the way out. */
    val backgroundBlocked: StateFlow<Boolean> = mutableBlocked

    /** Reads the rules; called once from the application. */
    fun initialize(context: Context) {
        preferences = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
        mutableRules.value = CamSceneRule.listFromJson(preferences.getString(KEY_RULES, "[]").orEmpty())
        mutableBackground.value = preferences.getBoolean(KEY_BACKGROUND, true)
        mutableBlocked.value = preferences.getBoolean(KEY_BLOCKED, false)
    }

    /** Adds [rule], or replaces the one with its id. */
    fun save(rule: CamSceneRule) {
        val others = mutableRules.value.filter { existing -> existing.id != rule.id }
        val index = mutableRules.value.indexOfFirst { existing -> existing.id == rule.id }
        val next = others.toMutableList()
        if (index < 0) {
            next.add(rule)
        } else {
            next.add(index, rule)
        }
        store(next)
    }

    /** Deletes the rule with [id]. */
    fun delete(id: String) {
        store(mutableRules.value.filter { rule -> rule.id != id })
    }

    /** Turns background watching on or off. */
    fun setBackground(on: Boolean) {
        preferences.edit { putBoolean(KEY_BACKGROUND, on) }
        mutableBackground.value = on
    }

    /** Records whether Android refused to start background watching. */
    fun setBackgroundBlocked(blocked: Boolean) {
        preferences.edit { putBoolean(KEY_BLOCKED, blocked) }
        mutableBlocked.value = blocked
    }

    /** Saves and publishes [rules]. */
    private fun store(rules: List<CamSceneRule>) {
        preferences.edit { putString(KEY_RULES, CamSceneRule.listToJson(rules)) }
        mutableRules.value = rules
    }
}
