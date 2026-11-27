plugins {
    alias(libs.plugins.tirivia.android.library)
}

android { namespace = "com.tirivia.core.protocol" }

dependencies {
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // runs protocol/testvectors (OptionShuffleVectorTest)
}
