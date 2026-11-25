# ADR-0010: TV simulcast, where the app is the game and TV runs deliberately behind it

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context
The show may also air on a TV channel. A TV broadcast carries none of the app's timed-metadata cues, and its delay varies by platform (terrestrial, cable, satellite, IPTV). If TV runs *ahead* of the app stream, TV viewers hear the host read a question before their app timer starts. That gives them extra time to think or look up the answer, which is unfair to app-only players.

The Master Plan forbids microphone monitoring, which rules out syncing phones to the TV audio.

## Decision
1. **The app is the only place to play.** Questions, timers, answers and reveals ride only in the app's IVS stream (ADR-0002). Nothing in the game depends on the TV signal.
2. **TV is always behind the app.** The TV station's feed passes through a broadcast delay set **above the app's p99 latency**: about 8–10 s to start, tuned from rehearsal measurements. The delay value lives in the producer's run-of-show settings, and on-air control belongs to the TV station.
3. **The reveal cannot leak.** The host reveals only after the server has closed the question (timer + `L_max`). Even an undelayed TV feed therefore shows the reveal after every answer window has closed. This is already enforced (SPEC §6.3).
4. **TV companion mode in the apps.** A player taps "Watching on TV". The app then mutes and minimises the video, or switches to the audio-only rendition, while still receiving the cues. This avoids double audio and saves mobile data.
5. **TV drives installs.** The TV feed carries a QR code and "play in the app" calls to action. Producers control these as vMix graphics.

## Rejected
- **Audio sync, where the phone listens to the TV:** it breaks the no-microphone rule.
- **Server-timed question push, independent of the video:** some viewers would see questions before the host reads them, which breaks fairness.

## Consequences
- No game logic changes. The work is a client mode, a broadcast-delay agreement with the station, and rehearsal checks.
- **To verify in Phase 2a:** whether cues arrive in the IVS **audio-only** rendition (UNVERIFIED). If they don't, companion mode plays the lowest video rendition, muted and hidden.
- **Legal input:** prize games on TV may also fall under Albania's Audiovisual Media Authority (AMA). This is UNVERIFIED and has been added to counsel's list.
