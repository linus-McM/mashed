---
name: desloppify
description: >
  Codebase health scanner and technical debt tracker powered by desloppify.
  Use this skill whenever the user asks to improve code quality, scan for
  technical debt, get a health score, set up desloppify, clean up slop,
  or create a codebase cleanup plan. Also trigger when the user mentions
  "slop", "sloppy code", "code health", "code quality score", "technical
  debt scan", "desloppify", or wants to systematically improve their
  codebase quality. This skill handles Python venv management automatically
  so it works seamlessly in non-Python repos (TypeScript, Go, Rust, etc.).
  Desloppify supports 29 languages.
---

# Desloppify — Codebase Health Scanner

Desloppify combines mechanical detection (dead code, duplication, complexity, naming)
with LLM-powered subjective review (abstractions, error handling, module boundaries)
to produce a health score. A strict score above 98 correlates with code a senior
engineer would call beautiful.

The scoring resists gaming — the only way to improve it is to actually make the code better.

## Venv Convention

Desloppify is a Python tool, but the project being scanned is likely not Python.
All desloppify commands run inside a dedicated venv at `.venv-desloppify/`.

**Every bash block that calls desloppify must start with:**
```bash
source .venv-desloppify/bin/activate
```

This does not interfere with the project's own toolchain (npm, go, cargo, etc.).

## Setup

Check if desloppify is already set up before doing anything:

```bash
if [ -d .venv-desloppify ] && source .venv-desloppify/bin/activate && command -v desloppify &>/dev/null; then
  echo "desloppify ready — $(desloppify --version 2>/dev/null || echo 'installed')"
else
  echo "NEEDS_SETUP"
fi
```

If `NEEDS_SETUP`:

1. **Create the venv and install:**
```bash
python3 -m venv .venv-desloppify
source .venv-desloppify/bin/activate
pip install --upgrade "desloppify[full]"
```

2. **Update .gitignore** — add `.venv-desloppify/` and `.desloppify/` if not already present.

3. **Exclude non-source directories** before scanning. Check what exists and exclude
   obvious ones (vendor, node_modules, dist, build, generated code, worktrees).
   Ask the user about anything questionable.
```bash
source .venv-desloppify/bin/activate
desloppify exclude node_modules
desloppify exclude dist
desloppify exclude build
desloppify exclude .venv-desloppify
```

4. **Set up the git pre-commit hook** for automatic scanning. Copy `scripts/pre-commit-hook.sh`
   from this skill's own directory (the directory containing this SKILL.md) and install it:

```bash
mkdir -p .githooks
# Use this skill's directory — the one containing this SKILL.md file
cp ~/.claude/skills/desloppify/scripts/pre-commit-hook.sh .githooks/pre-commit
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

If `.githooks/pre-commit` already exists, append only the body of `pre-commit-hook.sh`
(skip the `#!/usr/bin/env bash` shebang and any `set -euo pipefail` lines to avoid
duplicating them) after the existing content. This preserves other hooks already in place.

## Phase 1: Scan — Understand the Codebase

```bash
source .venv-desloppify/bin/activate
desloppify scan --path .
desloppify status
```

After scanning, **always run `desloppify next`** — it tells you exactly what to do, in priority order. Don't interpret scan output yourself or ask the user what to prioritize. Just follow `next`.

If subjective dimensions need review, the scan output will say so. For manual review:
```bash
desloppify review --prepare    # then follow the review workflow
```

Share the scan results and score with the user before proceeding to fixes.

## Phase 2: Plan — Shape the Queue

After scan (and optionally review), triage stages appear in the execution queue.
Complete them in order — `next` tells you what each stage expects:

```bash
source .venv-desloppify/bin/activate
desloppify next
desloppify plan triage --stage observe --report "themes and root causes..."
desloppify plan triage --stage reflect --report "comparison against completed work..."
desloppify plan triage --stage organize --report "summary of priorities..."
desloppify plan triage --complete --strategy "execution plan..."
```

For automated triage: `desloppify plan triage --run-stages --runner claude`

Then shape the queue — the plan determines what `next` shows you:

```bash
desloppify plan queue                    # compact execution queue view
desloppify plan reorder <pat> top        # reorder priorities
desloppify plan cluster create <name>    # group related issues
desloppify plan skip <pat>              # defer items
```

## Phase 3: Execute — Grind the Queue

This is the core loop. Don't be lazy — large refactors and small fixes deserve equal energy.

```bash
source .venv-desloppify/bin/activate

# 1. Get the next item from the execution queue
desloppify next

# 2. Fix the issue in code — properly, not minimally

# 3. Resolve it (next shows the exact command including required attestation)

# 4. Commit logical batches
git add <files> && git commit -m "desloppify: fix <description>"
desloppify plan commit-log record

# 5. Repeat until the queue is empty
desloppify next
```

Score may temporarily drop after fixes — cascade effects are normal, keep going.
If `next` suggests an auto-fixer: `desloppify autofix <fixer> --dry-run` to preview, then apply.

**When the queue is clear, go back to Phase 1.** New issues surface, cascades resolve,
priorities shift. This is the cycle.

## Quick Reference

```bash
source .venv-desloppify/bin/activate
desloppify scan --path .                  # full scan
desloppify status                         # check scores
desloppify next                           # what to fix next
desloppify next --count 5                 # top 5 items
desloppify next --cluster <name>          # drill into a cluster
desloppify backlog                        # broader open work
desloppify show <pattern>                 # filter by file/detector/ID
desloppify show --status open             # all open findings
desloppify plan queue                     # execution queue
desloppify autofix <fixer> --dry-run      # preview auto-fix
desloppify exclude <path>                 # exclude a directory
desloppify config show                    # show all config
```

## Scoring

Overall score = **25% mechanical** + **75% subjective**.

- **Mechanical (25%)**: auto-detected — duplication, dead code, smells, unused imports, security.
- **Subjective (75%)**: LLM review — naming, error handling, abstractions, clarity. Starts at 0% until reviewed.
- **Strict score** is the north star: wontfix items count as open.

## Reviews

Four paths to get subjective scores:

- **Local runner (Claude)**: `desloppify review --prepare` then launch parallel subagents, then `desloppify review --import merged.json`
- **Cloud/external**: `desloppify review --external-start --external-runner claude` then follow session template then `--external-submit`
- **Automated (Codex)**: `desloppify review --run-batches --runner codex --parallel --scan-after-import`
- **Manual**: `desloppify review --prepare` then review per dimension then import

Review output must be JSON matching the format specified in `desloppify review --prepare` output.

## Post-Commit Scanning

The git pre-commit hook installed during setup runs `desloppify scan --path .` and
reports the strict score before each commit. This catches regressions early.

If the score drops significantly, the hook warns but does not block the commit —
the user decides whether to address it now or later.
