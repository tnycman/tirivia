# android: native Android app

Built with Kotlin 2.4, Jetpack Compose (Material 3) and Hilt. SDK levels: **minSdk 26**, **targetSdk 36**, **compileSdk 37**. Stack and best practices are in [`stack/07-android`](../stack/07-android/README.md).

```
android/
├── settings.gradle.kts           # module graph
├── gradle/libs.versions.toml     # version catalog (every version sourced: research or Maven metadata)
├── build-logic/convention/       # convention plugins: application, library, feature, compose, hilt
├── app/                          # shell: navigation, DI graph, per-app locales (sq, en)
├── core/
│   ├── model  common  data  datastore  ui  testing
│   ├── protocol                  # generated connect-kotlin + OptionShuffle + cue verification
│   ├── network                   # OkHttp, DPoP, retries, pinning (own API hosts only)
│   ├── attestation               # Play Integrity standard requests (warm-up ≤ 5/min, requestHash)
│   ├── security                  # Keystore/StrongBox keys; screenshot (34+), recording (35+), overlays (31+)
│   ├── video                     # IVS Player 1.58.1 wrapper; cue callbacks on the main thread
│   ├── analytics                 # Sentry 8.65.1
│   └── designsystem
├── feature/<name>/{api,impl}     # auth home liveshow reactions chat wallet leaderboards referrals profile
└── benchmarks/{baselineprofile,macrobenchmark}   # reaction layer must hold ≥ 55 fps on the reference phone
```

## Rules

- **Feature dependencies:** a feature's `impl` module may depend only on other features' `api` modules. `app` owns navigation.
- **Generated code:** all of it lives in `core:protocol`.
- **Chat:** the panel shows or hides based on `chat_enabled`. The value comes from `JoinShowResponse` at join and from `Pulse` while the show is live ([ADR-0009](../docs/adr/0009-chat-backend-switch.md)).
- **Integrity signals:** they are logged. Users of accessibility services are never penalised.

## Status

There is no JDK or Android SDK in this workspace, so the Gradle files and `OptionShuffle.kt` were written against the documented APIs but **have not been compiled yet**.

- **Phase 0:** compile both in CI and confirm the KSP/Kotlin pairing (all versions in `libs.versions.toml` were checked against Maven metadata on 2026-11-25).
- **Then:** confirm through the test vectors that the shuffle matches the Go implementation.
