# build-logic

Convention plugins, following the Now in Android pattern ([modularization patterns](https://developer.android.com/topic/modularization/patterns)). Each module applies one plugin instead of repeating Gradle configuration.

| Plugin id | Applies |
|---|---|
| `tirivia.android.application` | AGP application, compileSdk 37 / targetSdk 36 / minSdk 26, JDK 17, R8 full mode, `generateLocaleConfig = true` |
| `tirivia.android.library` | AGP library with the same SDK levels; `explicitApi()` in core modules |
| `tirivia.android.compose` | Compose compiler plugin and the Compose BOM; strong skipping is on by default |
| `tirivia.android.feature` | library + compose + hilt, with `core:ui`, `core:designsystem` and `core:network` |
| `tirivia.hilt` | Hilt and KSP |

Each plugin is implemented in `convention/src/main/kotlin/` and registered in `convention/build.gradle.kts`. This happens in Phase 0.
