package al.trivia.feature.auth.impl

import al.trivia.core.attestation.IntegrityTokens
import al.trivia.core.data.SignInRepository
import al.trivia.core.datastore.Sessions
import al.trivia.core.datastore.StoredSession
import al.trivia.core.v1.AttestRequest
import al.trivia.core.v1.AttestResponse
import al.trivia.core.v1.AuthServiceClientInterface
import al.trivia.core.v1.GetAttestationChallengeRequest
import al.trivia.core.v1.GetAttestationChallengeResponse
import al.trivia.core.v1.RefreshSessionRequest
import al.trivia.core.v1.RefreshSessionResponse
import al.trivia.core.v1.SignOutRequest
import al.trivia.core.v1.SignOutResponse
import al.trivia.core.v1.StartOtpRequest
import al.trivia.core.v1.StartOtpResponse
import al.trivia.core.v1.StartOtpResult
import al.trivia.core.v1.VerifyOtpRequest
import al.trivia.core.v1.VerifyOtpResponse
import al.trivia.core.v1.attestResponse
import al.trivia.core.v1.getAttestationChallengeResponse
import al.trivia.core.v1.session
import al.trivia.core.v1.startOtpResponse
import al.trivia.core.v1.verifyOtpResponse
import com.connectrpc.Code
import com.connectrpc.ConnectException
import com.connectrpc.Headers
import com.connectrpc.ResponseMessage
import com.google.protobuf.ByteString
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Before
import org.junit.Test

class SignInViewModelTest {
    private val auth = FakeAuth()
    private val sessions = FakeSessions()
    private var integrityFails = false
    private val integrity = object : IntegrityTokens {
        override suspend fun token(challenge: ByteArray): String {
            check(!integrityFails) { "integrity unavailable" }
            return "token"
        }
    }
    private lateinit var viewModel: SignInViewModel

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        viewModel = SignInViewModel(SignInRepository(auth, integrity), sessions)
    }

    @After
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun sendCodeNormalizesThePhoneAndMovesToTheCodeStep() {
        viewModel.onPhoneChange("+355 69 123-4567")
        viewModel.sendCode("sq-AL")
        assertEquals("+355691234567", auth.lastStart?.phoneE164)
        assertEquals("sq-AL", auth.lastStart?.locale)
        val state = viewModel.state.value
        assertEquals(SignInStep.Code, state.step)
        assertEquals("r1", state.otpRequestId)
        assertEquals(30, state.resendInSeconds)
        assertFalse(state.busy)
    }

    @Test
    fun startResultsBecomeErrors() {
        auth.startResult = StartOtpResult.START_OTP_RESULT_NUMBER_NOT_ALLOWED
        viewModel.sendCode("sq-AL")
        assertEquals(SignInStep.Phone, viewModel.state.value.step)
        assertEquals(SignInError.NumberNotAllowed, viewModel.state.value.error)
    }

    @Test
    fun verifySavesOnlyTheRefreshTokenAndFinishes() {
        viewModel.sendCode("sq-AL")
        viewModel.onCodeChange("12a3456789")
        assertEquals("123456", viewModel.state.value.code)
        viewModel.verify()
        assertEquals(StoredSession("u1", "refresh-1"), sessions.saved)
        assertEquals(SignInStep.Done, viewModel.state.value.step)
    }

    @Test
    fun wrongCodeStaysOnTheCodeStep() {
        viewModel.sendCode("sq-AL")
        auth.verifyError = Code.UNAUTHENTICATED
        viewModel.onCodeChange("000000")
        viewModel.verify()
        assertEquals(SignInStep.Code, viewModel.state.value.step)
        assertEquals(SignInError.WrongCode, viewModel.state.value.error)
        assertEquals(null, sessions.saved)
    }

    @Test
    fun unexpectedFailureIsShownAndReleasesBusy() {
        integrityFails = true
        viewModel.sendCode("sq-AL")
        assertEquals(SignInError.Unavailable, viewModel.state.value.error)
        assertFalse(viewModel.state.value.busy)
    }

    private class FakeSessions : Sessions {
        var saved: StoredSession? = null
        override val session: Flow<StoredSession> = MutableStateFlow(StoredSession.SIGNED_OUT)
        override suspend fun save(session: StoredSession) {
            saved = session
        }
        override suspend fun clear() {
            saved = StoredSession.SIGNED_OUT
        }
    }

    private class FakeAuth : AuthServiceClientInterface {
        var lastStart: StartOtpRequest? = null
        var startResult = StartOtpResult.START_OTP_RESULT_SENT
        var verifyError: Code? = null

        override suspend fun getAttestationChallenge(request: GetAttestationChallengeRequest, headers: Headers) =
            ok(getAttestationChallengeResponse { challenge = ByteString.copyFromUtf8("c") })

        override suspend fun attest(request: AttestRequest, headers: Headers) = ok(attestResponse { attestationTicket = "t" })

        override suspend fun startOtp(request: StartOtpRequest, headers: Headers): ResponseMessage<StartOtpResponse> {
            lastStart = request
            return ok(
                startOtpResponse {
                    result = startResult
                    otpRequestId = "r1"
                    resendAfterSeconds = 30
                },
            )
        }

        override suspend fun verifyOtp(request: VerifyOtpRequest, headers: Headers): ResponseMessage<VerifyOtpResponse> {
            verifyError?.let { return ResponseMessage.Failure(ConnectException(it), emptyMap(), emptyMap()) }
            val issued = session {
                accessToken = "access-1"
                refreshToken = "refresh-1"
                userId = "u1"
            }
            return ok(verifyOtpResponse { session = issued })
        }

        override suspend fun refreshSession(request: RefreshSessionRequest, headers: Headers): ResponseMessage<RefreshSessionResponse> =
            error("not used")

        override suspend fun signOut(request: SignOutRequest, headers: Headers): ResponseMessage<SignOutResponse> = error("not used")

        private fun <T> ok(message: T): ResponseMessage<T> = ResponseMessage.Success(message, emptyMap(), emptyMap())
    }
}
