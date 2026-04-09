# Story 7: Build System, Signing, and Cleanup

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** Story 6 (ptyhelper-06-main-integration)
**Status:** ready

## Description

Update the build system to compile both the main Wails binary and the `mashed-pty-helper` binary, bundle the helper into the `.app` package, sign both binaries with entitlements, and clean up all experimental code from the PTY investigation. This story makes the helper architecture production-ready: `just build` produces a working signed app bundle, `just dev` builds the helper and launches Wails with the helper available, and all dead code from the investigation sprint is removed.

## Developer Notes

### Architecture
- **Modified file:** `justfile` -- add `build-helper`, update `build`, `dev` recipes
- **Modified file:** `build/darwin/build-and-sign.sh` -- sign helper binary inside bundle
- **Deleted file:** `internal/terminal/spawner.go` (if it exists -- experimental code from investigation)
- **Modified files:** Various cleanup (remove debug logging, experimental code)

### Justfile Changes

New recipe: `build-helper`
```just
# Build only the PTY helper binary (for development)
build-helper:
    go build -o build/bin/mashed-pty-helper ./cmd/pty-helper
```

Updated `build` recipe:
```just
build:
    cd frontend && npm install && cd ..
    go build -o build/bin/mashed-pty-helper ./cmd/pty-helper
    PATH="$HOME/go/bin:$PATH" wails build
    cp -r fonts build/bin/mashed.app/Contents/Resources/fonts
    cp build/bin/mashed-pty-helper build/bin/mashed.app/Contents/MacOS/mashed-pty-helper
```

Updated `dev` recipe:
```just
dev: build-helper
    PATH="$HOME/go/bin:$PATH" wails dev
```

### build-and-sign.sh Changes

The script needs to sign the helper binary inside the app bundle. Sign order: inner binaries first, then the outer `.app` bundle.

```bash
# After go build completes and Wails packages the .app:
HELPER_BINARY="$SCRIPT_DIR/../bin/mashed.app/Contents/MacOS/mashed-pty-helper"
if [ -f "$HELPER_BINARY" ]; then
    codesign --force --sign - --entitlements "$ENTITLEMENTS" "$HELPER_BINARY" 2>/dev/null && \
        echo "[build-and-sign] Signed pty-helper with entitlements" || true
fi
# Then sign the main binary (existing logic)
codesign --force --sign - --entitlements "$ENTITLEMENTS" "$APP_BINARY" 2>/dev/null && \
    echo "[build-and-sign] Re-signed mashed binary with PTY entitlements" || true
```

### Cleanup Targets

From the plan's Phase 6 cleanup table:

| File | Action | Details |
|------|--------|---------|
| `internal/terminal/spawner.go` | Delete | If exists -- experimental channel spawner |
| `main.go` | Clean | Remove any TestFork debug code, old spawner startup |
| `app.go` | Clean | Remove old spawner param if present |
| `app_spawn.go` | Clean | Remove debug logging related to investigation |
| `frontend/src/views/NotificationFeed.svelte` | Clean | Remove debug logging from investigation |
| `build/darwin/build-and-sign.sh` | Rework | Per above signing changes |
| `frontend/wailsjs/runtime/runtime.js` | Keep | Valid fix for browser dev mode |
| `build/darwin/entitlements.plist` | Keep | Still needed for the helper |

### Technical Considerations
- The helper binary must be copied INTO the `.app` bundle at `Contents/MacOS/` for production
- In `wails dev` mode, the helper binary sits at `build/bin/mashed-pty-helper` (not in a bundle)
- `build-and-sign.sh` runs as a post-build hook from Wails -- it re-signs after Wails does its own signing
- The helper needs the SAME entitlements as the main binary (PTY/fork/exec entitlements)
- Sign order matters: sign inner binaries first, then the outer .app bundle
- `just dev` should auto-build the helper so developers don't have to remember a separate step

### Risks & Edge Cases
- Wails may re-sign the .app bundle after our script runs -- the background sleep+resign strategy in `build-and-sign.sh` handles this
- If the helper binary is missing when `just build` runs (first time), `go build ./cmd/pty-helper` handles it
- `just dev` depends on `build-helper` -- if helper build fails, dev mode won't have terminals (graceful degradation)
- macOS Gatekeeper may quarantine the helper binary in dev mode -- signing with Apple Dev cert prevents this

### Reference Files
- `justfile` -- current build recipes (full file)
- `build/darwin/build-and-sign.sh` -- current signing script
- `build/darwin/entitlements.plist` -- entitlements to apply to both binaries
- `main.go` -- check for any debug/experimental code to remove
- `app_spawn.go` -- check for debug logging to remove

