package al.trivia.core.network

import al.trivia.core.security.DeviceKey
import al.trivia.core.security.EcdsaJose
import org.json.JSONObject
import java.security.MessageDigest
import java.util.Base64
import java.util.UUID

/** Builds DPoP proof JWTs (RFC 9449 §4.2) signed with the device key. */
class DpopProofs(
    private val key: DeviceKey,
    private val nowSeconds: () -> Long = { System.currentTimeMillis() / 1000 },
) {
    private val header by lazy { encode(headerJson().toString()) }

    fun create(method: String, url: String, accessToken: String? = null, nonce: String? = null): String {
        val claims = JSONObject()
            .put("jti", UUID.randomUUID().toString())
            .put("htm", method)
            .put("htu", url.substringBefore('#').substringBefore('?'))
            .put("iat", nowSeconds())
        accessToken?.let { claims.put("ath", b64(sha256(it))) }
        nonce?.let { claims.put("nonce", it) }
        val signingInput = header + "." + encode(claims.toString())
        return signingInput + "." + b64(key.sign(signingInput.toByteArray(Charsets.US_ASCII)))
    }

    private fun headerJson(): JSONObject {
        val point = key.publicKey.w
        val jwk = JSONObject()
            .put("kty", "EC")
            .put("crv", "P-256")
            .put("x", b64(EcdsaJose.unsigned(point.affineX)))
            .put("y", b64(EcdsaJose.unsigned(point.affineY)))
        return JSONObject().put("typ", "dpop+jwt").put("alg", "ES256").put("jwk", jwk)
    }

    private fun sha256(value: String) = MessageDigest.getInstance("SHA-256").digest(value.toByteArray(Charsets.US_ASCII))

    private fun encode(json: String) = b64(json.toByteArray(Charsets.UTF_8))

    private fun b64(bytes: ByteArray) = Base64.getUrlEncoder().withoutPadding().encodeToString(bytes)
}
