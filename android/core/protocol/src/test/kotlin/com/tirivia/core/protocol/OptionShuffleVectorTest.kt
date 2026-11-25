package com.tirivia.core.protocol

import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File

/**
 * Runs protocol/testvectors/shuffle_vectors.json against the Android OptionShuffle.
 * The same file is run by Go (reference) and Swift; all must agree. The vector
 * "high_bytes_mask_check" catches the Kotlin signed-byte pitfall (bytes must be masked 0..255).
 * Uses the real org.json artifact (Android's bundled org.json is a stub in local unit tests).
 * The cue-verification runner is added with the Android CueVerifier in Phase 2a.
 */
class OptionShuffleVectorTest {
    private fun vectorsFile(): File {
        // Gradle runs unit tests with the module directory as working dir: android/core/protocol
        val candidates = listOf("../../../protocol/testvectors/shuffle_vectors.json", "protocol/testvectors/shuffle_vectors.json")
        return candidates.map(::File).firstOrNull { it.exists() }
            ?: error("shuffle_vectors.json not found from ${File(".").absolutePath}")
    }

    private fun hex(s: String): ByteArray = ByteArray(s.length / 2) { i -> s.substring(2 * i, 2 * i + 2).toInt(16).toByte() }

    @Test
    fun sharedShuffleVectors() {
        val suite = JSONObject(vectorsFile().readText())
        assertEquals(1, suite.getInt("format_version"))
        val vectors = suite.getJSONArray("vectors")
        assertTrue("runner must not silently skip vectors", vectors.length() >= 16)
        for (i in 0 until vectors.length()) {
            val v = vectors.getJSONObject(i)
            val expected = v.getJSONArray("order").let { a -> IntArray(a.length()) { a.getInt(it) } }
            val got = OptionShuffle.order(hex(v.getString("seed_hex")), v.getString("question_id"), v.getInt("n"))
            assertEquals(v.getString("name"), expected.toList(), got.toList())
        }
    }
}
