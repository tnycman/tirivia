plugins {
    alias(libs.plugins.trivia.android.library)
}

android { namespace = "al.trivia.core.protocol" }

dependencies {
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // runs protocol/testvectors (OptionShuffleVectorTest)
}
