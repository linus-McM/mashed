# AC Validation Report — ui-ast-U7 DecisionGroup

**Story:** `docs/stories/ui-ast-U7-decision-group.md`
**Validation mode:** single-story (invoked from `/team-sprint` Phase 4f)
**App:** `wails dev` at http://localhost:34115
**Run date:** 2026-04-22
**Gate:** `MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts tests/ac/ui-ast-decision-group.spec.ts`

## Summary

| Category | Count |
|----------|-------|
| Total ACs | 11 |
| UI-testable (Playwright) | 6 |
| Unit/Component-only (Vitest) | 5 |
| PASS | 11 |
| FAIL | 0 |
| BLOCKED | 0 |

**Outcome:** GATE PASSED — 11/11 ACs verified. 1 round. No fixes required in Phase 4f.

## AC Classification + Results

| AC | Description | Classification | Evidence | Result |
|----|-------------|----------------|----------|--------|
| AC-1 | DecisionGroup dispatches six widget types | Unit (Vitest) | `DecisionGroup.test.ts::AC1_dispatches_six_widget_types` ×6 | PASS |
| AC-2 | responses store accumulates keys | Unit (Vitest) | `astResponses.test.ts::AC2_responses_accumulate` | PASS |
| AC-3 | ShapeJSON submits full map as JSON | Playwright | `json-submit-full-map` L162 (4.7s) | PASS |
| AC-4 | Single group on non-JSON shape → plain string | Playwright | `single-group-plain-string-submit` L192 (2.2s) | PASS |
| AC-5 | Multi-group non-JSON → collapse banner visible | Playwright | `collapse-banner-visible` L212 (2.5s) | PASS |
| AC-6 | User switches active group under collapse | Playwright | `collapse-user-switches-active` L238 (2.4s) + FIX-#16 extension asserting payload of active group | PASS |
| AC-7 | Send disabled until required filled | Unit (Vitest) | `InputResponseModal.test.ts::AC7_send_disabled_until_required_filled` + `AC7_collapse_send_gated_on_active` | PASS |
| AC-8 | Cmd+Enter submits the form | Playwright | `cmd-enter-submits` L274 (2.4s) — Meta+Enter fires RespondToInput with same payload | PASS |
| AC-9 | Zero groups → Layer-1 fallback widget | Playwright | `zero-groups-fallback-layer1` L295 (2.5s) | PASS |
| AC-10 | `aria-labelledby` wires heading → widget | Unit (Vitest) | `DecisionGroup.test.ts::AC10_aria_labelledby_wired` | PASS |
| AC-11 | §7.1 — file widget path sent verbatim | Unit (Vitest) | `InputResponseModal.test.ts::AC11_file_widget_passthrough` | PASS |

## Playwright Run

```
Running 6 tests using 1 worker

✓  json-submit-full-map (AC-3)                 4.7s
✓  single-group-plain-string-submit (AC-4)     2.2s
✓  collapse-banner-visible (AC-5)              2.5s
✓  collapse-user-switches-active (AC-6)        2.4s
✓  cmd-enter-submits (AC-8)                    2.4s
✓  zero-groups-fallback-layer1 (AC-9)          2.5s

6 passed (17.5s)
```

## Fix Rounds

| Round | PASS | FAIL | BLOCKED | Action |
|-------|------|------|---------|--------|
| 1     | 6/6  | 0    | 0       | GATE PASSED — no fix loop |

## Notes

- AC-6 Playwright extended during FIX #16 to assert that `RespondToInput` receives the **user-selected** active group's value after a collapse-rule swap, not the first-required group's value. This also implicitly covers the bug fix from QG2.
- AC-11 file-widget passthrough is asserted via Vitest mock of `RespondToInput`. Backend `resolveFileInput` path-traversal rejection is out of scope for the frontend AC.
- Playwright suite is gated on `MASHED_E2E=1` so CI runs skip cleanly without `wails dev`.

## Final Status

All UI-testable ACs PASS against the live application. Unit-testable ACs PASS in Vitest. Phase 4f gate: **PASSED**. Unblocks Phase 5 (commit + story status + sprint report).
