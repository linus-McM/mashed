# AC Validation Report: ui-ast-U6 — AstNode dispatcher + node components + `pendingAst` derived store

**Generated:** 2026-04-22T07:50:00Z
**App URL:** http://localhost:34115
**Stories validated:** 1 / 1 (single-story mode — invoked from /team-sprint Phase 4f)

## Summary

| Story      | UI ACs | Passed | Failed | Blocked | Backend-Only |
|------------|--------|--------|--------|---------|--------------|
| ui-ast-U6  | 7      | 7      | 0      | 0       | 4            |

**Overall pass rate:** 7/7 (100%) on UI-testable ACs. 4/4 backend-only ACs verified via Vitest in QG1.

---

## Story: ui-ast-U6 — AstNode dispatcher + passive node components + `pendingAst` derived store

### AC-1: `PendingPrompt` wire format carries `structured?: string` — BACKEND-ONLY

**Classification:** backend-only (compile-time type guarantee; not UI-observable)
**Status:** BACKEND-ONLY (satisfied)
- **Note:** Type-level contract on the `PendingPrompt` interface. Validated by Vitest `AC1_structured_field_optional` in `frontend/src/stores/interactiveInput.test.ts:154`. Both `{...fields, structured: 'x'}` and `{...fields}` compile without error — the `?:` guarantee.

### AC-2: `pendingAst` derived store parses valid v1 AST — BACKEND-ONLY

**Classification:** backend-only (store projection behavior; indirectly visible via AC-5/AC-9)
**Status:** BACKEND-ONLY (satisfied)
- **Note:** Validated by Vitest `AC2_pending_ast_parses_valid_v1` in `interactiveInput.test.ts:162`. Store emits parsed UIAST with `version === '1'` when `pendingPrompt.structured` is a valid v1 JSON string. Downstream UI behavior covered by AC-5 and AC-9 Playwright tests below.

### AC-3: `pendingAst` emits null on malformed JSON — BACKEND-ONLY

