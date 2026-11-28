plugins {
    alias(libs.plugins.trivia.android.feature)
}

android { namespace = "al.trivia.feature.auth.impl" }

dependencies {
    implementation(project(":core:data"))
    implementation(project(":core:datastore"))
    implementation(libs.androidx.hilt.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)
    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
}
