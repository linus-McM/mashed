## Sprint Report — ui-ast-U7: `DecisionGroup` + response-map collection + §3.4 submit path

### Summary
- Tasks: 14/14 + 3 FIX (#15/#16/#17) — all closed
- Coverage (new): `astResponses.ts` 100%, `DecisionGroup.svelte` 95.55% lines. Documented exceptions: `AstNode.svelte` 45% (U6 Svelte-prop artifact carried forward), `InputResponseModal.svelte` ~71% (pre-existing gaps per U6 baseline).
- /simplify: verified on every modified file across all engineer and reviewer turns
- AC Validation: 11/11 acceptance criteria verified (5 Vitest + 6 Playwright live)

### AC Validation Results
| AC | Description | Method | Evidence | Result |
|----|-------------|--------|----------|--------|
| AC-1 | DecisionGroup dispatches six widget types | Vitest | `AC1_dispatches_six_widget_types` ×6 (DecisionGroup.test.ts) | PASS |
| AC-2 | responses store accumulates keys | Vitest | `AC2_responses_accumulate` + `AC2_returns_writable` + `AC2_per_instance_isolation` (astResponses.test.ts) | PASS |
| AC-3 | ShapeJSON submits full map as JSON | Playwright | `json-submit-full-map` (ui-ast-decision-group.spec.ts:162) | PASS |
| AC-4 | Single group on non-JSON shape → plain string | Playwright | `single-group-plain-string-submit` (spec.ts:192) | PASS |
| AC-5 | Multi-group non-JSON → collapse banner visible | Playwright | `collapse-banner-visible` (spec.ts:212) | PASS |
| AC-6 | User switches active group under collapse | Playwright | `collapse-user-switches-active` (spec.ts:238) + payload assertion via FIX #16 extension | PASS |
| AC-7 | Send disabled until required filled | Vitest | `AC7_send_disabled_until_required_filled` + `AC7_collapse_send_gated_on_active` (InputResponseModal.test.ts) | PASS |
| AC-8 | Cmd+Enter submits the form | Playwright | `cmd-enter-submits` (spec.ts:274) — Meta+Enter fires RespondToInput with same payload | PASS |
| AC-9 | Zero groups → Layer-1 fallback widget | Playwright | `zero-groups-fallback-layer1` (spec.ts:295) | PASS |
| AC-10 | `aria-labelledby` wires heading → widget | Vitest | `AC10_aria_labelledby_wired` (DecisionGroup.test.ts) | PASS |
| AC-11 | §7.1 — file widget path verbatim to RespondToInput | Vitest | `AC11_file_widget_passthrough` (InputResponseModal.test.ts) | PASS |

**All ACs validated:** YES
**Lead spot-check:** VERIFIED — Phase 4f ran `MASHED_E2E=1 playwright test` against live `wails dev` (6/6 pass, 17.5s); pre-flight re-ran full Vitest (675/675) and Go race suite (clean) prior to commit.

### Task Breakdown
| Task | Test Writer | Implementer | Coverage | /simplify | AC Validated | Status |
|------|-------------|-------------|----------|-----------|--------------|--------|
| T1 RED DecisionGroup tests | test-writer | — | N/A (tests) | yes | AC-1, AC-10 | completed |
| T2 RED store + Send-gate tests | test-writer | — | N/A | yes | AC-2, AC-7 | completed |
| T3 RED file passthrough test | test-writer | — | N/A | yes | AC-11 | completed |
| T4 GREEN astResponses + DecisionGroup | — | ui-engineer | 100% / 95.55% | yes | AC-1, AC-2, AC-10 | completed |
| T5 GREEN §3.4 submit + collapse banner | — | ui-engineer | — | yes | AC-2..9, AC-11 | completed |
| T6 GREEN AstNode → DecisionGroup wiring | — | ui-engineer | — | yes | AC-1 | completed |
| T7 Playwright AC suite | — | ui-engineer | 6/6 | yes | AC-3, 4, 5, 6, 8, 9 | completed |
| QG1 Build + coverage | — | lead | 652→675 tests | — | — | completed |
| QG2 Code review | — | reviewer | 1 HIGH + 2 MED + 4 LOW; HIGH+MED fixed via #16/#17, MED-1 false positive | yes | — | completed |
| QG3 Security audit | — | security-check | 10/10 PASS | — | — | completed |
| QG4 Design critique | — | ui-architect | 49/60 → 54/60 (via #15) | — | — | completed |
| #15 FIX DecisionGroup active border | — | ui-engineer | +2 tests | yes | — | completed |
| #16 FIX collapse Send-gate + submit group | — | ui-engineer | +2 tests | yes | AC-6, AC-7 | completed |
| #17 FIX tabindex rationale | — | ui-engineer | docs-only | yes | — | completed |
| QG5 Spec completion | — | spec-reviewer | 11/11 AC + DoD, 1 cosmetic gap, 1 documented deviation | — | — | completed |
| Phase 4f AC Validation | — | lead | 6/6 PASS, 1 round | — | — | completed |
| Phase 5 commit + story + report | — | lead | `b0efd13` | — | — | completed |

### Quality Gate Results
- Build: PASS (`frontend/npm run build` 52.52s; pre-commit hook `frontend-build` 53.57s; `go build ./...` clean; `go vet ./...` clean; `go test ./... -race -count=1 -short` all packages PASS)
- Test suite: 675 Vitest tests passing (54 files); coverage see Summary
- Coverage gate: PASS with two documented U6-baseline exceptions
- AC validation (unit): PASS — every AC has test evidence
- Code review (QG2): PASS after FIX #16 + FIX #17 — 0 CRIT, 0 HIGH outstanding, 0 MED outstanding; 4 LOW (accepted / follow-up)
- Security (QG3): PASS — 10/10 checks PASS, zero CRIT. §7.1 file passthrough verified (no frontend normalisation); `fallback_answer_shape` never used for submit format.
- Performance: N/A (no Go backend changes; frontend bundle unchanged except DecisionGroup component)
- Spec completion (QG5): PASS — 11/11 ACs, DoD complete, §3.4 rules 1-3 + FIX #16 active-group swap faithful, §6.3 modal layout faithful, §6.4 a11y wired with documented tabindex deviation
- UI verification (QG4): PASS — 54/60 aggregate after FIX #15, no individual dimension below 8, hard-rule audit clean (no new tokens, no hex literals, no px font-size literals, Send == `FreeTextWidget.btn-submit`, banner == `HintBanner warn`, Required pill == `.round-pill` idiom)

### AC Validation (Playwright) — Phase 4f
- **Report:** `docs/playwright_cli_US_validate/ui-ast-U7-decision-group-report.md`
- **Fix rounds:** 0 (gate passed on first run)
- **Final pass rate:** 6/6 (100%)
- **Accepted failures:** none
- **Accepted blocked:** none

| Round | PASS | FAIL | BLOCKED | Action |
|-------|------|------|---------|--------|
| 1     | 6    | 0    | 0       | GATE PASSED — no fix loop required |

### Files Changed
Commit: `b0efd13 feat(ui-ast/U7): DecisionGroup + response-map collection + §3.4 submit path` (58 files; +1646 / −20445, dominated by coverage-artifact churn)

**New source:**
- frontend/src/components/bmad/DecisionGroup.svelte
- frontend/src/stores/astResponses.ts
- frontend/src/components/bmad/__tests__/DecisionGroup.test.ts
- frontend/src/stores/__tests__/astResponses.test.ts
- tests/ac/ui-ast-decision-group.spec.ts

**Modified:**
- frontend/src/components/bmad/AstNode.svelte (routes `decision_group` → `DecisionGroup`, new optional `isGroupDisabled` + `isGroupActive` props)
- frontend/src/components/bmad/InputResponseModal.svelte (responses store, Send button, §3.4 onSend, collapse banner, active-group switch, Cmd+Enter)
- frontend/src/components/bmad/InputResponseModal.test.ts (U7 describe block: 6 new tests)
- frontend/src/components/bmad/__tests__/AstNode.test.ts (2 new AC-1 tests)
- frontend/src/stores/interactiveInput.ts (minor)
- frontend/src/types/uiAst.ts (type additions for AstNode props + widget discriminant)
- docs/stories/ui-ast-U7-decision-group.md (status → done, tasks + DoD checked, tabindex deviation noted)

**Docs added:**
- docs/playwright_cli_US_validate/ui-ast-U7-decision-group-report.md

### Open Follow-ups (non-blocking)
- **Reviewer LOW #1** — `AstNode.svelte:50` leaks `JSON.stringify(node)` for `decision_group` nodes missing both `prompt` and `heading`. Not a security issue; recommend empty placeholder.
- **Reviewer LOW #2** — Widget `prompt` prop duck-typed as `PendingPrompt`; define a minimal `WidgetPromptShape` interface in `uiAst.ts`.
- **Reviewer LOW #3** — AC-8 Playwright only exercises `Meta+Enter`; add `Control+Enter` variant for non-macOS coverage.
- **Reviewer LOW #4** — Collapse banner single-line content vs design brief two-line layout. AC-5 regex tolerates.
- **Security pre-existing** — `FileInputWidget.svelte:17,26` has `path.trim()` from `bmad-interactive-06`. Not a U7 regression; if §7.1 demands strict verbatim at widget layer, raise a separate ticket.
- **Spec §6.3 deviation** — Footer DOM order `[Cancel] [Send]` vs spec `Send → Cancel` in tab order. Design Brief §4 intentionally overrides (Send visually rightmost as primary). Recommend one-line note under story Developer Notes deviation list next sprint.
