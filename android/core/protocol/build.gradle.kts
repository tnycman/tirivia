plugins {
    alias(libs.plugins.trivia.android.library)
}

android {
    namespace = "al.trivia.core.protocol"
    // Committed output of protocol/buf.gen.kotlin.yaml (docs/research/16)
    sourceSets.named("main") {
        java.directories += "src/gen/java"
        kotlin.directories += "src/gen/kotlin"
    }
}

dependencies {
    api(libs.protobuf.kotlin.lite)
    api(libs.connect.kotlin)
    testImplementation(libs.junit)
    testImplementation(libs.org.json) // runs protocol/testvectors (OptionShuffleVectorTest)
}
