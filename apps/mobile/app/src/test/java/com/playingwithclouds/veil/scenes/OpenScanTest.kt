package com.playingwithclouds.veil.scenes

import com.playingwithclouds.veil.data.VideoSegment
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/** A manifest with the broad heads: the rating in both graphs, the scene phase in the temporal graph only. */
private val phaseManifestJson = """
{
  "format": 1,
  "version": "test-phase",
  "taxonomyVersion": 2,
  "input": {"size": 224, "mean": [0.485, 0.456, 0.406], "std": [0.229, 0.224, 0.225]},
  "frame": {"file": "frame.onnx", "input": "images", "embedding": "embedding", "heads": ["rating", "position"]},
  "temporal": {"file": "temporal.onnx", "input": "embeddings", "window": 16, "frameSeconds": 0.5, "heads": ["rating", "position", "scene_phase"]},
  "heads": [
    {"name": "rating", "kind": "exclusive", "labels": [{"name": "sfw"}, {"name": "suggestive"}, {"name": "nudity"}, {"name": "sexual_solo"}, {"name": "sexual_partner"}]},
    {"name": "position", "kind": "multi", "labels": [{"name": "cowgirl"}]},
    {"name": "scene_phase", "kind": "exclusive", "labels": [{"name": "intro"}, {"name": "foreplay"}, {"name": "main"}, {"name": "climax"}, {"name": "outro"}, {"name": "transition"}]}
  ]
}
""".trimIndent()

/** A master playlist as the backend's proxy writes it: root-relative variant URIs, an audio-only rendition. */
private val masterPlaylist = """
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=2500000,AVERAGE-BANDWIDTH=2000000,RESOLUTION=1280x720,CODECS="avc1.64001f,mp4a.40.2"
/api/stream/manifest?s=abc&u=https%3A%2F%2Fcdn.example%2F720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=64000,CODECS="mp4a.40.2"
/api/stream/manifest?s=abc&u=https%3A%2F%2Fcdn.example%2Faudio.m3u8
#EXT-X-STREAM-INF:AVERAGE-BANDWIDTH=9000000,BANDWIDTH=400000,RESOLUTION=426x240,CODECS="avc1.42c015,mp4a.40.2"
/api/stream/manifest?s=abc&u=https%3A%2F%2Fcdn.example%2F240.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=900000,RESOLUTION=640x360
360/index.m3u8
""".trimIndent()

class OpenScanTest {

    private val manifest = SceneModelManifest.parse(phaseManifestJson)

    /** One frame's rating, from its label's name, sure of it. */
    private fun rated(label: String): Map<String, FloatArray> {
        val names = listOf("sfw", "suggestive", "nudity", "sexual_solo", "sexual_partner")
        return mapOf("rating" to FloatArray(names.size) { index -> if (names[index] == label) 1f else 0f })
    }

    @Test
    fun theLowestVariantWithVideoIsPickedAndResolved() {
        val variants = HlsVariants.parse(masterPlaylist)
        assertEquals(listOf(2_500_000L, 64_000L, 400_000L, 900_000L), variants.map { variant -> variant.bandwidth })
        assertEquals(
            "http://127.0.0.1:47831/api/stream/manifest?s=abc&u=https%3A%2F%2Fcdn.example%2F240.m3u8",
            HlsVariants.lowestUrl(masterPlaylist, "http://127.0.0.1:47831/api/stream/manifest?s=abc&u=https%3A%2F%2Fcdn.example%2Fmaster.m3u8"),
        )
        val mediaPlaylist = "#EXTM3U\n#EXT-X-TARGETDURATION:6\n#EXTINF:6.0,\nseg0.ts\n#EXT-X-ENDLIST"
        assertNull(HlsVariants.lowestUrl(mediaPlaylist, "http://127.0.0.1:47831/a.m3u8"))
    }

    @Test
    fun relativeVariantsResolveBesideTheMaster() {
        val playlist = "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=900000\n360/index.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=300000\nlow/index.m3u8"
        assertEquals("https://cdn.example/v/low/index.m3u8", HlsVariants.lowestUrl(playlist, "https://cdn.example/v/master.m3u8"))
    }

    @Test
    fun wideSpacingKeepsOnlyTheTemporalOnlyHeads() {
        val video = mapOf("position" to floatArrayOf(0.9f), "scene_phase" to floatArrayOf(1f, 0f, 0f, 0f, 0f, 0f))
        assertEquals(video, FrameProbabilities.videoAtSpacing(video, 2.0, manifest))
        assertEquals(setOf("scene_phase"), FrameProbabilities.videoAtSpacing(video, ScanPlan.MIN_STEP_SECONDS, manifest)?.keys)
    }

