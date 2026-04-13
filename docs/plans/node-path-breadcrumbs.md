# Plan: Node path breadcrumbs + artifact auto-propagation

**Date:** 2026-04-13
**Status:** draft — awaiting approval
**Owner:** lead
**Related:** `internal/bmad/artifacts.go`, `internal/bmad/executor.go`, `frontend/src/components/bmad/*.svelte`

---

## Problem

BMAD workflow canvas shows symbolic artifact names (`brainstorm-notes`, `PRD.md`) on every process node but never surfaces the *resolved filesystem path*. Users can't tell:

1. Which file a File Loader actually points at.
2. Where a BMAD node's output will land.
3. Whether the downstream File Loader is correctly wired to receive the upstream output.

Consequence: workflows run blind. Users eyeball `_bmad-output/` manually to verify what got written. Artifact contract mismatches only surface at execution time.

## Goal

Every node displays both layers of truth:
- **Symbolic artifact name** (the contract) — already visible.
- **Resolved breadcrumb path** (the reality) — new. Format: `.../<basename>`, dim `—` when unset.

Auto-propagate resolved output paths from upstream BMAD nodes into downstream File Loaders so the canvas stays coherent without manual path copy-paste.

---

## Phase 1 — UI breadcrumbs on every node *(S)*

### Discovery (Task 1.0)

Identify the Svelte component rendering the **File Loader** node. Candidates: `ProcessNode.svelte`, `CommandNode.svelte`, or an undiscovered node type. Grep the frontend for `File Loader` string + `file-path` symbolic output to confirm.

### Task 1.1 — Add breadcrumb row to `ProcessNode.svelte`

Beneath the existing `in:` / `out:` rows (which show symbolic names), add:

```
│ IN:  brainstorm-notes     │
│      .../brainstorm.md    │   ← new breadcrumb row, dim, smaller
│ OUT: prd                  │
│      —                    │   ← dim em-dash when unresolved
```

Render path from `node.data.config.inputPath` / `outputPath` (extend struct if not present). Format: `.../<basename>` when value exists, `—` otherwise. Style: 9–10px mono, `--text-muted` color, reuse `.artifact-label` / `.artifact-list` typography fixed this session.

### Task 1.2 — Same treatment in `CommandNode.svelte` and File Loader component

Consistent breadcrumb row per the pattern above. File Loader's output path is its *loaded* file; breadcrumb confirms at a glance what got selected in the browse dialog.

### Task 1.3 — Multi-input rendering

When a node accepts multiple inputs (e.g. `["prd", "sprint-status"]`), render one breadcrumb row per symbolic name. Labels truncate with `text-overflow: ellipsis`; full path visible on hover via `title` attribute.

### Acceptance

- Every process node on canvas shows the breadcrumb row for every symbolic in/out.
- Unresolved paths render `—` (dim), not empty.
- Resolved paths are always `.../<basename>` — never absolute, never truncated mid-segment.
- Full path on hover.

---

## Phase 2 — Artifact auto-propagation *(M)*

### Task 2.1 — Executor writes resolved output path on completion

In `internal/bmad/executor.go`, when a BMAD process node transitions to `complete`:

```go
for _, artifactName := range node.Outputs {
    resolved := ResolveArtifactPath(artifactName, repoPath)  // already exists
    if resolved == "" {
        continue  // unmapped (e.g. "code", "tests") — skip
    }
    if !fileExists(resolved) {
        continue  // artifact didn't actually get written — caller logs warn
    }
    node.Data.Config.OutputPaths[artifactName] = resolved
}
```

Emit the existing `bmad:node:artifacts` event with the updated map.

### Task 2.2 — Downstream File Loader auto-fill

Frontend listens for `bmad:node:artifacts`. For each updated path, walk outgoing edges from the source node. For every target whose `data.config.inputPath` is **empty** AND whose accepted input artifact name matches, copy the resolved path.

