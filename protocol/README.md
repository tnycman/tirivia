# protocol: the shared contract

This is the single definition of every API message and of the timed-metadata cues. iOS, Android, the Go backend and the web consoles all generate their code from it. See [`stack/01-protocol`](../stack/01-protocol/README.md) and [ADR-0007](../docs/adr/0007-connect-rpc.md).

```
protocol/
├── buf.yaml                 # Buf v2: lint STANDARD, breaking FILE
├── buf.gen.yaml             # Go (local), Swift, TS (remote BSR plugins)
├── buf.gen.go.yaml          # Go only, offline (backend development)
├── buf.gen.kotlin.yaml      # Android, pinned remote plugins, into android/core/protocol/src/gen
├── proto/trivia/
│   ├── common/v1/           # Money, Eligibility, ReactionKind
│   ├── cue/v1/              # SignedCue / CueBody: question_open, reveal, pulse (incl. chat_enabled), burst, final
│   ├── game/v1/             # GameService (hot path)
│   ├── core/v1/             # AuthService, ProfileService, ContentService, WalletService
│   └── events/v1/           # Kinesis event envelope (written as proto3 JSON)
├── gen/                     # generated code (Go module trivia.al/protocol/gen)
├── testvectors/             # cross-platform conformance corpus (see its README)
└── scripts/install-tools.sh # pinned buf + plugins
```

## Commands

| Task | Command |
|---|---|
| Install tools | `sh scripts/install-tools.sh` (buf 1.75.0, protoc-gen-go 1.36.12, protoc-gen-connect-go 2.0.2) |
| Lint | `buf lint` |
| Format check | `buf format -d` |
| Breaking-change check | `buf breaking` against the PR base commit; a deliberate break needs the visible `buf skip breaking` label (docs/research/07). Used once, pre-release, for the core_services.proto split |
| Generate Go | `buf generate --template buf.gen.go.yaml` |
| Generate Kotlin (Android) | `buf generate --template buf.gen.kotlin.yaml` (network; CI fails if the committed output is stale) |
| Generate all | `buf generate` (needs network access for the remote Swift and TS plugins) |

## Verified on 2026-11-25

- `buf lint` passes with no findings.
- `buf format -d` shows no diff.
- The generated Go code builds with connect-go v2.0.2.
- The worst-case `question_open` cue measures **845 bytes**, inside the 900-byte budget (`backend/internal/platform/cuesign`).

## Rules

- Never renumber or reuse a field. When deleting one, `reserved` both its number and its name.
- Never put the correct answer in any message a client receives before `reveal`.
- Every new RPC or cue type ships with test vectors in the same pull request.
