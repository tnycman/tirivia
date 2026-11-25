# ADR-0001: Fully native apps

- **Status:** Accepted (locked decision, Master Plan v1.1 §1)
- **Date:** 2026-11-25

## Context
These features are only first-class in native code:
- device attestation (App Attest, Play Integrity)
- the IVS Player SDK and its timed-metadata callbacks
- smooth reaction animation over video on low-end Android
- OS signals for screen capture and app focus

## Decision
- iOS uses Swift and SwiftUI.
- Android uses Kotlin and Jetpack Compose.
- There is no cross-platform UI layer.
- The two apps are kept identical by the shared protobuf protocol and shared test vectors (`protocol/testvectors`).

## Consequences
- There are two codebases. The shared game logic is a small state machine.
- The test vectors are the contract between the two apps.
