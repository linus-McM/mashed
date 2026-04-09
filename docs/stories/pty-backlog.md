# PTY Sprint Backlog

## Sprint Goal

Replace tmux with direct PTY management in the terminal pipeline to fix broken scrolling, text selection, and clipboard in the Wails desktop app's terminal component.

## Stories

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | [ManagedSession with Scrollback Ring Buffer](pty-01-managed-session.md) | P0 | backend | M | none |
| 2 | [SessionManager Lifecycle Manager](pty-02-session-manager.md) | P0 | backend | M | Story 1 |
| 3 | [Bridge Refactor -- Route WebSockets to Managed Sessions](pty-03-bridge-refactor.md) | P0 | backend | L | Story 1, 2 |
| 4 | [App Spawn Integration & Frontend Terminal Cleanup](pty-04-app-integration-and-frontend-cleanup.md) | P0 | full-stack | L | Story 2, 3 |
| 5 | [Scan Loop Dual PID Lookup](pty-05-scan-dual-lookup.md) | P1 | backend | S | Story 2, 4 |

**Total Stories:** 5 (merged former Stories 4+6 into combined Story 4)
**Ready for Sprint:** Stories 1, 2, 3, 4, 5 (all status: ready)
**Recommended Sprint Order:** 1, 2, 3, 4, 5
**Superseded files:** `pty-04-app-spawn-integration.md`, `pty-06-frontend-cleanup.md` (replaced by `pty-04-app-integration-and-frontend-cleanup.md`)

## Dependency Graph

```
Story 1 (ManagedSession)
  |
  v
Story 2 (SessionManager)
  |
  v
Story 3 (Bridge)
  |
  v
Story 4 (App Integration + Frontend Cleanup)
  |
  v
Story 5 (Scan Dual Lookup)
```

## Scope Notes

- `internal/terminal/panes.go` is NOT modified -- still needed for BMAD executor + scan fallback
- `internal/bmad/executor.go` is NOT modified -- keeps its own tmux usage
- `internal/domain/types.go` `TerminalSession` struct is unchanged (PaneTarget field repurposed to hold session name)
- Story 4 now includes all frontend tmux cleanup (Terminal.svelte, sessions.js) alongside the backend App integration, eliminating back-and-forth on session naming

## Files Modified Per Story

| Story | New Files | Modified Files |
|-------|-----------|----------------|
| 1 | `internal/terminal/session.go`, `internal/terminal/session_test.go` | -- |
| 2 | `internal/terminal/manager.go`, `internal/terminal/manager_test.go` | -- |
| 3 | -- | `internal/terminal/bridge.go` |
| 4 | `app_spawn.go` (renamed from `app_tmux.go`) | `app.go`, `app_terminal_registry.go`, `app_terminal_registry_test.go`, `frontend/src/components/Terminal.svelte`, `frontend/src/lib/stores/sessions.js` |
| 5 | -- | `app_scan.go` |
