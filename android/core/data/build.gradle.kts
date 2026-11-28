plugins {
    alias(libs.plugins.trivia.android.library)
}

android { namespace = "al.trivia.core.data" }

dependencies {
    api(project(":core:protocol"))
    api(project(":core:attestation"))
    testImplementation(project(":core:network"))
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // Android's bundled org.json is a stub in local unit tests
    testImplementation(libs.okhttp.mockwebserver)
}