## Acceptance Criteria

AC-1: `just build-helper` compiles helper binary
- Given the source code exists at `cmd/pty-helper/main.go`
- When `just build-helper` is run
- Then `build/bin/mashed-pty-helper` is produced
- And it is executable

AC-2: `just build` produces a complete signed app bundle with helper
- Given all source code is present
- When `just build` is run
- Then `build/bin/mashed.app/Contents/MacOS/mashed-pty-helper` exists inside the bundle
- And both the main binary and helper binary are code-signed with entitlements
- And the app launches and can spawn terminals

AC-3: `just dev` builds helper and starts Wails dev mode
- Given the development environment is set up
- When `just dev` is run
- Then the helper binary is built first
- And Wails dev mode starts
- And the running app can find and launch the helper binary

AC-4: Experimental code is removed
- Given the codebase has leftover investigation code
- When cleanup is complete
- Then `internal/terminal/spawner.go` does not exist (if it existed)
- And no debug/investigation logging remains in `main.go`, `app_spawn.go`, or `NotificationFeed.svelte`
- And `go build ./...` still passes
- And all existing tests pass

AC-5: Both binaries are signed with entitlements
- Given `build/darwin/build-and-sign.sh` has been updated
- When a production build is run
- Then `codesign -d --entitlements -` on both binaries shows the PTY entitlements
- And the helper is signed BEFORE the outer .app bundle

## BDD Test Scenarios

### Scenario 1: Build system

```gherkin
Feature: Build system produces complete helper binary

  Scenario: Build helper standalone
    Given the Go source at cmd/pty-helper/ is valid
    When "just build-helper" is executed
    Then build/bin/mashed-pty-helper exists and is executable
    And "file build/bin/mashed-pty-helper" shows it is a Mach-O arm64 binary

  Scenario: Full build includes helper in bundle
    Given a clean build directory
    When "just build" is executed
    Then build/bin/mashed.app/Contents/MacOS/mashed-pty-helper exists
    And build/bin/mashed.app/Contents/MacOS/mashed exists
    And both binaries are Mach-O arm64
```

### Scenario 2: Signing

```gherkin
Feature: Code signing with entitlements

  Scenario: Helper binary is signed with entitlements
    Given a production build has been completed
    When codesign -dvvv is run on the helper binary inside the bundle
    Then it shows a valid signature
    And the entitlements include com.apple.security.cs.allow-jit

  Scenario: Main binary is signed after helper
    Given the build-and-sign.sh script runs
    Then the helper binary is signed first
    And the main .app bundle is signed second
```

### Scenario 3: Cleanup

```gherkin
Feature: Experimental code cleanup

  Scenario: No spawner.go exists
    Given the cleanup has been performed
    When "ls internal/terminal/spawner.go" is run
    Then the file does not exist

  Scenario: No debug investigation code in app files
    Given the cleanup has been performed
    When main.go, app_spawn.go, and NotificationFeed.svelte are searched for investigation markers
    Then no debug-only code related to the PTY investigation remains
    And "go build ./..." succeeds
    And "go test ./..." passes
```

## Tasks / Subtasks

- [ ] Task 1: Add justfile recipes (AC: AC-1, AC-3)
  - [ ] Add `build-helper` recipe to build `cmd/pty-helper` to `build/bin/`
  - [ ] Update `build` recipe to build helper, copy into `.app` bundle
  - [ ] Update `dev` recipe to depend on `build-helper`

- [ ] Task 2: Update build-and-sign.sh (AC: AC-2, AC-5)
  - [ ] Add signing step for the helper binary inside the bundle
  - [ ] Ensure helper is signed BEFORE the main .app bundle
  - [ ] Use same entitlements.plist for both binaries

- [ ] Task 3: Clean up experimental code (AC: AC-4)
  - [ ] Delete `internal/terminal/spawner.go` if it exists
  - [ ] Remove investigation debug code from `main.go`
  - [ ] Remove investigation debug logging from `app_spawn.go`
  - [ ] Remove investigation debug logging from `NotificationFeed.svelte`
  - [ ] Verify `go build ./...` and `go test ./...` still pass

- [ ] Task 4: Verify end-to-end (AC: AC-2, AC-3)
  - [ ] Run `just build` and verify helper is in bundle
  - [ ] Run `just dev` and verify terminal button works
  - [ ] Run the built `.app` and verify terminal button works
  - [ ] Verify code signing on both binaries

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `just build-helper` produces executable binary
- [ ] `just build` produces signed bundle with helper included
- [ ] `just dev` starts successfully with helper available
- [ ] All experimental/investigation code is removed
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Manual verification in both `wails dev` and `wails build` modes
