package com.playingwithclouds.veil.ui

import java.io.File
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Keeps screens on the design system: colours, corner shapes, type sizes and spacing come from
 * `ui/theme` tokens and `ui/design` components. Only those two packages may define raw values.
 */
class DesignTokenUsageTest {

    /** A forbidden pattern and the token to use instead. */
    private data class Rule(val pattern: Regex, val instead: String)

    private val rules = listOf(
        Rule(Regex("""RoundedCornerShape\("""), "VeilShapes"),
        Rule(Regex("""\bCircleShape\b"""), "VeilShapes.capsule"),
        Rule(Regex("""Color\(0x"""), "VeilColors"),
        Rule(Regex("""Color\.(White|Black)\b"""), "VeilColors"),
        Rule(Regex("""MaterialTheme\.colorScheme"""), "VeilColors"),
        Rule(Regex("""\b\d+(\.\d+)?\.sp\b"""), "MaterialTheme.typography"),
        Rule(Regex("""\b(padding|spacedBy|PaddingValues)\([^)]*\b\d+(\.\d+)?\.dp"""), "VeilSpacing"),
    )

    private val uiRoot = File("src/main/java/com/playingwithclouds/veil/ui")

    @Test
    fun screensUseDesignTokens() {
        assertTrue("UI sources not found at ${uiRoot.absolutePath}", uiRoot.isDirectory)
        val violations = mutableListOf<String>()
        for (file in screenSources()) {
            file.readLines().forEachIndexed { index, line ->
                for (rule in rules) {
                    if (rule.pattern.containsMatchIn(line)) {
                        violations.add("${file.relativeTo(uiRoot)}:${index + 1}: use ${rule.instead}: ${line.trim()}")
                    }
                }
            }
        }
        assertTrue("Raw design values outside ui/design and ui/theme:\n" + violations.joinToString("\n"), violations.isEmpty())
    }

    /** Every Kotlin file under `ui/` except the design kit and the theme. */
    private fun screenSources(): List<File> {
        val exempt = listOf(File(uiRoot, "design"), File(uiRoot, "theme"))
        return uiRoot.walkTopDown()
            .filter { file -> file.isFile && file.extension == "kt" }
            .filter { file -> exempt.none { directory -> file.startsWith(directory) } }
            .toList()
    }
}
