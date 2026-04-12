# skills-editor-02: Save frontmatter round-trip via new Wails binding

**Status:** ready
**Domain:** fullstack
**Size:** M
**Depends on:** skills-editor-01
**Phase:** 4a

## Description

Wire the Save button on the skill editor modal to a new Wails binding that reads the existing markdown file, parses out the YAML frontmatter, merges the edited fields, and writes it back — preserving the markdown body verbatim. Round-trip symmetry is the headline requirement: saving without changes must leave the file byte-identical (modulo deterministic YAML key ordering).

After save succeeds, the sidebar refreshes its copy of the asset (re-calls `ListAllMashedAssets`) so the UI reflects the new frontmatter immediately.

**This is a UI-facing story (save flow + toast) — route through ui-architect.**

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/assets.go` (or new `assets_write.go`):
    - Add `WriteMashedAssetFrontmatter(path string, asset MashedAssetInfo) error` — reads the file, splits frontmatter from body (respecting the existing `---` delimiter convention), marshals the updated frontmatter using `gopkg.in/yaml.v3`, rewrites the file atomically via a tempfile + rename.
  - `app.go` (BMAD section ~line 1547):
    - Add Wails binding `SaveMashedAsset(path string, asset bmad.MashedAssetInfo) error` — nil-check, delegate, return typed result. Follow the existing BMAD binding pattern (cerebrum Key Learning).
  - `frontend/src/components/bmad/SkillEditorModal.svelte`:
    - Wire Save button to the new binding
    - On success: close modal + emit `assetSaved` event + show toast
    - On error: display inline error + keep modal open
  - `frontend/src/views/WorkflowBuilder.svelte`:
    - Handle `assetSaved` → re-call `ListAllMashedAssets` and update `groupedMashedAssets`
  - `internal/bmad/assets_write_test.go` — NEW. Round-trip tests.
- **Atomic write pattern:**
  ```go
  tmp, err := os.CreateTemp(filepath.Dir(path), ".mashed-save-*")
  if err != nil { return fmt.Errorf("create temp: %w", err) }
  defer os.Remove(tmp.Name()) // cleanup on failure
  if _, err := tmp.Write(newContent); err != nil { return fmt.Errorf("write temp: %w", err) }
  if err := tmp.Close(); err != nil { return fmt.Errorf("close temp: %w", err) }
  if err := os.Rename(tmp.Name(), path); err != nil { return fmt.Errorf("rename: %w", err) }
  ```
- **YAML key ordering:** `yaml.v3` preserves order when the input is a `yaml.Node`, not a `map`. Parse into a `*yaml.Node`, mutate fields in place, marshal back. Using a plain map will shuffle keys and break round-trip symmetry.
- **Risks / gotchas:**
  - **Round-trip symmetry is load-bearing.** A user who opens and saves without edits must get a byte-identical file. Test it.
  - **Preserve the body verbatim.** The editor only touches frontmatter. The body (everything after the second `---`) is opaque and must not be re-serialised.
  - **Handle files without frontmatter gracefully** — if the source file lacks a frontmatter block (legacy skill), fail with a typed error `ErrNoFrontmatterBlock` and surface a user-readable message in the modal. Do not silently prepend one.
  - **Wails binding regeneration**: cerebrum notes `wails dev` regenerates bindings. Run it after adding `SaveMashedAsset`.
  - Path validation: reject paths outside `~/.claude/` and `{repo}/.claude/` to prevent arbitrary file writes. Use `filepath.Clean` and a whitelist check.
- **Error handling** (per user's memory `feedback_error_handling.md`): sentinel errors + wrapping + custom types where appropriate. `ErrNoFrontmatterBlock`, `ErrPathOutsideAllowedRoots` as sentinels. Wrap with `%w`.
- **Prerequisites already in place:**
  - `yaml.v3` is in go.mod (cerebrum Key Learning).
  - `MashedAssetInfo` type exists in `internal/bmad/assets.go`.
  - `ListAllMashedAssets` Wails binding exists from Phase 1.
  - skills-editor-01 has the modal shell and form state.

## Acceptance Criteria

**AC-1: Round-trip symmetry — unchanged save yields byte-identical file**
- Given a SKILL.md file with valid frontmatter and a multi-line body
- When the user opens the editor modal and clicks Save without changing any field
- Then the resulting file is byte-identical to the original

**AC-2: Round-trip preservation — edited field persists, others unchanged**
- Given a SKILL.md file with `mashedRole: skill` and other frontmatter fields
- When the user changes `mashedRole` to `command` and clicks Save
- Then the file's frontmatter has `mashedRole: command`
- And all other frontmatter fields are unchanged
- And the body section is byte-identical to the original

**AC-3: Save failure surfaces in the modal**
- Given a file that cannot be written (permission denied simulated via a file with 0400 mode)
- When the user clicks Save
- Then the modal stays open
- And an inline error message is displayed
- And no partial write occurred (atomic rename was rolled back)

**AC-4: Sidebar refreshes after successful save**
- Given the modal saved an asset whose `mashedRole` changed from `skill` to `command`
- When the sidebar receives the `assetSaved` event
- Then `ListAllMashedAssets` is re-invoked
- And the asset moves from the "Skills" group to the "Commands" group

**AC-5: Path outside allowed roots is rejected**
- Given a malicious payload with `path: "/etc/passwd"`
- When `SaveMashedAsset` is called
- Then it returns `ErrPathOutsideAllowedRoots`
- And no write occurred

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Save skill frontmatter

  Scenario: Unchanged save is byte-identical
    Given an existing SKILL.md with valid frontmatter and body
    When the editor saves without changing any field
    Then the file on disk is byte-identical to the original

  Scenario: Edit mashedRole persists
    Given a SKILL.md with mashedRole skill
    When the user changes mashedRole to command and saves
    Then the frontmatter now reads mashedRole command
    And other fields and body are unchanged

  Scenario: Write failure surfaces cleanly
    Given a file the process cannot write
    When the user clicks Save
    Then the modal displays an inline error
    And the file on disk is unchanged (atomic)

  Scenario: Sidebar refreshes after save
    Given a skill was re-roled to command
    When the assetSaved event fires
    Then the sidebar re-invokes ListAllMashedAssets
    And the asset appears in the Commands group

  Scenario: Reject path outside allowed roots
    Given a path of /etc/passwd
    When SaveMashedAsset is invoked
    Then it returns ErrPathOutsideAllowedRoots
    And no filesystem write occurred

  Scenario: File without frontmatter returns typed error
    Given a legacy SKILL.md with no frontmatter block
    When SaveMashedAsset is invoked
    Then it returns ErrNoFrontmatterBlock
```

## Tasks / Subtasks

- [ ] Task 1 — Backend `WriteMashedAssetFrontmatter` (AC-1, AC-2, AC-3, AC-5)
  - [ ] Parse source file with `yaml.Node` (preserves key order)
  - [ ] Mutate fields in place
  - [ ] Atomic tempfile + rename write
  - [ ] Path-allowlist guard returning `ErrPathOutsideAllowedRoots`
  - [ ] Sentinel `ErrNoFrontmatterBlock` for legacy files
- [ ] Task 2 — Wails binding `SaveMashedAsset` (AC-1, AC-2, AC-3, AC-5)
  - [ ] Follow existing BMAD binding pattern (nil-check → delegate → typed return)
  - [ ] Regenerate bindings via `wails dev`
- [ ] Task 3 — Frontend save flow (AC-3, AC-4)
  - [ ] Wire Save button to the new binding
  - [ ] Success: close modal, emit `assetSaved`, toast
  - [ ] Failure: inline error, modal stays open
- [ ] Task 4 — Sidebar refresh (AC-4)
  - [ ] Listen for `assetSaved` in `WorkflowBuilder.svelte`
  - [ ] Re-call `ListAllMashedAssets` and update store
- [ ] Task 5 — Tests (AC-1 through AC-5)
  - [ ] Round-trip unchanged test
  - [ ] Round-trip edited test (assert body byte-identical)
  - [ ] Write-failure test (0400 mode fixture)
  - [ ] Allowlist guard test
  - [ ] Legacy-file sentinel test
  - [ ] Playwright: end-to-end edit + save + sidebar refresh

## Definition of Done

