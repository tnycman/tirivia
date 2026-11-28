package al.trivia.feature.auth.impl

import al.trivia.core.v1.OtpChannel
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.input.KeyboardType

/** ID-1: the user picks the channel; the button order is set after the price/delivery test (PLAN L9). */
private val channels = listOf(
    OtpChannel.OTP_CHANNEL_SMS to R.string.channel_sms,
    OtpChannel.OTP_CHANNEL_WHATSAPP to R.string.channel_whatsapp,
    OtpChannel.OTP_CHANNEL_VIBER to R.string.channel_viber,
)

@Composable
internal fun PhoneStep(
    state: SignInState,
    onPhoneChange: (String) -> Unit,
    onChannelChange: (OtpChannel) -> Unit,
    onSend: () -> Unit,
) {
    OutlinedTextField(
        value = state.phone,
        onValueChange = onPhoneChange,
        label = { Text(stringResource(R.string.phone_label)) },
        singleLine = true,
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
        modifier = Modifier.fillMaxWidth(),
    )
    Text(stringResource(R.string.channel_label))
    channels.forEach { (channel, label) ->
        Row(
            Modifier.fillMaxWidth().selectable(
                selected = state.channel == channel,
                role = Role.RadioButton,
                onClick = { onChannelChange(channel) },
            ),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            RadioButton(selected = state.channel == channel, onClick = null)
            Text(stringResource(label))
        }
    }
    Button(onClick = onSend, enabled = !state.busy && state.phone.isNotBlank(), modifier = Modifier.fillMaxWidth()) {
        Text(stringResource(R.string.send_code))
    }
}