Rule (per architectural Q3): **empty-only fill**. Never clobber a user-edited path.

### Task 2.3 — Execution model

Path resolution runs on the Go side (authoritative). Graph traversal for auto-fill runs on the Svelte side (already has the xyflow graph). No new events needed; existing `bmad:node:artifacts` carries the payload.

### Acceptance

- Running a BMAD node to completion causes its `OUT:` breadcrumb to light up with the resolved path.
- The immediate downstream File Loader's `IN:` breadcrumb auto-fills to the same path.
- Manually editing the downstream path before or after upstream completes persists — no clobber.
- Unmapped artifact names (`"code"`, `"tests"`) show `—` in breadcrumb, never a stale or guessed path.

---

## Phase 2b — MultiFileLoader node type *(M)*

User preference: new node over loop-unrolling. Rationale: one node, one config panel, list-valued input. Loop control-flow would require pseudo-edges for path enumeration — messy.

### Task 2b.1 — Backend

- `internal/bmad/types.go`: add `NodeTypeMultiFileLoader NodeType = "multiFileLoader"`.
- `internal/bmad/executor.go`: synchronous execution (no tmux). Reads each configured path, emits one output per path with a naming scheme: `file[0]`, `file[1]`, … or caller-supplied labels.
- `internal/bmad/registry.go`: register the new node so it appears in the sidebar.

### Task 2b.2 — Frontend

- New component: `frontend/src/components/bmad/MultiFileLoaderNode.svelte`. Renders one breadcrumb per configured path.
- `NodeConfigPanel.svelte`: list editor. Reuse `ArrayEditorModal.svelte` (already exists) for add/edit/reorder.
- `ProcessSidebar.svelte`: surface in the Utilities group.

### Acceptance

- User can drag MultiFileLoader onto canvas, add N paths via ArrayEditorModal, each breadcrumb visible on the node.
- Each path emits as a numbered output; downstream nodes can wire to `file[0]`, `file[1]`, etc. symbolically.

---

## Phase 3 — Ollama signal router *(L, scoped)*

### Role

**Intermediary only.** Ollama does not generate content. It inspects BMAD output events and decides:

1. Which downstream node receives the output when auto-propagation is ambiguous (multiple targets accept the same artifact name).
2. Whether to fire a toast/snackbar based on output content (e.g. "BMAD flagged a blocker — show warning toast").

### Why LLM instead of deterministic

Phase 2 handles the easy case (one-to-one artifact routing). Phase 3 only kicks in when:
- Multiple downstream targets accept the same symbolic artifact (fan-out is ambiguous).
- The output content carries semantic signal not captured in the artifact map (e.g. BMAD emits an ok status but the body says "needs clarification" — toast should fire).

Deterministic code cannot read semantic content.

### Task 3.0 — Settings UI: Ollama lifecycle

- `frontend/src/views/Settings.svelte` gets an **Ollama** section.
- **State machine**: `not-installed` → `installed-stopped` → `starting` → `running` → `model-selected`.
  - `not-installed`: detect by probing `which ollama` via a new `IsOllamaInstalled()` binding. Show download link to https://ollama.com.
  - `installed-stopped`: show **Start Ollama server** button. Click → `StartOllamaServer()` spawns `ollama serve` as a background subprocess tracked in `app.go`. Poll `GET http://localhost:11434/api/version` until 200 or timeout.
  - `running`: call `GET /api/tags` → list installed models. Show dropdown populated with model names + param count + size.
  - `model-selected`: user picks from dropdown. Setting persists in `mashedConfig.OllamaModel` + `mashedConfig.OllamaEnabled`. A `Stop Ollama server` button is visible.
- **Wails bindings** (new, in `app.go`):
  - `IsOllamaInstalled() bool`
  - `StartOllamaServer() error`
  - `StopOllamaServer() error`
  - `GetOllamaStatus() (OllamaStatus, error)` — returns running + pid + port
  - `ListOllamaModels() ([]OllamaModel, error)` — parses `/api/tags`
  - `SetOllamaModel(name string) error`
