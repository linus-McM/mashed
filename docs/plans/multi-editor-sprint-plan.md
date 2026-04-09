# Sprint Plan: PTY Migration + Multi-Editor + Screenshot Inject

> **Source:** `docs/feasibility-multi-editor.md`, PTY stories
> **Date:** 2026-04-09
> **Execution:** Sequential — PTY sprint first, then Multi-Editor sprint
> **Total Stories:** 10 (5 PTY + 5 Multi-Editor)
> **Estimated Effort:** 8-12 agent sessions

---

## Execution Strategy

**Sequential sprints to avoid file collisions.** Both sprints modify `app.go`, `Terminal.svelte`, and `AgentDetail.svelte`. Landing PTY first establishes the new terminal architecture, then multi-editor builds on top of clean files.

---

## Sprint 1: PTY Migration (5 stories)

Replace tmux with direct PTY management. Fix broken scrolling, text selection, and clipboard.

| # | Story | Priority | Domain | Size | Depends On | Agent |
|---|-------|----------|--------|------|------------|-------|
| 1 | [ManagedSession with Scrollback Ring Buffer](../stories/pty-01-managed-session.md) | P0 | backend | M | none | golang-pro |
| 2 | [SessionManager Lifecycle Manager](../stories/pty-02-session-manager.md) | P0 | backend | M | 1 | golang-pro |
| 3 | [Bridge Refactor — Route WebSockets to Managed Sessions](../stories/pty-03-bridge-refactor.md) | P0 | backend | L | 1, 2 | golang-pro |
| 4 | [App Integration & Frontend Cleanup](../stories/pty-04-app-integration-and-frontend-cleanup.md) | P0 | full-stack | L | 2, 3 | golang-pro + frontend-design |
| 5 | [Scan Loop Dual PID Lookup](../stories/pty-05-scan-dual-lookup.md) | P1 | backend | S | 2, 4 | golang-pro |

### PTY Dependency Graph

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
Story 4 (App Integration + Frontend Cleanup)  ← merged pty-04 + pty-06
  |
  v
Story 5 (Scan Dual Lookup)
```

**Note:** Story 4 is a merged story (former pty-04 + pty-06). It handles both the backend App struct/spawn integration AND the frontend Terminal.svelte/sessions.js cleanup in one pass. This eliminates back-and-forth on session naming conventions.

---

## Sprint 2: Multi-Editor + Screenshot (5 stories)

Replace MonacoEditor with context-aware EditorRouter. Add screenshot-to-Claude-Code injection.

| # | Story | Priority | Domain | Size | Depends On | Agent |
|---|-------|----------|--------|------|------------|-------|
| 1 | [EditorRouter — Extension-Based Editor Switching](../stories/meditor-01-editor-router.md) | P0 | frontend | S | none | frontend-design |
| 2 | [ReadFileBase64 Wails Binding for Image Data](../stories/meditor-02-readfilebase64-binding.md) | P0 | backend | S | none | golang-pro |
| 3 | [ImageViewer Component with Panzoom](../stories/meditor-03-image-viewer.md) | P0 | frontend | M | 1, 2 | frontend-design |
| 4 | [MarkdownEditor with Milkdown Crepe WYSIWYG](../stories/meditor-04-markdown-editor.md) | P1 | frontend | L | 1 | typescript-pro |
| 5 | [Screenshot-to-Claude-Code — Full Stack](../stories/meditor-05-screenshot-fullstack.md) | P1 | full-stack | M | none (land last) | golang-pro + typescript-pro |

### Multi-Editor Dependency Graph

```
Story 1 (EditorRouter)       Story 2 (ReadFileBase64)
  |                            |
  |--- Story 3 (ImageViewer) --+
  |
  |--- Story 4 (MarkdownEditor)

