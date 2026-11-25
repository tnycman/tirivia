# Protocol test vectors: the shared conformance suite

Master Plan §5 calls for "one shared definition and conformance test suite for both clients and the server".

`connectrpc/conformance` is **not** that suite. It tests whether a Connect, gRPC or gRPC-Web *implementation* follows the protocol; it never exercises our own API ([research 02](../../docs/research/02-backend-protocol-data.md)). We still run it against the Go server, but our own API needs its own suite, defined here.

## What a vector is

Each vector is a directory with:

| File | Meaning |
|---|---|
| `input.json` | The message in canonical proto3 JSON |
| `expected.binpb` | The exact binary encoding |
| `meta.yaml` | Message type, description, expected outcome (`valid`, `reject:<reason>`), and for cues: key id, signature, max size |

## Suites

| Folder | What every implementation (Go, Swift, Kotlin) must do |
|---|---|
| `cue/` | Decode `t1.<base64url>` → `SignedCue`. Verify the Ed25519 signature against the test key. Reject unsigned, wrong-key, tampered, wrong-show and replayed (`seq` ≤ last) cues. **Assert the encoded text cue is 900 bytes or less** when built from maximum-length question content (240-byte question, 3 × 64-byte options, every character a 2-byte UTF-8 letter such as `ë`/`ç`; the Go reference implementation measures 845 B). |
| `answer/` | Build a `SubmitAnswerRequest` from a shuffled option list using `shuffle_seed` and check that the same `option_id` is submitted on every platform (the shuffle algorithm is specified in SPEC §6.3). |
| `errors/` | Map each Connect error code and `AnswerResult` to its user-facing state. |

## Runners

- Go: `backend/internal/platform/cuesign/cuesign_test.go` reads `../../protocol/testvectors/**`
- Swift: `ios/Packages/Core/Tests/CoreProtocolTests` (XCTest resources)
- Kotlin: `android/core/protocol/src/test` (JUnit, resources copied by Gradle task)

CI fails if any runner skips a vector (the vector count is asserted).

The generator script `protocol/scripts/gen-testvectors.go` (Phase 0 task) produces the vectors from a deterministic test key, which is never a production key.
