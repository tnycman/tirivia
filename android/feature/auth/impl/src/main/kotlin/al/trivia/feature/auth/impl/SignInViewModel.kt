package al.trivia.feature.auth.impl

import al.trivia.core.data.SignInRepository
import al.trivia.core.data.StartOutcome
import al.trivia.core.data.VerifyOutcome
import al.trivia.core.datastore.Sessions
import al.trivia.core.datastore.StoredSession
import al.trivia.core.v1.OtpChannel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SignInViewModel @Inject constructor(
    private val signIn: SignInRepository,
    private val sessions: Sessions,
) : ViewModel() {
    private val mutableState = MutableStateFlow(SignInState())
    val state: StateFlow<SignInState> = mutableState.asStateFlow()
    private var resendTimer: Job? = null

    fun onPhoneChange(value: String) = mutableState.update { it.copy(phone = value, error = null) }

    fun onChannelChange(channel: OtpChannel) = mutableState.update { it.copy(channel = channel) }

    fun onCodeChange(value: String) =
        mutableState.update { it.copy(code = value.filter(Char::isDigit).take(CODE_LENGTH), error = null) }

    fun changeNumber() = mutableState.update { it.copy(step = SignInStep.Phone, code = "", error = null) }

    fun sendCode(locale: String) = launchBusy {
        val current = mutableState.value
        when (val outcome = signIn.start(normalizePhone(current.phone), current.channel, locale)) {
            is StartOutcome.Sent -> {
                mutableState.update { it.copy(step = SignInStep.Code, otpRequestId = outcome.otpRequestId, code = "") }
                startResendTimer(outcome.resendAfterSeconds)
            }
            StartOutcome.RateLimited -> fail(SignInError.RateLimited)
            StartOutcome.NumberNotAllowed -> fail(SignInError.NumberNotAllowed)
            StartOutcome.SignupsPaused -> fail(SignInError.SignupsPaused)
            StartOutcome.AttestationFailed -> fail(SignInError.AttestationFailed)
            is StartOutcome.Failed -> fail(SignInError.Unavailable)
        }
    }

    fun verify() = launchBusy {
        val current = mutableState.value
        when (val outcome = signIn.verify(current.otpRequestId, current.code)) {
            is VerifyOutcome.SignedIn -> {
                sessions.save(StoredSession(outcome.session.userId, outcome.session.refreshToken))
                mutableState.update { it.copy(step = SignInStep.Done) }
            }
            VerifyOutcome.WrongCode -> fail(SignInError.WrongCode)
            VerifyOutcome.TooManyAttempts -> fail(SignInError.TooManyAttempts)
            is VerifyOutcome.Failed -> fail(SignInError.Unavailable)
        }
    }

    private fun fail(error: SignInError) = mutableState.update { it.copy(error = error) }

    private fun startResendTimer(seconds: Int) {
        resendTimer?.cancel()
        resendTimer = viewModelScope.launch {
            for (left in seconds downTo 0) {
                mutableState.update { it.copy(resendInSeconds = left) }
                if (left > 0) delay(1_000)
            }
        }
    }

    private fun launchBusy(block: suspend () -> Unit) {
        if (mutableState.value.busy) return
        mutableState.update { it.copy(busy = true, error = null) }
        viewModelScope.launch {
            try {
                block()
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                fail(SignInError.Unavailable)
            } finally {
                mutableState.update { it.copy(busy = false) }
            }
        }
    }
}