- [ ] All ACs verified by an automated test (Go + Playwright; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths, magic numbers, or color literals added
- [ ] Wails bindings regenerated and committed
- [ ] Existing tests still pass

## Design Brief

### Layout composition
Builds on the `SkillEditorModal` shell from skills-editor-01. This story adds three visual affordances:

1. **Save button loading state** — inside the existing footer, the Save button reserves its width and swaps label → spinner + "Saving…". `min-width: 120px` on the button so the modal chrome does not reflow.
2. **Inline error banner** — a new slot in the footer, ABOVE the action row, spanning full footer width. `padding: var(--sp-sm) var(--sp-md);` `border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent);` `background: color-mix(in srgb, var(--accent-red) 10%, transparent);`. Only renders when an error exists.
3. **Success toast** — app-level toast that replaces the modal on success, anchored `position: fixed; top: var(--sp-lg); right: var(--sp-lg);` so the sidebar refresh (the main success signal) is the primary feedback, not a dismissible modal alert.

Footer restructure:
```
modal-footer:
  [inline error banner — conditional, grid-column: 1 / -1]
  [action row — flex-end, gap: var(--sp-sm)]
```

### Typography plan
- **Error banner text**: `font-family: var(--font-ui)`, `font-size: var(--text-label)` (11px), `font-weight: 500`, `color: var(--accent-red)`. Prefix with a Lucide `AlertCircle` icon at 12px, same color, inline.
- **Save button "Saving…" state**: same `font-size: var(--text-body)` and `font-weight: 600` as idle state — no typography drift during the transition.
- **Toast body**: `font-family: var(--font-ui)`, `font-size: var(--text-body)`, `color: var(--text-primary)`.
- **Toast subtitle** (file path that was saved): `font-family: var(--font-mono)`, `font-size: var(--text-label)`, `color: var(--text-dim)`.

### Color strategy
- **Save button idle**: `var(--accent-green)` on `var(--bg-deepest)` text (from skills-editor-01).
- **Save button loading**: `background: color-mix(in srgb, var(--accent-green) 70%, transparent)` + disabled cursor. `opacity: 0.8` if preferred; `opacity` keeps the green readable.
- **Save button success flash** (optional, 200ms): `background: var(--accent-green)` + `box-shadow: var(--glow-spread) var(--accent-green)` — ONE pulse, then the modal unmounts. Uses the existing `--glow-spread` token.
- **Error banner**: `background: color-mix(in srgb, var(--accent-red) 10%, transparent)`, `border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent)`, `color: var(--accent-red)`. Alpha grammar matches `.btn-discard:hover` / `.artifact-icon.missing`.
- **Toast surface**: `var(--bg-elevated)`, border `1px solid color-mix(in srgb, var(--accent-green) 40%, transparent)`, box-shadow reused from modal (`0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent)`).
- **Toast icon**: Lucide `Check` in `var(--accent-green)`.
- **Sidebar refresh pulse** (post-save highlight): the refreshed asset row gets a 400ms background flash from `color-mix(in srgb, var(--accent-green) 15%, transparent)` back to `transparent`. This is the "I saw your change" signal and ties the modal commit to the sidebar state.

### Interaction model
- **Click Save** → button transitions to loading state (`var(--duration-short)`); Cancel button disables.
- **Save success** → modal fades out over `var(--duration-medium)`, toast fades in top-right over `var(--duration-medium)`, sidebar row flashes green. Toast auto-dismisses after `2500ms` or on click.
- **Save error** → button returns to idle, inline error banner slides down over `var(--duration-medium) var(--ease-enter)`, focus moves to Save button. Modal stays open.
- **Keyboard**: `Enter` on any single-line input triggers Save when `canSave` is true. `Escape` dismisses the error banner first (if visible), then the modal on the second press.
- **Focus management**: after error, focus returns to Save. After toast dismiss, focus returns to the triggering sidebar row.
- **Sidebar refresh timing**: the `assetSaved` event fires the moment the backend returns success. Sidebar refetches, renders, then triggers the flash highlight on the moved/changed row via `class:just-saved={asset.path === lastSavedPath}` and a 400ms timeout that clears the class.

### Component specs
```
.inline-error {
  grid-column: 1 / -1;
  padding: var(--sp-sm) var(--sp-md);
  background: color-mix(in srgb, var(--accent-red) 10%, transparent);
  border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent);
  color: var(--accent-red);
  font-size: var(--text-label);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: var(--sp-xs);
}
.btn-save { min-width: 120px; }
.btn-save.saving {
  opacity: 0.8;
  cursor: not-allowed;
  background: color-mix(in srgb, var(--accent-green) 70%, transparent);
}
.btn-save.saving::before {
  content: '';
  display: inline-block;
  width: 10px; height: 10px;
  margin-right: var(--sp-xs);
  border-radius: 50%;
  border: 2px solid var(--bg-deepest);
  border-top-color: transparent;
  animation: spin 700ms linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.save-toast {
  position: fixed;
  top: var(--sp-lg);
  right: var(--sp-lg);
  padding: var(--sp-sm) var(--sp-md);
  background: var(--bg-elevated);
  border: 1px solid color-mix(in srgb, var(--accent-green) 40%, transparent);
  border-radius: var(--radius-md);
  box-shadow: 0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  display: flex;
  align-items: center;
  gap: var(--sp-sm);
  z-index: 500;
}
.save-toast .icon { color: var(--accent-green); }
.save-toast .path { font-family: var(--font-mono); font-size: var(--text-label); color: var(--text-dim); }

.process-item.just-saved {
  animation: row-flash 400ms var(--ease-enter);
}
@keyframes row-flash {
  0%   { background: color-mix(in srgb, var(--accent-green) 15%, transparent); }
  100% { background: transparent; }
}
```
No toast token system exists — the toast spec hand-rolls values consistent with modal precedent. Shadow token still missing; reuse modal shadow verbatim.

### Signature elements
- **Three-way commit choreography**: button spin → modal dismiss → toast appear → sidebar row flash. Four separate places acknowledge the save. Each is under 200ms and none are blocking. This IS the "notification-first" product thesis playing out inside a single user action.
- **Row flash on the sidebar** is the quiet hero: the user sees their edit "land" in the list, not a modal yelling "saved!". Linear does this well; we go further by making the flash match the brand green exactly.
- **Mono file path in the toast** repeats the signature move from the modal title — path as first-class feedback.

### Distinguishing from existing patterns
- **vs a generic browser `alert()`** — no blocking, no OS chrome, all surfaces use the app's token system.
- **vs modal error dialogs** — errors live inline in the footer, the modal stays open, the form data persists. Users NEVER lose their edits on error.
- **vs the `NameWorkflowModal` save path** — NameWorkflowModal's save path navigates away entirely. This save path keeps the user in place and tells them "we heard you" via the sidebar flash.
- **Cross-story rule**: the sidebar row flash established here (400ms green fade) is the canonical "row acknowledged update" pattern. skills-watch-02 should reuse it for reactive reloads from filesystem events.
