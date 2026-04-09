# PTY Helper Sprint Backlog

**Sprint Goal:** Replace direct `pty.StartWithSize` in the Wails process with a separate `mashed-pty-helper` binary that owns all PTY lifecycle, communicating via Unix socket + SCM_RIGHTS fd passing. This eliminates the macOS "operation not permitted" error caused by Wails/Cocoa callback context contaminating fork/exec.

**Branch:** feature/xyflow_design

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On | File |
|---|-------|----------|--------|------|------------|------|
| 1 | Protocol Types | P0 | backend | S | none | `ptyhelper-01-protocol-types.md` |
| 2 | Helper Binary (Server) | P0 | backend | L | Story 1 | `ptyhelper-02-helper-binary.md` |
| 3 | SCM_RIGHTS End-to-End Proof | P0 | backend | S | Story 1, 2 | `ptyhelper-03-scmrights-proof.md` |
| 4 | Helper Client | P0 | backend | M | Story 1 | `ptyhelper-04-client.md` |
| 5 | SessionManager Rework | P0 | backend | M | Story 4 | `ptyhelper-05-session-manager-rework.md` |
| 6 | main.go Integration | P0 | fullstack | M | Story 4, 5 | `ptyhelper-06-main-integration.md` |
| 7 | Build System & Cleanup | P1 | fullstack | M | Story 6 | `ptyhelper-07-build-system-and-cleanup.md` |

**Total Stories:** 7
**Sprint Complete:** Stories 1, 2, 3, 4, 5, 6, 7 (all status: done)

## Recommended Sprint Order

**Phase A (Foundation, parallelizable):**
- Story 1: Protocol Types (no deps, blocks everything)

**Phase B (Core implementation, partially parallelizable):**
- Story 2: Helper Binary -- depends on Story 1
- Story 4: Helper Client -- depends on Story 1 (can run in parallel with Story 2)

**Phase C (Validation):**
- Story 3: SCM_RIGHTS Proof -- depends on Story 1 + 2 (validates architecture before integration)

**Phase D (Integration):**
- Story 5: SessionManager Rework -- depends on Story 4
- Story 6: main.go Integration -- depends on Story 4 + 5

**Phase E (Polish):**
- Story 7: Build System & Cleanup -- depends on Story 6

## Dependency Graph

```
Story 1 (Protocol)
  |         \
  v          v
Story 2    Story 4
(Server)   (Client)
  |          |
  v          |
Story 3     |
(Proof)     |
            v
         Story 5
    (SessionManager)
            |
            v
         Story 6
    (main.go Integration)
            |
            v
         Story 7
    (Build & Cleanup)
```

## Key Files Created/Modified

### New Files
- `cmd/pty-helper/main.go` -- helper binary entry point
- `internal/terminal/helper/protocol.go` -- shared protocol types
- `internal/terminal/helper/server.go` -- helper server logic
- `internal/terminal/helper/client.go` -- client for main process
- `internal/terminal/helper/protocol_test.go` -- protocol tests
- `internal/terminal/helper/server_test.go` -- server tests
- `internal/terminal/helper/client_test.go` -- client tests
- `internal/terminal/helper/integration_test.go` -- E2E proof

### Modified Files
- `main.go` -- helper launch, wait, dial, shutdown
- `app.go` -- `NewApp` accepts `*helper.Client`
- `internal/terminal/manager.go` -- use helper client instead of direct pty.Start
- `internal/terminal/session.go` -- support remote sessions (nil cmd)
- `justfile` -- build-helper, updated build/dev recipes
- `build/darwin/build-and-sign.sh` -- sign both binaries
- `frontend/src/views/NotificationFeed.svelte` -- visible error on spawn failure
- `frontend/src/views/AgentDetail.svelte` -- visible error on spawn failure

### Deleted Files
- `internal/terminal/spawner.go` (if exists -- experimental code)

## Risk Summary

| Risk | Stories Affected | Mitigation |
|------|-----------------|------------|
| SCM_RIGHTS macOS arm64 quirks | 1, 2, 3 | Story 3 is early validation |
| fd/response ordering mismatch | 4 | Mutex serialization on Client.Spawn |
| Helper binary not found in dev | 6, 7 | resolveHelperPath with fallback |
| Orphan helper after parent crash | 2 | kqueue EVFILT_PROC + polling fallback |
| Stale socket from crash | 2 | os.Remove before Listen |
