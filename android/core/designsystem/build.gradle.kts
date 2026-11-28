plugins {
    alias(libs.plugins.trivia.android.library)
    alias(libs.plugins.trivia.android.compose) // generated tokens use Color, dp, sp
}

android { namespace = "al.trivia.core.designsystem" }
