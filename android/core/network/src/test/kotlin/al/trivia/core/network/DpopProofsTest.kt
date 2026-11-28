package al.trivia.core.network

import al.trivia.core.security.DeviceKey
import al.trivia.core.security.EcdsaJose
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.math.BigInteger
import java.security.AlgorithmParameters
import java.security.KeyFactory
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.PublicKey
import java.security.Signature
import java.security.interfaces.ECPublicKey
import java.security.spec.ECGenParameterSpec
import java.security.spec.ECParameterSpec
import java.security.spec.ECPoint
import java.security.spec.ECPublicKeySpec
import java.util.Base64
import java.util.UUID

/** Vectors are RFC 9449 §4.1 Figure 2 / Figure 4 and §7.1 Figures 13-14. */
class DpopProofsTest {
    private class SoftwareKey(private val pair: KeyPair) : DeviceKey {
        override val publicKey = pair.public as ECPublicKey
        override fun sign(data: ByteArray): ByteArray = EcdsaJose.derToJose(
            Signature.getInstance("SHA256withECDSA").run {
                initSign(pair.private)
                update(data)
                sign()
            },
        )
    }

    private val key = SoftwareKey(
        KeyPairGenerator.getInstance("EC").apply { initialize(ECGenParameterSpec("secp256r1")) }.generateKeyPair(),
    )
    private val proofs = DpopProofs(key) { 1_700_000_000 }

    @Test
    fun rfcExampleVerifiesWithThisHarness() {
        assertTrue(verifies(RFC_PROOF, publicKey(RFC_X, RFC_Y)))
    }

    @Test
    fun headerCarriesOnlyThePublicJwk() {
        val rfcPublic = publicKey(RFC_X, RFC_Y)
        val rfcKey = object : DeviceKey {
            override val publicKey: ECPublicKey = rfcPublic
            override fun sign(data: ByteArray) = ByteArray(64)
        }
        val header = part(DpopProofs(rfcKey).create("POST", URL), 0)
        assertEquals("dpop+jwt", header.getString("typ"))
        assertEquals("ES256", header.getString("alg"))
        val jwk = header.getJSONObject("jwk")
        assertEquals(setOf("kty", "crv", "x", "y"), jwk.keySet())
        assertEquals("EC", jwk.getString("kty"))
        assertEquals("P-256", jwk.getString("crv"))
        assertEquals(RFC_X, jwk.getString("x"))
        assertEquals(RFC_Y, jwk.getString("y"))
    }

    @Test
    fun claimsAndSignatureFollowRfc9449() {
        val jwt = proofs.create("POST", "$URL?q=1#frag", accessToken = RFC_TOKEN, nonce = "n-1")
        val claims = part(jwt, 1)
        assertEquals("POST", claims.getString("htm"))
        assertEquals(URL, claims.getString("htu"))
        assertEquals(1_700_000_000L, claims.getLong("iat"))
        assertEquals(RFC_ATH, claims.getString("ath"))
        assertEquals("n-1", claims.getString("nonce"))
        assertEquals(4, UUID.fromString(claims.getString("jti")).version())
        assertTrue(verifies(jwt, key.publicKey))
    }

    @Test
    fun optionalClaimsOmittedAndJtiUnique() {
        val first = part(proofs.create("GET", URL), 1)
        val second = part(proofs.create("GET", URL), 1)
        assertFalse(first.has("ath"))
        assertFalse(first.has("nonce"))
        assertNotEquals(first.getString("jti"), second.getString("jti"))
    }

    private fun part(jwt: String, index: Int) = JSONObject(String(decode(jwt.split('.')[index]), Charsets.UTF_8))

    private fun verifies(jwt: String, publicKey: PublicKey): Boolean {
        val (header, payload, signature) = jwt.split('.')
        return Signature.getInstance("SHA256withECDSAinP1363Format").run {
            initVerify(publicKey)
            update("$header.$payload".toByteArray(Charsets.US_ASCII))
            verify(decode(signature))
        }
    }

    private fun publicKey(x: String, y: String): ECPublicKey {
        val params = AlgorithmParameters.getInstance("EC")
            .apply { init(ECGenParameterSpec("secp256r1")) }
            .getParameterSpec(ECParameterSpec::class.java)
        val point = ECPoint(BigInteger(1, decode(x)), BigInteger(1, decode(y)))
        return KeyFactory.getInstance("EC").generatePublic(ECPublicKeySpec(point, params)) as ECPublicKey
    }

    private fun decode(value: String) = Base64.getUrlDecoder().decode(value)

    private companion object {
        const val URL = "https://server.example.com/token"
        const val RFC_X = "l8tFrhx-34tV3hRICRDY9zCkDlpBhF42UQUfWVAWBFs"
        const val RFC_Y = "9VE4jf_Ok_o64zbTTlcuNJajHmt6v9TDVrU0CdvGRDA"
        const val RFC_TOKEN = "Kz~8mXK1EalYznwH-LC-1fBAo.4Ljp~zsPE_NeO.gxU"
        const val RFC_ATH = "fUHyO2r2Z3DZ53EsNrWBb0xWXoaNy59IiKCAqksmQEo"
        const val RFC_PROOF =
            "eyJ0eXAiOiJkcG9wK2p3dCIsImFsZyI6IkVTMjU2IiwiandrIjp7Imt0eSI6IkVDIiwieCI6Imw4dEZyaHgtMzR0VjNoUklDUkRZOXpDa0RscEJoRjQyVVFVZldWQVdCRnMiLCJ5IjoiOVZFNGpmX09rX282NHpiVFRsY3VOSmFqSG10NnY5VERWclUwQ2R2R1JEQSIsImNydiI6IlAtMjU2In19" +
                ".eyJqdGkiOiItQndDM0VTYzZhY2MybFRjIiwiaHRtIjoiUE9TVCIsImh0dSI6Imh0dHBzOi8vc2VydmVyLmV4YW1wbGUuY29tL3Rva2VuIiwiaWF0IjoxNTYyMjYyNjE2fQ" +
                ".2-GxA6T8lP4vfrg8v-FdWP0A0zdrj8igiMLvqRMUvwnQg4PtFLbdLXiOSsX0x7NVY-FNyJK70nfbV37xRZT3Lg"
    }
}
