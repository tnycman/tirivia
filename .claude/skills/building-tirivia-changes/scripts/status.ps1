# Read-only branch status: size against the base branch, oversize files, correction count (SKILL.md rule 4).
# Usage: pwsh -File .claude/skills/building-tirivia-changes/scripts/status.ps1 -Base <branch this one is built on>
param([Parameter(Mandatory)][string]$Base, [int]$MaxLines = 100)

$branch = git rev-parse --abbrev-ref HEAD
$range = "$Base..HEAD"
$commits = @(git --no-pager log --format='%h %s' $range)
$stat = git --no-pager diff --shortstat "$Base...HEAD"
$files = @(git --no-pager diff --name-only --diff-filter=AM "$Base...HEAD")

"Branch      $branch (base: $Base)"
"Commits     $($commits.Count)"
"Size        $(if ($stat) { $stat.Trim() } else { 'no changes' })"

$big = foreach ($f in $files | Where-Object { $_ -like '*.go' -and (Test-Path $_) }) {
  $n = @(Get-Content $f).Count
  if ($n -gt $MaxLines) { "  $f ($n lines)" }
}
"Files over $MaxLines lines: $(@($big).Count)"
$big

$corrections = @($commits | Where-Object { $_ -match '(?i)codex review of|address .*review' })
"Corrections $($corrections.Count)"
$corrections | ForEach-Object { "  $_" }
if ($corrections.Count -ge 2) { 'ESCALATE    two or more correction rounds on this branch: stop and ask the owner (rule 1)' }
