# 02 Live video, studio chain and chat

## Role
Carries the live show from camera to phones with low latency. Questions and live counts travel inside the stream as signed timed metadata. Every show is recorded so it can be replayed and cut into clips. Chat runs alongside the stream.

## Decision
- **Amazon IVS Low-Latency**, using an **ADVANCED_HD** channel. The top rendition is 720p30, and IVS transcodes the lower renditions on the server.
- Timed metadata goes in with **PutMetadata** ([ADR-0002](../../docs/adr/0002-questions-in-video-stream.md)).
- Studio chain:
  - Primary ingest: **vMix/OBS → SRT**, connection priority 1.
  - Backup ingest: **RTMPS** on a second internet line, connection priority 2.
- Recording: IVS auto-record to S3, served through CloudFront with OAC.
- Chat: **Amazon IVS Chat** with a review Lambda. Chat is switched on and off from the backend ([ADR-0009](../../docs/adr/0009-chat-backend-switch.md)).
- Regions:
  - Primary: Frankfurt.
  - Standby: **Ireland**, with the channel and chat created in advance ([ADR-0003](../../docs/adr/0003-standby-region-ireland.md)).

## Pinned versions (2026-11-25)

| Item | Version | Source |
|---|---|---|
| IVS Player iOS | 1.58.1 (SPM local `Package.swift`; CocoaPods dropped Feb 2026) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/ios-getting-started.html |
| IVS Player Android | `com.amazonaws:ivs-player:1.58.1` | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/android-getting-started.html |
| Channel | `type=ADVANCED_HD`, `latencyMode=LOW`, `insecureIngest=false` | https://docs.aws.amazon.com/ivs/latest/LowLatencyAPIReference/API_CreateChannel.html |

## Hard limits (verified)

| Limit | Value | Adjustable |
|---|---|---|
| PutMetadata payload | 1 KB | No |
| PutMetadata rate | 5 TPS per channel, 155 per account | No |
| Concurrent views per region | 15,000 by default | Yes, increase must be requested |
| Chat connections per region | 50,000 by default | Yes |
| CreateChatToken | 200 TPS | No |
| Chat SendMessage per room | 100/s | No |
| Chat review Lambda timeout | 200 ms | No |
| Recording reconnect window | 0–300 s | — |

Source: https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/service-quotas.html and https://docs.aws.amazon.com/ivs/latest/ChatUserGuide/service-quotas.html

## Best practices
1. **Encoder profile.** Both encoders must use identical settings, otherwise stream takeover fails:
   - H.264 Main, 720p30, CBR
   - 2 s keyframe interval
   - scene-change detection **off**
   - VBV no larger than the average bitrate
   - **B-frames 0**
   - BT.709 colour
   - AAC-LC 48 kHz

   ([streaming config](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/streaming-config.html))
2. **Failover.** IVS Low-Latency has **no automatic ingest failover** ([resilience](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/security-resilience.html)). Instead:
   - Use stream takeover by `priority`.
   - Enable the failover slate.
   - Set `recordingReconnectWindowSeconds` to 90 so one show stays one recording.
3. **SRT latency setting:** at least 4× the round-trip time ([vMix](https://www.vmix.com/help23/SRT.html)). SRT uses port 9000 with a passphrase.
4. **Studio network.** Wired connections, 50% upload headroom, encoders on a dedicated VLAN, and no third-party relay in the path.
5. **Cues stay small.** Send IDs and short text, and fetch anything larger separately. AWS's own quiz sample puts the correct answer in metadata; **we do not.**
6. **Cue delivery.** All cues go through one token bucket of 4 TPS or less per channel. On 404 or 429, retry with backoff.
7. **Recordings.** Read file paths from `recording-ended.json`. Start processing only after the EventBridge Recording End event.
8. **Quotas.** Request quota increases early. Alarm on `ConcurrentViews` reaching 80% of the quota.

## Rules for this codebase
- All server deadlines are measured from `T_cue`, the time PutMetadata succeeded, never from the client's clock.
- The app reads `playback_url` and the chat endpoint from `JoinShow`. **Never hard-code a channel or region.**
- In replay or VOD mode the app ignores `question_open`, `reveal` and `final` cues.
- The chat panel's visibility follows `Pulse.chat_enabled` within one pulse.

## Changes vs Master Plan
- Standby region moves from Milan to Ireland, because IVS is not offered in Milan.
- Use ADVANCED_HD, not multitrack. Multitrack needs Enhanced RTMP, which is incompatible with our SRT ingest.
- Backup encoder failover is manual, by priority takeover, and has a runbook.
- New: chat cost line, room sharding, and lazy connection to chat.
- Wording corrections: Low-Latency is not WebRTC, and IVS Real-Time defaults to 10k viewers, with up to 25k on request.

## Must be verified in Phase 2a (UNVERIFIED)
- How cues behave across rendition switches.
- Cue behaviour on late join.
- Measured cue-to-render p99, which sets `L_max`.
- Which IVS channel types accept SRT.
- The range of the failover slate parameters.

## Code
- `backend/internal/studio` (cue signer and PutMetadata)
- `backend/internal/chatmod`
- `infra/modules/ivs-channel`
- [`docs/runbooks/`](../../docs/runbooks/)

## Sources
[docs/research/01-live-video-ivs.md](../../docs/research/01-live-video-ivs.md)

## TV simulcast (ADR-0010)
- The app stream is the only game clock. The TV feed is a watch-only simulcast with a **broadcast delay above the app p99 latency** (~8–10 s to start, tuned in rehearsal), so TV never runs ahead of an app player.
- The reveal cannot leak, because the host reveals only after the server closes the question.
- Apps offer a "Watching on TV" companion mode: muted and minimised video, or the audio-only rendition, still receiving cues.
- No microphone-based sync (forbidden by the Master Plan).
- **To verify in Phase 2a:** cues in the audio-only rendition.
