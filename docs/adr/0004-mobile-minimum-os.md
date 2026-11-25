# ADR-0004: Minimum iOS 16.0; Android minSdk 26

- **Status:** Accepted. Owner decision, 2026-11-25. The owner kept the Master Plan DEFAULT of iOS 16 instead of the researched recommendation of iOS 17.
- **Date:** 2026-11-25

## Context ([research 04](../research/04-native-mobile.md))

Three things favoured iOS 17:

- `@Observable` (the Observation framework) and `UITraitCollection.sceneCaptureState` both need iOS 17.0 or later.
- Xcode 27.2 can debug on devices only if they run iOS 17 or later.
- Adoption is high: iOS 26 is on 79% of all iPhones (Apple, June 2026).

The owner chose iOS 16 to maximise reach.

## Decision

**iOS deployment target: 16.0.** The consequences for the code are:

1. **State management:** `ObservableObject` with `@Published`, `@StateObject` and `@EnvironmentObject` across the app. Do not mix in `@Observable`, so there is one model.
   - Views that update at high frequency (the timer, the reaction totals) must publish narrowly scoped objects. Otherwise the whole screen redraws, because ObservableObject has no per-property tracking.
2. **Screen-capture detection** for decision AC-4 (block during questions):
   - Use `UIScreen.main.isCaptured` and `UIScreen.capturedDidChangeNotification`, which are available on iOS 11–27 and deprecated in 27.
   - On iOS 17 and later, prefer `sceneCaptureState` behind `#available(iOS 17, *)`.
   - Screenshots: `userDidTakeScreenshotNotification` on every version.
3. **Live Activities:**
   - The countdown Live Activity needs iOS 16.1 or later.
   - Push-to-start (17.2) and broadcast updates (18) are gated with `#available`.
   - On 16.x, the countdown starts locally when the user taps "Remind me".
4. **Testing:** Xcode 27 cannot debug on iOS 16 devices. The iOS 16 reference device receives TestFlight or ad-hoc builds and is used for manual and automated UI checks only. Debugging happens on iOS 17 or later.
5. **Revisit:** review the minimum each year. Raise it when iOS 16 drops below about 2% of active installs, measured from the app's own analytics.

**Android:**
- minSdk 26 (unchanged)
- targetSdk 36 (required by Play since 2026-08-31)
- compileSdk 37

## Consequences

- Wider reach on older iPhones.
- More code paths: capture detection and Live Activity start each need a 16.x fallback.
- Higher-frequency UI needs careful object scoping to stay smooth.
- The iOS 16 test device cannot be live-debugged, which slows bug fixes found only on iOS 16.
