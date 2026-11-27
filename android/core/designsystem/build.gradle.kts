plugins {
    alias(libs.plugins.tirivia.android.library)
    alias(libs.plugins.tirivia.android.compose) // generated tokens use Color, dp, sp
}

android { namespace = "com.tirivia.core.designsystem" }
