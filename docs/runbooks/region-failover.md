# Runbook: region failover from Frankfurt to Ireland

**Target:** restart the show within **5 minutes** (DEFAULT). **Standby:** eu-west-1, Ireland ([ADR-0003](../adr/0003-standby-region-ireland.md)). Milan cannot serve as the standby because it has no IVS or IVS Chat.

## Pre-conditions (already in place during the show window)

- **Compute:** the Ireland Fargate services are warm-scaled to 30–50% (DEFAULT).
- **Video and chat:** the IVS channel and chat rooms are pre-created in Ireland.
- **Database:** an Aurora Global Database secondary that is not headless.
- **Secrets:** Secrets Manager secrets are replicated to Ireland.
- **Traffic:** Route 53 ARC routing controls are configured.

## Steps

1. **Decide.** The incident commander confirms the Frankfurt impairment from AWS Health, the dashboard and synthetic checks. Do not fail over for an encoder-only problem; that case has its own [runbook](encoder-failover.md).
2. **Data.** Run the Aurora Global Database **managed failover** (or a switchover, if Frankfurt is still reachable) to the Ireland cluster.
3. **Live state.** Promote the Valkey Global Datastore with the scripted step, or rebuild live state from the last cue and the Kinesis stream.
4. **Traffic.** Flip the ARC routing control so that API DNS points to Ireland. ARC operates on the data plane only.
5. **Video.** Point both encoders at the **Ireland channel's** ingest endpoint and stream key. Then set the show's active channel in `shows.ivs_channel_arn` and `shows.region`.
6. **Apps.** No app release is needed. Apps call `GetShowState` or `JoinShow`, receive the Ireland `playback_url` and chat endpoint, and resync from the next `pulse`.
7. **Resume.** The producer resumes at the last completed REVEAL and re-releases any question that was open during the failure.
8. **Afterwards.** Fail back during a quiet window. Hold a post-incident review.

**Rehearse this quarterly** as an FIS game day (`route-table-disrupt-cross-region-connectivity` plus an Aurora failover).
