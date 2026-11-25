# ADR-0006: Batch reactions every 5 s with jitter, interval set by the server

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context ([research 02](../research/02-backend-protocol-data.md))
With every client posting a batch every 2 s, the load is:

| Players | Reaction requests/s |
|---|---|
| 300k | 150k |
| 600k | 300k |

That is **5–10 times the peak answer rate**. The ALB and the server fleet would have to be sized for reactions, not answers.

## Decision
- **Send only after a tap.** A client sends a batch only if the user tapped since the last batch.
- **Interval.** The default is **5 s ±1 s of random jitter**.
- **Server control.** `JoinShowResponse` sets the interval and the maximum taps per batch. `SubmitReactionsResponse` can change both during the show.
- **Pause during questions.** Reactions pause while a question is open (Master Plan).
- **Per-user cap.** The server enforces a cap on each user's reactions.
- **Rendering.** Phones render a sample sized in proportion to the totals in `pulse`. How often batches are sent does not change what users see.

## Consequences
- The upper bound becomes about 60k requests/s at 300k players, and less in practice.
- A user's own reactions still render instantly. Other users' reactions show up within one pulse, which is no slower than today because `pulse` runs every 3–5 s.
