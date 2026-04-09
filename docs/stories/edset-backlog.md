# Editor Settings Sprint Backlog

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | [edset-01] Backend Editor Settings Config | P0 | backend | S | none |
| 2 | [edset-02] Editor Settings Store, Monaco Integration, and Settings UI | P0 | fullstack | L | edset-01 |
| 3 | [edset-03] Terminal Cursor Sync from Editor Settings | P1 | frontend | S | edset-02 |
| 4 | [edset-04] Auto-Load Bundled Themes from ./themes/ | P1 | fullstack | M | none |

**Total Stories:** 4
**Ready for Sprint:** Stories 1, 2, 3, 4 (all status: ready)
**Recommended Sprint Order:** 1, 4, 2, 3

### Sprint Order Rationale

- **Story 1 first:** Foundation -- defines the Go struct, config persistence, and Wails-bound methods. Story 2 cannot start without it.
- **Story 4 in parallel with 1:** Has no dependencies on editor settings; can be developed concurrently by a second engineer. Backend refactor of theme_scanner.go + frontend themeInit.js are self-contained.
- **Story 2 after 1:** The core vertical slice -- creates the store, wires Monaco, builds Settings UI. Largest story (L) so it starts as soon as Story 1 completes.
- **Story 3 after 2:** Small follow-on that extends the editor settings store to Terminal.svelte. Quick to implement once the store exists.

### Parallelism Opportunities

| Timeslot | Engineer A | Engineer B |
|----------|-----------|-----------|
| Slot 1 | Story 1 (backend, S) | Story 4 (fullstack, M) |
| Slot 2 | Story 2 (fullstack, L) | -- |
| Slot 3 | Story 3 (frontend, S) | -- |

### File Change Summary

| File | Stories | Change Type |
|------|---------|-------------|
| `app.go` | 1 | Add EditorSettings struct, mashedConfig field, 3 methods |
| `theme_scanner.go` | 4 | Refactor scanVSIXDirectory, add bundledThemesDir, ListBundledThemes, ReadBundledThemeFile |
| `frontend/src/lib/stores/editorSettings.js` | 2 | **NEW** -- writable store + helpers |
| `frontend/src/lib/themeInit.js` | 4 | Add loadBundledThemes() |
| `frontend/src/components/MonacoEditor.svelte` | 2 | Use store in getEditorOptions() + reactive update |
| `frontend/src/components/Terminal.svelte` | 3 | Map cursor settings from store |
| `frontend/src/views/Settings.svelte` | 2 | Add Editor section with 13 controls |
| `frontend/src/App.svelte` | 2, 4 | Init editor settings + call loadBundledThemes |

Story files written to: `docs/stories/`
