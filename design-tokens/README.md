# design-tokens: one source of truth for colour, type, spacing and motion

Tokens are written once in the **W3C Design Tokens Community Group Format 2025.10**, the first stable version, published 2025-10-28 ([spec](https://www.designtokens.org/TR/2025.10/format/)). **Style Dictionary 5.8.0** generates the iOS, Android and web code from them.

> **All values are neutral placeholders** until the product name and brand are chosen (decision LG-3). App code must use only the **semantic** names (`theme.*`, `overlay.*` and the base scales). Re-skinning then means changing the primitives with no app code changes.

```
design-tokens/
├── tokens/
│   ├── primitives.tokens.json    palette (placeholder), spacing, radius, touch target, type, motion
│   ├── theme-light.tokens.json   semantic colours, light
│   ├── theme-dark.tokens.json    semantic colours, dark (same names)
│   └── overlay.tokens.json       live-show overlay drawn over video (theme-independent)
└── scripts/
    ├── contrast-check.mjs        WCAG 2.2 AA gate for every text/background pair
    └── build.mjs                 Style Dictionary build (+ --check for CI)
```

## Generated outputs (never edit by hand)

| Platform | Files |
|---|---|
| iOS | `ios/Packages/Core/Sources/DesignSystem/Generated/Tokens{Base,Light,Dark,Overlay}.swift` |
| Android | `android/core/designsystem/src/main/kotlin/al/trivia/core/designsystem/generated/Tokens{Base,Light,Dark,Overlay}.kt` |
| Web | `web-consoles/packages/ui/src/tokens/tokens-*.css`, imported via `@trivia/ui/tokens.css`. Dark mode follows the OS, and `data-theme` can force light or dark |

## Rules

1. **Author dimensions in `rem`** (1 rem = 16 pt iOS / 16 dp Android / 1 rem web). This is verified: the built-in Swift and Compose transforms treat numbers as rem, so a `px` value of 16 came out as **256**.
2. **Text sizes go under `font.size.*`.** On Android they become `sp`, which respects the user's font-scale setting; every other dimension becomes `dp`. On iOS, scale text with Dynamic Type (`UIFontMetrics`).
3. **Durations use DTCG `{value, unit}` objects.** Custom transforms emit `TimeInterval` (seconds) in Swift and `Long` (milliseconds) in Kotlin. The built-ins printed `[object Object]`.
4. **Every new text/background pair is added to `contrast-check.mjs`.** CI fails below WCAG 2.2 AA: 4.5 : 1 for normal text, 3 : 1 for large text. Overlay text is checked over a **pure-white video frame** behind the 70% scrim, the worst case.
5. **Touch targets are at least 48 (`touch.min`).** That covers Apple's 44 pt and Material's 48 dp.
6. **Fonts:** native apps use the system font (SF Pro / Roboto), which renders Albanian **ë, ç, Ë, Ç**. Any brand font must be checked for these glyphs before adoption.

## Verified on 2026-11-25

**Build output:**

| Token | iOS | Android | Web |
|---|---|---|---|
| `space.md` | 16 pt | 16.00 dp | 1rem |
| `font.size.body` | 17 pt | 17.00 sp | — |
| `motion.reactionFloat` | 2.0 s | 2000L | 2s |
| overlay scrim | — | `0xb3000000` | `rgba(0, 0, 0, 0.7)` |

**Contrast gate:** all **25 pairings** meet WCAG 2.2 AA. On its first run the gate caught 4 failures, and each was fixed in the tokens:

- red text on the dark background (3.0 : 1) → a separate lighter `primaryText` token
- three overlay colours over bright video → the scrim was raised from 60% to 70%, and the overlay "wrong" red was lightened

**Web:** the studio console builds with the tokens, including automatic dark mode.

**Not yet compiled:** the Swift and Kotlin outputs. That happens in the macOS and JDK CI.

## Commands

```sh
npm install
npm run build    # regenerate platform files
npm run check    # contrast gate + stale check (CI)
```
