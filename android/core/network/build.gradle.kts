plugins {
    alias(libs.plugins.trivia.android.library)
}

android { namespace = "al.trivia.core.network" }

dependencies {
    api(project(":core:security"))
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // Android's bundled org.json is a stub in local unit tests
}
