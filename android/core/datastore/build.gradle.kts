plugins {
    alias(libs.plugins.trivia.android.library)
}

android { namespace = "al.trivia.core.datastore" }

dependencies {
    api(libs.androidx.datastore)
    implementation(libs.androidx.datastore.tink)
    implementation(libs.tink.android)
    testImplementation(libs.junit)
}
