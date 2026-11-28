package al.trivia.feature.auth.impl

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

/** Phone, then code, then signed in (PLAN Phase 1, ID-1, ID-2). */
@Composable
fun SignInRoute(viewModel: SignInViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val locale = LocalConfiguration.current.locales[0].toLanguageTag()
    Column(Modifier.fillMaxSize().padding(24.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text(stringResource(R.string.sign_in_title), style = MaterialTheme.typography.headlineSmall)
        when (state.step) {
            SignInStep.Phone -> PhoneStep(
                state = state,
                onPhoneChange = viewModel::onPhoneChange,
                onChannelChange = viewModel::onChannelChange,
                onSend = { viewModel.sendCode(locale) },
            )
            SignInStep.Code -> CodeStep(
                state = state,
                onCodeChange = viewModel::onCodeChange,
                onVerify = viewModel::verify,
                onResend = { viewModel.sendCode(locale) },
                onChangeNumber = viewModel::changeNumber,
            )
            SignInStep.Done -> Text(stringResource(R.string.sign_in_done))
        }
        state.error?.let { Text(stringResource(it.messageRes()), color = MaterialTheme.colorScheme.error) }
    }
}

private fun SignInError.messageRes(): Int = when (this) {
    SignInError.RateLimited -> R.string.error_rate_limited
    SignInError.NumberNotAllowed -> R.string.error_number_not_allowed
    SignInError.SignupsPaused -> R.string.error_signups_paused
    SignInError.AttestationFailed -> R.string.error_attestation_failed
    SignInError.WrongCode -> R.string.error_wrong_code
    SignInError.TooManyAttempts -> R.string.error_too_many_attempts
    SignInError.Unavailable -> R.string.error_unavailable
}
