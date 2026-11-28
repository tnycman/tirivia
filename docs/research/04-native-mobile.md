# Research 04: Native mobile (iOS + Android)

> Checked against primary sources on **2026-11-25**. **UNVERIFIED** means we could not confirm it from a primary source.
> Feeds: [`stack/06-ios`](../../stack/06-ios/README.md), [`stack/07-android`](../../stack/07-android/README.md), [ADR-0004](../adr/0004-mobile-minimum-os.md).

## 1. Verified facts and limits

| Fact | Value | Source |
|---|---|---|
| Current Xcode | **Xcode 27.2**, Swift 6.4, iOS 27.2 SDK. **On-device debugging iOS 17+ only** | https://developer.apple.com/documentation/xcode-release-notes/xcode-27_2-release-notes.md |
| Swift stable | **6.4.3** | https://www.swift.org/install/ |
| App Store build requirement | Xcode 26+ / iOS 26 SDK since 2026-04-28 | https://developer.apple.com/news/upcoming-requirements/ |
| iOS adoption | iOS 26: 79% of all iPhones, 86% of iPhones from the last 4 years (Jun 2026). No figures for iOS 17/18 | https://developer.apple.com/support/app-store/ |
| `@Observable` | **iOS 17.0+** | https://developer.apple.com/documentation/observation.md |
| `UIScreen.isCaptured` | Deprecated in iOS 27. Also true during AirPlay mirroring | https://developer.apple.com/documentation/uikit/uiscreen/iscaptured.md |
| `sceneCaptureState` | **iOS 17.0+** | https://developer.apple.com/documentation/uikit/uitraitcollection/scenecapturestate.md |
| App Attest throttling | Ramp to at most 10M users/day; **keep `attestKey` under 100 req/s** across the install base; assertions are not limited | https://developer.apple.com/documentation/devicecheck/preparing-to-use-the-app-attest-service.md |
| App Attest risk metric | Approximate count of attested keys per device over 30 days | https://developer.apple.com/documentation/devicecheck/assessing-fraud-risk.md |
| APNs token auth | ES256 JWT, refreshed between every 20 and every 60 min | https://developer.apple.com/documentation/usernotifications/establishing-a-token-based-connection-to-apns.md |
| APNs request | HTTP/2; priority 10/5/1; `apns-collapse-id` ≤ 64 B; payload ≤ 4 KB | https://developer.apple.com/documentation/usernotifications/sending-notification-requests-to-apns.md |
| Time-sensitive interruption | iOS 15+; breaks through Focus | https://developer.apple.com/documentation/usernotifications/unnotificationinterruptionlevel/timesensitive.md |
| Live Activities push-to-start | iOS 17.2+. Broadcast channels need iOS 18+ and cannot start an activity | https://developer.apple.com/documentation/activitykit/starting-and-updating-live-activities-with-activitykit-push-notifications.md |
| **Albanian on iOS** | **iOS system UI is not localized into Albanian or Macedonian** (Apple engineer on the forums) | https://developer.apple.com/forums/thread/850074 |
| Kotlin | **2.4.21** | https://kotlinlang.org/docs/releases.html |
| AGP / Gradle / JDK | **9.5.1** (max API 37, needs Gradle ≥ 9.6, JDK 17) / **9.9.2** | https://developer.android.com/build/releases/gradle-plugin · https://gradle.org/releases/ |
| Compose | **BOM 2026.11.01**, UI 1.13.0, Material 3 1.5.0 | https://developer.android.com/develop/ui/compose/bom/bom-mapping |
| Compose minSdk | 21 → 23 (1.10). AndroidX is moving to 24 | https://developer.android.com/jetpack/androidx/versions/all-channel |
| Play target API | **API 36** required since 2026-08-31 | https://developer.android.com/google/play/requirements/target-sdk |
| Android 8.0+ reach | API 26+ 99.0%; API 28+ 97.0% (Oct 2026; third-party compilation of Studio data) | https://composables.com/android-distribution-chart.md |
| **Play Integrity quota** | **10,000 requests/day by default** (plus 10k decrypts); an increase takes up to a week | https://developer.android.com/google/play/integrity/setup |
| Play Integrity lib | `com.google.android.play:integrity:1.6.0`; standard requests need API 23+ | same |
| MEETS_STRONG_INTEGRITY | On Android 13+ the security patch must be from the last year | same |
| `appAccessRiskVerdict` | Values `*_CAPTURING`, `*_CONTROLLING`, `*_OVERLAYS`; verified accessibility tools are exempt | same |
| Screenshot detection | API 34; the user sees a system notice | https://developer.android.com/about/versions/14/features/screenshot-detection |
| Screen-recording detection | API 35 `addScreenRecordingCallback` | https://developer.android.com/about/versions/15/features |
| Hide overlays | `setHideOverlayWindows`, API 31 | https://developer.android.com/reference/android/view/Window |
| FLAG_SECURE | Reliable on only about 70% of devices on Android ≤11 | https://developer.android.com/security/fraud-prevention/activities |
| Strong skipping | Default since Kotlin 2.0.20 | https://developer.android.com/develop/ui/compose/performance/stability/strongskipping |
| Baseline Profiles | About 30% faster from first launch | https://developer.android.com/topic/performance/baselineprofiles/overview |
| FCM | HTTP v1. **InstanceId server APIs (topics, device groups) fail after 2027-09-29** | https://firebase.google.com/docs/cloud-messaging/send/v1-api · https://firebase.google.com/docs/cloud-messaging/troubleshooting |
| Sentry | Cocoa **9.35.0** (iOS 15+); Android **8.65.1** | https://github.com/getsentry/sentry-cocoa/releases · https://github.com/getsentry/sentry-java/releases |
| OWASP MASVS | **v2.1.0**: 24 controls in 8 groups | https://github.com/OWASP/masvs |
| Instagram Stories | Facebook App ID required; `instagram-stories://share`; Android intent `com.instagram.share.ADD_TO_STORY` | https://developers.facebook.com/docs/instagram-platform/sharing-to-stories/ |
| TikTok Share Kit | Client key; media via the photo library (iOS) or FileProvider (Android) | https://developers.tiktok.com/docs/en/share-kit-ios-quickstart-v2 |
| WhatsApp | URL scheme sends text only; images go through the share sheet | https://faq.whatsapp.com/5913398998672934 |

