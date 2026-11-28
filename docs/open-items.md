# Open items

These carry forward from Master Plan v1.1 §11. Items added by this research pass are marked **(NEW)**.

## Outside inputs

1. Albanian counsel: how the show is classified, how prize tax and withholding work, and the official rules. A position from the AMLF if one can be obtained.
2. **Before the first cash show** (all countries are eligible per LG-1): counsel views on North Macedonia (2026 permit plus an 18% fee on giveaways), Kosovo (prize games have no legal basis since 2019) and the main diaspora countries.
3. An Albanian bank partner for payouts.
4. Delivery and price tests for SMS, WhatsApp and Viber in AL, XK and MK. Twilio's SIM-swap data does **not** cover these countries.
5. ~~A low-end Android reference phone~~ Chosen: Samsung Galaxy A07 4G, plus Xiaomi Redmi A5 as the worst case (OP-4). Buy both.
6. Studio, host and production team.
7. Launch sponsors.
8. **(NEW)** KYC vendor proof of concept with real AL/XK/MK documents (Veriff, Sumsub).
9. **(NEW)** An EU Art. 27 representative and a DPIA (GDPR applies to the EU diaspora).
10. **(NEW)** Confirm with Google Play policy support that free-entry prize contests are allowed. Play's policy does not address them directly.
11. **(NEW)** TV broadcast partner, if the show is simulcast: feed hand-off, broadcast delay above the app p99 latency, and on-air control (ADR-0010, PLAN L16).
12. **(NEW)** Counsel: do Audiovisual Media Authority (AMA) rules apply to prize games shown on TV? (PLAN L17)

## Owner's call

All owner decisions were settled on 2026-11-25. They are recorded in the [decision register](decision-register.md) (60 items). Still open:

1. **Product name and brand** (LG-3). `trivia` is a working codename. The name sets the visual design and the store listings.
2. **Prize currency** (PR-1). Decided together with the bank partner. The ledger is multi-currency-ready until then.
3. **Chat in production.** Chat is built, tested and switched from the back office, off by default. Decide at the Phase 7 go/no-go ([ADR-0009](adr/0009-chat-backend-switch.md)).
4. **Producers** set each weekly show's day and time, the per-question timers (6–10 s), themes and chat slow mode in the studio console. These are not owner decisions.
5. **(NEW) AC-6 can only be enforced per install.** Neither platform gives a device ID that survives a reinstall. Play Integrity device recall requires Play Billing volume, Google forbids fingerprinting, and App Attest keys are reset on reinstall. Today the limit is the install (the hardware-bound key) plus the ID check before any cash-out (ID-5). Accept this, or require ID verification earlier ([research 08](research/08-device-attestation.md)).

## Must be verified in testing (UNVERIFIED in research)

- How IVS cues behave across rendition switches and when a viewer joins late (Phase 2a).
- Whether IVS cues are carried in the audio-only rendition (TV companion mode, ADR-0010).
- The actual TV broadcast delay per platform (terrestrial, cable, IPTV), which must always trail the app p99 latency.
- Which IVS channel types accept SRT ingest.
- Parameter ranges for the IVS failover slate.
- Which surface the IVS Android player view uses (affects reaction overlay performance).
- ~~Whether pgx protocol-level prepared statements pin RDS Proxy sessions.~~ Settled: they are multiplexed, but we use exec mode anyway ([research 07](research/07-persistence-sourced-values.md)).
- App Attest validation-category byte order, confirmed on the first TestFlight build ([research 08](research/08-device-attestation.md)).
- The Play Integrity verdict on the Galaxy A07 and Redmi A5 reference phones (Android 13+ `MEETS_DEVICE_INTEGRITY`).
- The exact SwiftPM setting for default MainActor isolation.
- CLDR plural categories for Albanian.
