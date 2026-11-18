package com.playingwithclouds.veil.cams

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class CamRuleEngineTest {

    private val deepthroat = setOf("Deepthroat")
    private val nothing = emptySet<String>()

    private val notifyRule = CamSceneRule(id = "n", tags = setOf("deepthroat"), seenSeconds = 10, notify = true, cooldownMinutes = 5)

    /** Feeds frames every 5 s from [start] with what each one shows; returns every action. */
    private fun feed(engine: CamRuleEngine, rules: List<CamSceneRule>, frames: List<Set<String>>, start: Long = 0, room: String = "room"): List<Pair<Long, CamRuleAction>> {
        val out = mutableListOf<Pair<Long, CamRuleAction>>()
        for ((index, seen) in frames.withIndex()) {
            val at = start + index * 5_000L
            engine.observe(rules, room, at, seen).forEach { action -> out.add(at to action) }
        }
        return out
    }

    @Test
    fun aRuleFiresOnceItsSceneRanLongEnough() {
        val actions = feed(CamRuleEngine(), listOf(notifyRule), listOf(deepthroat, deepthroat, deepthroat, deepthroat))
        assertEquals(listOf(10_000L to CamRuleAction.Notify("n", "room")), actions)
    }

    @Test
    fun oneMissedFrameDoesNotBreakTheSceneButAGapDoes() {
        val tolerated = feed(CamRuleEngine(), listOf(notifyRule), listOf(deepthroat, nothing, deepthroat))
        assertEquals(1, tolerated.size)
        val broken = feed(CamRuleEngine(), listOf(notifyRule), listOf(deepthroat, nothing, nothing, nothing, deepthroat, deepthroat))
        assertTrue(broken.isEmpty())
    }

    @Test
    fun theCooldownKeepsItQuiet() {
        val frames = List(80) { deepthroat }
        val actions = feed(CamRuleEngine(), listOf(notifyRule), frames)
        // Fires at 10 s, then again once 5 minutes have passed.
        assertEquals(listOf(10_000L, 310_000L), actions.map { (at, _) -> at })
    }

    @Test
    fun rulesOfOtherRoomsAndSwitchedOffOnesStayOut() {
        val other = notifyRule.copy(id = "o", roomId = "elsewhere")
        val off = notifyRule.copy(id = "x", enabled = false)
        assertTrue(feed(CamRuleEngine(), listOf(other, off), List(4) { deepthroat }).isEmpty())
        assertFalse(other.appliesTo("room"))
        assertTrue(other.appliesTo("elsewhere"))
    }

    @Test
    fun aFixedRecordingStartsBackInTheBufferAndStopsInTime() {
        val rule = CamSceneRule(id = "r", tags = deepthroat, seenSeconds = 10, notify = false, record = true, recordMinutes = 1, cooldownMinutes = 60)
        val engine = CamRuleEngine()
        val actions = feed(engine, listOf(rule), List(20) { deepthroat })
        assertEquals(10_000L to CamRuleAction.StartRecording("r", "room", 10 + CamRuleEngine.RECORD_LEAD_SECONDS), actions[0])
        assertEquals(70_000L to CamRuleAction.StopRecording("r", "room"), actions[1])
        assertEquals(2, actions.size)
        assertFalse(engine.isRecording("room"))
    }

    @Test
    fun aRecordingUntilGoneEndsWhenTheSceneIsGoneLongEnough() {
        val rule = CamSceneRule(id = "g", tags = deepthroat, seenSeconds = 0, notify = false, record = true, recordStop = RecordStop.UNTIL_GONE, goneMinutes = 1)
        val engine = CamRuleEngine()
        val frames = List(10) { deepthroat } + List(13) { nothing }
        val actions = feed(engine, listOf(rule), frames)
        assertEquals(CamRuleAction.StartRecording("g", "room", CamRuleEngine.RECORD_LEAD_SECONDS), actions[0].second)
        // Last seen at 45 s, so it stops a minute later.
        assertEquals(105_000L to CamRuleAction.StopRecording("g", "room"), actions[1])
    }

    @Test
    fun aRecordingStopsWhenItsRuleGoesAway() {
        val rule = CamSceneRule(id = "r", tags = deepthroat, seenSeconds = 0, notify = false, record = true, recordMinutes = 30)
        val engine = CamRuleEngine()
        feed(engine, listOf(rule), listOf(deepthroat))
        assertTrue(engine.isRecording("room"))
        assertEquals(listOf(CamRuleAction.StopRecording("r", "room")), engine.observe(emptyList(), "room", 5_000, deepthroat))
    }

    @Test
    fun onlyOneRecordingRunsPerRoom() {
        val first = CamSceneRule(id = "a", tags = deepthroat, seenSeconds = 0, notify = false, record = true)
        val second = first.copy(id = "b")
        val actions = feed(CamRuleEngine(), listOf(first, second), listOf(deepthroat))
        assertEquals(1, actions.count { (_, action) -> action is CamRuleAction.StartRecording })
    }

    @Test
    fun backgroundWatchingTakesNamedRoomsFirstThenTheBusiest() {
        val rules = listOf(CamSceneRule(id = "every"), CamSceneRule(id = "named", roomId = "quiet"))
        val candidates = listOf(
            WatchCandidate("busy", live = true, followed = true, autoRecorded = false, viewers = 900),
            WatchCandidate("quiet", live = true, followed = false, autoRecorded = false, viewers = 3),
            WatchCandidate("offline", live = false, followed = true, autoRecorded = false, viewers = 0),
            WatchCandidate("stranger", live = true, followed = false, autoRecorded = false, viewers = 5000),
            WatchCandidate("recorded", live = true, followed = false, autoRecorded = true, viewers = 50),
            WatchCandidate("small", live = true, followed = true, autoRecorded = false, viewers = 10),
        )
        assertEquals(listOf("quiet", "busy", "recorded"), CamWatchPlan.targets(rules, candidates))
        assertEquals(listOf("quiet"), CamWatchPlan.targets(listOf(CamSceneRule(id = "named", roomId = "quiet")), candidates))
    }

    @Test
    fun backgroundWatchingNeedsBatteryAndWifiWhenAsked() {
        assertTrue(CamWatchPlan.allowed(true, batteryPercent = 50, charging = false, wifiOnly = true, metered = false))
        assertFalse(CamWatchPlan.allowed(true, batteryPercent = 15, charging = false, wifiOnly = false, metered = false))
        assertTrue(CamWatchPlan.allowed(true, batteryPercent = 15, charging = true, wifiOnly = false, metered = false))
        assertFalse(CamWatchPlan.allowed(true, batteryPercent = 80, charging = false, wifiOnly = true, metered = true))
        assertTrue(CamWatchPlan.allowed(true, batteryPercent = 80, charging = false, wifiOnly = false, metered = true))
        assertFalse(CamWatchPlan.allowed(false, batteryPercent = 80, charging = true, wifiOnly = false, metered = false))
    }

    @Test
    fun rulesSurviveStorage() {
        val rule = CamSceneRule(id = "s", roomId = "room", roomName = "Room", tags = setOf("Deepthroat", "Facial"), record = true, recordStop = RecordStop.UNTIL_GONE)
        val everyRoom = CamSceneRule(id = "e")
        assertEquals(listOf(rule, everyRoom), CamSceneRule.listFromJson(CamSceneRule.listToJson(listOf(rule, everyRoom))))
        assertTrue(CamSceneRule.listFromJson("not json").isEmpty())
    }

    @Test
    fun yuvGreyBecomesGrey() {
        val width = 4
        val height = 2
        val planes = YuvPlanes(
            y = ByteArray(width * height) { 126.toByte() },
            u = ByteArray(width) { 128.toByte() },
            v = ByteArray(width) { 128.toByte() },
            width = width,
            height = height,
            yRowStride = width,
            uvRowStride = width,
            uvPixelStride = 2,
        )
        val pixels = YuvFrames.toArgb(planes)
        val red = (pixels[0] shr 16) and 0xFF
        val green = (pixels[0] shr 8) and 0xFF
        val blue = pixels[0] and 0xFF
        assertEquals(red, green)
        assertEquals(green, blue)
        assertEquals(128, red)
        assertEquals(width * height, pixels.size)
    }
}