    @Test
    fun theCoverJoinsFewStillsInTheMiddle() {
        assertEquals(listOf("a", "cover", "b"), RoughPass.pictures(listOf("a", "b"), "cover"))
        assertEquals(listOf("cover"), RoughPass.pictures(emptyList(), "cover"))
        assertEquals(listOf("a", "b", "c", "d"), RoughPass.pictures(listOf("a", "b", "c", "d"), "cover"))
        assertEquals(emptyList<String>(), RoughPass.pictures(emptyList(), null))
    }

    @Test
    fun theStreamReplacesTheRoughReadingWhereItIsScanned() {
        val rough = listOf(TimedSegment("Blowjob", 0.0, 300.0), TimedSegment("Cowgirl", 300.0, 600.0))
        val stream = listOf(TimedSegment("Doggystyle", 120.0, 200.0))
        val merged = RoughPass.merge(stream, rough, listOf(100.0..250.0))
        assertEquals(
            listOf(
                TimedSegment("Blowjob", 0.0, 100.0),
                TimedSegment("Doggystyle", 120.0, 200.0),
                TimedSegment("Blowjob", 250.0, 300.0),
                TimedSegment("Cowgirl", 300.0, 600.0),
            ),
            merged,
        )
        assertEquals(stream, RoughPass.merge(stream, null, listOf(100.0..250.0)))
        assertEquals(stream, RoughPass.merge(stream, rough, listOf(0.0..600.0)))
    }

    @Test
    fun broadLabelsComeFromThePhaseThenTheRating() {
        val intro = rated("sexual_partner") + ("scene_phase" to floatArrayOf(0.9f, 0f, 0.1f, 0f, 0f, 0f))
        val talkingInMain = rated("sfw") + ("scene_phase" to floatArrayOf(0f, 0f, 1f, 0f, 0f, 0f))
        assertEquals("Intro", best(BroadLabels.distribution(intro, manifest)))
        assertEquals("Talking", best(BroadLabels.distribution(talkingInMain, manifest)))
        assertEquals("Solo", best(BroadLabels.distribution(rated("sexual_solo"), manifest)))
        assertNull(BroadLabels.distribution(mapOf("position" to floatArrayOf(0.9f)), manifest))
    }

    @Test
    fun stillsCountOneByOne() {
        val timeline = ScanTimeline(listOf(0.0, 100.0, 200.0), 100.0, 0, SegmentRules.STILLS)
        timeline.add(0, floatArrayOf(0f), rated("sfw") + ("position" to floatArrayOf(0f)))
        timeline.add(1, floatArrayOf(0f), rated("sexual_partner") + ("position" to floatArrayOf(0.9f)))
        timeline.add(2, floatArrayOf(0f), rated("sexual_partner") + ("position" to floatArrayOf(0f)))
        assertEquals(listOf(TimedSegment("Cowgirl", 100.0, 200.0)), timeline.segments(manifest))
        assertEquals(listOf(TimedSegment("Talking", 0.0, 100.0), TimedSegment("Sex", 100.0, 300.0)), timeline.broadSegments(manifest))
    }

    @Test
    fun everySectionIsTitledFromStartToEnd() {
        val sections = SceneSections.sections(
            listOf(TimedSegment("Blowjob", 60.0, 180.0), TimedSegment("Doggystyle", 400.0, 600.0)),
            listOf(TimedSegment("Talking", 0.0, 70.0), TimedSegment("Sex", 150.0, 420.0), TimedSegment("Posing", 300.0, 305.0)),
            700.0,
        )
        assertEquals(listOf("Talking", "Blowjob", "Sex", "Doggystyle"), sections.map { section -> section.title })
        assertEquals(0.0, sections.first().startSeconds, 0.0)
        assertEquals(700.0, sections.last().endSeconds, 0.0)
        for ((previous, next) in sections.zipWithNext()) {
            assertEquals(previous.endSeconds, next.startSeconds, 0.0)
        }
    }

    @Test
    fun chaptersHaveNoUntitledStretchEvenWithoutBroadLabels() {
        val stored = listOf(
            VideoSegment("a", "Blowjob", 100.0, 200.0, SceneSections.ORIGIN_AI, "Blowjob"),
            VideoSegment("b", "Cowgirl", 500.0, 650.0, SceneSections.ORIGIN_AI, "Cowgirl"),
        )
        val chapters = SceneSections.chapters(stored, durationSeconds = 900.0)
        assertTrue(chapters.all { chapter -> chapter.title.isNotEmpty() })
        assertEquals(0.0, chapters.first().seconds, 0.0)
        assertEquals(900.0, chapters.last().endSeconds ?: 0.0, 0.0)
        assertTrue(SceneSections.chapters(emptyList(), durationSeconds = 900.0).isEmpty())
    }

    /** The title with the highest value. */
    private fun best(values: FloatArray?): String? {
        if (values == null) {
            return null
        }
        return BroadLabels.titles[values.indices.maxBy { index -> values[index] }]
    }
}
