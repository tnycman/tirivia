# 07 Android app

## Role
The native Android client. It has the same features as iOS and must hold its frame rate on a **low-end reference phone** while video plays.

## Decision
- **Kotlin**, **Jetpack Compose** with Material 3, and **Hilt** for dependency injection.
- **Play Integrity standard requests** for device attestation.
- **IVS Player SDK** for video and **connect-kotlin** for the API.
- Modules follow the **Now in Android** pattern.
- SDK levels: minSdk **26**, targetSdk **36**, compileSdk **37** ([ADR-0004](../../docs/adr/0004-mobile-minimum-os.md)).

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| Kotlin | 2.4.21 | https://kotlinlang.org/docs/releases.html |
| AGP / Gradle / JDK | 9.5.1 / 9.9.2 / 17 | https://developer.android.com/build/releases/gradle-plugin |
| Compose BOM / Material 3 | 2026.11.01 / 1.5.0 | https://developer.android.com/develop/ui/compose/bom/bom-mapping |
| Hilt (androidx) | 1.4.0 | https://developer.android.com/jetpack/androidx/releases/hilt |
| Play Integrity | 1.6.0 | https://developer.android.com/google/play/integrity/setup |
| IVS Player | 1.58.1 | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/android-getting-started.html |
| connect-kotlin | 0.10.0 | https://github.com/connectrpc/connect-kotlin/releases |
| Sentry Android | 8.65.1 | https://github.com/getsentry/sentry-java/releases |
| Play target API requirement | 36 (since 2026-08-31) | https://developer.android.com/google/play/requirements/target-sdk |

## Best practices
1. **Modularisation.**
   - Split code into `core:*` and `feature:*:{api,impl}` modules. An `impl` module may depend only on other features' `api` modules.
   - The app module wires navigation.
   - Prefer `implementation` dependencies over `api`.
   - Keep versions in a version catalog and shared build logic in convention plugins under `build-logic/`.
   - Sources: [patterns](https://developer.android.com/topic/modularization/patterns), [NiA](https://github.com/android/nowinandroid/blob/main/docs/ModularizationLearningJourney.md).
2. **Play Integrity.**
   - Warm up the token provider at app launch, no more than 5 times a minute. Request a token on demand with `requestHash`.
   - Decode the token on the server and never cache verdicts.
   - Opt in to `appAccessRiskVerdict`, which reports capturing, controlling and overlay apps.
   - Collect telemetry first, then enforce.
   - The **default quota is 10k requests/day**, so request an increase in Phase 0.
   - Source: [standard requests](https://developer.android.com/google/play/integrity/standard).
3. **Compose performance.**
   - Defer state reads into lambda modifiers (`offset {}`, `graphicsLayer {}`, `drawBehind {}`).
   - Use `drawWithCache`. No backwards writes.
   - Strong skipping is on by default.
   - Source: [best practices](https://developer.android.com/develop/ui/compose/performance/bestpractices).
4. **Reaction layer.**
   - One full-screen `Canvas` with a single frame loop.
   - Particles live in a **preallocated pool of primitive arrays**, not one `State` per particle.
   - Emoji are pre-rasterised to `ImageBitmap` once.
   - Live particles are capped, and new ones are dropped when frames run slow.
   - **Gate: ≥ 55 fps on the reference phone**, measured with Macrobenchmark `FrameTimingMetric`.
5. **Video surface.** SurfaceView is preferred over TextureView ([media3](https://developer.android.com/media/media3/ui/surface)). Which surface the IVS view uses internally is UNVERIFIED; confirm it in Phase 2a.
6. **Startup performance.**
   - Ship Baseline Profiles and Startup Profiles covering launch → join → live screen. They make code about 30% faster from first launch ([overview](https://developer.android.com/topic/performance/baselineprofiles/overview)).
   - Use R8 full mode.
7. **Integrity signals.**

   | Signal | API |
   |---|---|
   | Screenshot | `Activity.ScreenCaptureCallback` (API 34+) |
   | Screen recording | `WindowManager.addScreenRecordingCallback` (API 35+) |
   | Overlays | `setHideOverlayWindows(true)` on the live screen (API 31+) |
   | Accessibility services | Enabled-services list |

   **Log these signals; never penalise accessibility users.**
8. **Push.** Send through FCM HTTP v1 to individual tokens, not topics. Ask for `POST_NOTIFICATIONS` in context, on "Remind me". ([FCM v1](https://firebase.google.com/docs/cloud-messaging/send/v1-api))
9. **Language.** Use `generateLocaleConfig = true` and `AppCompatDelegate.setApplicationLocales` (sq/en). ([app languages](https://developer.android.com/guide/topics/resources/app-languages))
10. **Credentials.**
    - The DPoP key is a Keystore EC P-256 key, StrongBox-backed where available, with user authentication required for the step-up key.
    - Tokens are encrypted with a Keystore key before they are written to DataStore.

## Rules for this codebase
- No feature module may depend on another feature's `impl`.
- Generated protocol code lives only in `core:protocol`.
- All dependencies come from `libs.versions.toml`. Renovate keeps it current.
- Lint, detekt and ktlint must pass in CI.

## Code
[`android/`](../../android/)

## Sources
[docs/research/04-native-mobile.md](../../docs/research/04-native-mobile.md)

## TV companion mode (ADR-0010)
A "Watching on TV" toggle on the live screen mutes and minimises the video, or switches to the IVS audio-only rendition if cues are carried there (to verify in Phase 2a). It keeps receiving cues, so the question overlay, timer and answers work normally while the player watches the show on TV.
