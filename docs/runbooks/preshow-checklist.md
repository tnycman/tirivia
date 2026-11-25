# Runbook: pre-show checklist

Run this for every live show, and for every rehearsal from Phase 2b onward. **T** is the show's start time as scheduled by the producers in the studio console. Scaling actions are booked automatically from it.

| When | Check | Owner |
|---|---|---|
| T-24 h | Questions approved by two people in the vault. Show-dependent questions placed late. Every question within the content limits (≤ 240 B / 64 B) | Show team |
| T-4 h | **ALB LCU reservation set** to the load-test `PeakLCUs`. It can take hours to apply, so confirm it is active | Platform |
| T-4 h | Quota headroom checked: IVS concurrent views, chat connections (if chat is on), Fargate vCPU, Play Integrity daily quota | Platform |
| T-2 h | Deploy freeze starts. No prod applies or migrations until SETTLED | Platform |
| T-60 | The orchestrator has raised MinCapacity, re-sharded KDS, scaled out Valkey and warmed the Ireland standby | Platform (automatic; verify) |
| T-45 | Encoders up. Primary SRT (`priority=1`) and backup RTMPS (`priority=2`) use **identical** encode profiles. Failover slate enabled | Studio |
| T-30 | Test cue sent on the rehearsal show. Signature verifies on reference iOS and Android devices. Cue-to-render latency within `L_max` | Studio + QA |
| T-30 | **TV simulcast only:** the station confirms the broadcast delay is on and set to the run-of-show value (above the app p99 latency). Compare a test cue on a reference phone with the TV picture: TV must be **later** | Studio + TV station |
| T-20 | LOBBY open. **Feature switches confirmed in the studio console**: chat (ON/OFF as decided for this show), reactions, extra lives, sign-ups. Country eligibility table matches counsel status | Producer |
| T-20 | Live-show dashboard green: joins, PutMetadata errors 0, Valkey per-shard CPU, pinned RDS Proxy sessions | On-call |
| T-5 | Host briefed on the focus-loss rule announcement and the burst moments | Producer |
| T+end | SETTLED. Winners pass risk review before any claim card is issued | Risk |

## Chat switch quick reference

- **Where:** studio console → Feature switches. Every change needs a reason, which is written to the audit log.
- **Scope:** *Toggle globally* applies to all shows. *Toggle for this show* overrides the global value for one show only.
- **Effect:** apps show or hide the chat panel within one `pulse` (≤ 5 s). No app release or rejoin is needed.
- **Default:** off. Chat delivery costs nothing while it is off.
- **Kill switch:** to stop chat during a show, toggle it off for that show. It takes effect within 5 s.