## 2. iOS best practices adopted

1. **Swift 6 language mode, migrated module by module**: leaf modules first, upcoming-feature flags, then complete checking. Use default MainActor isolation for UI packages. Network and attestation code stays `nonisolated` or uses actors. ([migration guide](https://raw.githubusercontent.com/swiftlang/swift-migration-guide/main/Guide.docc/MigrationStrategy.md), [Swift 6.2](https://www.swift.org/blog/swift-6.2-released/))
2. **Use Observation (`@Observable`)**. Views update only for the properties they read, which matters for an overlay that updates many times per second. ([Apple](https://developer.apple.com/documentation/swiftui/migrating-from-the-observable-object-protocol-to-the-observable-macro.md))
3. **Plain `@Observable` view models per feature, plus one explicit show state machine.** TCA 1.26.2 is a viable alternative but needs more setup. This choice is our judgement, not Apple guidance.
4. **App Attest**:
   - Check `isSupported` first.
   - Use one key per user per device.
   - Retry with the same key on `serverUnavailable`; on any other error, discard the key.
   - Challenges must be at least 16 bytes.
   - The server runs Apple's 9 attestation checks plus the assertion checks.
   - **Attest at sign-in and send an assertion at show join**, to stay under 100 req/s. ([establishing integrity](https://developer.apple.com/documentation/devicecheck/establishing-your-app-s-integrity.md), [validating](https://developer.apple.com/documentation/devicecheck/validating-apps-that-connect-to-your-server.md))
5. **Capture and focus signals**: watch `sceneCaptureState` and `userDidTakeScreenshotNotification`. `scenePhase` `.inactive` means foreground but paused (it also fires when Notification Center opens); `.background` means the app is not visible. ([scenePhase](https://developer.apple.com/documentation/swiftui/scenephase.md))
6. **APNs**:
   - Use a .p8 key and long-lived HTTP/2 connections.
   - Mark show reminders `timeSensitive` and use `collapse-id`.
   - Start the countdown Live Activity with push-to-start (iOS 17.2+), and on iOS 18+ update it over a broadcast channel.
7. **String Catalogs** (`.xcstrings`) with plural variants, plus an **in-app language picker**, because iOS has no Albanian system UI.

## 3. Android best practices adopted

1. **Modularisation** follows the official patterns and Now in Android:
   - Split into `core:*` and `feature:*`, with each feature divided into `:api` and `:impl`; an `impl` module may depend only on other features' `api`.
   - Keep versions in a version catalog and shared build config in convention plugins under `build-logic`. ([patterns](https://developer.android.com/topic/modularization/patterns), [NiA](https://github.com/android/nowinandroid/blob/main/docs/ModularizationLearningJourney.md))
2. **Hilt** (Jetpack-recommended, compile-time graph). ([DI](https://developer.android.com/training/dependency-injection))
3. **Play Integrity standard requests**:
   - Warm up at app launch and request on demand with `requestHash`.
   - Decode on the server and never cache verdicts.
   - Opt in to `appAccessRiskVerdict`.
   - Collect telemetry first, then enforce in tiers. ([standard](https://developer.android.com/google/play/integrity/standard))
4. **Compose performance**: defer state reads into lambda modifiers (`offset{}`, `graphicsLayer{}`, `drawBehind{}`); use `drawWithCache` and `CompositingStrategy.ModulateAlpha`. ([best practices](https://developer.android.com/develop/ui/compose/performance/bestpractices))
5. **Reaction layer**:
   - One full-screen `Canvas` driven by a single frame loop.
   - Particles in a preallocated pool of primitive arrays, not one State object per particle.
   - Emoji pre-rasterised to `ImageBitmap`.
   - A cap on live particles, with extras dropped on slow frames.
   - This is our design. The `withFrameNanos` reference page returned 404, so that API detail is UNVERIFIED.
6. **Video surface**: SurfaceView is preferred over TextureView ([media3 surface](https://developer.android.com/media/media3/ui/surface)). Which surface the IVS view uses internally is **UNVERIFIED**; check it in phase 2a.
7. **Baseline Profiles and Startup Profiles** covering launch → join → live screen with reactions, built with Macrobenchmark. R8 full mode.
8. **Integrity signals**: screenshot callback (API 34+), recording callback (API 35+), `setHideOverlayWindows` (API 31+), and the list of enabled accessibility services. **These are logged and never used to penalise accessibility users.**
9. **`POST_NOTIFICATIONS`**: ask in context, for example from a "Remind me" button.
10. **Per-app language**: `generateLocaleConfig = true` and `AppCompatDelegate.setApplicationLocales`.

## 4. Conflicts with Master Plan v1.1 and adopted fixes

| # | Plan | Finding | Adopted fix |
|---|---|---|---|
| 1 | Minimum iOS 16 (DEFAULT) | iOS 16 blocks `@Observable` and `sceneCaptureState`, and Xcode 27 cannot debug on it | **Minimum iOS 17.0** ([ADR-0004](../adr/0004-mobile-minimum-os.md)). Push-to-start (17.2) and broadcast channels (18) are gated with `#available`. |
| 2 | Play Integrity (quota not mentioned) | Default 10k/day; one show could exceed it | Quota request filed in **Phase 0**; one standard request per join; quota alerts. |
| 3 | IVS views | 15k per region by default | Quota request (see research 01). |
| 4 | "Correct answer is never sent" | AWS's sample quiz puts `correctIndex` in metadata | Our cue schema has no correct-answer field until `reveal` ([cue.proto](../../protocol/proto/trivia/cue/v1/cue.proto)). |
| 5 | Attestation to join a prize show | `attestKey` must stay under 100 req/s | **Attest at sign-in, assertion at join** (`AuthService.Attest` carries both). |
| 6 | Albanian primary | iOS has no Albanian system UI | In-app language picker. **Default `sq` when region is AL/XK/MK** (product decision; listed in open items). |
| 7 | Focus loss voids final-three answers | `.inactive` fires on harmless events; Android shows a system notice on screenshot | Void only on `.background` or on `.inactive` longer than a threshold. The threshold is a DEFAULT to tune in rehearsal. The rule is explained before each show. |
| 8 | FCM direct | InstanceId topic and group APIs end 2027-09-29 | Send to tokens via HTTP v1; no server-side topics. |
| 9 | Connect clients | connect-kotlin is still 0.x | Pin the version; protocol test vectors guard against regressions. |
| 10 | Minimum Android 8.0 (DEFAULT) | 99.0% reach; Play Integrity needs API 23+, IVS 6.0+ | **Keep minSdk 26.** compileSdk 37, targetSdk 36. |

## 5. Versions

| Tool | Version |
|---|---|
| Xcode / Swift | 27.2 / 6.4 (Swift 6 language mode) |
| iOS deployment target | 17.0 |
| Kotlin / AGP / Gradle / JDK | 2.4.21 / 9.5.1 / 9.9.2 / 17 |
| Compose BOM / Material 3 | 2026.11.01 / 1.5.0 |
| compile / target / min SDK | 37 / 36 / 26 |
| Hilt (androidx) | 1.4.0 |
| Play Integrity | 1.6.0 |
| IVS Player | 1.58.1 |
| Sentry | Cocoa 9.35.0 / Android 8.65.1 |
| connect-swift / connect-kotlin | 1.2.3 / 0.10.0 |

## 6. MASVS v2 controls in scope

- STORAGE-1/2: tokens kept in Keychain/Keystore; no PII in logs, Sentry or session replays.
- CRYPTO-1/2.
- AUTH-1/3: OTP, plus step-up for wallet claims.
- NETWORK-1/2: TLS everywhere; **pin only our own API hosts**, never IVS or CDN hosts.
- PLATFORM-3: overlays and FLAG_SECURE.
- CODE-1/2: minimum OS and forced update.
- RESILIENCE-1..4: App Attest and Play Integrity, anti-tamper, anti-static and anti-dynamic analysis.
- PRIVACY-1..4.

## 7. Not verified

- Whether iOS 27 requires the scene lifecycle.
- The time-sensitive entitlement docs.
- CLDR plural categories for Albanian (`sq`).
- Whether the iOS per-app language setting lists a language that has no system UI.
- The 2027 store requirement bumps.
