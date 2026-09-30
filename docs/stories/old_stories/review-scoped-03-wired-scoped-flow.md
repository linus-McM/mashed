# Story 3: Wire Selection to Scoped Advice and Enriched Plan

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** Story 1 (review-scoped-01), Story 2 (review-scoped-02)
**Status:** done

## Description

Wire the frontend file selection (Story 2) to the backend `StreamScopedAdvice` method (Story 1), enabling the full scoped review flow: select files, get scoped advice, re-generate with prior context, and create a refactor plan enriched with file summaries and scoped advice. This completes the per-file selection feature end-to-end.

## Developer Notes

### Architecture

- **File to modify:** `/Users/dev/Development/mashed/frontend/src/views/SummarisationModal.svelte`
- **Binding to import:** `StreamScopedAdvice` from `../../wailsjs/go/main/App.js` (auto-generated after Story 1 + `wails dev` restart)
- **No backend changes** -- Story 1 provides the backend; this story wires it.

### State Changes

Add to the `<script>` section:

```typescript
// Prior advice for incremental context chaining
let priorAdvice = '';
```

### Function Changes

**`getAdvice()` (line 125):**
Replace `StreamAdvice(repoPath, selectedMode, selectedModel)` with:
```typescript
function getAdvice() {
  if (!canGetAdvice) return;
  // Save current advice as prior context for re-generation
  if (adviceText) {
    priorAdvice = adviceText;
  }
  adviceText = '';
  adviceError = '';
  adviceLoading = true;
  const filePaths = Array.from(selectedFiles);
  const context = buildAdditionalContext();
  StreamScopedAdvice(repoPath, selectedMode, selectedModel, filePaths, context);
}
```

**New `buildAdditionalContext()` function:**
```typescript
function buildAdditionalContext(): string {
  const parts: string[] = [];

  // Include summaries for selected files
  const selectedSummaries = files
    .filter(f => selectedFiles.has(f.path))
    .map(f => `- **${f.path}** (+${f.added}/-${f.removed}): ${f.summary}`)
    .join('\n');
  if (selectedSummaries) {
    parts.push('## File Summaries\n' + selectedSummaries);
  }

  // Include prior advice if re-generating
  if (priorAdvice) {
    parts.push('## Previous Advice\n' + priorAdvice);
  }

  return parts.join('\n\n');
}
```

**`createRefactorPlan()` (line 133):**
Enrich the advice text passed to `SpawnRefactorPlan`:
```typescript
async function createRefactorPlan() {
  if (!canCreatePlan) return;
  planLoading = true;
  planError = '';
  try {
    const enrichedAdvice = buildEnrichedAdvice();
    planPath = await SpawnRefactorPlan(repoPath, enrichedAdvice);
  } catch (e) {
    planError = e?.message || 'Failed to create plan';
  }
  planLoading = false;
}
```

**New `buildEnrichedAdvice()` function:**
```typescript
function buildEnrichedAdvice(): string {
  const parts: string[] = [];

  // Selected files list
  const fileList = Array.from(selectedFiles).join(', ');
  parts.push('## Scoped Files\n' + fileList);

  // File summaries for selected files
  const summaries = files
    .filter(f => selectedFiles.has(f.path))
    .map(f => `- **${f.path}**: ${f.summary}`)
    .join('\n');
  if (summaries) {
    parts.push('## File Summaries\n' + summaries);
  }

  // The advice text
  parts.push('## Code Review Advice\n' + adviceText);

  return parts.join('\n\n');
}
```

### Import Changes

Update the import line (line 7):
```typescript
import { StreamCodeReviewSummary, ListAdviceModes, StreamAdvice, StreamScopedAdvice, SpawnRefactorPlan, ListModels } from '../../wailsjs/go/main/App.js';
```

Note: `StreamAdvice` can remain imported for backward compatibility but is no longer called from this modal. It may be removed if no other component uses it.

### Technical Considerations

