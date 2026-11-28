pluginManagement {
    includeBuild("build-logic")
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral() // IVS Player is on Maven Central
    }
}

rootProject.name = "tirivia"

include(":app")

// Core modules (Now in Android pattern; stack/07-android)
include(
    ":core:model",
    ":core:common",
    ":core:protocol",      // generated connect-kotlin + javalite; shuffle + cue verification
    ":core:network",       // OkHttp, DPoP proofs, retries, pinning (own API hosts only)
    ":core:attestation",   // Play Integrity standard requests
    ":core:security",      // Keystore keys, capture/overlay/accessibility signals
    ":core:video",         // IVS player wrapper, cue callbacks
    ":core:analytics",     // Sentry, events
    ":core:data",
    ":core:datastore",
    ":core:designsystem",
    ":core:ui",
    ":core:testing",
)

// Feature modules: each split into :api and :impl; impl may depend only on other features' api.
listOf("auth", "home", "liveshow", "reactions", "chat", "wallet", "leaderboards", "referrals", "profile")
    .forEach { feature ->
        include(":feature:$feature:api")
        include(":feature:$feature:impl")
    }

// :benchmarks:baselineprofile and :benchmarks:macrobenchmark join with the reaction benchmark (Phase 3, L11).
