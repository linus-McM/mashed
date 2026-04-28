# Story bmad-rollout-09: Skill-Prompt Alignment Prep PR (Optional / Pre-Phase 2)

**Priority:** P2-medium
**Domain:** backend (skill markdown only)
**Estimated Complexity:** M
**Depends On:** none (can run in parallel with story 01; recommended to ship BEFORE story 02 per Open Decision #4)
**Status:** ready

## Description

Pulled forward from plan §"Open decisions #4". Audit and align the embedded BMAD skill markdown for each of the 26 processes touched by stories 02–08 so claude's facilitator copy ("ask the user for X") matches the modal prompt vocabulary the helper renders. This is a separate prep PR before Phase 2 so the helper rollout doesn't bundle skill-copy edits with registry-shape changes — a clean diff for reviewers and a clean rollback if the modal copy proves awkward in smoke testing.

This story does NOT change Go code or registry shape. It edits skill markdown only.

## Developer Notes

### Architecture
- Skill markdown lives under the BMAD skill bundles. Locate by:
  ```bash
  rg -l 'ask the user|ask user|user will|please ask' --type=md
  ```
  Likely paths: `internal/bmad/skillgen.go` generates `${repo}/.claude/skills/bmad-*/SKILL.md` from registry templates; alternatively the source-of-truth markdown lives under `internal/bmad/testdata/skills/` or alongside the skill template. Confirm by reading `internal/bmad/skillgen.go` before editing.
- For each of the 26 IDs, the skill should:
  1. Use the same prompt vocabulary as the helper-rendered `Prompt` (e.g. "Provide feedback or type 'done' when ready.").
  2. Reference the same accept tokens the Gate honours (`done`, `wrap up`, `complete` plus per-process additions).
  3. Avoid asking via the legacy "ask claude to ask" pattern that previously triggered the autonomous-fallback snackbar.
- The 26 IDs are exactly the union of stories 02–08 rollout sets: 6 + 3 + 4 + 1 + 6 + 3 + 3 = 26.

### Process / methodology
1. Generate the helper-rendered `Prompt`/`HelpText` per ID by reading the spec tables in stories 02–08.
2. For each ID, open the SKILL.md (or template) and locate the user-facing question copy.
3. Rewrite the copy to mirror the helper Prompt; reference the AcceptTokens explicitly so claude reminds the user of the exit verbs.
4. Keep the diff minimal — do NOT restructure the skill, only update the Q&A copy.

### Audit checklist (per ID)
- [ ] Prompt verb matches the helper Prompt verb (e.g. "Provide feedback" not "Please tell me").
- [ ] Accept tokens listed verbatim in the skill so claude prompts them in-pane (e.g. "Type `done`, `wrap up`, or `ship` when ready.").
- [ ] Reject tokens (`abort`, `cancel`) referenced if the skill ever offers a back-out.
- [ ] No "ask claude to ask" loops (those triggered the legacy snackbar; the modal owns prompting now).

### Risks & migration notes

- **Risk #3 (Skill prompts vs modal vocabulary)**: This is the explicit subject of plan §"Migration risks" #3. Mismatched copy means the modal asks one thing while claude asks another in-pane — confusing for the user. This story is the resolution for Risk #3.
- **Risk #5 (Frontend modal copy)**: Aligning skill copy reduces the smoke-run typo surface in stories 02–08. Each downstream story still does its own Playwright smoke; this story makes those smoke runs cleaner.
- **Open Decision #4 (resolution)**: Plan recommends a separate prep PR. This story is that prep PR.

### Reference Files
- `internal/bmad/skillgen.go` — confirms whether SKILL.md is generated or source-controlled.
- `internal/bmad/skillgen_test.go:142, 184` — test patterns showing the SKILL.md output path layout.
- `internal/bmad/templates.go` — string-only references; not edited here.
- Stories 02–08 — source of truth for per-process Prompt/HelpText/AcceptTokens.

## Acceptance Criteria

**AC-1: Each of the 26 rollout IDs has its skill markdown audited**
- Given the 26 process IDs from stories 02–08
- When the skill markdown for each is read
- Then the prompt copy explicitly mentions the matching helper Prompt verb
- And the matching AcceptTokens are listed verbatim in the skill

**AC-2: No "ask claude to ask" patterns remain in the 26 skills**
- Given the skill markdown after this story
- When `rg "please ask the user|ask the user to provide|user will be asked"` is run against the 26 skill files
- Then zero matches remain (or each match has a comment explaining the autonomous-fallback rationale)

**AC-3: Skill markdown for non-rollout processes is untouched**
- Given the skill markdown for processes NOT in the rollout (sprint-planning, sprint-status, util-file-loader, util-multi-file-loader, brainstorming, product-brief, advanced-elicitation, party-mode)
- When `git diff` is run
- Then no changes are present in those files

**AC-4: SKILL.md generation pipeline (if any) still works**
- Given `internal/bmad/skillgen.go` is invoked
- When SKILL.md is regenerated for any rolled-out process
- Then the generator output matches the committed skill markdown byte-for-byte
- Or the generator template is updated alongside the markdown so regeneration is idempotent

## BDD Test Scenarios

```gherkin
Feature: Skill prompt alignment

  Scenario: dev-story skill mentions "done" and "ship" tokens
    When the bmad-dev-story SKILL.md is read
    Then it contains the text "done" within an accept-token reference
    And it contains the text "ship" within an accept-token reference
    And it does not contain "ask the user" without a fallback comment

  Scenario: code-review skill mentions "approved" and "ship" tokens
    When the bmad-code-review SKILL.md is read
    Then it contains "approved" and "ship" as exit verbs
    And the prompt copy aligns with "Reply with notes, or 'approved'/'done' to finish the review."

  Scenario: Untouched processes have no skill diff
    When git diff is run against skill markdown for non-rollout processes
    Then no changes are present in sprint-planning, sprint-status, util-file-loader, util-multi-file-loader, brainstorming, product-brief, advanced-elicitation, party-mode

  Scenario: Skill regeneration is idempotent
    Given internal/bmad/skillgen.go runs against any rolled-out process
    Then the generated SKILL.md matches the committed file byte-for-byte
```

## Tasks / Subtasks

- [ ] Task 1: Locate skill markdown source of truth (AC-1, AC-4)
  - [ ] Read `internal/bmad/skillgen.go` to understand whether SKILL.md is generated, source-controlled, or both.
  - [ ] Identify the editable file per process (the `.md` template if generated, or the SKILL.md itself if source-controlled).
- [ ] Task 2: Audit + edit per ID (AC-1, AC-2)
  - [ ] For each of the 26 rollout IDs, open the editable file.
  - [ ] Update prompt copy and accept-token references per the per-process specs in stories 02–08.
  - [ ] Run `rg "please ask the user|ask the user to provide"` per file; resolve or comment.
- [ ] Task 3: Regression test — non-rollout skills untouched (AC-3)
  - [ ] `git diff --stat` should list only the 26 rollout process files (or templates).
- [ ] Task 4: Skill generator parity check (AC-4)
  - [ ] If `internal/bmad/skillgen.go` regenerates SKILL.md, run the generator and confirm zero diff post-regeneration.
- [ ] Task 5: Smoke verification (manual)
  - [ ] `wails dev`, drop a `bmad-dev-story` node, run, observe that claude's in-pane question copy matches the modal `Prompt`.
  - [ ] Verify the user sees consistent vocabulary across modal + claude pane.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated checks (or manual read-through where automation is impractical)
- [ ] 80%+ coverage on any modified Go code (this story should not modify Go code; if generator changes are needed, cover them)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on any modified Go code; no CRITICAL/HIGH issues
- [ ] AC validation table populated in PR description
- [ ] Smoke run on at least one rolled-out process verifying modal + pane vocabulary alignment
- [ ] Status flipped to `done` by sprint lead
