# Research 01: Live video (Amazon IVS), studio chain, chat

> Verified against primary sources on **2026-11-25**. Anything we could not confirm is marked **UNVERIFIED**; treat it as a test item, not a fact.
> Feeds into: [`stack/02-live-video`](../../stack/02-live-video/README.md), [`SPEC.md`](../../SPEC.md) §5 and §7, [ADR-0003](../adr/0003-standby-region-ireland.md).

## 1. Verified facts and limits

| Fact | Value | Source |
|---|---|---|
| PutMetadata payload | **1 KB max**, not adjustable (whether 1000 or 1024 bytes is UNVERIFIED, so budget 900 B or less) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/service-quotas.html |
| PutMetadata rate | **5 TPS per channel, 155 TPS per account**, not adjustable. AWS: "batch your data into a single PutMetadata call" | https://docs.aws.amazon.com/ivs/latest/LowLatencyAPIReference/API_PutMetadata.html |
| PutMetadata errors | `ChannelNotBroadcasting` (404) when offline; `ThrottlingException` (429) | same |
| Delivery timing | Inserted "as soon as possible… no guarantee as to when… however, all viewers see the metadata at the same point in the stream". Carried as ID3 in segments, so it is also kept in recordings | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/metadata.html |
| Player callbacks | iOS `player(_:didOutputCue:)` → `IVSTextMetadataCue.text`; Android `onCue(Cue)` → `TextMetadataCue.text`. They fire for **live and recorded** content | same |
| Callback thread | iOS: main queue. Android: main thread. Players are not thread-safe | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/ios-getting-started.html · https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/android-getting-started.html |
| Late join or rejoin | A cue fires only when playback reaches its segment, so past cues are missed. **UNVERIFIED (inferred from the mechanism)** | metadata.html |
| Cues across rendition switches | **UNVERIFIED.** Test in phase 2a | — |
| Concurrent views | **15,000 per region by default**, adjustable | service-quotas.html |
| Concurrent streams / channels | 100 / 5,000 per region (adjustable). 1 stream key per channel | same |
| Channel types | STANDARD (≤1080p, 8.5 Mbps in; multitrack 15 Mbps); **ADVANCED_HD (720p top, server transcode)**; ADVANCED_SD (480p top); BASIC (transmux, single rendition) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/channel-types.html |
| ADVANCED_HD ladder (30 fps) | HIGHER preset: 720p 2.3 Mbps / 480p 1.3 / 360p 0.7 / 160p 0.27 / audio-only. CONSTRAINED: 1.9 / 0.8 / 0.4 / 0.22 | same |
| Advanced channel keyframes | IVS sets them itself, with at most 2 s between keyframes | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/player.html |
| Multitrack (client-side ABR) | STANDARD only; Enhanced RTMP + OBS 30.2+ + NVIDIA/AMD GPU. **Not possible over SRT** | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/multitrack-video-setup.html |
| Ingest | RTMPS 443 (TLS 1.2+), RTMP, **SRT on port 9000** with passphrase. H.264 + AAC-LC | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/streaming-config.html |
| SRT launch | 5 Apr 2024 | https://aws.amazon.com/about-aws/whats-new/2024/04/amazon-ivs-low-latency-streaming-srt-ingest |
| Automatic ingest failover | **None** for Low-Latency: "streamers should configure their encoders… to automatically re-ingest" | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/security-resilience.html |
| Stream takeover | `?priority=N` (RTMPS) or `streamid=#!::u=<key>,priority=N` (SRT). The higher priority wins. Both streams need **identical resolution, codecs and tracks**. The recording has no break | streaming-config.html |
| Failover slate (26 Oct 2026) | Default animated slate on an unexpected drop, with EventBridge start/end events. The docs disagree on parameter ranges (UNVERIFIED) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/failover-slate.html |
| Redundant ingest | **Real-Time stages only** (Apr 2026) | https://aws.amazon.com/about-aws/whats-new/2026/04/amazon-ivs-real-time-streaming-redundant-ingest |
| **IVS Low-Latency regions** | us-east-1, us-west-2, ap-south-1, ap-northeast-2, ap-northeast-1, **eu-central-1, eu-west-1**. **Not eu-south-1 (Milan)** | https://docs.aws.amazon.com/general/latest/gr/ivs.html |
| **IVS Chat regions** | The same 7. **Not Milan** | same |
| Control plane vs data plane | Control plane is regional; ingest and playback are global (edge PoPs) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/what-is.html |
| Latency | Low-Latency: under 5 s with the IVS player. Real-Time: under 300 ms | same |
| Keyframe and startup | 2 s keyframe gives about 6–7 s start; 1 s gives about 3–4 s | streaming-config.html |
| Auto-record layout | `ivs/v1/<acct>/<channel>/<y>/<m>/<d>/<h>/<min>/<rec_id>/` with `events/recording-*.json`, `media/hls/…`, `media/thumbnails`. **Read paths from the JSON; do not hard-code them** | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/record-to-s3.html |
| Recording reconnect window | 0–300 s (default 0). The merge needs the same resolution, fps and codec | https://docs.aws.amazon.com/ivs/latest/LowLatencyAPIReference/API_CreateRecordingConfiguration.html |
| Clips | No Low-Latency clip API found. Use byte-range playlists or MediaConvert post-processing | record-to-s3.html · https://aws.amazon.com/blogs/media/awse-using-amazon-ivs-and-mediaconvert-in-a-post-processing-workflow |
| Recording cost | No IVS charge; S3 and CloudFront only | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/costs.html |
| Chat review handler | Lambda in the **same account and region**, **200 ms timeout (fixed)**, `fallbackResult` ALLOW/DENY | https://docs.aws.amazon.com/ivs/latest/ChatUserGuide/chat-message-review-handler.html |
| Chat quotas | **50,000 connections per account per region** (adjustable); SendMessage **100/s per room (fixed)**; **CreateChatToken 200 TPS (fixed)** | https://docs.aws.amazon.com/ivs/latest/ChatUserGuide/service-quotas.html |
| Chat token | session 1–180 min (default 60) | https://docs.aws.amazon.com/ivs/latest/ChatAPIReference/API_CreateChatToken.html |
| Real-Time viewer cap | 10,000 per stage by default; **up to 25,000 on request** | https://docs.aws.amazon.com/ivs/latest/RealTimeUserGuide/service-quotas.html |
| Input price per hour | ADVANCED_HD **$0.85**; STANDARD $2.00 | https://aws.amazon.com/ivs/pricing/ |
| Europe output, first 10k h | SD **$0.036**, HD **$0.072**, FHD $0.144 per viewer-hour, falling with volume (over 500k h: 0.024/0.048/0.096) | same |
| Chat price | $0.56 per 1k sent; **$0.008 per 1k delivered** | same |
| Player SDK | **iOS 1.58.1 / Android 1.58.1** (9 Nov 2026) | https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/doc-history.html |
| Player minimum OS | iOS 15+, Android 6.0+ | player.html |
| Distribution | iOS: SPM (local `Package.swift` from player.live-video.net) or xcframework; **CocoaPods dropped Feb 2026**. Android: `com.amazonaws:ivs-player` on Maven Central | getting-started pages |
| iOS background playback | Not supported; pause when the app goes to the background | ios-getting-started.html |

