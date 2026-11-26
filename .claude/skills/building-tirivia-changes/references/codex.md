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

## Review prompt template

```
Read-only code review. Do not edit, commit or push.
Repo: <path>. Review commit <sha> on branch <branch> with `git --no-pager show <sha>`.
What it fixes: <finding or task, one paragraph>. Design source: <research file, register row>.
Check adversarially: <3-6 specific risks: atomicity, retries, auth, boundaries, migrations>.
Project rules: no guessed constants, one-line comments, small files, module boundaries.
Out of scope: <known, tracked items>.
Report findings ranked by severity with file:line, a concrete failure scenario and a fix;
mark each verified or suspicion; say plainly if it is sound.
Wait for the Codex job to finish and report its real findings.
```

## Implementation prompt additions

Add the exact behavior required for each bug and the regression test each needs, plus the store implementations to cover (memory and real). Give the verification commands to run, and set `GOCACHE` and `GOTMPDIR` to folders inside the worktree if Go cannot create its temp directory.

## Handling results

- Keep Codex's severity order, its file:line references, and its verified-versus-suspicion labels.
- Check every finding yourself. Note which ones you confirmed in the code, and decline the ones that don't hold up, giving the reason.
- Then follow the owner rule in SKILL.md step 4: present the findings and ask, unless the owner has already said to fix this round.
- Each fix round gets its own commit ("address Codex review of <sha>"). Re-request a review only when the fix is large or risky.
