package al.trivia.core.attestation

/** Produces the Play Integrity token sent with `Attest` for a server challenge. */
interface IntegrityTokens {
    suspend fun token(challenge: ByteArray): String
}
