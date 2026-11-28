# 13. Building the apps without a Mac or store accounts

Owner state on 2026-11-27: Windows workstation only, no Mac with Xcode, no Apple Developer Program membership, no Google Play Console account. Today `ios/` and `android/` hold module layouts, generated design tokens and the option-shuffle vectors. There is no Xcode project, no Android app source and no mobile CI workflow.

## Primary sources (verbatim, read 2026-11-27)

### CI for iOS

| Source | Text |
|---|---|
| [GitHub Actions billing](https://docs.github.com/en/billing/managing-billing-for-your-products/managing-billing-for-github-actions/about-billing-for-github-actions) | "The use of standard GitHub-hosted runners is free:" ... "In public repositories". "Larger runners are billed for public repositories." |
| [actions/runner-images](https://github.com/actions/runner-images) | Labels: `macos-26` (Arm64), `xcode-27` or `xcode-27-xlarge` |
| [Xcode 27 arm64 image](https://github.com/actions/runner-images/blob/main/images/macos/xcode-27-arm64-Readme.md) | "OS Version: macOS 27.0.1"; Xcode 27.2 (build 27B52), 27.1 (default), 27.0 installed |
| [macOS 26 arm64 image](https://github.com/actions/runner-images/blob/main/images/macos/macos-26-arm64-Readme.md) | Xcode 26.0.1 to 26.6 only; **no Xcode 27.2** |
| [XcodeGen](https://github.com/yonaskolb/XcodeGen) | "generates your Xcode project using your folder structure and a project spec" (YAML or JSON). Latest release 2.49.0. 2.48.0: "requires macOS 14 and Swift 6.1 to build and run" |
| [Tuist](https://tuist.dev/docs/) | "Describe Xcode projects in Swift and generate them consistently." Free or open-source status: UNVERIFIED (guide page returned 403) |

### Apple

| Source | Text |
|---|---|
| [Membership overview](https://developer.apple.com/support/compare-memberships/) | Free account: on-device testing needs signing in "to your Apple Account in Xcode" (a Personal Team); "up to 10 App IDs, which expire after 7 days"; "up to 3 devices, which expire after 7 days"; "To build more advanced app capabilities and distribute apps, you'll need to be an Apple Developer Program member" |
| [Supported capabilities (iOS)](https://developer.apple.com/help/account/reference/supported-capabilities-ios) | App Attest is listed; whether a free account gets it: **UNVERIFIED** (the table's check marks did not survive extraction) |
| [DeviceCheck: establishing integrity](https://developer.apple.com/documentation/devicecheck/establishing-your-app-s-integrity), `isSupported` | Simulator behaviour: **UNVERIFIED** (the pages rendered titles only) |
| [Enroll](https://developer.apple.com/programs/enroll/) | "99 USD per membership year". Individual: an Apple Account "with two-factor authentication". Organization: "must have a D-U-N-S Number"; "We do not accept DBAs, fictitious business names, trade names, or branches." |

### Google

| Source | Text |
|---|---|
| [Play Integrity setup](https://developer.android.com/google/play/integrity/setup) | "you must have a Google Cloud project, which is necessary to begin making requests"; default "10,000 total requests" a day; "To be eligible for a quota increase, your app must be available on Google Play" |
| [Verdicts](https://developer.android.com/google/play/integrity/verdicts) | PLAY_RECOGNIZED: "The app and certificate match the versions distributed by Google Play." UNRECOGNIZED_VERSION: "The certificate or package name does not match Google Play records." The device verdict is empty on "an emulator that does not pass Google Play integrity checks" |
| [Play Console registration](https://support.google.com/googleplay/android-developer/answer/6112435) | "There is a US$25 one-time registration fee." |
| [Testing requirements](https://support.google.com/googleplay/android-developer/answer/14151465) | "Developers with personal accounts created after November 13, 2023, must run a closed test for their app with a minimum of 12 testers who have been opted in continuously for at least 14 days." Organization accounts are not mentioned (UNVERIFIED) |
| [Android Studio install](https://developer.android.com/studio/install) | "Windows machines with ARM-based CPUs aren't currently supported." Windows 10 64-bit, 16 GB RAM with the emulator, CPU virtualization (summary, not verbatim) |

### In this repo

- `backend/cmd/core/authdev.go`: with `AUTH_DEV_MODE=1`, sign-in runs with dev attestation and log-only code delivery. The sign-in screens can therefore be exercised on a simulator or emulator before real attestation is possible.

## What the sources settle

1. iOS can be compiled and unit-tested in CI at no cost (public repo), but only on the `xcode-27` image; `macos-26` lacks Xcode 27.2. That `xcode-27` (not `-xlarge`) counts as a standard free runner is inferred from the naming: confirm on the first run.
2. With no Mac, the Xcode project has to come from a spec that CI turns into a project (XcodeGen), not a hand-edited `.xcodeproj`.
3. Installing on a real iPhone needs Xcode on a Mac, even with a free account. Without a Mac, real-iPhone testing waits for a Mac, or for the Developer Program and TestFlight (TestFlight's need for membership is UNVERIFIED here).
4. Android builds, unit tests and the emulator can run on this Windows machine (x86-64 only) and in Linux CI.
5. A sideloaded Android build gets `UNRECOGNIZED_VERSION`, so under AC-5 it plays for points only. The prize-eligible path needs the app on Google Play (Play Console, US$25).
6. PLAN Phase 1's exit test ("sign-up works on real devices in all three countries") needs both store accounts. PLAN lists them as L13, blocking Phase 6, which is too late.

## Organization accounts (owner chose organization, OP-8)

| Source | Text |
|---|---|
| [Play Console account types](https://support.google.com/googleplay/android-developer/answer/13628312) | "You can't create an organization developer account without a valid D-U-N-S number." Applying to D&B "can take up to 30 days" and "is a completely free process". The account also needs an organization phone, an organization website, contact and developer email and phone, each verified "using a one-time password" |
| [Play developer verification](https://support.google.com/googleplay/android-developer/answer/10841920) | Organization: "D-U-N-S number", "Official government identity document", "Official organization document"; "Website verification (for organization accounts)" |
| [Apple D-U-N-S](https://developer.apple.com/support/D-U-N-S/) | Free number from D&B: "allow up to 5 business days"; then "up to 2 business days for Apple to receive your information" |
| [Apple enroll](https://developer.apple.com/programs/enroll/) | The enroller "must have the legal authority to bind your organization"; "Your work email address needs to [be] associated with your organization's domain name"; the website "must be publicly available and functional" (no social-media links or registrar placeholder pages) |

Not answered by the pages read: whether Google's 12-tester, 14-day closed test applies to organization accounts, and whether an account type can be changed after creation (UNVERIFIED).

## OPEN

- When to buy a Mac, or whether to rely on CI plus TestFlight.
