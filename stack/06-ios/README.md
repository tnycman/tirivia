# 06 iOS app

## Role
The native iPhone client:
- live video with the question overlay, timer and answer bars
- floating reactions and chat
- wallet, leaderboards and sharing
- push notifications and a Live Activity countdown

## Decision
- **Swift 6 language mode** with **SwiftUI** and `ObservableObject` state (iOS 16 target; ADR-0004).
- **Minimum iOS 16.0** ([ADR-0004](../../docs/adr/0004-mobile-minimum-os.md)).
- **App Attest** for device integrity.
- **IVS Player SDK** for video.
- **connect-swift** for the API.
- Feature modules are **local Swift packages**.

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| Xcode | 27.2 (Swift 6.4; requires macOS Tahoe 26.6+) | https://developer.apple.com/documentation/xcode-release-notes/xcode-27_2-release-notes.md |
| Swift toolchain | 6.4.3 | https://www.swift.org/install/ |
| Deployment target | iOS 16.0 (owner decision) | ADR-0004 |
| IVS Player | 1.58.1 (SPM, local Package.swift) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/ios-getting-started.html |
| connect-swift / swift-protobuf | 1.2.3 / 1.41.0 | GitHub releases |
| Sentry Cocoa | 9.35.0 | https://github.com/getsentry/sentry-cocoa/releases |
| App Store build requirement | Xcode 26+ / iOS 26 SDK (since 2026-04-28) | https://developer.apple.com/news/upcoming-requirements/ |

## Best practices

1. **Migrate to Swift 6 mode one module at a time.** Start with leaf modules: enable upcoming-feature flags, then complete concurrency checking, then switch the module to Swift 6 mode.
   - UI packages default to MainActor isolation.
   - Network and attestation code is `nonisolated` or actor-based.
   - Sources: [migration guide](https://raw.githubusercontent.com/swiftlang/swift-migration-guide/main/Guide.docc/MigrationStrategy.md), [Swift 6.2](https://www.swift.org/blog/swift-6.2-released/).
2. **Use `ObservableObject` view models, scoped narrowly.** The deployment target is iOS 16, and `@Observable` needs iOS 17 (ADR-0004).
   - `ObservableObject` re-renders a view on *any* `@Published` change ([Apple](https://developer.apple.com/documentation/swiftui/migrating-from-the-observable-object-protocol-to-the-observable-macro.md)). So split fast-changing state into its own small objects: the timer, the reaction totals and the viewer count.
   - Never put fast state in one large screen model.
   - Migrate to `@Observable` when the minimum moves to 17.
3. **Model the show with one explicit state machine** in `LiveShow` that mirrors SPEC §6.2. Plain view models are used everywhere else; we do not use TCA, to avoid its extra setup cost.
4. **Video**
   - IVS delegate callbacks arrive on the main queue. Parse and verify cues off the main actor, then publish the result to it.
   - IVS does not support background playback, so pause the player when the app goes to the background.
5. **App Attest**
   - Check `isSupported` before anything else.
   - Use one key per user per device and persist its `keyId`.
   - On `serverUnavailable`, retry with the same key. On any other error, discard the key and generate a new one.
   - **Attest once at sign-in, then send an assertion when joining a show.**
   - Source: [Apple](https://developer.apple.com/documentation/devicecheck/establishing-your-app-s-integrity.md).
6. **Integrity signals**
   - Screen recording: `sceneCaptureState`. It also reports true during AirPlay mirroring.
   - Screenshots: `userDidTakeScreenshotNotification`.
   - App focus: `scenePhase`. `.inactive` is common and usually harmless; `.background` means the app is no longer visible.
7. **Device key for DPoP**
   - A Secure Enclave P-256 key, used for ES256 DPoP proofs, with `.biometryCurrentSet` required for the step-up key.
   - Tokens are stored in the Keychain with `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly`.
8. **Reactions rendering**
   - A single `TimelineView(.animation)` containing a `Canvas`, drawing from a preallocated particle pool.
   - At most about 30 new particles per second.
   - Glyph images are resolved once and reused.
9. **Push notifications**
   - APNs show reminders use the `timeSensitive` interruption level.
   - The countdown is a Live Activity started locally on iOS 16.1+, by push-to-start on iOS 17.2+, and updated over a broadcast channel (iOS 18+).
10. **Localisation**
    - Strings live in `Localizable.xcstrings` with plural variants.
    - **iOS has no Albanian system UI, so the app has its own language picker** (sq/en). ([forum](https://developer.apple.com/forums/thread/850074))
11. **MASVS**
    - No PII in logs or Sentry.
    - TLS pinning only for our own API hosts.
    - A forced-update check at launch.

## Rules for this codebase
- Feature packages depend only on `Core`. Navigation is wired in the app shell.
- Generated protocol code lives only in `CoreProtocol`.
- Every user-facing string goes through the String Catalog. No hard-coded strings.
- Accessibility: Dynamic Type and VoiceOver labels on every control. Accessibility users are never penalised by the integrity signals.

## Code
[`ios/`](../../ios/)

## Sources
[docs/research/04-native-mobile.md](../../docs/research/04-native-mobile.md)

## TV companion mode (ADR-0010)
A "Watching on TV" toggle on the live screen mutes and minimises the video, or switches to the IVS audio-only rendition if cues are carried there (to verify in Phase 2a). It keeps receiving cues, so the question overlay, timer and answers work normally while the player watches the show on TV.
