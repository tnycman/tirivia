package al.trivia.feature.auth.impl

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType

@Composable
internal fun CodeStep(
    state: SignInState,
    onCodeChange: (String) -> Unit,
    onVerify: () -> Unit,
    onResend: () -> Unit,
    onChangeNumber: () -> Unit,
) {
    Text(stringResource(R.string.code_sent_to, normalizePhone(state.phone)))
    OutlinedTextField(
        value = state.code,
        onValueChange = onCodeChange,
        label = { Text(stringResource(R.string.code_label)) },
        singleLine = true,
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
        modifier = Modifier.fillMaxWidth(),
    )
    Button(onClick = onVerify, enabled = !state.busy && state.code.length == CODE_LENGTH, modifier = Modifier.fillMaxWidth()) {
        Text(stringResource(R.string.verify))
    }
    TextButton(onClick = onResend, enabled = !state.busy && state.resendInSeconds == 0) {
        Text(
            if (state.resendInSeconds > 0) {
                stringResource(R.string.resend_in, state.resendInSeconds)
            } else {
                stringResource(R.string.resend)
            },
        )
    }
    TextButton(onClick = onChangeNumber, enabled = !state.busy) {
        Text(stringResource(R.string.change_number))
    }
}
