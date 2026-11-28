# 15. iOS shell: what it takes to compile in CI

`ios/` has the two package manifests, `OptionShuffle.swift` and the vector test, and it has never been compiled. `Vendor/AmazonIVSPlayer` is empty but required by `Packages/Core`, there is no Xcode project, and no CI. Sources read 2026-11-28 (verbatim unless marked).

| Source | Text |
|---|---|
| [Xcode 27 runner image](https://github.com/actions/runner-images/blob/main/images/macos/xcode-27-arm64-Readme.md) | "27.2 \| 27B52 \| /Applications/Xcode_27.2.app"; the default is 27.1, so CI selects 27.2. iOS 27.2 simulators include "iPhone 17". Homebrew 7.0.16; XcodeGen, Mint and SwiftLint are not installed |
| [IVS iOS getting started](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/ios-getting-started.html) | "We recommend that you integrate the player SDK via Swift Package Manager"; download the `Package.swift` into a local `AmazonIVSPlayer` directory |
| [IVS 1.58.1 Package.swift](https://player.live-video.net/1.58.1/Package.swift) | `.binaryTarget(name: "AmazonIVSPlayer", url: ".../1.58.1/AmazonIVSPlayer.xcframework.zip", checksum: "1ff3bbd4...1455")`, `.iOS("15.0")`. SwiftPM checks the checksum on download |
| [IVS release notes](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/release-notes.html) | "November 9, 2026 ... iOS 1.58.1". 1.59.0 followed on November 26, 2026; the pin stays at 1.58.1 (stack/06) |
| [XcodeGen](https://github.com/yonaskolb/XcodeGen) | Install: `brew install xcodegen`, Mint, `make install`, or `swift run xcodegen` from a clone. 2.49.0 is the latest, commit `45ad520`. Release assets did not load: UNVERIFIED |
| [XcodeGen ProjectSpec](https://github.com/yonaskolb/XcodeGen/blob/master/Docs/ProjectSpec.md) | "deploymentTarget"; `info:` "will generate and write an `Info.plist`"; settings such as `CODE_SIGNING_ALLOWED: NO` and `SWIFT_VERSION`. Local `packages:` with `path:` and `- package:` dependencies are from the summary only: CI confirms |

## Settled by the sources

1. CI runs on the `xcode-27` image with `xcode-select` pointing at Xcode 27.2.
2. The AWS-published `Package.swift` for 1.58.1 is committed under `Vendor/AmazonIVSPlayer`. The binary itself is never committed, and SwiftPM verifies it by checksum.
3. XcodeGen is built from its 2.49.0 tag (`swift run`), so the version is pinned. Homebrew would install whatever is latest.
4. The simulator build needs no signing (`CODE_SIGNING_ALLOWED: NO`), so no Apple account is needed.
