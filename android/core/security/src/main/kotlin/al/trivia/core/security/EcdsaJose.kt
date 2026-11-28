package al.trivia.core.security

import java.math.BigInteger

/** P-256 encodings: DER signatures (RFC 3279 §2.2.3) to JWS R‖S (RFC 7518 §3.4), full-size coordinates (RFC 7518 §6.2.1). */
object EcdsaJose {
    const val COORDINATE_BYTES = 32

    fun derToJose(der: ByteArray): ByteArray {
        require(der.size >= 8 && der[0] == SEQUENCE && (der[1].toInt() and 0xFF) == der.size - 2) { "not a DER ECDSA signature" }
        val (r, next) = integer(der, 2)
        val (s, end) = integer(der, next)
        require(end == der.size) { "trailing bytes in DER signature" }
        return unsigned(r) + unsigned(s)
    }

    /** Big-endian, left-padded to [COORDINATE_BYTES]. */
    fun unsigned(value: BigInteger): ByteArray {
        require(value.signum() >= 0 && value.bitLength() <= COORDINATE_BYTES * 8) { "value out of range" }
        val bytes = value.toByteArray().takeLast(COORDINATE_BYTES).toByteArray()
        return ByteArray(COORDINATE_BYTES - bytes.size) + bytes
    }

    private fun integer(der: ByteArray, at: Int): Pair<BigInteger, Int> {
        require(at + 2 <= der.size && der[at] == INTEGER) { "expected DER INTEGER" }
        val start = at + 2
        val end = start + (der[at + 1].toInt() and 0xFF)
        require(end > start && end <= der.size) { "bad DER INTEGER length" }
        return BigInteger(1, der.copyOfRange(start, end)) to end
    }

    private const val SEQUENCE: Byte = 0x30
    private const val INTEGER: Byte = 0x02
}
