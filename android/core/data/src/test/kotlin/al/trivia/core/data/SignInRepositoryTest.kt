package al.trivia.core.data

import al.trivia.core.attestation.IntegrityTokens
import al.trivia.core.network.DpopProofs
import al.trivia.core.network.connectClient
import al.trivia.core.security.DeviceKey
import al.trivia.core.security.EcdsaJose
import al.trivia.core.v1.AttestRequest
import al.trivia.core.v1.AuthServiceClient
import al.trivia.core.v1.OtpChannel
import al.trivia.core.v1.Platform
import al.trivia.core.v1.StartOtpRequest
import al.trivia.core.v1.StartOtpResult
import al.trivia.core.v1.attestResponse
import al.trivia.core.v1.getAttestationChallengeResponse
import al.trivia.core.v1.session
import al.trivia.core.v1.startOtpResponse
import al.trivia.core.v1.verifyOtpResponse
import com.connectrpc.Code
import com.google.protobuf.ByteString
import com.google.protobuf.MessageLite
import kotlinx.coroutines.runBlocking
import mockwebserver3.MockResponse
import mockwebserver3.MockWebServer
import mockwebserver3.RecordedRequest
import okhttp3.OkHttpClient
import okio.Buffer
import org.json.JSONObject
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.security.KeyPairGenerator
import java.security.Signature
import java.security.interfaces.ECPublicKey
import java.security.spec.ECGenParameterSpec
import java.util.Base64

/** Real Connect client and DPoP interceptor against a local HTTP server speaking the Connect protocol. */
class SignInRepositoryTest {
    private val server = MockWebServer().apply { start() }
    private val keyPair = KeyPairGenerator.getInstance("EC").apply { initialize(ECGenParameterSpec("secp256r1")) }.generateKeyPair()
    private val deviceKey = object : DeviceKey {
        override val publicKey = keyPair.public as ECPublicKey
        override fun sign(data: ByteArray): ByteArray = EcdsaJose.derToJose(
            Signature.getInstance("SHA256withECDSA").run {
                initSign(keyPair.private)
                update(data)
                sign()
            },
        )
    }
    private val integrity = object : IntegrityTokens {
        override suspend fun token(challenge: ByteArray) = "integrity-" + challenge.decodeToString()
    }
    private val repository = SignInRepository(
        AuthServiceClient(connectClient(server.url("/").toString(), OkHttpClient(), DpopProofs(deviceKey))),
        integrity,
    )

    @After
    fun close() = server.close()

    @Test
    fun startAttestsThenSendsCode() = runBlocking {
        reply(getAttestationChallengeResponse { challenge = ByteString.copyFromUtf8("c1") })
        reply(attestResponse { attestationTicket = "t1" })
        reply(
            startOtpResponse {
                result = StartOtpResult.START_OTP_RESULT_SENT
                otpRequestId = "r1"
                channelUsed = OtpChannel.OTP_CHANNEL_SMS
                resendAfterSeconds = 30
            },
        )

        val outcome = repository.start("+355691234567", OtpChannel.OTP_CHANNEL_SMS, "sq-AL")

        assertEquals(StartOutcome.Sent("r1", OtpChannel.OTP_CHANNEL_SMS, 30), outcome)
        assertEquals("/trivia.core.v1.AuthService/GetAttestationChallenge", server.takeRequest().target)
        val attest = server.takeRequest()
        assertEquals("/trivia.core.v1.AuthService/Attest", attest.target)
        val attestBody = AttestRequest.parseFrom(attest.body!!.toByteArray())
        assertEquals(Platform.PLATFORM_ANDROID, attestBody.platform)
        assertEquals("c1", attestBody.challenge.toStringUtf8())
        assertEquals("integrity-c1", attestBody.playIntegrityToken)
        assertProofMatches(attest)
        val start = StartOtpRequest.parseFrom(server.takeRequest().body!!.toByteArray())
        assertEquals("t1", start.attestationTicket)
        assertEquals("+355691234567", start.phoneE164)
        assertEquals("sq-AL", start.locale)
    }

    @Test
    fun startReportsServerResults() = runBlocking {
        reply(getAttestationChallengeResponse { challenge = ByteString.copyFromUtf8("c1") })
        reply(attestResponse { attestationTicket = "t1" })
        reply(startOtpResponse { result = StartOtpResult.START_OTP_RESULT_RATE_LIMITED })
        assertEquals(StartOutcome.RateLimited, repository.start("+355691234567", OtpChannel.OTP_CHANNEL_SMS, "sq-AL"))
    }

    @Test
    fun refusedAttestationIsReported() = runBlocking {
        reply(getAttestationChallengeResponse { challenge = ByteString.copyFromUtf8("c1") })
        replyError(403, "permission_denied")
        assertEquals(StartOutcome.AttestationFailed, repository.start("+355691234567", OtpChannel.OTP_CHANNEL_SMS, "sq-AL"))
    }

    @Test
    fun verifyReturnsTheSessionAndSignsTheCall() = runBlocking {
        val issued = session {
            accessToken = "a"
            refreshToken = "r"
            userId = "u"
        }
        reply(verifyOtpResponse { session = issued })
        val outcome = repository.verify("r1", "123456")
        assertEquals("r", (outcome as VerifyOutcome.SignedIn).session.refreshToken)
        val request = server.takeRequest()
        assertEquals("/trivia.core.v1.AuthService/VerifyOtp", request.target)
        assertProofMatches(request)
    }

    @Test
    fun verifyMapsServerCodes() = runBlocking {
        replyError(401, "unauthenticated")
        assertEquals(VerifyOutcome.WrongCode, repository.verify("r1", "000000"))
        replyError(429, "resource_exhausted")
        assertEquals(VerifyOutcome.TooManyAttempts, repository.verify("r1", "000000"))
        replyError(503, "unavailable")
        assertEquals(VerifyOutcome.Failed(Code.UNAVAILABLE), repository.verify("r1", "000000"))
    }

    private fun reply(message: MessageLite) = server.enqueue(
        MockResponse.Builder()
            .addHeader("Content-Type", "application/proto")
            .body(Buffer().write(message.toByteArray()))
            .build(),
    )

    private fun replyError(status: Int, code: String) = server.enqueue(
        MockResponse.Builder()
            .code(status)
            .addHeader("Content-Type", "application/json")
            .body(JSONObject().put("code", code).put("message", code).toString())
            .build(),
    )

    private fun assertProofMatches(request: RecordedRequest) {
        val (header, payload, signature) = request.headers["DPoP"]!!.split('.')
        val claims = JSONObject(String(Base64.getUrlDecoder().decode(payload), Charsets.UTF_8))
        assertEquals("POST", claims.getString("htm"))
        assertEquals(server.url(request.target).toString(), claims.getString("htu"))
        val verified = Signature.getInstance("SHA256withECDSAinP1363Format").run {
            initVerify(keyPair.public)
            update("$header.$payload".toByteArray(Charsets.US_ASCII))
            verify(Base64.getUrlDecoder().decode(signature))
        }
        assertTrue(verified)
    }
}
