plugins {
    alias(libs.plugins.trivia.android.library)
}

android { namespace = "al.trivia.core.network" }

dependencies {
    api(project(":core:security"))
    api(libs.connect.kotlin)
    api(libs.okhttp)
    implementation(libs.connect.kotlin.okhttp)
    implementation(libs.connect.kotlin.google.javalite.ext)
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // Android's bundled org.json is a stub in local unit tests
}
