plugins {
    alias(libs.plugins.tirivia.android.application)
    alias(libs.plugins.tirivia.android.compose)
    alias(libs.plugins.tirivia.hilt)
}

android {
    namespace = "com.tirivia.app"
    defaultConfig {
        applicationId = "com.tirivia.app"
        versionCode = 1
        versionName = "0.1.0"
    }
}

dependencies {
    implementation(project(":core:designsystem"))
    implementation(libs.androidx.activity.compose)
}
