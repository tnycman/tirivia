package al.trivia.core.protocol

import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

/**
 * Per-player option order (SPEC §6.3 step 7). Must match the Go reference
 * (backend/internal/platform/shuffle) bit for bit; protocol/testvectors/answer pins it.
 *
 *     k = HMAC-SHA256(key = seed, msg = questionId)
 *     for i = n-1 downTo 1: j = k[n-1-i] mod (i+1); swap(idx[i], idx[j])
 */
object OptionShuffle {
    const val MAX_OPTIONS = 33

    /** Returns a permutation of 0 until n: display position p shows option `order[p]`. */
    fun order(seed: ByteArray, questionId: String, n: Int): IntArray {
        require(n in 1..MAX_OPTIONS) { "option count out of range" }
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(SecretKeySpec(seed, "HmacSHA256"))
        val k = mac.doFinal(questionId.toByteArray(Charsets.UTF_8))
        val idx = IntArray(n) { it }
        for (i in n - 1 downTo 1) {
            // Bytes are signed in Kotlin: mask to 0..255 to match Go/Swift.
            val j = (k[n - 1 - i].toInt() and 0xFF) % (i + 1)
            val tmp = idx[i]
            idx[i] = idx[j]
            idx[j] = tmp
        }
        return idx
    }
}