Story 5 (Screenshot Full Stack) — independent, lands last
```

**Note:** Story 5 is a merged story (former meditor-05 + meditor-06). It handles both the Go backend (TakeScreenshot, SetActiveContext) and frontend (Terminal.svelte listener, AgentDetail.svelte context calls) in one pass.

---

## Recommended Execution Order

### With 2 Engineers

| Phase | Engineer A | Engineer B |
|-------|-----------|-----------|
| **PTY Sprint** | | |
| 1 | pty-01 (ManagedSession) | — |
| 2 | pty-02 (SessionManager) | — |
| 3 | pty-03 (Bridge) | — |
| 4 | pty-04 (App + Frontend Cleanup) | — |
| 5 | pty-05 (Scan Dual Lookup) | — |
| **Multi-Editor Sprint** | | |
| 6 | meditor-01 (EditorRouter) | meditor-02 (ReadFileBase64) |
| 7 | meditor-03 (ImageViewer) | meditor-05 (Screenshot Full Stack) |
| 8 | meditor-04 (MarkdownEditor) | — |

**Wall-clock:** 8 phases

### With 3 Engineers

| Phase | Engineer A | Engineer B | Engineer C |
|-------|-----------|-----------|-----------|
| **PTY Sprint** | | | |
| 1 | pty-01 (ManagedSession) | — | — |
| 2 | pty-02 (SessionManager) | — | — |
| 3 | pty-03 (Bridge) | — | — |
| 4 | pty-04 (App + Frontend) | — | — |
| 5 | pty-05 (Scan) | — | — |
| **Multi-Editor Sprint** | | | |
| 6 | meditor-01 (EditorRouter) | meditor-02 (ReadFileBase64) | meditor-05 (Screenshot) |
| 7 | meditor-03 (ImageViewer) | meditor-04 (MarkdownEditor) | — |

**Wall-clock:** 7 phases

---

## Key Integration Points

### AgentDetail.svelte (lines 547-555)
Touched by: pty-04 (session naming), meditor-01 (MonacoEditor -> EditorRouter swap), meditor-05 (SetActiveContext). Sequential execution eliminates conflicts.

### Terminal.svelte
Touched by: pty-04 (remove stripControlSequences, update connection msg), meditor-05 (add screenshot:inject listener). Sequential execution means meditor-05 adds to already-cleaned file.

### app.go (App struct)
Touched by: pty-04 (add manager field, update NewApp/shutdown), meditor-05 (add context fields, SetActiveContext, modify TakeScreenshot). Different struct sections — clean merge.

---

## Collision Avoidance Summary

| File | PTY Sprint | Multi-Editor Sprint | Conflict Risk |
|------|-----------|-------------------|---------------|
| `app.go` | pty-04: manager field, lifecycle | meditor-05: context fields, TakeScreenshot | **None** (different sections, sequential) |
| `Terminal.svelte` | pty-04: remove stripping, clean lifecycle | meditor-05: add event listener | **None** (additive after cleanup) |
| `AgentDetail.svelte` | pty-04: session naming update | meditor-01: swap editor, meditor-05: add context call | **None** (sequential) |
| `sessions.js` | pty-04: remove `:0.0` stripping | — | **None** |

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Crepe API differs from docs | Story 4 rework | Medium | Verify API before building full component |
| Frontmatter mangling by Milkdown | Data loss in .md files | Medium | Test with frontmatter-heavy files; fall back to Monaco |
| Large base64 data URIs in DOM | UI lag for big images | Low | 10 MB cap on ReadFileBase64 |
| PTY sprint delays block multi-editor | Schedule slip | Medium | Multi-editor stories 1-4 are PTY-independent; only story 5 needs PTY to land first |

---

## Quality Gates

Each story must pass before merge:
1. All acceptance criteria satisfied
2. All BDD scenarios have corresponding automated tests
3. 80%+ code coverage on new/modified files
4. `go build ./...` and `go vet ./...` pass (backend)
5. `go test ./... -race` passes (backend)
6. `/simplify` run on all modified code
7. Code review: no CRITICAL/HIGH issues

---

## Story Files

### PTY Sprint
- `docs/stories/pty-01-managed-session.md`
- `docs/stories/pty-02-session-manager.md`
- `docs/stories/pty-03-bridge-refactor.md`
- `docs/stories/pty-04-app-integration-and-frontend-cleanup.md` (merged)
- `docs/stories/pty-05-scan-dual-lookup.md`
- `docs/stories/pty-backlog.md`

### Multi-Editor Sprint
- `docs/stories/meditor-01-editor-router.md`
- `docs/stories/meditor-02-readfilebase64-binding.md`
- `docs/stories/meditor-03-image-viewer.md`
- `docs/stories/meditor-04-markdown-editor.md`
- `docs/stories/meditor-05-screenshot-fullstack.md` (merged)
- `docs/stories/meditor-backlog.md`

### Superseded (no longer active)
- ~~`docs/stories/pty-04-app-spawn-integration.md`~~ → replaced by `pty-04-app-integration-and-frontend-cleanup.md`
- ~~`docs/stories/pty-06-frontend-cleanup.md`~~ → merged into `pty-04-app-integration-and-frontend-cleanup.md`
- ~~`docs/stories/meditor-05-screenshot-backend.md`~~ → replaced by `meditor-05-screenshot-fullstack.md`
- ~~`docs/stories/meditor-06-screenshot-frontend.md`~~ → merged into `meditor-05-screenshot-fullstack.md`

---

## Deferred to Phase 3 Sprint

- SVG dual-mode (image-first with source toggle)
- Keyboard shortcut parity (Escape to close)
- Crepe plugin additions (slash commands, tables, Mermaid diagrams)
- Large markdown file fallback (> 100 KB -> Monaco)
- Image viewer enhancements (rotate, flip, metadata)
