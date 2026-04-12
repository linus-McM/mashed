# Sprint Backlog: Multi-Editor & Screenshot Inject

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | [EditorRouter -- Extension-Based Editor Switching](meditor-01-editor-router.md) | P0 | frontend | S | none |
| 2 | [ReadFileBase64 Wails Binding for Image Data](meditor-02-readfilebase64-binding.md) | P0 | backend | S | none |
| 3 | [ImageViewer Component with Panzoom](meditor-03-image-viewer.md) | P0 | frontend | M | Story 1, Story 2 |
| 4 | [MarkdownEditor with Milkdown Crepe WYSIWYG](meditor-04-markdown-editor.md) | P1 | frontend | L | Story 1 |
| 5 | [Screenshot-to-Claude-Code — Full Stack](meditor-05-screenshot-fullstack.md) | P1 | full-stack | M | none (but land after PTY sprint) |

**Total Stories:** 5 (merged former Stories 5+6 into combined Story 5)
**Ready for Sprint:** Stories 1, 2, 3, 4, 5 (all status: ready)
**Recommended Sprint Order:** 1, 2, 3, 4, 5
**Superseded files:** `meditor-05-screenshot-backend.md`, `meditor-06-screenshot-frontend.md` (replaced by `meditor-05-screenshot-fullstack.md`)

**Execution order:** This sprint runs AFTER the PTY sprint completes. PTY stories land first to establish the new terminal architecture, then multi-editor stories build on top.

## Dependency Graph

```
Story 1 (EditorRouter)       Story 2 (ReadFileBase64)
  |                            |
  |--- Story 3 (ImageViewer) --+
  |
  |--- Story 4 (MarkdownEditor)

Story 5 (Screenshot Full Stack) — independent, but lands after PTY sprint
```

## Parallelization Notes

- **Stories 1, 2 can run in parallel** — zero dependencies, different domains
- Story 3 depends on both Story 1 and Story 2 — start after both complete
- Story 4 depends only on Story 1 — can start alongside Story 3
- Story 5 is independent but should land last (modifies Terminal.svelte and AgentDetail.svelte which PTY sprint also touches)
- **Maximum wall-clock path:** 2 serial steps: (1 + 2 parallel) -> (3 + 4 parallel) -> 5
- **With 2 engineers:** Engineer A: Story 1 -> Story 3 -> Story 5. Engineer B: Story 2 -> Story 4.
- **With 3 engineers:** Engineer A: Story 1 -> Story 3. Engineer B: Story 2 -> Story 4. Engineer C: Story 5 (after PTY lands).

## Agent Assignments

| Story | Primary Agent | Supporting Agent |
|-------|--------------|-----------------|
| 1 (EditorRouter) | frontend-design | go-svelte-test |
| 2 (ReadFileBase64) | golang-pro | go-svelte-test |
| 3 (ImageViewer) | frontend-design | go-svelte-test |
| 4 (MarkdownEditor) | typescript-pro | go-svelte-test |
| 5 (Screenshot Full Stack) | golang-pro + typescript-pro | go-svelte-test |

## Files Created/Modified

| File | Action | Stories |
|------|--------|---------|
| `frontend/src/components/EditorRouter.svelte` | New | 1, 3, 4 |
| `frontend/src/views/AgentDetail.svelte` | Modify (swap MonacoEditor -> EditorRouter) | 1, 5 |
| `app_git.go` | Add `ReadFileBase64` method | 2 |
| `frontend/src/components/ImageViewer.svelte` | New | 3 |
| `frontend/src/components/MarkdownEditor.svelte` | New | 4 |
| `frontend/src/styles/crepe-mashed.css` | New | 4 |
| `app.go` | Add context fields, SetActiveContext, modify TakeScreenshot | 5 |
| `main.go` | Update menu callback for screenshot | 5 |
| `screenshot_test.go` | Update tests for new signature | 5 |
| `frontend/src/components/Terminal.svelte` | Add `screenshot:inject` listener | 5 |
| `frontend/src/App.svelte` | Clear context on `goBack()` | 5 |
| `frontend/package.json` | Add `panzoom`, `@milkdown/crepe` deps | 3, 4 |

## Open Questions

1. **Image viewer features:** Start with panzoom. Upgrade to PhotoSwipe if users request rotate/flip.
2. **SVG handling:** Phase 3 scope. SVGs render as images in ImageViewer for now.
3. **Multiple terminals open:** Active tab only via SetActiveContext.

## Phase 3 Scope (Not in this sprint)

- SVG dual-mode (image + source toggle)
- Keyboard shortcut parity (Escape to close editor)
- Crepe plugin additions (slash commands, tables, code block highlighting)
- Large file fallback (Crepe -> Monaco for markdown files > 100 KB)
