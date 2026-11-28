package al.trivia.core.data

import al.trivia.core.v1.OtpChannel
import al.trivia.core.v1.Session
import com.connectrpc.Code

sealed interface StartOutcome {
    data class Sent(val otpRequestId: String, val channel: OtpChannel, val resendAfterSeconds: Int) : StartOutcome
    data object RateLimited : StartOutcome
    data object NumberNotAllowed : StartOutcome
    data object SignupsPaused : StartOutcome
    data object AttestationFailed : StartOutcome
    data class Failed(val code: Code?) : StartOutcome
}

sealed interface VerifyOutcome {
    data class SignedIn(val session: Session) : VerifyOutcome
    data object WrongCode : VerifyOutcome
    data object TooManyAttempts : VerifyOutcome
    data class Failed(val code: Code) : VerifyOutcome
}
