# Sprint Backlog: Persistent Terminal Sessions

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | Terminal Session Registry -- Domain Type & Backend API | P0 | backend | M | none |
| 2 | Startup Session Recovery from tmux | P0 | backend | S | Story 1 |
| 3 | Register Sessions on Spawn & Emit Wails Events | P0 | backend | S | Story 1 |
| 4 | Svelte Sessions Store & Event Subscription | P1 | frontend | S | Story 1, Story 3 |
| 5 | Session Tab Bar in AgentDetail View | P1 | frontend | L | Story 4 |

**Total Stories:** 5
**Ready for Sprint:** Stories 1, 2, 3, 4, 5 (all status: ready)
**Recommended Sprint Order:** 1, 2, 3, 4, 5

## Dependency Graph

```
Story 1 (Registry + Type + API)
  |--- Story 2 (Startup Recovery)
  |--- Story 3 (Spawn Registration + Events)
          |--- Story 4 (Svelte Store + Event Wiring)
                  |--- Story 5 (Tab Bar UI)
```

## Parallelization Notes

- Stories 2 and 3 can run in parallel after Story 1 completes (both depend only on Story 1, no overlap).
- Stories 4 and 5 are sequential (5 depends on 4's store).
- Maximum wall-clock path: Story 1 -> Story 3 -> Story 4 -> Story 5 (4 serial steps).
- With 2 engineers: assign Story 2 to engineer B while engineer A does Story 3, then both converge on frontend.

## Files Created/Modified

| File | Action | Stories |
|------|--------|---------|
| `internal/domain/types.go` | Add `TerminalSession` struct | 1 |
| `app.go` | Add `terminalSessions` field, init in `NewApp()`, call `recoverSessions()` in `startup()` | 1, 2 |
| `app_terminal_registry.go` | New file: `registerSession`, `ListRepoSessions`, `KillTerminalSession`, `recoverSessions` | 1, 2 |
| `app_terminal_registry_test.go` | New file: tests for registry and recovery | 1, 2 |
| `app_tmux.go` | Modify `spawnTmuxSession` signature, add registration + events. Modify `KillAgent` for deregistration. | 3 |
| `frontend/src/lib/stores/sessions.js` | New file: `repoSessions` store, `refreshSessions`, `addSession`, `removeSession` | 4 |
| `frontend/src/App.svelte` | Add `terminal:session:added` and `terminal:session:removed` event handlers | 4 |
| `frontend/src/views/AgentDetail.svelte` | Add tab bar, session switching, kill/spawn handlers | 5 |
| `frontend/src/views/NotificationFeed.svelte` | Add `addSession` calls after spawn | 5 |

Story files written to: `docs/stories/sessions-01-terminal-session-registry.md` through `docs/stories/sessions-05-session-tab-bar.md`
