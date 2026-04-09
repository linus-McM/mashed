# Menu Bar Backlog: Native macOS Menu Bar for mashed

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | Native macOS Menu Bar Construction | P0 | backend | M | none |
| 2 | Frontend Menu Event Handlers & Navigation | P0 | frontend | S | Story 1 |

**Total Stories:** 2
**Ready for Sprint:** Stories 1, 2 (all status: ready)
**Recommended Sprint Order:** 1, 2

## Dependency Graph

```
Story 1 (Menu Bar + TakeScreenshot backend)
  |
  v
Story 2 (Frontend event handlers + AboutModal + toast)
```

## Parallelization Notes

- **Story 1** must complete first -- it defines the menu structure and emits the events that Story 2 consumes.
- **Story 2** is purely frontend and can begin as soon as Story 1's event names are finalized (the event contract is defined in Story 1's notes: `menu:navigate`, `menu:about`, `screenshot:taken`, `menu:open-workspace`).

## Sprint Goal

Replace the minimal Edit-only native menu with a full 5-menu macOS menu bar (mashed, File, Edit, View, Help) with keyboard shortcuts, screenshot capture, and frontend navigation wired through Wails events.

## Story Files

- `docs/stories/menu-01-native-menu-bar.md`
- `docs/stories/menu-02-frontend-event-handlers.md`