- **Graceful shutdown**: if Mashed started Ollama, Mashed stops it on app quit. If Ollama was already running when Mashed started, leave it alone (track with a `startedByUs` flag).

### Task 3.1 — Ollama client

- New package: `internal/ollama/client.go`. HTTP client to `localhost:11434`.
- Config: model name (from Settings), timeout (default 2s for routing calls), enabled flag.
- Exposed via `app.go` as `IsOllamaAvailable()` (covers installed + running + model-selected) + `AskOllama(prompt, schema)` where `schema` constrains the JSON response shape.

### Task 3.2 — Integration

- Subscribe to `bmad:node:artifacts` + `bmad:node:status` events on the Go side.
- When ambiguous downstream target found OR output content matches a semantic trigger regex, call Ollama with a tight prompt. Parse JSON response: `{action: "populate-path" | "toast" | "noop", target?: nodeId, toastLevel?: string, toastText?: string}`.
- Apply the decision.

### Task 3.3 — Observability

- Log every Ollama call + decision to `~/.mashed/logs/ollama-router.jsonl`.
- Surface latency + token count in NotificationFeed so users can see what the router is deciding.

### Acceptance

- Ollama off by default. Phase 2 still works without it.
- When enabled, ambiguous fan-out gets routed consistently (same input → same decision within a session, until model or prompt changes).
- All LLM decisions logged with the prompt, the response, and the action taken.

---

## Architectural decisions (locked)

1. **Artifact path storage shape** — **resolved**. `OutputPaths map[string]string` (artifact name → absolute path) on `WorkflowNode.Data.Config`. `json:",omitempty"` so old saved workflows still parse. Risks to mitigate: (a) concurrent writes during parallel execution → wrap map mutation in the executor's existing node mutex; (b) stale paths if user moves files between runs → clear on next `start`; (c) drift if `artifacts.go` map changes → old workflows ignore unknown keys, re-populate on next run. All tractable.
2. **Re-run behavior** — **resolved**. Default deterministic: clear breadcrumbs on `start`, populate on `complete`. Ollama (Phase 3) may override — e.g. if re-run produced semantically equivalent output, router can choose to keep old path. Off by default.
3. **MultiFileLoader artifact names** — **resolved**. User-supplied labels. BMAD references files by name, so labels propagate verbatim into downstream prompts. Config UI: one row per file, two fields — `label` + `path`. Empty label falls back to positional `file[0]`, `file[1]`… as a safety net so the node still runs without manual labeling.
4. **Ollama model selection** — **resolved**. Managed via Settings UI (see Phase 3 Task 3.0 below), not hardcoded.

---

## Scope checklist (in this plan)

- [x] Phase 1: breadcrumb UI on all nodes
- [x] Phase 2: auto-propagation Go → Svelte (`OutputPaths` map, node mutex, empty-only fill)
- [x] Phase 2b: MultiFileLoader node type (label + path pairs)
- [x] Phase 3: Ollama signal router — Settings lifecycle (Task 3.0) + client (3.1) + integration (3.2) + logging (3.3). Deferred, scoped, opt-in.

## Out of scope

- Moving BMAD output files around on disk. Artifact path resolution remains via `artifacts.go`'s canonical map.
- Node-to-node communication other than file paths (no streaming, no pub/sub).
- Ollama doing content generation. Intermediary only.

---

## Execution order recommendation

1. Phase 1 standalone — instant visual win, zero backend risk.
2. Phase 2 after Phase 1 lands — surfaces the data model question (#1 above).
3. Phase 2b after Phase 2 — needs the same path-storage plumbing.
4. Phase 3 last — only after Phase 2 surfaces evidence of ambiguity that deterministic code can't handle.

Each phase ships independently. Grade each with `/skill-validator` → report-only acceptance before moving on.
