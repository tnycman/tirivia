# build-logic

Convention plugins, following the Now in Android pattern ([modularization patterns](https://developer.android.com/topic/modularization/patterns)). Each module applies one plugin instead of repeating Gradle configuration.

| Plugin id | Applies |
|---|---|
| `trivia.android.application` | AGP application, compileSdk 37.1 / targetSdk 36 / minSdk 26 (from the version catalog), `generateLocaleConfig = true` |
| `trivia.android.library` | AGP library, compileSdk 37.1 / minSdk 26 |
| `trivia.android.compose` | Compose compiler plugin, `buildFeatures.compose`, the Compose BOM, UI and Material 3 |
| `trivia.android.feature` | library + compose + hilt, with `core:ui`, `core:designsystem` and `core:network` |
| `trivia.hilt` | Hilt and KSP |

Each plugin is in `convention/src/main/kotlin/` and registered in `convention/build.gradle.kts`. AGP 9 compiles Kotlin itself, so no plugin applies `org.jetbrains.kotlin.android` ([research 14](../../docs/research/14-android-shell-build.md)). `explicitApi()` for core modules is not set yet.
