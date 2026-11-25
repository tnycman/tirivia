plugins {
    alias(libs.plugins.android.test)
}

// Macrobenchmark FrameTimingMetric gate: reaction layer >= 55 fps on the reference phone.
android { namespace = "com.tirivia.benchmarks.macro" }
