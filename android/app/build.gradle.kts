plugins {
    alias(libs.plugins.trivia.android.application)
    alias(libs.plugins.trivia.android.compose)
    alias(libs.plugins.trivia.hilt)
}

android {
    namespace = "al.trivia.app"
    defaultConfig {
        applicationId = "al.trivia.app"
        versionCode = 1
        versionName = "0.1.0"
    }
}

dependencies {
    implementation(project(":core:designsystem"))
    implementation(project(":core:data"))
    implementation(project(":core:datastore"))
    implementation(project(":core:network"))
    implementation(project(":feature:auth:impl"))
    implementation(libs.androidx.activity.compose)
}
