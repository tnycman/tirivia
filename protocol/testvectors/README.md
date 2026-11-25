# Protocol test vectors: the shared conformance suite

Master Plan §5 asks for "one shared definition and conformance test suite for both clients and the server". This folder is that suite: language-neutral JSON files. Every platform runs the same files, and CI fails if any platform disagrees.

`connectrpc/conformance` is a different thing. It tests whether a Connect protocol *implementation* behaves correctly, and never exercises our own API ([research 02](../../docs/research/02-backend-protocol-data.md)). We run it against the Go server as well.

## Files

| File | Contents |
|---|---|
| `cue_vectors.json` | 14 signed timed-metadata cues with the outcome every client must produce (see below) |
| `shuffle_vectors.json` | 16 per-player option-order cases (SPEC §6.3.7), including a non-ASCII question id and the high-byte case that catches Kotlin's signed-byte pitfall |
| `runners/node/verify-vectors.mjs` | An independent implementation using only Node built-ins. It proves the corpus does not depend on Go behaviour |

**The keys are TEST ONLY.** They are derived from a public seed and must never be used in any environment.

### Cue vectors

The 14 cues cover:

- **Valid:** a Albanian-language question, the worst-case 845 B question, a pulse, a reveal.
- **Valid, but resync needed:** a gap in the sequence numbers (`valid:gap_resync`).
- **Must be rejected:**
  - unsigned JSON
  - bad base64
  - wrong prefix version
  - tampered payload
  - forged signature under the right key id
  - unknown key id
  - oversize
  - a cue for another show
  - a replayed sequence number

## Runner algorithm (identical on every platform)

1. **Verify** the text cue with `public_keys`. Map each failure to its expectation value:

   | Check | Expectation on failure |
   |---|---|
   | Size is at most `max_text_bytes` | `reject:too_large` |
   | Prefix is `t1.`, the rest is base64url, and it decodes as a `SignedCue` | `reject:malformed` |
   | `key_id` is in `public_keys` | `reject:unknown_key` |
   | Ed25519 signature over the exact `body` bytes | `reject:bad_signature` |
   | `body` decodes as a `CueBody` | `reject:malformed` |

2. **Check the sequence** against `context`:
   - `show_id` differs from `expected_show_id`: `reject:wrong_show`
   - `seq` is at most `last_seq`: `reject:replay`
   - `seq` is greater than `last_seq + 1`: `valid:gap_resync`. Apply the cue, then call `GetShowState`.
   - Otherwise: `valid`.
3. **Compare fields.** For valid vectors, compare `fields` against the decoded body using the canonical names. The full list is in `fields()` in `backend/internal/platform/cuesign/vectors_test.go`.

An `expect` value of `a|b` means either outcome is acceptable. For example, tampering may break the framing or only the signature.

## Runners

| Platform | Location | Status |
|---|---|---|
| Go (reference) | `backend/internal/platform/cuesign/vectors_test.go`, `backend/internal/platform/shuffle/vectors_test.go` | Passing |
| Node (independent) | `runners/node/verify-vectors.mjs` | Passing. A mutation test confirmed it fails on a corrupted vector |
| Swift | `ios/Packages/Core/Tests/CoreProtocolTests/VectorTests.swift` (shuffle). CI copies the JSON into `Vectors/` | Written; compiled in macOS CI (no Xcode on this host). Cue runner arrives with the iOS cue verifier in Phase 2a |
| Kotlin | `android/core/protocol/src/test/.../OptionShuffleVectorTest.kt` (shuffle) | Written; compiled in CI (no JDK on this host). Cue runner arrives with the Android cue verifier in Phase 2a |

Every runner asserts a minimum vector count, so a runner that silently skips vectors fails.

## Regenerating

```sh
cd backend
go run ./cmd/testvectors          # rewrite the files (deterministic: Ed25519 + deterministic protobuf)
go run ./cmd/testvectors -check   # CI: fails if the committed files are stale
```

A new cue kind or RPC ships with its vectors in the same pull request (build standards item 5).
