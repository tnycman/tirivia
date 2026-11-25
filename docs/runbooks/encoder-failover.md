# Runbook: primary encoder failure

**Background.** IVS Low-Latency has **no automatic ingest failover** ([research 01](../research/01-live-video-ivs.md)). Continuity relies on three things:
- stream takeover by priority
- the failover slate
- the recording reconnect window

**Setup (prepared in Phase 2b, re-checked at T-45)**

| Item | Setting |
|---|---|
| Primary | vMix → **SRT** `srt://<ingest>:9000?streamid=#!::u=<key>,priority=1&passphrase=<p>` on line A |
| Backup | OBS or second encoder → **RTMPS** `rtmps://<ingest>:443/app/<key>?priority=2` on line B |
| Both | Identical H.264 Main 720p30 CBR, 2 s GOP, B-frames 0, AAC-LC 48 kHz. Takeover fails if resolution, codecs or track count differ |
| Channel | Failover slate enabled. `recordingReconnectWindowSeconds` = 90 |

**Steps**

1. **Detect.** Any of these signals a failure:
   - the IVS stream-health alarm or EventBridge "Stream Failover Slate Start"
   - the primary vMix shows a disconnect
   - viewers see the slate
2. **Hold the game.** The producer keeps the show in its current phase:
   - Do not release a question while the stream is down. PutMetadata returns `404 ChannelNotBroadcasting`.
   - If a question is open, studio extends nothing. The server deadline is already anchored on `T_cue`.
3. **Take over.** Start or unmute the backup encoder. Because it has higher priority, IVS switches to it. Viewers may buffer briefly and do not need to reload.
4. **Verify.** Watch the stream on the reference devices. Send a test `pulse` and confirm it renders.
5. **Resume.** Continue the run of show.
6. **After the show.** Write an incident note covering:
   - timeline
   - root cause (line or encoder)
   - whether the recording merged within the 90 s window

**Escalation.** If both encoders fail for more than 2 minutes, go to [region-failover](region-failover.md) only if AWS itself is impaired. Otherwise announce a delay and restart the show (target: within 5 minutes).
