# Tech stack, by function

Each folder covers one function of the system. Every folder README follows the same layout:

1. **Role**: what this layer does in the product
2. **Decision**: what we use, and the Master Plan or ADR reference
3. **Pinned versions**: checked on 2026-11-25, each with its source
4. **Best practices**: numbered, each cited to a primary source
5. **Rules for this codebase**: the conventions reviewers enforce
6. **Changes vs Master Plan**: where the research changed something
7. **Code location**
8. **Sources**: links to the full research report

| # | Function | Core choice | Code |
|---|---|---|---|
| [01](01-protocol/) | Protocol / API contract | Protobuf + Connect RPC, Buf v2 | `protocol/` |
| [02](02-live-video/) | Live video, studio, chat | Amazon IVS Low-Latency (ADVANCED_HD), IVS Chat | `infra/modules/ivs-*`, `backend/internal/studio`, `docs/runbooks` |
| [03](03-backend-go/) | Backend services | Go 1.27 modular monolith (`game`, `core`) | `backend/` |
| [04](04-data/) | Data | Aurora PostgreSQL 17, Valkey 9.1, Kinesis → Firehose → S3 → Athena | `backend/migrations`, `infra/modules/*` |
| [05](05-auth-identity/) | Auth and identity | Own OTP, App Attest / Play Integrity, EdDSA JWT + DPoP | `backend/internal/auth` |
| [06](06-ios/) | iOS app | Swift 6 mode, SwiftUI, iOS 16+ | `ios/` |
| [07](07-android/) | Android app | Kotlin 2.4, Compose, minSdk 26 | `android/` |
| [08](08-web-consoles/) | Web consoles and website | React 19 + Vite SPAs behind a BFF; Astro | `web-consoles/` |
| [09](09-infrastructure-aws/) | Cloud infrastructure | AWS multi-account, Terraform 1.17, Fargate ARM64 | `infra/` |
| [10](10-ci-cd/) | CI/CD and supply chain | GitHub Actions + OIDC, lockfiles, attestations | `.github/` |
| [11](11-observability/) | Observability | OpenTelemetry, ADOT, CloudWatch/AMP/Grafana, Sentry | cross-cutting |
| [12](12-security/) | Security baseline | ASVS 5.0, MASVS 2.1, API Top 10 2023 | cross-cutting |
| [13](13-anti-cheat/) | Anti-cheat | 4 layers + risk scoring | `backend/internal/risk`, apps |
| [14](14-testing-quality/) | Testing and quality | Test vectors, k6, FIS, Macrobenchmark | everywhere |
| [15](15-compliance-privacy/) | Compliance and privacy | Law 124/2024, GDPR, store rules | `docs/` |

The full evidence behind each choice is in [`docs/research/`](../docs/research/).
