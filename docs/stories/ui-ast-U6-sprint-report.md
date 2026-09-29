## Sprint Report — ui-ast-U6: AstNode dispatcher + node components + `pendingAst` derived store

### Summary
- Tasks: 15/15 completed + 1 FIX (#16) — all closed
- Coverage: 98% interactiveInput.ts · 92% MarkdownBlock · 100% HintBanner · 81% SummaryCard · 88% CodeBlock · 94% ComparisonTable · 92% linkSanitiser (lines). AstNode 38% + InputResponseModal 75% documented exceptions (accepted by QG5 — Svelte prop-line artifact and pre-existing gaps respectively).
- /simplify: verified on every modified file across all engineer and reviewer turns
- AC Validation: 11/11 acceptance criteria verified

### AC Validation Results
| AC | Description | Method | Evidence | Result |
|----|-------------|--------|----------|--------|
| AC-1 | `PendingPrompt.structured` is `string \| undefined` | Vitest | `AC1_structured_field_optional` (interactiveInput.test.ts:154) | PASS |
| AC-2 | `pendingAst` parses valid v1 AST | Vitest | `AC2_pending_ast_parses_valid_v1` (interactiveInput.test.ts:162) | PASS |
| AC-3 | `pendingAst` null on malformed JSON | Vitest | `AC3_pending_ast_null_on_malformed` (interactiveInput.test.ts:169) | PASS |
| AC-4 | `pendingAst` null on unknown version | Vitest | `AC4_pending_ast_null_on_unknown_version` (interactiveInput.test.ts:174) | PASS |
| AC-5 | AstNode renders each of six node shapes | Vitest + Playwright | `AC5_six_node_shapes` (AstNode.test.ts:52) + `six-node-shapes` (ui-ast-rendering.spec.ts:80) | PASS |
| AC-6 | `javascript:` href stripped to plain text | Vitest + Playwright | `AC6_javascript_href_stripped` (MarkdownBlock.test.ts:60) + `javascript-href-stripped-dom` (spec:114) | PASS |
| AC-7 | `data:` and `file:` schemes rejected | Vitest + Playwright | `AC7_data_and_file_schemes_rejected` (MarkdownBlock.test.ts:74) + `data-and-file-schemes-rejected-dom` (spec:130) | PASS |
| AC-8 | https link routes through confirm → BrowserOpenURL | Vitest + Playwright | `AC8_https_link_routes_through_confirm` ×2 (MarkdownBlock.test.ts:85,98) + `https-link-confirm-dialog` (spec:152) | PASS |
| AC-9 | Modal renders AST above Layer-1 widget; legacy when null | Vitest + Playwright | `AC9_renders_ast_above_widget`/`AC9_layer1_unchanged_when_null`/`AC9_respects_prefers_reduced_motion` (InputResponseModal.test.ts:160,172,180) + `modal-renders-ast-above-widget`/`modal-layer1-unchanged-when-null` (spec:230,252) | PASS |
| AC-10 | Unknown node type falls back to markdown, no throw | Vitest | `AC10_unknown_type_markdown_fallback` + `AC10_unknown_type_does_not_throw` (AstNode.test.ts:57,63) | PASS |
| AC-11 | CodeBlock copy writes verbatim + "Copied" label | Vitest + Playwright | `AC11_copy_writes_verbatim` + `AC11_copy_label_flips_to_copied` (CodeBlock.test.ts:31,41) + `code-copy-button-works` (spec:264) | PASS |

**All ACs validated:** YES
**Lead spot-check:** VERIFIED — re-ran Playwright suite against live `wails dev` in Phase 4f (7/7 pass, 15.6s); re-ran full Vitest (652/652) in Phase 5 pre-flight.

### Task Breakdown
| Task | Test Writer | Implementer | Coverage | /simplify | AC Validated | Status |
|------|-------------|-------------|----------|-----------|--------------|--------|
| T1 RED store tests | test-writer | — | N/A (tests) | yes | AC-1..4 | completed |
| T2 GREEN store + types | — | ui-engineer | 98% | yes | AC-1..4 | completed |
| T3 RED node + dispatcher tests | test-writer | — | N/A | yes | AC-5,10,11 | completed |
| T4 GREEN five passive components | — | ui-engineer | 81–100% | yes | AC-5,11 | completed |
| T5 GREEN AstNode dispatcher | — | ui-engineer | 38% (artifact) | yes | AC-5,10 | completed |
| T6 GREEN §7.2 sanitisation | — | ui-engineer | 92% | yes | AC-6,7,8 | completed |
| T7 GREEN modal integration | — | ui-engineer | +∆ (75% legacy) | yes | AC-9 | completed |
| T8 Playwright AC suite | — | ui-engineer | 7/7 | yes | AC-5,6,7,8,9,11 | completed |
| QG1 Build + coverage | — | lead | 648→652 tests | — | — | completed |
| QG2 Code review | — | reviewer | 0 CRIT/HIGH, 2 MED fixed | yes | — | completed |
| QG3 Security audit | — | security-check | 10/10 checks PASS | — | — | completed |
| QG4 Design critique | — | ui-architect | 55/60 → 60/60 (via #16) | — | — | completed |
| #16 FIX mono tonal label + ease-exit | — | ui-engineer | 17/17 pass | yes | — | completed |
| QG5 Spec completion | — | spec-reviewer | 11/11 AC + DoD | — | — | completed |
| Phase 4f AC Validation | — | lead | 7/7 PASS, 1 round | — | — | completed |

### Quality Gate Results
- Build: PASS (frontend `npm run build` 16.70s; pre-commit hook `frontend-build` 22.85s also clean; `go build ./...` clean; `go vet ./...` clean; `go test ./... -race -count=1 -short` all packages PASS)
- Test suite: 652 tests passing (52 files); coverage see Summary
- Coverage gate: PASS with two documented exceptions
- AC validation (unit): PASS — all ACs verified by implementing agents, spot-checked by lead via pre-flight Vitest re-run
- Code review: PASS (QG2) — 2 MEDIUM fixes applied, no CRIT/HIGH
- Security: PASS (QG3) — 10/10 checks PASS, zero CRIT/HIGH. Sanitiser empirically verified against 12 attack variants
- Performance: N/A (no Go backend changes; frontend bundle unchanged except markdown-it addition)
- Spec completion: PASS (QG5) — 11/11 ACs, DoD complete, spec §6.1/§6.2/§7.2 fidelity confirmed
- UI verification: PASS (QG4) — 60/60 aggregate score after #16, cross-story coherence PASS ×3

### AC Validation (Playwright) — Phase 4f
- **Report:** `docs/playwright_cli_US_validate/ui-ast-U6-ast-dispatcher-report.md`
- **Fix rounds:** 1 (max 3)
- **Final pass rate:** 7/7 (100%)
- **Accepted failures:** none
- **Accepted blocked:** none

| Round | PASS | FAIL | BLOCKED | Action |
|-------|------|------|---------|--------|
| 1     | 7    | 0    | 0       | GATE PASSED — no fix loop required |

### Files Changed
Commit: `9a4a61b feat(ui-ast/U6): AstNode dispatcher + passive node components + pendingAst store` (26 files, +1719/-34)

**New:**
- frontend/src/types/uiAst.ts
- frontend/src/components/bmad/AstNode.svelte
- frontend/src/components/bmad/MarkdownBlock.svelte
- frontend/src/components/bmad/HintBanner.svelte
- frontend/src/components/bmad/SummaryCard.svelte
- frontend/src/components/bmad/CodeBlock.svelte
- frontend/src/components/bmad/ComparisonTable.svelte
- frontend/src/components/bmad/linkSanitiser.ts
- frontend/src/components/bmad/__tests__/AstNode.test.ts
- frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts
- frontend/src/components/bmad/__tests__/HintBanner.test.ts
- frontend/src/components/bmad/__tests__/SummaryCard.test.ts
- frontend/src/components/bmad/__tests__/CodeBlock.test.ts
- frontend/src/components/bmad/__tests__/ComparisonTable.test.ts
- frontend/src/components/bmad/__tests__/linkSanitiser.test.ts
- frontend/src/components/bmad/__tests__/mountSvelte.ts
- tests/ac/ui-ast-rendering.spec.ts
- docs/playwright_cli_US_validate/ui-ast-U6-ast-dispatcher-report.md

**Modified:**
- frontend/package.json + frontend/bun.lock (markdown-it@14.1.1 + @types/markdown-it@14.1.2)
- frontend/src/stores/interactiveInput.ts (+structured field, +pendingPrompt singleton, +pendingAst derived, +DEV test seam)
- frontend/src/stores/interactiveInput.test.ts (+4 AC-1..4 cases)
- frontend/src/components/bmad/InputResponseModal.svelte (+ast-region, staggered fly-in, reduce-motion guard, onDestroy cleanup)
- frontend/src/components/bmad/InputResponseModal.test.ts (+3 AC-9 cases)
- frontend/src/views/WorkflowBuilder.svelte (1-line: `structured: payload.structured` passthrough in annotatePrompt)
- docs/stories/ui-ast-U6-ast-dispatcher.md (Status → done, tasks + DoD checked)

### Non-goals held
No DecisionGroup, no response-map collection, no "View raw" toggle, no JSON/collapse submit path, no Monaco bundle change.

### Follow-ups parked (not in scope for U6)
- U7: DecisionGroup component + response-map collection + Cmd+Enter submit path
- U7: Wire `turn_summary` to snackbar (spec §11 Q3)
- U8: "View raw" toggle; replace native `window.confirm` with custom Svelte dialog
- Frontend: extract `prefersReducedMotion` helper (currently duplicated across 5 files — flagged by simplify, out of U6 scope)
- TS types: model UIAST envelope fields (`turn_summary`, `generated_by`, `diagnostics`, `fallback_answer_shape`) when U7/U8 consume them
- Pre-existing failure: `QuestionResponseModal.test.ts` flagged during T6 but not observed in final Vitest run — monitor on next story
- Regenerated `frontend/wailsjs/go/**` bindings + tracked `frontend/coverage/**` + `cover.out` remain in working tree out of U6 scope; left for a subsequent chore/maintenance commit
