// swift-tools-version: 6.0
// Feature modules. Rule: a feature depends only on Core; the app shell wires navigation.
import PackageDescription

let swift6: [SwiftSetting] = [.swiftLanguageMode(.v6)]

func feature(_ name: String, _ core: [String]) -> Target {
    .target(
        name: name,
        dependencies: core.map { .product(name: $0, package: "Core") },
        swiftSettings: swift6
    )
}

let package = Package(
    name: "Features",
    platforms: [.iOS(.v16)],
    products: [
        .library(name: "Features", targets: [
            "Auth", "Home", "LiveShow", "Reactions", "Chat", "Wallet",
            "Leaderboards", "Referrals", "Profile",
        ]),
    ],
    dependencies: [.package(path: "../Core")],
    targets: [
        feature("Auth", ["CoreNetwork", "CoreAttestation", "CoreSecurity", "DesignSystem"]),
        feature("Home", ["CoreNetwork", "DesignSystem"]),
        // Show state machine mirrors SPEC §6.2; timer starts on cue render; resync on seq gap.
        feature("LiveShow", ["CoreNetwork", "CoreVideo", "CoreSecurity", "DesignSystem"]),
        // TimelineView + Canvas particle pool, <= ~30 new particles/s (stack/06-ios).
        feature("Reactions", ["CoreNetwork", "DesignSystem"]),
        // Visibility follows Pulse.chat_enabled (backend switch, ADR-0009); lazy connect on open.
        feature("Chat", ["CoreNetwork", "DesignSystem"]),
        // Step-up (biometric unlock of the DPoP key) before claims (ADR-0005).
        feature("Wallet", ["CoreNetwork", "CoreSecurity", "DesignSystem"]),
        feature("Leaderboards", ["CoreNetwork", "DesignSystem"]),
        feature("Referrals", ["CoreNetwork", "DesignSystem"]),
        feature("Profile", ["CoreNetwork", "DesignSystem"]),
    ]
)
