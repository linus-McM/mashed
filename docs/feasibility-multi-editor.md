# Feasibility Report: Multi-Editor File Viewer System

> **Date:** 2026-04-08
> **Scope:** Replace single Monaco editor with context-aware editor switching (Markdown / Code / Image)
> **Current branch:** `feature/xyflow_design`

---

## 1. Current Architecture

**File selection flow:**
```
FileTree.svelte → dispatch('select', {path}) → AgentDetail.svelte → selectedFile state → MonacoEditor.svelte
```

**Key integration point:** `AgentDetail.svelte:497-504`
```svelte
{#if selectedFile}
  <div class="editor-pane" style="flex: 0 0 {editorFraction * 100}%">
    <MonacoEditor
      filePath={selectedFile.path}
      repoPath={agent.repoPath}
      mode={selectedFile.isBinary ? 'source' : (fileTab === 'all' ? 'source' : 'diff')}
      editable={!selectedFile.isBinary}
    />
  </div>
{/if}
```

**Monaco setup:** Lazy-loaded ESM import (~670 lines), custom theme integration, diff mode, auto-save, hover-to-explain. Properly destroys editor + models on file switch via `destroyEditor()`.

**No existing markdown or image viewer components exist.**

---

## 2. Proposed Architecture

### 2.1 Editor Router Component

A new `EditorRouter.svelte` component replaces the direct `<MonacoEditor>` usage. It inspects the file extension and mounts the appropriate viewer:

```
selectedFile.path
    │
    ├── .md / .mdx         → MarkdownEditor.svelte (Milkdown Crepe — WYSIWYG)
    ├── .png/.jpg/.svg/...  → ImageViewer.svelte (PhotoSwipe or panzoom)
    └── everything else     → MonacoEditor.svelte (existing, unchanged)
```

```svelte
<!-- EditorRouter.svelte -->
<script>
  export let filePath;
  export let repoPath;
  export let mode;
  export let editable;

  $: ext = filePath?.split('.').pop()?.toLowerCase() ?? '';
  $: editorType = getEditorType(ext);

  function getEditorType(ext) {
    const markdown = ['md', 'mdx', 'markdown'];
    const image = ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico'];
    if (markdown.includes(ext)) return 'markdown';
    if (image.includes(ext)) return 'image';
    return 'code';
  }
</script>

{#if editorType === 'markdown'}
  <MarkdownEditor {filePath} {repoPath} {editable} />
{:else if editorType === 'image'}
  <ImageViewer {filePath} {repoPath} />
{:else}
  <MonacoEditor {filePath} {repoPath} {mode} {editable} />
{/if}
```

**Key principle:** Each viewer is a separate Svelte component. Svelte's `{#if}` block automatically destroys the previous component and mounts the new one when `editorType` changes, handling lifecycle cleanup naturally.

### 2.2 Change in AgentDetail.svelte

Minimal — replace:
```svelte
<MonacoEditor filePath={selectedFile.path} ... />
```
with:
```svelte
<EditorRouter filePath={selectedFile.path} ... />
```

---

## 3. Markdown Editor Options

### Recommendation: **Milkdown** (Primary) or **Tiptap via svelte-tiptap** (Alternative)

