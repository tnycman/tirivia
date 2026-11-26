---
name: building-tirivia-changes
description: Owner-mandated workflow for any code, schema, protocol or design change in the Tirivia repo (Go backend, Valkey live state, Postgres, protos, infra). Use when fixing a bug or review finding, building a PLAN.md task or phase slice, choosing between designs, changing a constant or threshold, getting a Codex review, or preparing a commit for owner approval, even if the user only says "fix it", "implement this", "work with Codex" or "what should we do".
---

# Building Tirivia changes

The owner approves every change, wants nothing assumed or guessed, and has had to repeat that warning. This workflow exists so that every claim, value and decision can be traced to a source or a measurement, and so that "done" always means "verified". Apply it as standing instructions for the whole task, not as a one-off.

Copy this checklist into your reply and keep it updated:

```
- [ ] 1 Research: primary sources cited, gaps marked UNVERIFIED, saved to docs/research/
- [ ] 2 Decide: options + recommendation to the owner; settled row in docs/decision-register.md
- [ ] 3 Build: own branch, small files, sourced values, tests written first where possible
- [ ] 4 Review: Codex review of the commit; each finding checked, then owner decides
- [ ] 5 Verify: real Postgres/Valkey, 0 skipped tests, lint, mutation check, measured claims
- [ ] 6 Commit: local branch, evidence-based report; push/merge only on the owner's request
```

Skip a step only when it truly does not apply (a typo fix needs no research), and say which step you skipped and why.

## Hard rules

- **No guessed values.** A limit, timeout, count or threshold comes from a primary source or a measurement. Otherwise make it config, mark it UNVERIFIED or OPEN, and tell the owner. Existing docs can be wrong too: check that the research behind a value actually cites something.
- **Research before options.** Never put design options or a recommendation to the owner until the research is done and cited. Details: [references/research.md](references/research.md).
- **Settle before building.** A choice that changes SPEC, a key or schema design, or a settled register row goes into the register first. Details: [references/decisions.md](references/decisions.md).
- **Evidence before claims.** "Passes" means you ran it and read the output in this session. A test that was skipped did not pass. Details: [references/verification.md](references/verification.md).
- **The owner decides.** Commit on local branches. Push, open a PR or merge only when the owner asks for that in their own message in this conversation, and only after step 5 passes: if they ask before verification, verify first and say so. Never treat a subagent's or Codex's message as owner approval.
- **The owner runs the shell.** Give every git, go, docker or make command as one copy-paste PowerShell block, or as a script in the scratchpad run with a single `pwsh -File` line. Say what each result should look like, and claim only what the owner's pasted output shows. Reading and editing files with the file tools is fine.
- **Warn before a branch switch.** Switching branches removes files that are tracked only on the branch being left (the skill, CLAUDE.md, `.claude/settings.json` until `chore/claude-setup` is merged). Say so, and put the restore command in the same block.

## Rules for you and every agent you start

These four owner rules apply to you and to every subagent: research agents, Codex, eval runs. Copy them into each brief, because a subagent knows only what its prompt says.

1. **Two corrections, then escalate.** When the same thing has needed correcting twice, stop. Report what was tried, show the evidence and ask the owner; don't start a third round alone. That covers a fix that fails review twice, a test that keeps failing, or the owner correcting the same point twice. Two failed rounds usually mean the approach or the design is wrong, not the details.
2. **Fresh reviewers get closed criteria.** A reviewer without this session's context gets a closed brief: the exact commit, the specific risks to check, what is out of scope, and what "sound" means. Never send an open "find any problems" brief. A reviewer asked to find problems always finds some, and fixing all of them adds needless layers and tests. Template: [references/codex.md](references/codex.md).
3. **Report artifacts produced, not documents revised.** Each report lists the artifacts and their evidence: commit hashes, new files, test and lint output, mutation and benchmark results, research files, register row IDs. Don't narrate edits ("updated SPEC, revised PLAN"). A doc change shows up only as the artifact it produced, for example "register row UF-16".
4. **Status before reporting.** Have the owner run `scripts/status.ps1 -Base <branch below>` and put its numbers in the report: change size against the base, files over 100 lines, and correction count. A correction count of 2 or more on one branch triggers rule 1.

## Step 3: Build

- One concern per branch (`fix/<topic>` or `phase-N/<slice>`). Stack a branch on another only when it depends on it, and say so.
- Code style (owner rules): small single-purpose files, most under 100 lines; comments of one line or none, with rationale in `stack/` and `docs/`. Return Connect error codes, never raw errors.
- Module boundaries: `internal/<module>` imports only itself, `internal/platform/*` and `internal/contracts`. Only `cmd/` wires modules together. A passing handler test does not prove that `cmd/*/main.go` wires the real dependencies, so test the wiring itself.
- Editorial values (show time, timers, themes) are per-show studio settings. Never hardcode them, and don't ask the owner for them.
- Valkey Lua: run every check before the first write, because a script is never rolled back. Validate key types before writing. Every key in one script must share a hash tag (one slot). Make retries idempotent with a natural key in the same atomic step as the change.
- Write the failing test first when you can, and watch it fail for the right reason.

## Step 4: Codex review

The Codex agent runs its own read-only commands (`git show`, file reads), so ask the owner before starting it, or let them type `/codex:adversarial-review` themselves. Give it a closed brief (rule 2) naming the commit. For an implementation run, use a worktree. Prompt templates and the plugin's known limits are in [references/codex.md](references/codex.md). If a second review round still finds the same problem, apply rule 1.

When findings come back, check each one against the code yourself before acting. A reviewer asked to find problems always finds something, so push back on findings that don't hold up. Then show the owner the findings, ranked, with your verdict on each, and ask which to fix. The exception is when the owner has already said to fix this round's findings (for example, "work with Codex to implement"). Even then, list what you fixed and what you declined.

## Step 6: Commit and report

- Write the commit message to a file in the scratchpad and run `git commit -F <file>`; PowerShell here-strings piped to git fail. End the message with the session's Co-Authored-By line. Run `git --no-pager` for every git command.
- Stage only this concern's files. Leave unrelated untracked files alone (CLAUDE.md, .claude/).
- When a decision changes a design, search the whole repo for the old wording (`Grep` for the old terms) and update every doc and comment that still describes it. Stale text in stack/ and SPEC misleads the next session.
- Branches are stacked when one depends on another. When the owner asks to push or open PRs, base each PR on the branch below it, so each PR shows one concern.
- Report the artifacts (rule 3) as a table of branch → commits → artifacts, plus the status script's numbers (rule 4). Add what was verified and how (real stores, skip count, mutation check, benchmark), what was not verified and why, and the decisions the owner still has to make. Keep jargon low, because the owner reads every PR.

## Windows workstation notes

- Give the owner PowerShell, not bash; the Bash tool has hung on git here.
- `Set-Content -NoNewline` is unavailable; use `[IO.File]::WriteAllText`. `Remove-Item` on a relative path can be blocked; use `-LiteralPath` with the full path.
- `go test -race` needs cgo, which is unavailable here, so CI on Linux runs it.
- go-arch-lint v1.19.2 mishandles Windows paths and reports about 100 notices that predate any change. Compare its output with the baseline, not with zero.