## 2. Best practices

1. **Encoder:** H.264 Main, CBR, 2 s keyframe interval, scene-change detection off, VBV no larger than the average bitrate, zerolatency tune, BT.709, AAC-LC at 48 kHz. 720p at up to 4.5 Mbps. ([streaming-config](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/streaming-config.html))
2. **B-frames = 0.** Real-Time RTMP rejects B-frames, so keep the profile portable. ([rt-rtmp-publishing](https://docs.aws.amazon.com/ivs/latest/RealTimeUserGuide/rt-rtmp-publishing.html))
3. **Studio network:** wired connections, 50% upload headroom, dedicated VLAN, no third-party relay.
4. **SRT latency:** at least 4 times the RTT (vMix guidance). ([vMix SRT](https://www.vmix.com/help23/SRT.html), [OBS SRT](https://obsproject.com/wiki/Streaming-With-SRT-Or-RIST-Protocols))
5. **Cues stay small.** Send IDs and short text in metadata and fetch anything large. AWS sample quiz demos: [iOS](https://github.com/aws-samples/amazon-ivs-player-ios-sample), [Android](https://github.com/aws-samples/amazon-ivs-player-android-sample), [blog](https://dev.to/aws/creating-interactive-live-streaming-experiences-using-timed-metadata-with-amazon-ivs-2kp6).
6. **Quotas:** request the concurrent-view and chat-connection increases early, and alarm on `ConcurrentViews`.
7. **Recordings:** process only after the EventBridge Recording End event. Serve from a private S3 bucket through CloudFront with OAC.
8. **Audio-only rendition** comes with ADVANCED channels and acts as a weak-network fallback.

## 3. Conflicts with Master Plan v1.1 and the fix adopted

| Plan says | Docs say | Adopted fix |
|---|---|---|
| Milan (eu-south-1) standby | **IVS and IVS Chat are not available in Milan** | **Standby region = Ireland (eu-west-1)**: [ADR-0003](../adr/0003-standby-region-ireland.md). The app reads its playback URL and chat ARN from `/shows/{id}/join`, never from constants. |
| "IVS Low-Latency over WebRTC" (wording) | Low-Latency is HLS plus ID3; WebRTC is the Real-Time product | Wording corrected in SPEC. The decision itself stands. |
| Real-Time "caps at 25,000" | 10k by default, 25k per stage on request | Wording corrected. The decision stands. |
| Backup encoder (implied auto failover) | No automatic ingest failover in Low-Latency | Primary SRT `priority=1`, backup RTMPS `priority=2`, **identical encode params**. Failover slate on. `recordingReconnectWindowSeconds` set to 60–120. Runbook: [`encoder-failover`](../runbooks/encoder-failover.md). |
| SRT plus "adaptive lower renditions" | Client-side ABR needs E-RTMP and so cannot use SRT | **ADVANCED_HD channel** (server transcode, $0.85/h). Test HIGHER vs CONSTRAINED presets. |
| Every viewer gets the cue at once | Insert time is not guaranteed; late joiners miss past cues | `pulse` carries `{show_id, seq, phase, question_id, remaining_ms}` so a client can resync. **The server deadline is anchored on PutMetadata success plus the latency allowance plus the timer**, not on the client clock. Cues are idempotent by `seq`. Retry on 404 or 429. |
| Pulse and other cues | 5 TPS per channel | One shared token bucket of 4 TPS or less per channel; burst events are combined into a single cue. |
| "Well under 1 KB" | The limit includes the signature | **Budget 900 B or less**, with a CI test on the encoded size. Ed25519 signature (64 B raw). |
| Replays | Cues fire again on VOD | Replay mode ignores gameplay cues. |
| 300k/600k viewers | 15k views per region by default | File a quota request for **at least 600k in eu-central-1 and eu-west-1**. |
| IVS Chat for 300k | 50k connections; tokens limited to 200 TPS (300k tokens take about 25 min) | Chat is built in full, toggled from the back office, **default OFF**, and its production use is a go/no-go decision ([ADR-0009](../adr/0009-chat-backend-switch.md)). When it is on, a player connects only when they open the chat panel, sessions last 180 min, rooms are sharded, and the review Lambda has provisioned concurrency. |
| Chat cost (missing from the plan) | $0.008 per 1k delivered, so one message to 300k costs about $2.40 | **Sharded rooms of about 5k or less and a curated/sampled host feed.** Cost line added to SPEC §12. |
| "$0.036 SD, about double HD" | Confirmed: $0.036 / $0.072 | Kept. 300k × 0.5 h at HD first-tier is about $9.4k per show (upper bound). |

## 4. Versions

- IVS Player iOS **1.58.1** (SPM local package), Android **1.58.1** (`com.amazonaws:ivs-player`). Pin both and follow the [release notes](https://docs.aws.amazon.com/ivs/latest/LowLatencyUserGuide/release-notes.html) (about one release a month).
- Channel: `type=ADVANCED_HD`, `latencyMode=LOW`, `insecureIngest=false`, recording config and failover slate attached.
- vMix and OBS versions: UNVERIFIED; record the exact versions in `docs/runbooks/studio-chain.md` at phase 2b.

## 5. What could not be verified

- Cue behaviour across rendition switches and on late join (test in phase 2a).
- Which IVS channel types accept SRT (the docs state no restriction).
- The failover-slate parameter range.
- Whether IVS output tiers pool across resolutions.
