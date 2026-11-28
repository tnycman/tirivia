package al.trivia.core.data

import al.trivia.core.attestation.IntegrityTokens
import al.trivia.core.v1.AuthServiceClientInterface
import al.trivia.core.v1.OtpChannel
import al.trivia.core.v1.Platform
import al.trivia.core.v1.StartOtpResult
import al.trivia.core.v1.attestRequest
import al.trivia.core.v1.getAttestationChallengeRequest
import al.trivia.core.v1.startOtpRequest
import al.trivia.core.v1.verifyOtpRequest
import com.connectrpc.Code
import com.connectrpc.ConnectException
import com.connectrpc.getOrThrow

/** Sign-in against `AuthService`: challenge, attest, send code, verify (backend/internal/auth/api, research 17). */
class SignInRepository(
    private val auth: AuthServiceClientInterface,
    private val integrity: IntegrityTokens,
) {
    suspend fun start(phoneE164: String, channel: OtpChannel, locale: String): StartOutcome = try {
        val res = auth.startOtp(
            startOtpRequest {
                this.phoneE164 = phoneE164
                attestationTicket = attest()
                preferredChannel = channel
                this.locale = locale
            },
        ).getOrThrow()
        when (res.result) {
            StartOtpResult.START_OTP_RESULT_SENT -> StartOutcome.Sent(res.otpRequestId, res.channelUsed, res.resendAfterSeconds)
            StartOtpResult.START_OTP_RESULT_RATE_LIMITED -> StartOutcome.RateLimited
            StartOtpResult.START_OTP_RESULT_NUMBER_NOT_ALLOWED -> StartOutcome.NumberNotAllowed
            StartOtpResult.START_OTP_RESULT_SIGNUPS_PAUSED -> StartOutcome.SignupsPaused
            StartOtpResult.START_OTP_RESULT_ATTESTATION_REQUIRED -> StartOutcome.AttestationFailed
            else -> StartOutcome.Failed(null)
        }
    } catch (e: ConnectException) {
        if (e.code == Code.PERMISSION_DENIED) StartOutcome.AttestationFailed else StartOutcome.Failed(e.code)
    }

    suspend fun verify(otpRequestId: String, code: String): VerifyOutcome = try {
        val res = auth.verifyOtp(
            verifyOtpRequest {
                this.otpRequestId = otpRequestId
                this.code = code
            },
        ).getOrThrow()
        VerifyOutcome.SignedIn(res.session)
    } catch (e: ConnectException) {
        when (e.code) {
            Code.UNAUTHENTICATED -> VerifyOutcome.WrongCode
            Code.RESOURCE_EXHAUSTED -> VerifyOutcome.TooManyAttempts
            else -> VerifyOutcome.Failed(e.code)
        }
    }

    private suspend fun attest(): String {
        val challenge = auth.getAttestationChallenge(
            getAttestationChallengeRequest { platform = Platform.PLATFORM_ANDROID },
        ).getOrThrow().challenge
        return auth.attest(
            attestRequest {
                platform = Platform.PLATFORM_ANDROID
                this.challenge = challenge
                playIntegrityToken = integrity.token(challenge.toByteArray())
            },
        ).getOrThrow().attestationTicket
    }
}
