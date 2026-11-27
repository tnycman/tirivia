# Working with Codex (step 4)

Codex runs through the `codex:codex-rescue` subagent of the OpenAI Codex plugin. It gives a second, independent read. It is never the authority on whether a finding is real.

## Known limits of the plugin

- Its sandbox cannot write to the repo's shared `.git`, so it cannot create branches or commit. For an implementation run:
  1. Spawn the agent with `isolation: worktree`.
  2. Before Codex starts, create and check out the branch in that worktree yourself (`git -C <worktree> switch -c fix/<topic>`).
  3. Tell Codex not to commit.
  4. You review the diff, test it on real stores, and commit.
- Read-only review runs cannot create Go's build directory, so Codex reviews by reading code and does not run the tests. Run them yourself.
- Long jobs move to the background, and the agent may hand back "still running" with a task output file. Read that file (it is the Codex log, not the agent transcript) for the result. Don't start a duplicate job. You'll be notified when the job finishes.
- One task per run. Each prompt says what "done" looks like and ends with "wait for the Codex job to finish and report its real result".
- Never tell the agent to wait for every Codex process to exit. A loop polling `tasklist` for codex never ended and blocked the session for hours.
- Once the agent has handed back, it cannot report again until its background job ends. Later replies are lost, so wait for the "finished" notice and then ask for the final findings in one message.
- Give it the four rules from SKILL.md in the prompt: two corrections then escalate, closed criteria, report artifacts, no shell commands beyond read-only inspection.

## Review prompt template (closed criteria)

Every review brief is closed. It names the commit, lists the exact checks, says what is out of scope, and defines "sound". The reviewer reports only against that list.

```
Read-only code review. Do not edit, commit or push. Run only read-only git and file commands.
Repo: <path>. Review commit <sha> on branch <branch> with `git --no-pager show <sha>`.
What it fixes: <finding or task, one paragraph>. Design source: <research file, register row>.
Check exactly these, nothing else:
  1. <specific risk, e.g. "a retry after a lost reply cannot double-count">
  2. <...>  (3-6 items: atomicity, retries, auth, boundaries, migrations)
  N. Project rules: no guessed constants, one-line comments, small files, module boundaries.
Out of scope (do not report): <known, tracked items>; style preferences; hypothetical future features.
Sound means: every numbered check holds for the code as committed.
For each numbered check answer HOLDS or FAILS. For FAILS give file:line, a concrete
failure scenario, a fix, and whether it is verified or a suspicion.
Wait for the Codex job to finish, then report once with these answers.
```

## Implementation prompt additions

Add the exact behavior required for each bug and the regression test each needs, plus the store implementations to cover (memory and real). Give the verification commands to run, and set `GOCACHE` and `GOTMPDIR` to folders inside the worktree if Go cannot create its temp directory.

## Handling results

- Keep Codex's severity order, its file:line references, and its verified-versus-suspicion labels.
- Check every finding yourself. Note which ones you confirmed in the code, and decline the ones that don't hold up, giving the reason.
- Then follow the owner rule in SKILL.md step 4: present the findings and ask, unless the owner has already said to fix this round.
- Each fix round gets its own commit, with a subject containing "address Codex review of <sha>". `scripts/status.ps1` counts these commits as corrections, so keep the wording. Re-request a review only when the fix is large or risky. If the same check FAILS again after a fix, stop and escalate (rule 1).
