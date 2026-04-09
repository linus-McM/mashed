# Skill Validation Report: team-sprint

**Date:** 2026-04-09 10:30
**Skill Path:** /Users/linus/Development/mashed/.claude/skills/team-sprint/
**Validator Version:** 1.1

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 9 | 0 | 0 | 0 |
| Functional | 5 | 0 | 0 | 0 |
| Efficiency | 5 | 0 | 0 | 0 |
| Instruction Compliance | 10 | 5 | 3 | 0 |
| Agent Simulation | 1 | 2 | 3 | 0 |
| **Total** | **30** | **7** | **6** | **0** |

**Overall Grade: D** (6 failures — needs fixes before deployment)

## Structural Validation

- PASS: SKILL.md exists
- PASS: YAML frontmatter present with `---` delimiters
- PASS: `name` field present (`team-sprint`)
- PASS: `description` field present (73 words, substantive)
- PASS: SKILL.md body non-empty (56 lines)
- PASS: All 7 file references resolve to existing files
- PASS: No orphan files (all files referenced)
- PASS: No sensitive files
- PASS: Directory follows conventions (`references/` for docs)

## Functional Validation

- PASS: No `scripts/` directory (no scripts to validate)
- PASS: `references/ac-validation-protocol.md` — valid, 47 lines
- PASS: `references/agent-prompts.md` — valid, 50 lines
- PASS: `references/quality-gates.md` — valid, 52 lines
- PASS: `references/scaling-errors-antipatterns.md` — valid, 43 lines

## Efficiency Analysis

| Metric | Value | Status |
|--------|-------|--------|
| SKILL.md lines | 67 | PASS |
| SKILL.md tokens | ~340 | PASS |
| team-sprint.md lines | 177 | PASS |
| Total skill lines | 477 | PASS |
| Heavy directives | 10 | PASS |
| Progressive disclosure | References separated into files | PASS |

## Instruction Compliance

### Critical Instructions: 10/13 clear (77% compliance-ready)

| Instruction | Ambiguity | Buried | Conflict | Missing Why | Overload | Implicit | Drift |
|-------------|-----------|--------|----------|-------------|----------|----------|-------|
| Phase 0: Decompose | OK | OK | OK | OK | - | OK | OK |
| Phase 1: Pick story | OK | OK | OK | OK | - | OK | - |
| Phase 2a: TeamCreate | OK | OK | OK | OK | - | OK | - |
| Phase 2b: QG tasks | OK | OK | OK | OK | - | WARN | - |
| Phase 2c: Spawn engineers | OK | OK | OK | OK | - | OK | - |
| Phase 3: TDD loop | OK | OK | OK | OK | - | OK | - |
| Phase 4: Quality gates | OK | OK | OK | OK | - | OK | - |
| **Phase 5a: Commit** | **WARN** | OK | **FAIL** | **WARN** | - | **WARN** | **FAIL** |
| Phase 5b: Story status | OK | OK | OK | OK | - | OK | - |
| Phase 5c: Report/next | OK | OK | OK | OK | - | OK | - |
| /simplify gate | OK | OK | OK | OK | - | OK | - |
| AC Validation | OK | OK | OK | OK | - | OK | OK |
| 80% coverage | OK | OK | OK | OK | - | OK | - |

### Failure Modes Detected

**FAIL-1: Pre-flight checks positioned AFTER commit commands (Conflict + Drift)**
The skill lists `git add` → `git commit` → then "Pre-flight (ALL must pass)". This ordering implies commit first, verify after. A proper workflow must verify before committing. An agent following the literal text order will commit broken code, then discover failures with no rollback path.

**FAIL-2: No gate ensuring commit happens before next story (Drift)**
Phase 5b (update story status) and 5c (report/next) have no dependency check verifying a commit was actually created. An agent can skip 5a entirely, mark the story done, and proceed with uncommitted code in the working tree.

**FAIL-3: No mechanism to determine which files to add (Ambiguity + Implicit)**
The skill says `git add {specific files}` but provides no instruction for how to collect the list of changed files. Multi-agent sprints produce changes across many files — the lead has no convention for gathering this information.

**WARN-1: Missing Co-Authored-By trailer**
Commit message format `feat({storyId}): {description}` doesn't include Claude Code's Co-Authored-By convention.

**WARN-2: No rollback on pre-flight failure**
If pre-flight fails after commit, no instruction to reset or amend.

**WARN-3: Lead ownership not explicit for commit step**
Phase 5 completion tasks have `role: "lead"` in Phase 2b but the commit instructions don't reiterate "the lead runs these commands directly."

**WARN-4: Missing git status/diff review before staging**
No instruction to review changes before `git add`.

**WARN-5: Missing `go vet` in Phase 5a pre-flight**
Phase 4 quality-gates.md includes `go vet` but Phase 5a's pre-flight omits it.

### Suggested Rewrites

**Phase 5a — Before (current):**
```markdown
### 5a: Commit
git add {specific files}
git commit -m "feat({storyId}): {description}"

Pre-flight (ALL must pass): git status --porcelain (empty), go build ./..., 
go test ./... -coverprofile=cover.out -count=1 (zero failures), 
go tool cover -func=cover.out (>= 80%), go test ./... -race -count=1 -short (no races).
```

**Phase 5a — After (proposed):**
```markdown
### 5a: Commit (BLOCKING — do not proceed to 5b until commit succeeds)

The lead runs this sequence directly. This is the gate between "code complete" and "story done" — 
skipping it means uncommitted work that the next story cycle will conflict with.

**Step 1 — Identify changed files:**
git status
git diff --stat

Review the output. Only stage files related to this story.

**Step 2 — Pre-flight (ALL must pass before committing):**
go build ./...
go vet ./...
go test ./... -coverprofile=cover.out -count=1
go tool cover -func=cover.out | grep {modified files}  # >= 80%
go test ./... -race -count=1 -short

If any check fails: fix the issue (spawn build-fixer if needed), re-run pre-flight. 
Do not proceed until all pass.

**Step 3 — Stage and commit:**
git add {files from Step 1}
git commit -m "feat({storyId}): {description}

Co-Authored-By: Claude <noreply@anthropic.com>"

**Step 4 — Verify:**
git log --oneline -1  # confirm commit exists
git status            # confirm working tree is clean

If Step 4 shows uncommitted changes, return to Step 1.
```

## Agent Simulation

**Test prompt:** "Walk through Phase 5 as the sprint lead at end of a story cycle."

| Check | Result | Detail |
|-------|--------|--------|
| Agent found commit instructions | PASS | Located Phase 5a |
| Agent identified ordering problem | FAIL | Pre-flight listed after commit — agent would commit first |
| Agent knew which files to add | FAIL | No mechanism provided — agent said "would have to invent a strategy" |
| Agent found commit gate | FAIL | No dependency check preventing skip of 5a |
| Agent identified missing git review | WARN | No diff review before staging |
| Agent identified missing rollback | WARN | No recovery path if pre-flight fails post-commit |

## Recommendations (ordered by severity)

1. **FAIL** — Rewrite Phase 5a with correct ordering: pre-flight → stage → commit → verify
2. **FAIL** — Add BLOCKING gate language to Phase 5a preventing skip to 5b
3. **FAIL** — Add file collection step (git status + git diff --stat) before staging
4. **WARN** — Add Co-Authored-By trailer to commit message format
5. **WARN** — Add rollback instructions (reset + fix + re-commit) for pre-flight failures
6. **WARN** — Make lead ownership explicit in Phase 5a header
7. **WARN** — Add `go vet` to Phase 5a pre-flight
8. **WARN** — Add git diff review step before staging
