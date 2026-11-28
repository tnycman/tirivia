package al.trivia.feature.auth.impl

import al.trivia.core.v1.OtpChannel

enum class SignInStep { Phone, Code, Done }

enum class SignInError { RateLimited, NumberNotAllowed, SignupsPaused, AttestationFailed, WrongCode, TooManyAttempts, Unavailable }

data class SignInState(
    val step: SignInStep = SignInStep.Phone,
    val phone: String = "+355",
    val channel: OtpChannel = OtpChannel.OTP_CHANNEL_SMS,
    val code: String = "",
    val otpRequestId: String = "",
    val resendInSeconds: Int = 0,
    val busy: Boolean = false,
    val error: SignInError? = null,
)

/** Keeps "+" and digits only; the server validates the number (libphonenumber, E.164). */
internal fun normalizePhone(input: String): String = input.filter { it.isDigit() || it == '+' }

internal const val CODE_LENGTH = 6 // ID-2
