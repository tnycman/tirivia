# 11 Observability

## Role
Watch a 20-minute live show as it happens and catch problems before players feel them. Afterwards, explain every incident.

## Decision
- **OpenTelemetry** in the Go services, exporting through an **ADOT collector** sidecar.
- Backend telemetry lands in **CloudWatch**, **Amazon Managed Prometheus** and **Managed Grafana**.
- **Sentry** reports crashes and performance from the mobile apps.
- One **live-show dashboard** and on-call alerts (Master Plan §8).

## Pinned versions (2026-11-25)

| Tool | Version | Source |
|---|---|---|
| OTel Go | v1.47.0 (traces, metrics and logs stable) | https://github.com/open-telemetry/opentelemetry-go/releases |
| OTel semantic conventions | v1.45.0 (HTTP and DB spans stable) | https://opentelemetry.io/docs/specs/semconv/ |
| ADOT collector | v0.51.0 (pin the image tag) | https://github.com/aws-observability/aws-otel-collector/releases |
| Sentry Cocoa / Android | 9.35.0 / 8.65.1 | GitHub releases |

## Best practices
1. **ADOT sidecar:**
   - Configuration comes from SSM through `AOT_CONFIG_CONTENT`.
   - The app container starts after the collector (`dependsOn: START`).
   - Pin the image version.
   - Source: [ADOT ECS](https://aws-otel.github.io/docs/setup/ecs).
2. **Logs:**
   - Write structured JSON to stdout, collected into CloudWatch.
   - The OTLP log bridge is optional. The official pages disagree on whether the Go logs SDK is stable.
3. **Metrics:**
   - Use the semantic-convention names for HTTP and DB.
   - Use custom metrics for game events: `answers_accepted_total`, `cue_putmetadata_latency_ms` and the like.
4. **Cardinality:**
   - Never label a metric with user_id or question_id.
   - Exemplars and traces carry those IDs instead.
5. **Dashboard placement:**
   - Host it in Frankfurt.
   - Keep a **CloudWatch cross-region dashboard** as the fallback, because Managed Grafana is not available in every region.
6. **Mobile:**
   - Scrub PII in Sentry (`beforeSend`).
   - Do not record session replay on screens that show PII or the wallet.

## Live-show dashboard (SPEC §11)

| Group | Signals |
|---|---|
| Players | joins/s, players remaining, viewer count (server-side), IVS `ConcurrentViews` vs quota |
| Game | answers/s, acceptance rate by `AnswerResult`, answer ack p50/p99, tally/recount mismatches (`live.Verify`) |
| Cues | PutMetadata latency and errors (404/429), cue size, seq gaps reported by apps |
| Engage | reactions/s, capped batches, burst meter |
| Chat | switch state, connections vs quota, review Lambda p99 and fallbacks |
| Platform | Fargate tasks vs desired, ALB 5xx and `PeakLCUs`, Valkey CPU per shard, RDS Proxy pinned sessions, KDS throttles |
| Auth | OTP sent, verified, **conversion rate** (pumping alarm), attestation failures |
| Quotas | Fargate vCPU, IVS, chat, Play Integrity daily use |

## Alerts (DEFAULT thresholds, tuned in rehearsal)
- Answer ack p99 above 300 ms for 30 s.
- PutMetadata error rate above 1%.
- Any quota above 80%.
- OTP conversion rate below 35%, the Vonage reference value.
- Seq-gap reports above 0.5% of players.

## Code
- `backend/internal/platform/otel`
- `infra/modules/observability`

## Sources
[docs/research/02-backend-protocol-data.md](../../docs/research/02-backend-protocol-data.md), [docs/research/03-aws-infra-devops.md](../../docs/research/03-aws-infra-devops.md)
