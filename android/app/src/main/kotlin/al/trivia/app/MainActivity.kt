package al.trivia.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Text
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import al.trivia.core.designsystem.generated.TokensLight

// Shell screen until the sign-in flow lands (PLAN Phase 1).
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            Box(Modifier.fillMaxSize().background(TokensLight.themeBg), contentAlignment = Alignment.Center) {
                Text(stringResource(R.string.app_name), color = TokensLight.themePrimaryText)
            }
        }
    }
}