- **Wails binding availability:** `StreamScopedAdvice` will only exist in `App.js` / `App.d.ts` after Story 1 is merged and `wails dev` is restarted. If developing in parallel, the import will show a TypeScript error until bindings regenerate.
- **Context string size:** `buildAdditionalContext` can produce large strings if many files with long summaries are selected. The backend `StreamScopedAdvice` passes this via stdin (not CLI args), so there is no 256KB arg limit concern.
- **Prior advice accumulation:** Only one level of prior advice is kept (the most recent `adviceText` before re-generation). This is intentional -- deeper chains would make the context unwieldy.
- **Plan enrichment truncation:** `SpawnRefactorPlan` already truncates adviceText to 10000 chars (line 341 of `app_review.go`). The enriched string may be longer than raw advice, but the truncation is still appropriate as a safety net.

### Risks & Edge Cases

- **Selection changed between advice and plan:** If the user changes their file selection after getting advice but before creating a plan, `buildEnrichedAdvice` uses the current selection (not the selection at advice time). This is acceptable -- the user sees which files are selected.
- **Re-generate with fewer files:** If the user first selects 5 files, gets advice, then selects only 2 files and re-generates, the prior advice (from 5 files) is prepended as context. The new advice will naturally focus on the 2 selected files.
- **adviceText contains markdown:** The advice text is raw markdown from Claude. When embedded in `buildEnrichedAdvice`, the heading levels may conflict. Using `## Code Review Advice` as a wrapper heading mitigates this.

### Reference Files

- `/Users/dev/Development/mashed/frontend/src/views/SummarisationModal.svelte` -- primary file
- `/Users/dev/Development/mashed/app_review.go` -- `StreamScopedAdvice` signature (Story 1), `SpawnRefactorPlan` input handling (line 324)
- `/Users/dev/Development/mashed/frontend/wailsjs/go/main/App.d.ts` -- verify binding exists after Story 1

## Acceptance Criteria

AC-1: Scoped advice uses selected files
- Given the user has selected 2 of 5 file cards in the modal
- When the user selects a methodology and clicks "Get Advice"
- Then `StreamScopedAdvice` is called with exactly those 2 file paths
- And the streamed advice is scoped to those files only

AC-2: File summaries included in advice context
- Given the user has selected files with AI-generated summaries
- When "Get Advice" is clicked
- Then the `additionalContext` parameter includes a "File Summaries" section with each selected file's summary

AC-3: Prior advice chained on re-generation
- Given the user has already generated advice text "First pass: refactor error handling"
- When the user selects different/additional files and clicks "Get Advice" again
- Then the previous advice is saved as `priorAdvice`
- And the `additionalContext` parameter includes a "Previous Advice" section with the prior text
- And the advice panel clears and streams new advice

AC-4: Enriched plan includes file context
- Given advice text has been generated for selected files
- When the user clicks "Create Refactor Plan"
- Then `SpawnRefactorPlan` receives an enriched string containing:
  - "Scoped Files" section listing selected file paths
  - "File Summaries" section with per-file summaries
  - "Code Review Advice" section with the full advice text

AC-5: End-to-end flow
- Given a repo with code changes
- When the user opens Summarise, waits for summaries, selects 2 files, picks a methodology, clicks "Get Advice", waits for advice, then clicks "Create Refactor Plan"
- Then each step succeeds and the plan file is created at the returned path
- And the plan link is clickable in the UI

## BDD Test Scenarios

### Scenario 1: Scoped advice call

```gherkin
Feature: Wire file selection to StreamScopedAdvice

  Scenario: Selected files passed to backend
    Given the modal has files ["main.go", "util.go", "config.go", "test.go", "doc.go"]
    And the user has selected "main.go" and "util.go"
    And the user has chosen methodology "Clean Code"
    When the user clicks "Get Advice"
    Then StreamScopedAdvice is called with filePaths ["main.go", "util.go"]
    And the model parameter matches the selected model

  Scenario: Additional context includes file summaries
    Given "main.go" has summary "Refactored error handling"
    And "util.go" has summary "Added helper functions"
    And both are selected
    When getAdvice() builds the additional context
    Then the context contains "## File Summaries"
    And contains "main.go" with "Refactored error handling"
    And contains "util.go" with "Added helper functions"
```

### Scenario 2: Incremental re-generation

