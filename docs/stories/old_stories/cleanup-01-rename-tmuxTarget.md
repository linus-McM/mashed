# Story: Rename tmuxTarget to sessionTarget

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Rename the legacy `tmuxTarget` field to `sessionTarget` across the entire codebase. The field now holds a PTY session name (e.g., `"mashed-repo-1712694523"`), not a tmux pane target. The rename is cosmetic but prevents confusion as the codebase grows.

## Scope

### Go types (source of truth)
- `internal/domain/types.go` — `NotificationEvent.TmuxTarget`, `TerminalSession.TmuxTarget` (JSON tag + field)
- `internal/bmad/types.go` — `WorkflowNode.TmuxTarget`
- `internal/bmad/executor.go` — `NodeStatusEvent.TmuxTarget`, references in `executeProcessNode`, `emitEvent`

### Go logic
- `app_scan.go` — `resolveTmuxTarget()` function and all call sites
- `app_spawn.go` — `KillAgent` parameter and internal usage
- `internal/agent/engine.go` — agent event construction

### Go tests
- `app_terminal_registry_test.go` — `TestStory5_ResolveTmuxTarget`
- `internal/bmad/executor_test.go`, `types_nodetype_test.go`, `registry_test.go`

### Frontend
- `App.svelte` — `agent?.tmuxTarget` references
- `NotificationFeed.svelte` — `tmuxTarget: target` in spawn functions, `KillAgent` call
- `AgentDetail.svelte` — session matching and Terminal prop
- `WorkflowBuilder.svelte` — node data references
- `NodeConfigPanel.svelte` — reactive `tmuxTarget` binding

### Wails bindings
- After renaming Go JSON tags, regenerate `frontend/wailsjs/` bindings

## Acceptance Criteria

- AC-1: Zero occurrences of `tmuxTarget` or `TmuxTarget` in the codebase (case-insensitive grep)
- AC-2: All Go tests pass with `-race`
- AC-3: Frontend builds without errors
- AC-4: Terminal spawn, tab switching, and kill still work end-to-end

## Tasks

- [ ] Task 1: Rename Go struct fields and JSON tags
- [ ] Task 2: Rename Go functions (`resolveTmuxTarget` → `resolveSessionTarget`)
- [ ] Task 3: Update all Go test references
- [ ] Task 4: Regenerate Wails bindings
- [ ] Task 5: Rename all frontend references
- [ ] Task 6: Verify end-to-end

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `go test ./... -race` passes
- [ ] `wails dev` compiles without errors
- [ ] No `tmux` references remain except in git history and comments about the migration
