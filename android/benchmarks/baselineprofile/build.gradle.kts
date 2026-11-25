plugins {
    alias(libs.plugins.android.test)
}

// Generates Baseline + Startup Profiles: launch -> join -> live screen with reactions.
android { namespace = "com.tirivia.benchmarks.baselineprofile" }
