package al.trivia.core.security

import java.security.interfaces.ECPublicKey

/** The device's P-256 signing key; the private half never leaves the key store. */
interface DeviceKey {
    val publicKey: ECPublicKey

    /** ECDSA with SHA-256 over [data], as the 64-byte JWS R‖S value (RFC 7518 §3.4). */
    fun sign(data: ByteArray): ByteArray
}