| Criterion | Milkdown | Tiptap + svelte-tiptap | CodeMirror 6 + markdown |
|---|---|---|---|
| Svelte support | Framework-agnostic headless design; community Svelte wrapper | Community `svelte-tiptap` package | Direct — CM6 is vanilla JS |
| Editing mode | WYSIWYG (ProseMirror) | WYSIWYG (ProseMirror) | Source-only (closest to Obsidian's actual editor) |
| Plugin ecosystem | Slash commands, tooltips, math, diagrams | 150+ extensions; markdown ext available | CM6 extensions (syntax, fold, autocomplete) |
| Bundle size (gzip) | ~40-50 kB core + plugins | ~30-45 kB core | ~25-30 kB |
| Maintenance | Active (MIT) | Active, well-funded (MIT core) | Active (MIT) |
| "Obsidian feel" | Rich WYSIWYG, less Obsidian-like | Rich WYSIWYG, less Obsidian-like | Most Obsidian-like (Obsidian uses CM6 internally) |

**Trade-off decision needed:**

- **If "Obsidian-like" means WYSIWYG rendering** (headers render large, links are clickable, images inline): Use **Milkdown** or **Tiptap**. These give a Notion/Obsidian Live Preview experience.

- **If "Obsidian-like" means source editing with live preview** (what Obsidian actually does internally): Use **CodeMirror 6** with `@codemirror/lang-markdown` + `@lezer/markdown`. This is literally the same engine Obsidian uses. Bundle is smallest. You could add a split-pane preview with `markdown-it` for rendered output.

- **If you want both modes** (source + WYSIWYG toggle): Use **Milkdown** which supports switching between markup and WYSIWYG modes via its `@milkdown/plugin-slash` and dual-mode plugins.

### My recommendation: **CodeMirror 6 for markdown**

Rationale:
1. Closest to actual Obsidian UX (same engine)
2. Smallest bundle (~25 kB vs Monaco's ~2 MB)
3. You already have Monaco's theme system — CM6 themes can mirror it
4. Future option: add live preview pane without replacing the editor
5. No framework-specific adapter needed (vanilla JS, wrap in Svelte action)

---

## 4. Image Viewer Options

### Recommendation: **panzoom** (minimal) or **PhotoSwipe v5** (full-featured)

| Criterion | panzoom | PhotoSwipe v5 | Viewer.js |
|---|---|---|---|
| Bundle (gzip) | ~3 kB | ~15 kB | ~20 kB |
| Features | Zoom + pan only | Zoom, pan, gestures, swipe | Zoom, pan, rotate, flip |
| Svelte integration | Trivial (Svelte `use:action`) | Easy wrapper | Wrappable |
| Dark mode | Inherits (no UI) | Themed | Themed |
| License | MIT | MIT | MIT |

**Recommendation: Start with panzoom** (~3 kB). It gives zoom/pan which covers 90% of the use case. Wrap it as a Svelte action:

```svelte
<script>
  import panzoom from 'panzoom';
  function zoomable(node) {
    const instance = panzoom(node, { maxZoom: 10, minZoom: 0.1 });
    return { destroy: () => instance.dispose() };
  }
</script>

<div class="image-viewer">
  <img use:zoomable src={imageUrl} alt={fileName} />
</div>
```

If users later need rotate/flip/metadata, upgrade to PhotoSwipe v5 or Viewer.js.

---

## 5. Technical Concerns & Mitigations

### 5.1 Monaco Memory Management

**Risk:** Monaco is heavy (~2 MB). If the user switches rapidly between .md → .ts → .md → .ts, you don't want Monaco re-initializing each time.

**Mitigation:** Keep the `monacoModule` import cached at module scope (it already is — line 410 of MonacoEditor.svelte does `monacoModule = await import(...)`). The module itself stays in memory after first load. Only the editor *instance* is created/destroyed per file. This is already the current behavior.

**Enhancement:** Consider lazy-loading Monaco only when a code file is first selected (not on app startup). Currently it loads on `onMount` of MonacoEditor. With EditorRouter, it won't mount until a code file is picked.

### 5.2 Markdown Editor Lazy Loading

**Approach:** Dynamic `import()` for the markdown editor module, identical to how Monaco is loaded:

```javascript
// Inside MarkdownEditor.svelte onMount
const { EditorView } = await import('@codemirror/view');
const { markdown } = await import('@codemirror/lang-markdown');
```

**Bundle impact:** CM6 markdown adds ~25-30 kB gzip. Since it's lazy-loaded, it doesn't affect initial app load time.

### 5.3 Shared Concerns Across All Editors

| Concern | Current (Monaco) | Must replicate in new editors |
|---|---|---|
| Auto-save (800ms debounce) | Yes | Yes — same `WriteFile` Wails binding |
| Theme integration | Custom `defineAllThemes` | CM6: create matching theme. Image: CSS vars only |
| Font settings | `$currentMonoFont`, `$currentFontSize` stores | CM6: `EditorView.theme`. Image: N/A |
| Resize handling | `ResizeObserver` on container | CM6: built-in with `EditorView.updateListener`. Image: CSS `object-fit` |
| Cmd+S save | `editor.addCommand` | CM6: `keymap.of([...])`. Image: N/A |
| Diff mode | `createDiffEditor` | CM6: possible with `@codemirror/merge`. Image: N/A |
| Editable toggle | `readOnly` option | CM6: `EditorView.editable`. Image: N/A |

### 5.4 SVG Files — Ambiguity

SVGs are both images and code. Options:
- **Default to image viewer** for visual preview, with a "View Source" toggle that switches to Monaco
- Add a toggle in the editor header (like the existing Source/Diff toggle)

### 5.5 Binary/Unknown Files

Current behavior: Monaco opens with `isBinary` flag → read-only. This should remain the fallback for unrecognized file types. EditorRouter's `else` branch handles this naturally.

---

## 6. Implementation Phases

### Phase 1: EditorRouter + Image Viewer (Low risk, high impact)
1. Create `EditorRouter.svelte` with extension-based routing
2. Create `ImageViewer.svelte` with panzoom (~3 kB)
3. Swap `<MonacoEditor>` for `<EditorRouter>` in AgentDetail.svelte
4. Add `ReadFileBase64` or equivalent Wails binding for image data (or use file:// protocol if Wails allows)
5. **Estimated effort:** 1-2 sessions

### Phase 2: Markdown Editor (Medium risk, medium effort)
1. Install `@codemirror/view`, `@codemirror/state`, `@codemirror/lang-markdown`, `@lezer/markdown`
2. Create `MarkdownEditor.svelte` with lazy-loaded CM6
3. Wire auto-save, theme mirroring, font stores
4. Add markdown-specific toolbar (bold, italic, heading, link — optional)
5. **Estimated effort:** 2-3 sessions

### Phase 3: Polish & Edge Cases
1. SVG dual-mode (image + source toggle)
2. Markdown diff view (if needed — `@codemirror/merge`)
3. Markdown live preview pane (split view with `markdown-it` renderer)
4. Keyboard shortcut parity (Cmd+S, Escape to close)
5. **Estimated effort:** 1-2 sessions

---

## 7. Bundle Size Impact

| Component | Size (gzip) | Loading |
|---|---|---|
| Monaco (current) | ~800 kB (tree-shaken ESM) | Lazy on first code file |
| CodeMirror 6 + markdown | ~25-30 kB | Lazy on first .md file |
| panzoom | ~3 kB | Lazy on first image |
| **Total new** | **~28-33 kB** | **All lazy-loaded** |

Net impact on initial load: **zero** (all dynamic imports).

---

## 8. Screenshot-to-Claude-Code Auto-Inject

### 8.1 Overview

When the user takes a screenshot via the OS menu (`Cmd+Shift+S`), the image is saved into the active repo's `.screenshots/` directory (gitignored), and the file path is automatically injected into the currently active Claude Code terminal session — as if the user typed and submitted it.

### 8.2 Current State

- `TakeScreenshot()` in `app.go:182` calls `screencapture -i -x`, saves to `~/Desktop/`
- Menu item wired at `main.go:53` (`Cmd+Shift+S`)
- Emits `screenshot:taken` Wails event with path
- `App.svelte:92` shows a toast on receipt
- Terminal input flows via WebSocket: `xterm.js onData → ws.send(binary) → Go bridge → tmux pane`
- `pasteToTerminal()` in `Terminal.svelte:130` wraps text in bracketed paste sequences (`\x1b[200~` ... `\x1b[201~`)
- Session tab bar (Stories 1-5 done) tracks sessions per repo with `paneTarget` IDs

### 8.3 Architecture

```
Cmd+Shift+S
  → screencapture -i -x {repoPath}/.screenshots/screenshot-{timestamp}.png
  → Go: emit "screenshot:inject" event with { path, paneTarget }
  → Terminal.svelte: EventsOn("screenshot:inject") listener
  → if event.paneTarget matches this terminal's paneTarget:
      → ws.send(bracketed paste of path + Enter)
  → Claude Code CLI receives the image path as input
```

### 8.4 Go Backend Changes

**Modify `TakeScreenshot` signature** — accept `repoPath` and `paneTarget`:

```go
func (a *App) TakeScreenshot(repoPath, paneTarget string) (string, error) {
    dir := filepath.Join(repoPath, ".screenshots")
    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", fmt.Errorf("screencapture: %w", err)
    }

    filename := fmt.Sprintf("screenshot-%s.png", time.Now().Format("20060102-150405"))
    path := filepath.Join(dir, filename)

    ctx := a.ctx
    if ctx == nil {
        ctx = context.Background()
    }

    cmd := exec.CommandContext(ctx, "screencapture", "-i", "-x", path)
    if err := cmd.Run(); err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
            return "", nil // user cancelled
        }
        return "", fmt.Errorf("screencapture: %w", err)
    }

    // Emit scoped event — only the matching terminal responds
    runtime.EventsEmit(a.ctx, "screenshot:inject", map[string]string{
        "path":       path,
        "paneTarget": paneTarget,
    })
    // Also emit general event for toast
    runtime.EventsEmit(a.ctx, "screenshot:taken", path)

    return path, nil
}
```

**Track active repo/pane** — the menu callback needs the current repo and terminal context. Two approaches:

- **Option A (recommended):** Frontend sends `activeRepoPath` and `activePaneTarget` to Go via a setter method called when the view changes. Menu callback reads these cached values.
- **Option B:** Menu emits a Wails event requesting screenshot, frontend handles calling `TakeScreenshot` with the right args. Downside: extra round-trip.

```go
// Option A: cached active context
func (a *App) SetActiveContext(repoPath, paneTarget string) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.activeRepoPath = repoPath
    a.activePaneTarget = paneTarget
}

// Menu callback uses cached values:
file.AddText("Take Screenshot", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
    a.mu.Lock()
    rp, pt := a.activeRepoPath, a.activePaneTarget
    a.mu.Unlock()
    if rp == "" {
        log.Println("screenshot: no active repo")
        return
    }
    if _, err := app.TakeScreenshot(rp, pt); err != nil {
        log.Printf("screenshot failed: %v", err)
    }
})
```

### 8.5 Frontend Changes

**Terminal.svelte** — add scoped event listener after WebSocket opens:

```svelte
// Inside onMount, after ws.onopen:
const unsubScreenshot = EventsOn('screenshot:inject', (data) => {
    if (data.paneTarget !== paneTarget) return; // not for this terminal
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const encoder = new TextEncoder();
    // Bracketed paste (same as pasteToTerminal) + Enter for auto-submit
    ws.send(encoder.encode('\x1b[200~' + data.path + '\x1b[201~'));
    ws.send(encoder.encode('\r'));
});

// In onDestroy:
if (unsubScreenshot) unsubScreenshot();
```

**AgentDetail.svelte** — call `SetActiveContext` when the view mounts or the active session changes:

```svelte
import { SetActiveContext } from '../../wailsjs/go/main/App.js';

$: if (agent?.repoPath && activeSession?.paneTarget) {
    SetActiveContext(agent.repoPath, activeSession.paneTarget);
}
```

**App.svelte** — clear context when leaving detail view:

```svelte
function goBack() {
    SetActiveContext('', '');
    currentView = 'feed';
}
```

### 8.6 .gitignore

Add `.screenshots/` to the repo's `.gitignore`. Can be done automatically when `TakeScreenshot` creates the directory:

```go
gitignorePath := filepath.Join(repoPath, ".gitignore")
data, _ := os.ReadFile(gitignorePath)
if !strings.Contains(string(data), ".screenshots/") {
    f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err == nil {
        f.WriteString("\n# mashed screenshots\n.screenshots/\n")
        f.Close()
    }
}
```

### 8.7 Claude Code CLI Compatibility

Claude Code CLI accepts image file paths pasted into the input prompt. When it receives a path ending in `.png`/`.jpg`/etc., it reads the file and attaches it as visual context. The bracketed paste + Enter sequence mimics the user typing the path and pressing Enter.

**Caveat:** If Claude Code is mid-output or not at its input prompt when the screenshot is taken, the injected path may land in an unexpected place. Mitigation: the `screencapture` dialog itself takes focus, so Claude Code is unlikely to be actively streaming when the user finishes selecting the area.

### 8.8 Implementation Estimate

| Task | Effort |
|---|---|
| Modify `TakeScreenshot` (save to repo, emit scoped event) | 30 min |
| Add `SetActiveContext` Go method + menu wiring | 30 min |
| Terminal.svelte scoped listener | 15 min |
| AgentDetail.svelte context tracking | 15 min |
| Auto-append `.gitignore` | 15 min |
| Testing (manual: screenshot → verify path appears in Claude Code) | 30 min |
| **Total** | **~2 hours** |

---

## 9. Open Questions for User Decision

**Multi-Editor:**
1. **Markdown editing mode:** WYSIWYG (Milkdown/Crepe) vs. source editing (CM6)? Crepe gives the richest Obsidian-like WYSIWYG; CM6 gives the closest-to-Obsidian source editing.
2. **Markdown preview pane:** Needed at all? If yes, side-by-side or toggle?
3. **Image viewer features:** Just zoom/pan (panzoom), or also rotate/flip (PhotoSwipe)?
4. **SVG handling:** Image-first with source toggle, or code-first with preview toggle?
5. **Diff mode for markdown:** Use CM6's merge extension, or fall back to Monaco for diffs?
6. **Scope of "repo workspace":** Only the AgentDetail view, or also the WorkflowBuilder's file interactions?

**Screenshot Inject:**
7. **Context tracking approach:** Option A (frontend pushes `SetActiveContext` to Go on view change) vs Option B (menu emits request event, frontend calls `TakeScreenshot` with args)?
8. **Multiple terminals open:** If the user has multiple Terminal tabs, should the screenshot inject into the active tab only, or offer a quick-picker?
9. **Non-agent terminals:** Should screenshot inject work for plain shell terminals too, or only Claude Code agent sessions?

---

## 10. Verdict

**Feasible. Low-to-medium complexity.**

The current architecture already has clean separation between file selection and editor rendering. The `selectedFile` → `<MonacoEditor>` pattern in AgentDetail.svelte is the single integration point. Inserting an EditorRouter component between them is a minimal, non-breaking change.

The screenshot-to-Claude-Code injection is straightforward: the WebSocket bridge already handles terminal input, the session tab system already tracks `paneTarget` IDs, and `screencapture -i` blocks until the user finishes selecting. The scoped Wails event pattern (`screenshot:inject` with `paneTarget` matching) ensures only the correct terminal responds.

**Combined implementation estimate:**
- Multi-editor (Phases 1-3): 4-7 sessions
- Screenshot inject: ~2 hours
- Go backend: `ReadFileBase64` + modified `TakeScreenshot` + `SetActiveContext`
- Frontend: `EditorRouter.svelte` + `MarkdownEditor.svelte` + `ImageViewer.svelte` + Terminal.svelte listener
