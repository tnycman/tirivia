# Research 06: Which AWS region should be the standby?

> Sources checked on 2026-11-25. Answers the owner's question: "Are there regions closer and faster than Milan or Ireland?"
> Feeds into [ADR-0003](../adr/0003-standby-region-ireland.md) and decision TS-10 in the [decision register](../decision-register.md).

## Comparison

**How to read the latency columns**
- *From Tirana* and *From Skopje* are WonderNetwork two-week averages between city servers, not AWS data centres. Treat them as a proxy.
- *To Frankfurt* is the inter-region round trip from cloudping.co. Its percentile is not stated, so treat it as indicative.
- Fargate launch rate is written as tasks per minute per service / burst / sustained per second.

| Region | IVS + IVS Chat | Fargate launch rate | Aurora Global DB | Valkey Global Datastore | Managed Grafana | FIS | From Tirana | From Skopje | To Frankfurt |
|---|---|---|---|---|---|---|---|---|---|
| **Frankfurt** (primary) | Yes | 500 / 100 / 20 | Yes | Yes | Yes | Yes | 33 ms | 33 ms | — |
| Zurich | **No** | **125 / 25 / 5** | Yes | Yes | No | Yes | 24 ms | 39 ms | 8 ms |
| **Ireland** | **Yes** | 500 / 100 / 20 | Yes | Yes | Yes | Yes | 46 ms | 56 ms | 21 ms |
| London | No | 500 / 100 / 20 | Yes | Yes | Yes | Yes | 51 ms | 50 ms | 14 ms |
| Paris | No | 500 / 100 / 20 | Yes | Yes | No | Yes | 34 ms | 43 ms | 11 ms |
| Milan | **No** | **125 / 25 / 5** | Yes | Yes | No | Yes | 24 ms | 35 ms | 12 ms |
| Spain | No | **125 / 25 / 5** | Yes | Yes | No | Yes | 64 ms (noisy) | 63 ms | 33 ms |
| Stockholm | No | 500 / 100 / 20 | Yes | Yes | No | Yes | 56 ms | 55 ms | 25 ms |
| Athens Local Zone (GA Jul 2026) | No | No Fargate | No RDS | No ElastiCache | — | — | 23 ms | 73 ms | parent region is Frankfurt |
| European Sovereign Cloud (GA Jan 2026) | UNVERIFIED | UNVERIFIED | Separate partition, not interoperable | — | — | — | — | — | 13–16 ms |

## Sources

- IVS regions: https://docs.aws.amazon.com/general/latest/gr/ivs.html
- Fargate quotas: https://docs.aws.amazon.com/general/latest/gr/ecs-service.html and https://docs.aws.amazon.com/AmazonECS/latest/developerguide/throttling.html
- Aurora Global DB: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Concepts.Aurora_Fea_Regions_DB-eng.Feature.GlobalDatabase.html
- Global Datastore: https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Redis-Global-Datastores-Getting-Started.html
- Managed Grafana: https://docs.aws.amazon.com/grafana/latest/userguide/what-is-Amazon-Managed-Service-Grafana.html
- FIS: https://docs.aws.amazon.com/general/latest/gr/fis.html
- Latency: https://wondernetwork.com/pings/Tirana, https://wondernetwork.com/pings/Skopje, https://www.cloudping.co/
- Local Zones: https://aws.amazon.com/about-aws/global-infrastructure/localzones/locations/ and https://aws.amazon.com/about-aws/whats-new/2026/07/aws-local-zone-athens-greece/
- Sovereign Cloud: https://www.aws.eu/faq/

## Findings

1. **Only Frankfurt and Ireland in Europe run IVS and IVS Chat.** No other region can host the video or chat standby.
2. **Zurich and Spain have the same low Fargate launch cap as Milan.** The per-service limit of 125 tasks/min cannot be raised, so they fail the 5-minute restart test for the same reason Milan does.
3. **Paris is the closest region with full Fargate launch rates.** It is about 34 ms from Tirana, against 46 ms for Ireland. But it has no IVS, IVS Chat or Managed Grafana, so the video, chat and dashboards would still need Ireland. That means operating two standby regions.
4. **The latency gain is small, and only applies during a failover.**
   - Ireland adds about 13 ms over Frankfurt from Tirana and about 23 ms from Skopje.
   - The answer window is 6–10 s plus a 6 s latency allowance, and the video itself runs 2–5 s behind live.
   - Video and chat are delivered from AWS's global edge locations wherever the region is, so the region choice does not change what viewers see.
5. **The Athens Local Zone cannot be the standby.** It is an extension of Frankfurt, not an independent region, and it has no Fargate, RDS or ElastiCache. It could later serve as a latency edge.
6. **The European Sovereign Cloud cannot be the standby.** It is a separate partition that cannot interoperate with Frankfurt, so cross-region replication (Aurora Global DB, Global Datastore) cannot reach it.
7. **No AWS region or Local Zone closer to Albania has been announced.**

**Conclusion:** Ireland is the only region that can be the standby for every part of the system, video and chat included. The only alternative is a split standby: backend in Paris, video and chat in Ireland. It is about 12 ms faster for the backend during a failover, at the cost of running and rehearsing two standby regions.