```gherkin
Feature: Prior advice context chaining

  Scenario: First advice generation has no prior context
    Given no previous advice has been generated (priorAdvice is empty)
    When the user clicks "Get Advice" for the first time
    Then buildAdditionalContext does not include "## Previous Advice"

  Scenario: Re-generation includes prior advice
    Given the user previously generated advice "Check null returns in main.go"
    And adviceText is "Check null returns in main.go"
    When the user selects new files and clicks "Get Advice" again
    Then priorAdvice is set to "Check null returns in main.go"
    And adviceText is cleared to empty string
    And buildAdditionalContext includes "## Previous Advice" with "Check null returns in main.go"

  Scenario: Only most recent prior advice is kept
    Given the user has generated advice twice (first: "Advice A", second: "Advice B")
    When the user clicks "Get Advice" a third time
    Then priorAdvice contains only "Advice B" (not "Advice A")
```

### Scenario 3: Enriched plan

```gherkin
Feature: Enriched refactor plan context

  Scenario: Plan includes file list and summaries
    Given selected files are "main.go" and "util.go"
    And "main.go" summary is "Error handling refactor"
    And "util.go" summary is "New helpers"
    And adviceText is "Priority 1: fix error wrapping"
    When the user clicks "Create Refactor Plan"
    Then SpawnRefactorPlan receives a string containing "## Scoped Files"
    And containing "main.go, util.go"
    And containing "## File Summaries"
    And containing "## Code Review Advice"
    And containing "Priority 1: fix error wrapping"
```

### Scenario 4: End-to-end flow

```gherkin
Feature: Complete scoped review flow

  Scenario: Full happy path
    Given the Summarise modal is open for a repo with 5 changed Go files
    And summaries have finished streaming
    When the user clicks "Select All"
    And deselects 3 files (keeping 2 selected)
    And selects methodology "Security Review"
    And clicks "Get Advice"
    And waits for advice to finish streaming
    And clicks "Create Refactor Plan"
    Then a plan file path is returned
    And the plan link appears in the UI

  Scenario: Advice then deselect all prevents plan creation
    Given the user has generated advice with 2 files selected
    When the user clicks "Select None"
    Then the "Create Refactor Plan" button becomes disabled
    And the advice text remains visible in the panel
```

## Tasks / Subtasks

- [ ] Task 1: Import `StreamScopedAdvice` binding (AC: 1)
  - [ ] Subtask 1a: Add `StreamScopedAdvice` to the import from `../../wailsjs/go/main/App.js`
  - [ ] Subtask 1b: Verify binding exists in `App.d.ts` (requires Story 1 + `wails dev` restart)

- [ ] Task 2: Add `priorAdvice` state and `buildAdditionalContext` function (AC: 2, 3)
  - [ ] Subtask 2a: Add `let priorAdvice = ''` state variable
  - [ ] Subtask 2b: Implement `buildAdditionalContext()` that assembles selected file summaries + prior advice

- [ ] Task 3: Rewire `getAdvice()` to use `StreamScopedAdvice` (AC: 1, 3)
  - [ ] Subtask 3a: Save current `adviceText` to `priorAdvice` before clearing
  - [ ] Subtask 3b: Build `filePaths` array from `selectedFiles` Set
  - [ ] Subtask 3c: Call `StreamScopedAdvice` with filePaths and additional context

- [ ] Task 4: Implement `buildEnrichedAdvice` and rewire `createRefactorPlan` (AC: 4)
  - [ ] Subtask 4a: Implement `buildEnrichedAdvice()` with scoped files, summaries, and advice sections
  - [ ] Subtask 4b: Replace `adviceText` with `enrichedAdvice` in `SpawnRefactorPlan` call

- [ ] Task 5: End-to-end browser test (AC: 5)
  - [ ] Subtask 5a: Manual test: open Summarise, select files, get scoped advice, create plan
  - [ ] Subtask 5b: Verify plan file is created and link is clickable

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass via manual browser test or Playwright
- [ ] End-to-end flow verified: select files -> get advice -> create plan
- [ ] `go build ./...` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Prior advice chaining verified with at least 2 consecutive advice generations
