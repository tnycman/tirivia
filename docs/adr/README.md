# Architecture Decision Records

These follow Michael Nygard's lightweight ADR format. Records are append-only: to change a decision, write a new ADR that supersedes the old one. Copy [`0000-template.md`](0000-template.md) to start one.

| ADR | Decision | Status |
|---|---|---|
| [0001](0001-fully-native-apps.md) | Fully native iOS and Android apps with no cross-platform layer | Accepted (Master Plan, locked) |
| [0002](0002-questions-in-video-stream.md) | Questions and live counts delivered as signed IVS timed metadata | Accepted (Master Plan) |
| [0003](0003-standby-region-ireland.md) | Standby region is Ireland (eu-west-1), not Milan | **Accepted (owner, 2026-11-25)** |
| [0004](0004-mobile-minimum-os.md) | Minimum iOS 16.0; Android minSdk 26 | **Accepted (owner, 2026-11-25)** |
| [0005](0005-device-bound-tokens.md) | DPoP device-bound tokens, refresh rotation, step-up before claims | **Accepted (owner, 2026-11-25)** |
| [0006](0006-reaction-batching.md) | Reaction batches every 5 s with jitter; the server sets the interval | **Accepted (owner, 2026-11-25)** |
| [0007](0007-connect-rpc.md) | Connect RPC as the "protobuf over HTTPS" transport | **Accepted (owner, 2026-11-25)** |
| [0008](0008-terraform-not-opentofu.md) | Terraform 1.17, not OpenTofu | **Accepted (owner, 2026-11-25)** |
| [0009](0009-chat-backend-switch.md) | Chat is built for launch and switched on or off from the backend | **Accepted (owner, 2026-11-25)** |
| [0010](0010-tv-simulcast.md) | TV simulcast: app is the game, TV delayed behind the app, companion mode | **Accepted (owner, 2026-11-25)** |
