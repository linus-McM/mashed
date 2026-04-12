# skills-watch-01: fsnotify backend watcher for skills/commands directories

**Status:** ready
**Domain:** backend
**Size:** M
**Depends on:** none (Phase 1 is shipped)
**Phase:** 4b

## Description

Subscribe to filesystem changes under the four mashed-asset source directories:

- `{repoPath}/.claude/skills/`
- `{repoPath}/.claude/commands/`
- `~/.claude/skills/`
- `~/.claude/commands/`

On any create/write/remove/rename event, debounce 500 ms, then emit a Wails event `bmad:assets:changed` so the frontend can re-fetch via `ListAllMashedAssets`. The backend never sends asset data on the event — it just signals "something changed" and lets the frontend pull the authoritative list.

This story is backend-only; skills-watch-02 consumes the event.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/asset_watcher.go` — NEW. Owns the fsnotify watcher lifecycle.
    - `type AssetWatcher struct { ... }`
    - `NewAssetWatcher(ctx context.Context, repoPath string, emit func(eventName string, data any)) (*AssetWatcher, error)`
    - `(*AssetWatcher).Start() error` — begins watching; runs until ctx cancel
    - `(*AssetWatcher).Stop()` — idempotent cleanup
  - `app.go` — instantiate the watcher in the same place the repo-scan watcher is set up (cerebrum Key Learning: "Use the pattern already used for repo scanning"). Pass `runtime.EventsEmit` as the emit function.
  - `internal/bmad/asset_watcher_test.go` — NEW. Tests using `t.TempDir()` and real fsnotify events.
- **fsnotify dependency:** confirm it's already in go.mod via the repo-scan watcher. If not, add `github.com/fsnotify/fsnotify`.
- **Debounce pattern:**
  ```go
  var timer *time.Timer
  for {
      select {
      case <-ctx.Done():
          return
      case ev := <-watcher.Events:
          if !relevantExtension(ev.Name) { continue } // .md only
          if timer != nil { timer.Stop() }
          timer = time.AfterFunc(500*time.Millisecond, func() {
              emit("bmad:assets:changed", nil)
          })
      case err := <-watcher.Errors:
          log.Printf("bmad: asset watcher error: %v", err)
      }
  }
  ```
- **Recursive vs non-recursive:** fsnotify's `Watcher.Add` is non-recursive. Skills/commands live one level deep under `skills/`, `commands/` (each skill is a directory containing `SKILL.md`). Options:
  1. Walk at startup, add each subdirectory, handle `Create` events on new subdirectories by adding them dynamically.
  2. Just watch the parent directory and rescan on any event.
  - **Choose option 1** for correctness — option 2 misses events inside newly-created subdirectories until the next rescan.
- **Risks / gotchas:**
  - **Nonexistent directories are not errors** (Phase 1 invariant from plan §"Parser rules"). The watcher must skip missing directories silently and continue — do NOT fail `Start()` if one of the four roots doesn't exist.
  - **New subdirectory handling:** when a user creates a new skill directory under `~/.claude/skills/`, the watcher must add it dynamically so subsequent writes inside it fire events. Test this.
  - **Debounce coalesces rapid burst events** from editors (vim writes multiple times per save). 500 ms is the plan default.
  - **Extension filter:** only `.md` files (and optionally any file if the directory itself changes). Ignore `.swp`, `.DS_Store`, etc.
  - **Cleanup on shutdown:** the watcher goroutine must exit on ctx cancel. No leaked goroutines.
- **Prerequisites already in place:**
  - `ListAllMashedAssets` Wails binding from Phase 1 (frontend re-fetches via this).
  - Repo-scan watcher pattern in `app.go` (use as template).

## Acceptance Criteria

**AC-1: Single file write emits one debounced event**
- Given the watcher is running on a temp directory containing `skills/my-skill/SKILL.md`
- When the file is modified once
- Then exactly one `bmad:assets:changed` event is emitted after approximately 500 ms
- And no second event is emitted unless another write occurs

**AC-2: Burst of writes coalesces to one event**
- Given the watcher is running
- When five writes occur within 200 ms of each other
- Then exactly one event is emitted approximately 500 ms after the last write

**AC-3: New subdirectory is watched dynamically**
- Given the watcher is running on `~/.claude/skills/` which initially has one subdirectory
- When a new subdirectory `new-skill/` is created
- And a file `new-skill/SKILL.md` is written inside it 1 second later
- Then a `bmad:assets:changed` event is emitted after the file write

**AC-4: Nonexistent root directory is silently skipped**
- Given the watcher is configured with four roots and two of them do not exist on disk
- When the watcher starts
- Then `Start()` returns nil (no error)
- And the two existing roots are watched normally

**AC-5: Watcher cleans up on context cancel**
- Given the watcher is running
- When the context is cancelled
- Then `Stop()` completes within 1 second
- And no goroutines are leaked (`goleak.VerifyNone(t)` passes)

**AC-6: Non-markdown files are ignored**
- Given the watcher is running
- When `.DS_Store` or `foo.swp` is created inside a watched directory
- Then zero events are emitted

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Asset filesystem watcher

  Scenario: Single write debounced to one event
    Given the watcher is watching a skills directory
    When a SKILL.md file is written once
    Then bmad:assets:changed fires exactly once after ~500ms

  Scenario: Write burst coalesces
    Given the watcher is watching a skills directory
    When a SKILL.md is written five times within 200ms
    Then bmad:assets:changed fires exactly once ~500ms after the last write

  Scenario: New subdirectory joins the watch set
    Given the watcher is running on ~/.claude/skills
    When a new subdirectory "new-skill" is created
    And SKILL.md is written inside it after 1 second
    Then bmad:assets:changed fires after the file write

  Scenario: Nonexistent roots are silently skipped
    Given the watcher is configured with two missing roots and two existing roots
    When Start is invoked
    Then it returns nil
    And the existing roots are watched

  Scenario: Graceful shutdown on context cancel
    Given the watcher is running
    When ctx is cancelled
    Then Stop completes within 1 second
    And no goroutines leak

  Scenario: Non-markdown files ignored
    Given the watcher is running
    When .DS_Store is created inside a watched directory
    Then no events fire
```

## Tasks / Subtasks

- [ ] Task 1 — `AssetWatcher` type + lifecycle (AC-4, AC-5)
  - [ ] Struct with fsnotify watcher, emit fn, ctx, internal goroutine
  - [ ] `Start` walks each root, adds subdirectories, skips missing roots
  - [ ] `Stop` cancels ctx and closes watcher
- [ ] Task 2 — Event loop with debounce (AC-1, AC-2, AC-6)
  - [ ] Extension filter (`.md` only; plus directory create events)
  - [ ] Debounce via `time.AfterFunc`
  - [ ] Emit `bmad:assets:changed`
- [ ] Task 3 — Dynamic subdirectory handling (AC-3)
  - [ ] On `Create` event for a directory, `watcher.Add` it
  - [ ] Handle the race where a file is written immediately inside the new dir
- [ ] Task 4 — App wiring (AC-1 through AC-6)
  - [ ] Instantiate in `app.go` alongside the repo-scan watcher
  - [ ] Pass `runtime.EventsEmit` as the emit fn
  - [ ] Cancel on app shutdown
- [ ] Task 5 — Tests (AC-1 through AC-6)
  - [ ] Each AC as a focused test using `t.TempDir()` and real fsnotify
  - [ ] Goroutine leak test with `uber-go/goleak`
  - [ ] Flake-proof timing: wait up to 1s for the debounced event

## Definition of Done

- [ ] All ACs verified by an automated test
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass
