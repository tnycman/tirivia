# ios: native iPhone app

Swift 6 language mode with SwiftUI (`ObservableObject` state model). Requires **iOS 16.0 or later** ([ADR-0004](../docs/adr/0004-mobile-minimum-os.md)). Built with **Xcode 27.2**. Stack and best practices: [`stack/06-ios`](../stack/06-ios/README.md).

```
ios/
├── App/                          # thin Xcode app target (created in Phase 0 with Xcode 27.2)
│   ├── TriviaApp/                #   app shell: navigation, DI, scene lifecycle, language picker
│   ├── LiveActivity/             #   widget extension: show countdown (local start on iOS 16.1+, push-to-start on 17.2+)
│   └── Resources/Localizable.xcstrings   # sq (primary), en
├── Packages/
│   ├── Core/                     # CoreProtocol (generated), CoreNetwork, CoreAttestation,
│   │                             # CoreSecurity, CoreAnalytics, CoreVideo, DesignSystem, CoreTesting
│   └── Features/                 # Auth, Home, LiveShow, Reactions, Chat, Wallet,
│                                 # Leaderboards, Referrals, Profile
├── Vendor/AmazonIVSPlayer/       # local SPM package for IVS Player 1.58.1 (downloaded)
└── .swiftlint.yml
```

## Module rules

- A feature module depends only on `Core`. The app shell is the only place that wires navigation between features.
- `CoreProtocol` contains generated code only.
- Cue verification and the option shuffle live in `CoreSecurity`. They must pass every vector in `protocol/testvectors`.

## Chat (backend switch)

- The chat panel is shown or hidden from `JoinShowResponse.chat_enabled`. After that, `Pulse.chat_enabled` updates it while the show runs.
- The app connects to chat only when the user opens the panel. If chat is switched off, it disconnects ([ADR-0009](../docs/adr/0009-chat-backend-switch.md)).

## Status

This workspace runs on Windows, which has no Xcode or Swift toolchain. `.github/workflows/ios-ci.yml` runs on GitHub's `xcode-27` image with Xcode 27.2. It tests the Core package on the iPhone 17 simulator (including the shared shuffle vectors), generates `Trivia.xcodeproj` from `project.yml` with XcodeGen 2.49.0, and builds the app unsigned for the simulator ([research 15](../docs/research/15-ios-shell-build.md)). The project file, the generated `Info.plist` and the copied vectors are not committed.

`Vendor/AmazonIVSPlayer/Package.swift` is AWS's published 1.58.1 manifest. SwiftPM downloads the binary and checks its checksum.
