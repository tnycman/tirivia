# ADR-0003: Standby region is Ireland (eu-west-1), not Milan

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context
The plan has Frankfurt as primary, Milan as standby, and a 5-minute target for restarting a show. Research found that Milan cannot fill that role:

- **Amazon IVS (Low-Latency and Real-Time) and IVS Chat are not offered in eu-south-1.** The only EU regions are eu-central-1 and eu-west-1 ([AWS endpoints](https://docs.aws.amazon.com/general/latest/gr/ivs.html)). A standby in Milan could not carry video or chat.
- **Fargate in Milan launches too slowly for a cold failover.** The limits ([ECS quotas](https://docs.aws.amazon.com/general/latest/gr/ecs-service.html)) are:

  | Fargate launch limit | Milan | Frankfurt / Ireland |
  |---|---|---|
  | Tasks per minute, per service | 125 (not adjustable) | 500 |
  | Burst | 25 | not listed |
  | Sustained, per second | 5 | 20 |

  At Milan's rates, a cold failover cannot reliably finish in 5 minutes.
- Amazon Managed Grafana is not available in Milan.

## Decision
**eu-west-1 (Ireland) is the standby for every plane:**

- **Video and chat:** the IVS channel, recording configuration, an S3 bucket in the same region, chat rooms and the review Lambda are created in advance.
- **Compute:** Fargate services are pre-scaled to 30–50% of primary during the show window (DEFAULT).
- **Database:** an Aurora Global Database secondary (not headless).
- **Valkey:** either rebuilt from durable state, or run as a Global Datastore with a scripted promote.
- **Secrets and storage:** Secrets Manager secrets are replicated, and S3 cross-region replication covers KYC documents and media.
- **Failover:** Route 53 ARC routing controls.
- **Account:** the standby lives in the prod account, because Global Datastore works only within one account.

## Consequences
- Balkan players see slightly higher latency during a failover than they would from Milan.
- The warm standby costs about 30–50% of show-window compute.
- The SCP region allow-list becomes eu-central-1, eu-west-1, and us-east-1 (for global services).
