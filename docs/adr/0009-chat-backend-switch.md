# ADR-0009: Chat ships at launch and is switched on or off from the backend

- **Status:** Accepted (owner instruction, 2026-11-25)
- **Date:** 2026-11-25

## Context
The owner's instruction (2026-11-25):

- Build chat in full.
- Toggle it on and off from the back-office admin.
- Keep it as a working function for testing.
- If production launches without chat, that is fine.

The Master Plan already lists chat as a kill switch.

IVS Chat limits ([research 01](../research/01-live-video-ivs.md)):
- `CreateChatToken`: 200 TPS, fixed.
- Connections: 50k per region by default.
- Messages: 100 per second per room, fixed.
- Review Lambda timeout: 200 ms.
- Cost: $0.008 per 1,000 messages delivered.

## Decision
**The switch.** A `feature_switches` key, `chat.enabled`, scoped globally or per show. Operators toggle it from the studio console, and every change is audit-logged.

**How apps learn the state:**
- At join: `JoinShowResponse.chat_enabled`.
- During the show: **`Pulse.chat_enabled`**, applied within 5 s. No app release or rejoin needed.

**When chat is OFF:**
- The server mints no tokens.
- `GetChatToken` returns `FAILED_PRECONDITION`.
- Apps disconnect and hide the chat panel.
- The review Lambda falls back to DENY.

**When chat is ON:**
- The chat panel is visible.
- An app connects only when the user opens the panel.
- Rooms are sharded at 5k members or fewer.
- Slow mode is set per room.

**Default and launch.**
- The switch defaults to **OFF** in code (`switches.defaults`). If no database row exists, chat is off.
- Chat is turned on for testing and rehearsals (Phase 3 exit test).
- Production use is decided at the Phase 7 go/no-go. Launching without chat needs no code change.

## Consequences
- While the switch is off, chat costs nothing and uses no quota.
- When it is on, token minting and connection counts scale with the number of people who open chat, not with all viewers. That keeps quotas and cost bounded.
- The quota request covers 30% of peak viewers (DEFAULT).
