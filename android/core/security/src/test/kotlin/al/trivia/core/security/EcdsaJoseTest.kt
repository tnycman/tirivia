package al.trivia.core.security

import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Test
import java.math.BigInteger
import java.security.KeyPairGenerator
import java.security.Signature
import java.security.spec.ECGenParameterSpec

/** The JDK's P1363 verifier is an independent check of the R‖S form (RFC 7518 §3.4). */
class EcdsaJoseTest {
    private val keyPair = KeyPairGenerator.getInstance("EC")
        .apply { initialize(ECGenParameterSpec("secp256r1")) }
        .generateKeyPair()

    @Test
    fun convertedSignaturesVerifyAsJws() {
        repeat(256) { i ->
            val message = "message $i".toByteArray()
            val der = Signature.getInstance("SHA256withECDSA").run {
                initSign(keyPair.private)
                update(message)
                sign()
            }
            val jose = EcdsaJose.derToJose(der)
            assertEquals(64, jose.size)
            val verified = Signature.getInstance("SHA256withECDSAinP1363Format").run {
                initVerify(keyPair.public)
                update(message)
                verify(jose)
            }
            assertTrue("signature $i", verified)
        }
    }

    @Test
    fun stripsSignByteAndPadsShortValues() {
        val r = ByteArray(32) { 0x80.toByte() }
        val s = ByteArray(31) { 0x01 }
        val der = byteArrayOf(0x30, 68, 0x02, 33, 0) + r + byteArrayOf(0x02, 31) + s
        assertArrayEquals(r + byteArrayOf(0) + s, EcdsaJose.derToJose(der))
    }

    @Test
    fun rejectsMalformedDer() {
        EcdsaJose.derToJose(byteArrayOf(0x30, 6, 0x02, 1, 1, 0x02, 1, 1))
        listOf(
            byteArrayOf(0x31, 6, 0x02, 1, 1, 0x02, 1, 1),
            byteArrayOf(0x30, 7, 0x02, 1, 1, 0x02, 1, 1, 0),
            byteArrayOf(0x30, 6, 0x02, 1, 1, 0x03, 1, 1),
            byteArrayOf(0x30, 6, 0x02, 9, 1, 0x02, 1, 1),
        ).forEach { bad ->
            assertThrows(IllegalArgumentException::class.java) { EcdsaJose.derToJose(bad) }
        }
    }

    @Test
    fun coordinatesAreFullSize() {
        assertArrayEquals(ByteArray(31) + byteArrayOf(1), EcdsaJose.unsigned(BigInteger.ONE))
        val max = BigInteger.ONE.shiftLeft(256).subtract(BigInteger.ONE)
        assertArrayEquals(ByteArray(32) { -1 }, EcdsaJose.unsigned(max))
        assertThrows(IllegalArgumentException::class.java) { EcdsaJose.unsigned(BigInteger.ONE.shiftLeft(256)) }
    }
}
