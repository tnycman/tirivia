# ADR-0002: Questions and counts as signed timed metadata in the video stream

- **Status:** Accepted (Master Plan v1.1 §5), with the constraints below from [research 01](../research/01-live-video-ivs.md)
- **Date:** 2026-11-25

## Context
Every viewer watches with a different delay. If questions travel on a separate channel, viewers with less delay get more time to answer.

IVS `PutMetadata` places each cue "at the same point in the stream" for every viewer. Its limits:
- payload of 1 KB or less
- 5 TPS per channel
- no guarantee of exactly when the cue is inserted
- cues stay in recordings

## Decision
- Every gameplay cue is a `SignedCue` signed with Ed25519, sent as `t1.<base64url>`.
- Cues must fit a 900-byte budget, checked in CI.
- At most 4 TPS go out, through a single token bucket.
- **The server sets every deadline**, measured from when `PutMetadata` succeeded plus a latency allowance.
- The `pulse` cue carries enough state for a late joiner to resync.
- The correct answer is never sent before `reveal`.

## Consequences
- There are no persistent connections to scale.
- The latency allowance `L_max` has to be measured in rehearsals.
- Replay mode must ignore gameplay cues.