**Classification:** backend-only
**Status:** BACKEND-ONLY (satisfied)
- **Note:** Validated by Vitest `AC3_pending_ast_null_on_malformed` in `interactiveInput.test.ts:169`. `structured: "not-json"` → store emits `null`. Malformed input triggers Layer-1 fallback; no Playwright assertion needed since observable UI state is identical to "no structured field present" (covered by AC-9's `modal-layer1-unchanged-when-null`).

### AC-4: `pendingAst` emits null on unknown version — BACKEND-ONLY

**Classification:** backend-only
**Status:** BACKEND-ONLY (satisfied)
- **Note:** Validated by Vitest `AC4_pending_ast_null_on_unknown_version` in `interactiveInput.test.ts:174`. `{"version":"2"}` → store emits `null`. Same degradation path as AC-3 — UI invariance to Layer-1 fallback is directly observable via AC-9's null-case Playwright test.

### AC-5: `AstNode` renders each of the six node shapes — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Playwright test `six-node-shapes (AC-5) — all passive types render via AstNode dispatcher` at `tests/ac/ui-ast-rendering.spec.ts:80` passed in 2.5s against live `wails dev`. Seeded a UIAST fixture with one of each {markdown, hint, summary, code, table, unknown-type} via `window.__mashedEmitBmadEvent('bmad:node:awaiting_input', ...)`. Asserted inside `[data-testid="ast-region"]`:
  - `[data-testid="markdown-block"]` visible (×2: the `markdown` node + the `unknown` fallback)
  - `[data-testid="hint-banner"]` visible
  - `[data-testid="summary-card"]` visible
  - `[data-testid="code-block"]` visible
  - `[data-testid="comparison-table"]` visible
- Also covered by Vitest `AC5_six_node_shapes` in `AstNode.test.ts:52`.
- **Screenshot:** captured by Playwright runner (test passed; no failure screenshot needed)

### AC-6: §7.2 security — `javascript:` href is stripped to plain text — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Playwright test `javascript-href-stripped-dom (AC-6) — no javascript: anchors in rendered markdown` at `tests/ac/ui-ast-rendering.spec.ts:114` passed in 2.0s. Rendered `[click me](javascript:alert(1))` → assertion `locator('[data-testid="ast-region"] a[href^="javascript:"]').count() === 0` held. Text "click me" remained visible in DOM. Zero JavaScript execution.
- Also covered by Vitest `AC6_javascript_href_stripped` in `MarkdownBlock.test.ts:60`.

### AC-7: §7.2 security — `data:` and `file:` schemes are rejected identically — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Playwright test `data-and-file-schemes-rejected-dom (AC-7) — data: and file: hrefs absent` at `tests/ac/ui-ast-rendering.spec.ts:130` passed in 2.4s. Rendered markdown with both `[x](data:text/html,...)` and `[y](file:///etc/passwd)`. Assertions:
  - `a[href^="data:"]` count = 0
  - `a[href^="file:"]` count = 0
  - Text "x" and "y" visible (degraded to plain text, not removed)
- Also covered by Vitest `AC7_data_and_file_schemes_rejected` in `MarkdownBlock.test.ts:74`.

### AC-8: §7.2 security — valid https link routes through confirm dialog — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Playwright test `https-link-confirm-dialog (AC-8) — confirm gates BrowserOpenURL` at `tests/ac/ui-ast-rendering.spec.ts:152` passed in 2.0s. `page.addInitScript` installed a setter on `window.runtime` that survived Wails' late assignment, wrapping `BrowserOpenURL` with a call counter. `window.confirm` stubbed to return true/false across two clicks on `[docs](https://example.com)`.
  - confirm=true → 1 BrowserOpenURL call recorded
  - confirm=false → no new call
- Also covered by Vitest `AC8_https_link_routes_through_confirm` in `MarkdownBlock.test.ts:85,98`.

### AC-9: `InputResponseModal` renders AST above Layer-1 widget when `pendingAst != null` — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Two Playwright tests at `tests/ac/ui-ast-rendering.spec.ts:230,252` passed:
  1. `modal-renders-ast-above-widget (AC-9) — ast-region precedes Layer-1 widget` (2.0s) — verified `compareDocumentPosition` confirms `ast-region` precedes the Layer-1 `FreeTextWidget` wrapper in DOM order.
  2. `modal-layer1-unchanged-when-null (AC-9) — no ast-region when structured missing` (1.9s) — seeded a PendingPrompt without `structured` field; `[data-testid="ast-region"]` count = 0; Layer-1 widget still rendered normally.
- Also covered by Vitest `AC9_renders_ast_above_widget`, `AC9_layer1_unchanged_when_null`, `AC9_respects_prefers_reduced_motion` in `InputResponseModal.test.ts:160,172,180`.

### AC-10: Unknown node type renders as markdown, not removed — PASS

**Classification:** UI-testable (via the six-shapes fixture)
**Status:** PASS
- **Evidence:** The AC-5 Playwright fixture included an unknown-type node `{type: 'snarkfish', content: 'hello'}`. Assertion in `six-node-shapes` test confirmed the unknown node rendered as a second `[data-testid="markdown-block"]` element (total count = 2), with no JavaScript exception surfaced to console.
- Also covered by Vitest `AC10_unknown_type_markdown_fallback` and `AC10_unknown_type_does_not_throw` in `AstNode.test.ts:57,63`.

### AC-11: `CodeBlock` copy-to-clipboard copies verbatim content — PASS

**Classification:** UI-testable
**Status:** PASS
- **Evidence:** Playwright test `code-copy-button-works (AC-11) — clipboard receives content, button flips to "Copied"` at `tests/ac/ui-ast-rendering.spec.ts:264` passed in 2.1s. Seeded a `code` node with content `"rm -rf ~"` and `copyable: true`. Clicked `[data-testid="code-copy-button"]`. Asserted:
  - `navigator.clipboard.readText()` returned `"rm -rf ~"` verbatim
  - Button's textContent contained `"Copied"` post-click
- Also covered by Vitest `AC11_copy_writes_verbatim` and `AC11_copy_label_flips_to_copied` in `CodeBlock.test.ts:31,41`.

---

## Phase 4f Validation Summary

- **Round 1:** 7/7 UI ACs PASS, 4 backend-only ACs satisfied via Vitest. No FAIL, no BLOCKED.
- **GATE PASSED** in a single round. No fix loop required.

## Notes

- Wails dev server (:34115) was already running when validation started; not stopped at end per skill convention (user started it for the opt-in Phase 4f, not this skill).
- Per-AC screenshots omitted since all 7 Playwright cases produced PASS results and visual evidence (DOM queries, counts, text contents) is fully captured in the assertion traces; no screenshots required by report template for PASS cases beyond what the test runner's HTML report already contains.
- Playwright run command:
  `MASHED_E2E=1 npx playwright test --config frontend/playwright.config.ts tests/ac/ui-ast-rendering.spec.ts --reporter=list`
- Total duration: 15.6s (7 tests, 1 worker).
