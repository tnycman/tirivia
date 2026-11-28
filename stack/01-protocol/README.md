# 01 Protocol and API contract

## Role
One definition of every message and RPC, shared by the Go server, the iOS app, the Android app and the web consoles. It also defines the signed timed-metadata cues carried inside the video stream.

## Decision
- **Protobuf over HTTPS** (Master Plan §5) using the **Connect protocol** ([ADR-0007](../../docs/adr/0007-connect-rpc.md)).
- **Buf v2** workspace for linting, breaking-change checks and code generation.
- **Our own test-vector suite** run by all three platforms (`protocol/testvectors`). `connectrpc/conformance` runs only against the server.

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| buf CLI | 1.75.0 | https://github.com/bufbuild/buf/releases |
| connect-go | v2.0.2 (fallback v1.21.1) | https://github.com/connectrpc/connect-go/releases |
| connect-swift | 1.2.3 | https://github.com/connectrpc/connect-swift/releases |
| swift-protobuf | 1.41.0 | https://github.com/apple/swift-protobuf/releases |
| connect-kotlin | 0.10.0 (pre-1.0, pinned) | https://github.com/connectrpc/connect-kotlin/releases |
| protobuf-es | 2.17.0 | https://github.com/bufbuild/protobuf-es/releases |
| connectrpc/conformance | 1.0.7 | https://github.com/connectrpc/conformance |

## Best practices
1. **Lint STANDARD and breaking FILE.** Never use a breaking category weaker than WIRE_JSON, because Connect clients can speak JSON. Run `buf breaking --against '.git#branch=main'` on every pull request. ([Buf breaking rules](https://buf.build/docs/breaking/rules/))
2. **Field evolution:**
   - Never reuse or renumber a field.
   - When a field is deleted, `reserved` both its number and its name.
   - Adding fields and enum values is safe.
   - Do not move existing fields into a oneof.
   ([proto3 guide](https://protobuf.dev/programming-guides/proto3/), [dos and don'ts](https://protobuf.dev/best-practices/dos-donts/))
3. **Enums:** every enum has `*_UNSPECIFIED = 0`. Use an enum rather than a bool when the value might later need more states.
4. **Tag numbers:** hot fields get tags 1–15, which encode in one byte. This matters for cues, which have a 900 B budget.
5. **Time types:** use `google.protobuf.Timestamp` and `Duration` for times. Keep `int64 *_unix_ms` only inside cues, where every byte counts.
6. **Naming:** lower_snake_case files and fields, TitleCase types, plural names for repeated fields, enum values prefixed with the enum name. ([style guide](https://protobuf.dev/programming-guides/style/))
7. **Versioned packages** (`trivia.game.v1`). A `v2` package is created only for a breaking change.
8. **Read-only RPCs** are marked `idempotency_level = NO_SIDE_EFFECTS` so Connect can serve them as cacheable GETs. ([Connect protocol](https://connectrpc.com/docs/protocol/))
9. **Idempotency.** Mutating RPCs accept the `Idempotency-Key` header and follow draft-09 semantics: 400 when the key is missing, 409 while the first request is in flight, 422 when the key is reused with a different payload. ([IETF draft](https://www.ietf.org/archive/id/draft-ietf-httpapi-idempotency-key-header-09.html))

## Rules for this codebase
- Cues: `t1.` + base64url(`SignedCue`), encoded size **≤ 900 bytes**, Ed25519 signature over the exact body bytes received. See [`cue.proto`](../../protocol/proto/trivia/cue/v1/cue.proto).
- The **correct answer is never placed in any message that a client receives before `reveal`.**
- Kinesis events are written as **canonical proto3 JSON**, because Firehose's Parquet conversion only accepts JSON.
- Generated code is never edited by hand. Regenerate it with `buf generate`.
- A new RPC or cue type must ship with test vectors in the same pull request.

## Changes vs Master Plan
- The "conformance test suite" is `protocol/testvectors`, our own suite. `connectrpc/conformance` checks the protocol only, not our API. ([research 02](../../docs/research/02-backend-protocol-data.md))
- REST-style paths map to Connect RPCs. See SPEC §6.5.

## Code
[`protocol/`](../../protocol/)

## Sources
[docs/research/02-backend-protocol-data.md](../../docs/research/02-backend-protocol-data.md), [docs/research/01-live-video-ivs.md](../../docs/research/01-live-video-ivs.md)
