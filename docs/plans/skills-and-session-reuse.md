# Skills, Commands, and workflow-chain execution

**Status** · Phase 1 **SHIPPED** in `d6d2dc6` · Phases 2–4 pending
**Last updated** · 2026-04-11
**Owner** · next engineering session

> **Read this file first, then the code.** A fresh agent picking this up
> should be able to start Phase 2 immediately after reading this doc
> plus the five files listed in "Read first" below — no re-scouting of
> the tree required.

---

## TL;DR

Three phases remain, each a single focused commit:

| Phase | Unit of work | Risk | Ships alone? |
|---|---|---|---|
| **2** | Make **commands** (not skills) draggable canvas nodes. Drop handler, new `command` node type, load/save/restore symmetry with process nodes. | Low | Yes (non-executable until Phase 3) |
| **3** | Executor rewrite: replace "claude exited" with "idle-prompt stable" as the completion signal for every node type, plus session reuse for command nodes in the middle of a chain. | **High — touches every running workflow** | No, ships with Phase 2 |
| **4** | Polish: skill editor modal, filesystem watch, richer frontmatter validation. | Low | Each item lands independently |

Phase 1 is already done and does NOT need revisiting. Phase 2 and 3
must ship together because Phase 2 alone produces a canvas that errors
out on run ("unknown node type command") until Phase 3 teaches the
executor what a command node is.

---

## Read first (current tree state)

Five files are load-bearing for Phases 2–3. A fresh agent should open
all five before writing any code:

| Path | Why it matters |
|---|---|
| `internal/bmad/assets.go` | Asset loader + `MashedAssetInfo`/`GroupedMashedAssets` types. Phase 2 adds nothing here. |
| `internal/bmad/executor.go` (the `executeNode` function ≈ line 881+) | Current node runner. Phase 3 rewrites its completion signal. Read the full body of `executeNode` first — it's the single most important function for Phase 3. |
| `internal/bmad/question.go` (`detectIdlePrompt`, `pollForIdle`) | The idle-prompt detector Phase 3 wires into the completion loop. Already implemented, already tested against a real fixture in `fixture_verify_test.go`. Phase 3 does NOT rebuild this — it *uses* it. |
| `internal/terminal/tmux_adapter.go` (`SendInput` ≈ line 524+) | The `tmux send-keys -H` hex-injection path Phase 3 uses to fire `/<command-name>\n` into a live claude pane. Already handles raw byte fidelity for backspace, arrow keys, UTF-8, etc. Phase 3 just calls it. |
| `frontend/src/views/WorkflowBuilder.svelte` | Canvas owner. Phase 2 registers a new node type, accepts the drop, threads load/save/restore symmetry. `nodeTypes` map is around line 63; drop handler is in `CanvasPane.svelte`. |

After those five, also skim:

- `frontend/src/components/bmad/ProcessSidebar.svelte` — drag source. Already serialises `application/mashed-asset` with a JSON body; Phase 2's drop handler reads that payload as-is.
- `frontend/src/components/bmad/ProcessNode.svelte` — existing process-node Svelte component. Phase 2's `CommandNode` should mirror its shape for visual consistency.
- `.claude/skills/mashed-refactor-asset/SKILL.md` — the on-ramp skill that populates users' assets with `mashedRole` so they have something to drag. Shipped in Phase 1 via a tightly-scoped `.gitignore` negation (see `.gitignore` top of file).

---

## The schema (canonical, as shipped in Phase 1)

Every skill and command lives in a markdown file with YAML frontmatter.
**Presence of any `mashed*` field is the "mashed-ready" signal** — there
is no separate `mashedReady: true` flag to drift out of sync.

```yaml
---
# Existing fields (already present today; the refactor skill preserves them)
name: simplify
description: "Review recent changes for reuse, quality, and efficiency."

# ── Mashed fields — presence of mashedRole promotes this asset to a
# ── first-class workflow primitive. All other mashed* fields are optional
# ── and inherit per-role defaults when omitted.
mashedRole: command         # command | skill | agent
mashedCompletion: idle      # idle | exit | "marker: DONE" | "timeout: 5m"
mashedInputs: []            # artifact globs the command consumes
mashedOutputs: []           # artifact globs the command produces
mashedChainable: single     # single | none | any
mashedSessionPinned: false  # true for skills that own a long-running tmux session
---
```

### Role → defaults

| Field | `command` | `skill` | `agent` |
|---|---|---|---|
| `mashedCompletion` | `idle` | N/A | N/A |
| `mashedChainable` | `single` | `none` | `none` |
| `mashedSessionPinned` | `false` | `true` | `false` |
| Draggable onto canvas? | **yes** | no (listed read-only) | no |

### Parser rules (locked in `internal/bmad/assets.go`)

- Assets whose frontmatter lacks `mashedRole` are **silently skipped**.
  Not shown in the sidebar at all.
- `mashedRole` with an unrecognised value → log warning, skip asset.
- Missing `description:` → fall back to the first non-blank,
  non-heading body line.
- Unknown YAML keys are silently ignored (tolerant parser).
- Nonexistent directories are not errors — they return empty groups so
  the sidebar empty-state works.
- One malformed sibling does not hide valid siblings — per-file parse
  errors are logged and the bad file is skipped.

These invariants are pinned by 25 tests in `assets_test.go`. If you
change parser behaviour in Phase 2 or 3, update those tests first.

---

## Phase 1 — SHIPPED (retrospective)

Commit `d6d2dc6 feat(bmad): Skills tab — mashed-ready asset loader + refactor skill`.

What actually landed, and how it differs from the earlier draft of this
doc (which was written before the schema discussion revised the scope):

| Area | Original draft said | What actually shipped | Why |
|---|---|---|---|
| Types | `SkillInfo` / `GroupedSkills` | `MashedAssetInfo` / `GroupedMashedAssets` | Loader scans both `skills/` AND `commands/` directories, so the type is about mashed assets generally, not just skills. |
| Kind discriminator | n/a | `MashedAssetKind` (`skill` \| `command`) | Required once the loader walks both layouts. |
| Grouping | Two groups (Local / Global) | Four groups (Local/Global × Commands/Skills) | Same reason — two kinds × two scopes. |
| Loader function | `LoadSkillsFromDir` | `LoadMashedAssetsFromDir(dir, kind, source)` | Generalised to either layout via a `kind` parameter. |
| Wails binding | `ListAllSkills` | `ListAllMashedAssets` | Scans all four directories in one call. |
| Drag payload | `application/bmad-skill` | `application/mashed-asset` (JSON body) | Renamed. Payload includes `{name, path, kind, source, role, description}`. |
| On-ramp skill | mentioned but TBD | `.claude/skills/mashed-refactor-asset/SKILL.md`, tracked via tightly-scoped .gitignore negation | Users get a populated tab after one refactor-skill invocation. |

**The doc sections below have been rewritten to match this actual shape.**

---

## Phase 2 — Command nodes on the canvas

**Single commit. Non-executable until Phase 3 lands on top.**

### ⚠ Correction from the pre-schema draft

The original draft called this "Phase 2: Skill nodes" and proposed a
`SkillNode.svelte` component. That was wrong. Per the schema revision:

- **Skills** are agent-pinned, long-running, and do NOT chain. They
  stay listed in the sidebar read-only and are never dragged onto the
  canvas.
- **Commands** are short, invocable, chainable, and ARE the canvas
  primitive.

Implement **command** nodes, not skill nodes. The sidebar already
correctly refuses drag on skill rows (`effectAllowed = 'none'` in
`ProcessSidebar.svelte`'s `onMashedAssetDragStart`), so the frontend
already enforces this.

### Backend changes (small)

- `internal/bmad/types.go`:
  - Extend the `NodeType` enum with a new variant:
    ```go
    NodeTypeCommand NodeType = "command"
    ```
    (Put it alongside the existing `NodeTypeProcess`, `NodeTypeCondition`, etc.)
  - **Do NOT add a new field to `WorkflowNode`**. The existing
    `Config map[string]string` is the right slot for per-node
    command metadata. Store the asset path, command name, and
    description as config keys:
    - `config["commandName"]`
    - `config["commandPath"]`
    - `config["commandDescription"]`
    This keeps `WorkflowDef` JSON-stable and doesn't require a Wails
    regeneration cycle for a single new field.

- `internal/bmad/executor.go:executeNode`:
  - Add a dispatch at the top that routes on `EffectiveType()`:
    ```go
    switch nodesCopy[idx].EffectiveType() {
    case NodeTypeCommand:
        // Phase 3 will implement this. For Phase 2, fail fast with a
        // clear error so users who pull Phase 2 before Phase 3 see
        // "command nodes not yet runnable" in the UI instead of a
        // silent hang.
        e.failNode(state, idx, nodeID)
        return
    default:
        // existing process-node path
    }
    ```
  - This explicit fail-fast is important — it prevents the
    still-broken command path from silently masquerading as a
    process node (`NodeType` zero value would default to process).

### Frontend changes

1. **New Svelte component** — `frontend/src/components/bmad/CommandNode.svelte`
   - Mirror the shape of `ProcessNode.svelte` for visual consistency.
   - Icon: `Terminal` or `Slash` from `lucide-svelte` to signal "slash command".
   - Body: command name + truncated description.
   - NO `in:` / `out:` artifact rows — commands don't declare I/O the
     way `ProcessDef` does. (If Phase 4 wants to surface `mashedInputs`
     and `mashedOutputs`, that's a future enhancement.)
   - Status badge (running/complete/failed) re-used from `ProcessNode`.

2. **Register the new node type** — `WorkflowBuilder.svelte`
   - In the `nodeTypes` map (around line 63), add:
     ```js
     const nodeTypes = {
       bmadProcess: ProcessNode,
       command: CommandNode,   // ← new
       condition: ConditionNode,
       // ...existing entries
     };
     ```
   - Import the new component at the top.

3. **Accept the drop** — in `CanvasPane.svelte`'s drop handler
   - Look up how `application/bmad-process` drops are currently
     handled. Add a parallel branch for `application/mashed-asset`:
     ```js
     const mashedAssetRaw = e.dataTransfer.getData('application/mashed-asset');
     if (mashedAssetRaw) {
       const asset = JSON.parse(mashedAssetRaw);
       if (asset.role !== 'command') return; // belt-and-braces — sidebar already blocks skill drags
       // ...compute position via project()
       $nodes = [...$nodes, {
         id: `cmd-${Date.now()}`,
         type: 'command',
         position: { x, y },
         data: {
           label: asset.name,
           nodeType: 'command',
           config: {
             commandName: asset.name,
             commandPath: asset.path,
             commandDescription: asset.description || '',
           },
           status: 'pending',
         },
       }];
       return;
     }
     ```

4. **Load/save/restore symmetry** — in `WorkflowBuilder.svelte`
   - `loadNodesEdges`: treat a node with `n.nodeType === 'command'`
     as `type: 'command'` in the Svelte-Flow representation (the
     ternary around line 515 currently maps unknown types to
     `bmadProcess` — extend it).
   - `saveWorkflow`: already writes `nodeType: n.data.nodeType || ''`
     which will carry `'command'` through — no change needed as long
     as the data shape above is consistent.
   - `snapshotCanvas`: no change needed because it already hashes
     `n.type`, which will differ for command vs process nodes.
   - `restoreForRepo`: no change needed because the overlay operates
     on node IDs, not types.

### Verification before shipping Phase 2

- Unit test: write a WorkflowDef with one command node, save it, load
  it back, confirm the round-trip preserves `nodeType: "command"`.
  Extends existing executor save/load tests.
- Frontend smoke: drag a mashed-ready command from the sidebar onto
  the canvas, confirm the new `CommandNode.svelte` renders, save the
  workflow, reload the app, reopen the workflow, confirm the command
  node is still on the canvas.
- Negative test: click Run on a workflow containing a command node,
  confirm the node transitions to `failed` with a clear error message
  (Phase 3 will replace the fail with real execution).

### Explicit NOT-in-scope for Phase 2

- No execution. A command node on the canvas is inert.
- No edge validation based on `mashedInputs`/`mashedOutputs`.
- No visual distinction for "session-pinned" vs not.
- No command editor.

---

## Phase 3 — Executor session-reuse + idle completion  ⚠ HIGH RISK

**This is the architecturally risky commit.** It changes how every
workflow node detects completion, not just commands. Budget a full
dedicated session for this. Do not try to ship it in the same session
as Phase 2 — fresh context, fresh review pass.

### Before you write ANY code, verify three empirical invariants

These are the things I cannot prove in this document because they
depend on Claude Code CLI behaviour that can change between versions.
Run all three manually in a scratch tmux window before writing
`executeCommandNode`. If any one fails, the implementation strategy
changes and you MUST stop and redesign.

**Verification 1 — slash-command injection via `send-keys -H`**

```bash
# Start a claude session in a named tmux pane
tmux new-session -d -s verify-1 'claude --dangerously-skip-permissions'
# Wait for claude to reach its initial ❯ prompt
sleep 8
# Inject /simplify\n as hex bytes (2f 73 69 6d 70 6c 69 66 79 0d)
tmux send-keys -H -t verify-1:0.0 2f 73 69 6d 70 6c 69 66 79 0d
# Observe the pane
tmux capture-pane -t verify-1:0.0 -p | tail -20
tmux kill-session -t verify-1
```

**Pass criterion**: the pane shows claude executing the `/simplify`
slash command (you'll see it recognise the command and start output).
**Fail modes**:
- Literal `/simplify` appears in the input buffer but isn't
  executed → claude CLI isn't parsing slash commands from piped
  stdin. Escape hatch: inject one byte at a time with a 50ms delay,
  or use `send-keys -l` for the name and a separate `Enter`.
- Claude echoes back weird bytes → the `\r` vs `\n` contract is
  wrong. Try `0a` instead of `0d`.
- Nothing happens → the claude pane isn't fully initialised yet.
  Try `sleep 15` instead of 8.

**Document the outcome in the commit message** for the Phase 3
commit so future debuggers know the injection strategy is validated.

**Verification 2 — pane stability at idle prompt**

```bash
tmux new-session -d -s verify-2 'claude --dangerously-skip-permissions'
sleep 10  # let claude fully boot
H1=$(tmux capture-pane -t verify-2:0.0 -p | sha256sum)
sleep 3
H2=$(tmux capture-pane -t verify-2:0.0 -p | sha256sum)
sleep 3
H3=$(tmux capture-pane -t verify-2:0.0 -p | sha256sum)
echo "H1=$H1"
echo "H2=$H2"
echo "H3=$H3"
tmux kill-session -t verify-2
```

**Pass criterion**: all three hashes are identical. The `pollForIdle`
state machine depends on this — if the pane mutates between polls (e.g.
a blinking cursor moves bytes around in the capture output), the
two-stable-polls guard never fires and nodes never complete.
**Fail mode**: the hashes differ. Likely causes: status-bar timer
updates, animated spinner. Escape hatch: strip the status bar from
the capture before hashing, or require three stable polls instead of two.

**Verification 3 — keeping the session alive after claude exits**

```bash
tmux new-session -d -s verify-3 -c "$HOME" 'bash -c "claude --dangerously-skip-permissions; exec bash"'
sleep 10
# Simulate claude exiting
tmux send-keys -t verify-3:0.0 '/exit' Enter
sleep 3
# Pane should now host bash, not be dead
tmux list-panes -t verify-3 -F '#{pane_dead} #{pane_current_command}'
tmux kill-session -t verify-3
```

**Pass criterion**: the list-panes output shows `0 bash` (or whatever
login shell) — the pane is alive and no longer running claude. The
`exec bash` fallback kept the session usable so a downstream command
node can reuse it.
**Fail mode**: `1` (pane dead). Likely cause: the user's `$SHELL`
isn't bash, or claude's exit code triggers bash's `errexit`. Try
`bash --norc -c '... ; exec bash'` or use `$SHELL` instead of a
literal `bash`.

### The problem, stated precisely

**Current `executeNode` completion path** — `internal/bmad/executor.go:942-989`:

```go
// Poll for completion.
ticker := time.NewTicker(e.pollInterval)
defer ticker.Stop()
var questionPollCounter int
for {
    select {
    case <-ctx.Done():
        e.failNode(state, idx, nodeID)
        return
    case <-ticker.C:
        out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
        if err != nil {
            // Session gone → capture output, complete
            captured, _ := e.captureOutput(ctx, target)
            state.exec.NodeOutputs[nodeID] = captured
            e.completeNode(state, idx, nodeID)
            return
        }
        if strings.TrimSpace(string(out)) == "1" {
            // Pane dead → capture output, complete
            captured, _ := e.captureOutput(ctx, target)
            state.exec.NodeOutputs[nodeID] = captured
            e.completeNode(state, idx, nodeID)
            return
        }
        // Signal detection: idle snackbar + question detection
        questionPollCounter++
        scanQuestion := questionPollCounter%questionScanStride == 0
        e.pollNodeSignals(ctx, state, nodeID, target, scanQuestion)
    }
}
```

The only two ways out of this loop are pane death or ctx cancel.
`pollNodeSignals` fires snackbar events but never breaks the loop.
Interactive skills keep the pane alive at `❯` forever, so the loop
also runs forever.

**Current `pollForIdle` state machine** — `internal/bmad/executor.go:1152-1200`:

```go
func (e *Executor) pollForIdle(state *execState, nodeID, target, captured string) {
    curHash := hashCapturedOutput(captured)
    isIdle := detectIdlePrompt(captured)

    state.mu.Lock()
    prevHash := state.lastOutputHash[nodeID]
    wasEmitted := state.idleEmitted[nodeID]
    state.lastOutputHash[nodeID] = curHash

    // Hash changed → pane activity. Dismiss if we previously emitted.
    if curHash != prevHash {
        if wasEmitted {
            state.idleEmitted[nodeID] = false
            state.mu.Unlock()
            e.emitEvent(EventIdleDismissed, map[string]string{...})
            return
        }
        state.mu.Unlock()
        return
    }

    // Hash unchanged AND prompt visible AND not already emitted → emit.
    if wasEmitted || !isIdle {
        state.mu.Unlock()
        return
    }
    state.idleEmitted[nodeID] = true
    state.mu.Unlock()
    e.emitEvent(EventIdle, IdleEvent{...})
}
```

This state machine already has the `prevHash == curHash && isIdle`
check Phase 3 needs. What's missing is:

1. It doesn't signal the node-runner loop. It only emits a frontend
   event. The loop in `executeNode` has no way to know an idle was
   detected.
2. It has no "started yet?" guard. A brand-new pane whose baseline
   capture coincidentally shows `❯` (because claude's startup banner
   rendered one) would fire idle on the second poll, completing the
   node before claude has done any work.

### The unified fix — three new pieces and one rewrite

**Piece 1 — New helper `waitForIdleCompletion`**

Signature and contract:

```go
// waitForIdleCompletion blocks until the tmux pane at `target` is
// considered "done" by the idle state machine, then returns nil.
// Returns an error if the context is cancelled, the pane dies before
// the state machine reaches completion, or `timeout` elapses.
//
// The state machine has three stages, enforced in order:
//
//  1. PRIMING — the very first poll captures the baseline output
//     hash. Idle cannot fire on this tick regardless of prompt state,
//     because a brand-new pane whose startup banner contains a `❯`
//     would otherwise fire instantly.
//
//  2. WAITING_FOR_WORK — every subsequent poll compares hash to the
//     baseline. Until the hash CHANGES (claude has produced new
//     output), idle cannot fire. If the hash never changes before
//     `timeout`, return ErrIdleTimeoutNoStart.
//
//  3. WATCHING_FOR_IDLE — once the hash has changed at least once,
//     poll for: hash unchanged across two consecutive polls AND
//     detectIdlePrompt(latest) == true. When that's true, return nil.
//
// Pane-death mid-wait is treated as COMPLETE (return nil) to preserve
// the pre-Phase-3 behaviour for process nodes whose claude exits
// cleanly. This is intentional — dead pane = claude finished AND
// exited, which is also "done".
//
// Timeout is a wall-clock upper bound on the total wait. Default
// caller value: 30*time.Minute.
func (e *Executor) waitForIdleCompletion(ctx context.Context, target string, timeout time.Duration) error
```

Implementation sketch:

```go
func (e *Executor) waitForIdleCompletion(ctx context.Context, target string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(e.pollInterval)
    defer ticker.Stop()

    type stage int
    const (
        stagePriming stage = iota
        stageWaitingForWork
        stageWatchingForIdle
    )

    var (
        st           = stagePriming
        baselineHash string
        lastHash     string
    )

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if time.Now().After(deadline) {
                return ErrIdleTimeoutNoStart
            }
            // Check pane liveness — dead pane = complete, preserves legacy.
            out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
            if err != nil || strings.TrimSpace(string(out)) == "1" {
                return nil
            }
            // Capture and hash.
            captured, err := e.captureQuestionOutput(ctx, target)
            if err != nil {
                continue // transient capture error, try next tick
            }
            curHash := hashCapturedOutput(captured)

            switch st {
            case stagePriming:
                baselineHash = curHash
                lastHash = curHash
                st = stageWaitingForWork

            case stageWaitingForWork:
                if curHash != baselineHash {
                    lastHash = curHash
                    st = stageWatchingForIdle
                }
                // else: still waiting for claude to produce output

            case stageWatchingForIdle:
                if curHash == lastHash && detectIdlePrompt(captured) {
                    return nil // DONE
                }
                lastHash = curHash
            }
        }
    }
}
```

**Piece 2 — New node-type dispatcher inside `executeNode`**

Current `executeNode` (line 881) builds the process-node command and
spawns a session unconditionally. Phase 3 wraps it in a dispatch:

```go
func (e *Executor) executeNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
    idx := nodeIndex[nodeID]
    state.mu.Lock()
    nodeType := state.exec.Nodes[idx].EffectiveType()
    state.mu.Unlock()

    switch nodeType {
    case NodeTypeCommand:
        e.executeCommandNode(ctx, state, nodeIndex, nodeID, repoPath, model)
    case NodeTypeProcess:
        e.executeProcessNode(ctx, state, nodeIndex, nodeID, repoPath, model)
    default:
        // condition/loop/merge/transform remain unchanged — they're
        // synchronous and don't touch tmux.
        e.executeDefaultNode(ctx, state, nodeIndex, nodeID, repoPath, model)
    }
}
```

**Piece 3 — `executeProcessNode` (current `executeNode` body, two changes)**

Lift the entire current body of `executeNode` into a new method
`executeProcessNode`. Two small changes:

1. **Wrap the claude invocation in `bash -c "... ; exec bash"`** so the pane stays alive after claude exits. Current line 911 builds `command` — change to:
   ```go
   innerCommand := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`, model, proc.SkillName, contextStr)
   command := fmt.Sprintf(`bash -c '%s; exec bash'`, strings.ReplaceAll(innerCommand, `'`, `'\''`))
   ```
   The shell-quoting escape for embedded single quotes is ugly but
   necessary — skill names and context strings can contain them.

2. **Replace the pane-death loop** (lines 942-989) with a call to
   `waitForIdleCompletion(ctx, target, 30*time.Minute)`. Keep the
   capture-on-complete step, keep `completeNode` on success, keep
   `failNode` on error/timeout. Keep the call to `pollNodeSignals` as
   a side-effect so the snackbar events still fire — but do it on a
   separate ticker (or split `waitForIdleCompletion` into a form that
   can call the existing `pollNodeSignals` as a hook). Simplest: have
   the wait function call `pollNodeSignals` internally on each tick.

**Piece 4 — `executeCommandNode` (new)**

Full implementation:

```go
func (e *Executor) executeCommandNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
    idx := nodeIndex[nodeID]

    // Mark running.
    state.mu.Lock()
    state.exec.Nodes[idx].Status = NodeRunning
    state.exec.CurrentNode = nodeID
    node := state.exec.Nodes[idx]
    commandName := node.Config["commandName"]
    state.mu.Unlock()
    e.emitEvent("bmad:node:status", NodeStatusEvent{
        ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning,
    })

    if commandName == "" {
        log.Printf("bmad: command node %s missing commandName in config", nodeID)
        e.failNode(state, idx, nodeID)
        return
    }

    // Find a live upstream session to reuse.
    target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, nodeID, repoPath, model)
    if err != nil {
        e.failNode(state, idx, nodeID)
        return
    }

    // Record this node's target so downstream command nodes can chain.
    state.mu.Lock()
    state.exec.Nodes[idx].TmuxTarget = target
    state.mu.Unlock()
    e.emitEvent("bmad:node:status", NodeStatusEvent{
        ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning, TmuxTarget: target,
    })

    // If we reused an existing session, inject the slash command.
    // If we spawned a new session, the command is already the initial
    // claude invocation argument — no injection needed.
    if reused {
        if err := e.injectSlashCommand(ctx, target, commandName); err != nil {
            log.Printf("bmad: failed to inject /%s into %s: %v", commandName, target, err)
            e.failNode(state, idx, nodeID)
            return
        }
    }

    // Wait for idle completion.
    if err := e.waitForIdleCompletion(ctx, target, 30*time.Minute); err != nil {
        log.Printf("bmad: command %s did not complete: %v", commandName, err)
        e.failNode(state, idx, nodeID)
        return
    }

    // Capture output (best-effort) and complete.
    captured, _ := e.captureOutput(ctx, target)
    state.mu.Lock()
    state.exec.NodeOutputs[nodeID] = captured
    state.mu.Unlock()
    e.completeNode(state, idx, nodeID)
}

// resolveCommandSession scans incoming edges for a live upstream
// session. Returns (target, reused, err). When reused==false the
// caller must spawn a fresh session for this command (first node in
// a chain, or all parents ended their sessions).
func (e *Executor) resolveCommandSession(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) (string, bool, error) {
    // Collect incoming-edge source node IDs.
    var upstream []*WorkflowNode
    state.mu.Lock()
    for _, edge := range state.exec.WorkflowEdges() { // helper or inline loop
        if edge.Target == nodeID {
            if srcIdx, ok := nodeIndex[edge.Source]; ok {
                upstream = append(upstream, &state.exec.Nodes[srcIdx])
            }
        }
    }
    state.mu.Unlock()

    // Filter to those with a live TmuxTarget.
    var live []*WorkflowNode
    for _, n := range upstream {
        if n.TmuxTarget == "" {
            continue
        }
        // Verify the session is still alive.
        out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", n.TmuxTarget, "-F", "#{pane_dead}")
        if err != nil || strings.TrimSpace(string(out)) != "0" {
            continue
        }
        live = append(live, n)
    }

    switch len(live) {
    case 0:
        // No upstream session to reuse — spawn a new one.
        // This branch re-uses executeProcessNode's spawn helper after
        // a small refactor that factors it out of executeProcessNode.
        target, err := e.spawnCommandSession(ctx, state, nodeID, repoPath, model, commandName)
        return target, false, err
    case 1:
        return live[0].TmuxTarget, true, nil
    default:
        // Multi-parent merge: pick the most recently started.
        // v1 is sequential-only; Phase 4 can add merge-node semantics.
        latest := live[0]
        for _, n := range live[1:] {
            if n.StartedAt > latest.StartedAt { // RFC3339 string compare is fine
                latest = n
            }
        }
        log.Printf("bmad: command node %s has %d live parents; reusing most recent (%s)", nodeID, len(live), latest.ID)
        return latest.TmuxTarget, true, nil
    }
}

// injectSlashCommand writes `/<name>\n` into the pane using
// tmux send-keys -H hex bytes. The TmuxAdapter already has a
// byte-fidelity implementation in SendInput — if it's only exposed
// on *TmuxAttachment, add a new method `SendInputToTarget(ctx, target, data)`
// that takes a string target instead of a bound attachment. The
// existing hex-formatting loop in tmux_adapter.go:524+ is the body
// to copy.
func (e *Executor) injectSlashCommand(ctx context.Context, target, name string) error {
    payload := []byte("/" + name + "\n")
    return e.tmuxAdapter.SendInputToTarget(ctx, target, payload)
}
```

**Piece 5 — Chain-tail cleanup at workflow terminal state**

Add a finalizer that runs when the workflow transitions to
complete/failed:

```go
// killWorkflowChainTails terminates every unique tmux session held
// by any node in the workflow. Called from the workflow runner's
// terminal-state transition so long-lived panes don't pile up.
//
// A "chain tail" is any surviving TmuxTarget at the moment the
// workflow ends. Because chained nodes share a target, this naturally
// dedupes to one kill per chain.
func (e *Executor) killWorkflowChainTails(ctx context.Context, state *execState) {
    state.mu.Lock()
    seen := map[string]struct{}{}
    for _, node := range state.exec.Nodes {
        if node.TmuxTarget == "" {
            continue
        }
        seen[bareSessionName(node.TmuxTarget)] = struct{}{}
    }
    state.mu.Unlock()

    for session := range seen {
        killCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        _, _ = e.runCmd(killCtx, "tmux", "kill-session", "-t", session)
        cancel()
    }
}

// bareSessionName is already in internal/bmad/cleanup.go.
```

Wire it into the workflow runner wherever `ExecComplete` or
`ExecFailed` is set. Run it before the runner's goroutine returns.

### State fields the executor needs

`execState` already has `lastOutputHash map[string]string` and
`idleEmitted map[string]bool` from the existing `pollForIdle`. Phase 3
does NOT need new per-node state — the `waitForIdleCompletion` state
machine uses local variables, not `execState` fields.

### Order of operations (do not reorder)

1. Do the three empirical verifications. Write their outcomes into
   the eventual commit message.
2. Add `ErrIdleTimeoutNoStart` as a sentinel in executor.go.
3. Write `waitForIdleCompletion` with unit tests driven by the
   existing `CommandRunner` mock. Cases: priming → work → stable →
   nil, priming → no-work → ErrIdleTimeoutNoStart, pane-dead during
   waiting-for-work → nil, pane-dead during watching-for-idle → nil,
   ctx cancel in each stage → ctx.Err.
4. Refactor `executeNode` body into `executeProcessNode`, wrap
   claude in `bash -c "... ; exec bash"`, swap the poll loop for a
   call to `waitForIdleCompletion`. Verify the full bmad test suite
   still passes — this is the change most likely to break existing
   tests because fixture `CommandRunner` mocks expect the old
   `#{pane_dead}` transition. Update mocks to fire the new signal
   (hash changing then stabilising with a trailing `❯`).
5. Add the `executeNode` dispatcher on `EffectiveType()`.
6. Write `executeCommandNode` + `resolveCommandSession` +
   `injectSlashCommand` + tests. Integration test: DAG with
   `process → command → command`, drive with mock runner, assert
   all three share the same `TmuxTarget` and complete in order.
7. Factor `spawnCommandSession` out of `executeProcessNode` so both
   node-types share the spawn path.
8. Add `killWorkflowChainTails`, wire into terminal-state
   transitions.
9. Add `SendInputToTarget(ctx, target, data)` helper on
   `*TmuxAdapter` if it doesn't already exist. Thin wrapper that
   builds the same argv as `TmuxAttachment.SendInput`.
10. Full test run. Manual smoke test in the Wails dev app: drag a
    command onto a canvas with a process node upstream, click Run,
    confirm the command injects and completes.

### Test coverage Phase 3 adds

- **Unit — `waitForIdleCompletion`** (driven via the `CommandRunner`
  mock): five cases listed in step 3 above. All pass without any
  real tmux invocation.
- **Unit — `resolveCommandSession`**: zero live parents, one live
  parent, multiple live parents (latest wins), parent TmuxTarget set
  but session dead (falls through to spawn).
- **Unit — `injectSlashCommand`**: confirms exact tmux argv
  (`send-keys -H -t <target> 2f 73 69 6d 70 6c 69 66 79 0d` for
  `/simplify\n`). This is a copy-paste of the existing
  `TestTmuxAttachment_AC3_SendInputUsesSendKeysHex` style.
- **Integration — `process → command → command → process`**: drive
  a 4-node DAG through the mock runner. Assert:
  - Node 1 (process): spawns target A.
  - Node 2 (command): reuses target A.
  - Node 3 (command): reuses target A.
  - Node 4 (process): spawns target B (does NOT reuse A — process
    nodes always spawn their own, per the design).
  - At workflow-complete: `kill-session` called on A and B.
- **Integration — chain-tail kill at fail**: DAG with one process
  node that fails mid-execution. Assert `kill-session` called on
  its target.
- **Regression — existing `CommandRunner` mock executor tests**:
  they rely on `#{pane_dead}` returning "1" after N ticks.
  `waitForIdleCompletion` re-uses pane-death as a completion signal,
  so those tests should keep passing unchanged — but run the full
  suite to confirm.

### Load-bearing warnings

- **The `bash -c "...; exec bash"` wrapper is a trust boundary.**
  Be precise about how the inner command is quoted. Single-quote
  escapes (`'\''`) are ugly but the only portable answer across
  user shells. Write a dedicated unit test that asserts the
  generated argv for a command containing single quotes, backticks,
  and dollar signs.
- **Existing executor tests will likely break** on the
  `executeProcessNode` change because the `CommandRunner` mock's
  expected call sequence (`tmux new-session ... ; list-panes ... ;
  list-panes ... ; pane_dead=1`) is hardcoded. Plan to update the
  shared mock helpers to simulate the new state transitions (hash
  changes + idle prompt + stable).
- **The `injectCommand` timing matters**: inject too early and
  claude isn't ready to parse slash commands. The `waitForIdleCompletion`
  state machine handles this for process nodes (first poll primes
  the baseline), but for command nodes that reuse an existing
  session, the pane is ALREADY at the idle prompt when we inject.
  The state machine will enter `stagePriming`, capture the
  post-injection hash, wait for it to change (claude responds to the
  slash command), and then watch for idle. This is correct —
  verified by walking the state machine against a simulated trace.
- **Multi-parent merge strategy is a PUNT, not a feature**. Document
  it as "first parent wins" in the code AND in the commit message.
  A real merge-node feature is Phase 4 or later.

---

## Phase 4 — Polish (incremental, each lands independently)

- **Skill editor modal**: Svelte modal to create/edit a SKILL.md with
  name, description, and mashed frontmatter fields. File I/O via a
  new wails binding.
- **Filesystem watch**: subscribe to changes under
  `{repoPath}/.claude/{skills,commands}/` and `~/.claude/{skills,commands}/`.
  Debounce 500 ms. Call `ListAllMashedAssets` on change and update
  `groupedMashedAssets`. Use the pattern already used for repo
  scanning.
- **Frontmatter validation badges**: warn in the sidebar when an
  asset's `description` is missing / too short / too long, or when
  `mashedInputs`/`mashedOutputs` reference paths that don't exist in
  the repo. Inline badge only, never blocking.
- **Rich metadata surfaced**: surface `mashedInputs`/`mashedOutputs`
  as collapsible detail rows on the CommandNode so the user can see
  what the command consumes/produces before running it.

---

## Decisions already made (don't re-litigate)

| # | Question | Answer | Source |
|---|---|---|---|
| 1 | Invocation syntax for a command in a live session | `/<command-name>\n` via `tmux send-keys -H` (hex-byte path) | User, schema discussion |
| 2 | Default "done" contract | Idle-prompt stable across two consecutive polls; overridable per-asset via `mashedCompletion` | User, schema discussion |
| 3 | Chain-tail cleanup | Explicit `tmux kill-session` at workflow terminal state on each chain tail. `CleanupStaleSessions` at startup remains as safety net. | User, schema discussion |
| 4 | Multi-parent command nodes | Schema-driven via `mashedChainable`. v1 supports `single` only. `any` is reserved for future merge-node design. | User, schema discussion |
| 5 | Skills vs commands as canvas primitive | Commands are chainable and draggable. Skills are agent-pinned, listed in sidebar read-only, never dragged. | User, schema discussion |
| 6 | `.claude/` visibility in git | Tightly-scoped negation in `.gitignore` — ONLY `.claude/skills/mashed-refactor-asset/` is tracked. Every other `.claude/` path stays personal/ignored. | Phase 1 commit |

---

## Open questions that remain

Only one is blocking, and it's only blocking Phase 3:

- **Empirical verification of `send-keys -H` slash-command injection**
  (see "Load-bearing verifications" in Phase 3). Must be tested before
  writing `executeCommandNode`. If the test fails, we need a different
  injection strategy and the plan for Phase 3 changes.

Non-blocking:

- **Seed `mashed*` frontmatter on the author's existing commands**.
  That's a one-time doc/annotation task, not a coding task. Users
  who want to populate their tab right now can run the
  `mashed-refactor-asset` skill that shipped in Phase 1.

---

## Proposed delivery order

| Order | Commit | Prereq | Notes |
|---|---|---|---|
| 1 | Phase 2 — command node canvas support | Phase 1 (shipped) | Small. Ships with explicit fail-on-run for command nodes so users who pull Phase 2 alone see a clear message. |
| 2 | Phase 3 — executor rewrite (idle completion + session reuse + command execution) | Phase 2 | Big. Touches every running workflow. Dedicated review pass. Do the three empirical verifications first. |
| 3 | Phase 4a — skill editor | Phase 1 | Independent. |
| 4 | Phase 4b — filesystem watch | Phase 1 | Independent. |
| 5 | Phase 4c — frontmatter validation badges | Phase 1 | Independent. |

Phases 2 and 3 are sequential and MUST ship in that order or the UI
creates unrunnable workflows.

---

## Appendix: the existing pieces Phase 3 reuses (precise locations)

If you're picking up Phase 3 in a fresh session, these are the exact
files and functions you'll be calling or extending. Read them before
writing new code.

### Idle detection pipeline (already built, already tested)

- `internal/bmad/question.go`:
  - `detectIdlePrompt(output string) bool` — tolerant `❯` line detector,
    handles NBSP, ANSI-stripped. Pinned by `question_test.go` and by
    `fixture_verify_test.go` (which runs against a real captured
    brainstorming session on disk).
  - `ansiStripRegex` — used by `detectIdlePrompt`.
- `internal/bmad/executor.go`:
  - `pollForIdle(state *execState, nodeID string, target string)` —
    output-hash state machine. Currently only emits `EventIdle` /
    `EventIdleDismissed` for the frontend snackbar. Phase 3 refactors
    this into a form that also signals the node-runner loop.

### Hex-byte input path (already built, already tested)

- `internal/terminal/tmux_adapter.go`:
  - `(TmuxAttachment).SendInput(data []byte)` — builds
    `tmux send-keys -H -t <target> <hex bytes>`. Handles backspace
    (0x7f), arrow keys (ESC sequences), UTF-8, and bracketed-paste
    markers correctly. Test coverage: `tmux_adapter_test.go`
    `TestTmuxAttachment_AC3_SendInputUsesSendKeysHex` with a table of
    7 cases including UTF-8 chevron and paste markers.
  - Phase 3 may need to add a helper that takes a `target` string
    directly (not bound to a live `*TmuxAttachment`) so the executor
    can inject into sessions it didn't open itself. Trivial to add.

### Cleanup path (already built)

- `internal/bmad/cleanup.go`:
  - `CleanupStaleSessions(ctx context.Context) error` — kills any
    `bmad-*` tmux session not referenced by the current executor
    state. Runs at app startup. Phase 3 adds an explicit chain-tail
    kill at workflow terminal state — this function stays as the
    safety net for crashes.

### Node status events (already emitted)

- `internal/bmad/executor.go`:
  - `NodeStatusEvent { ExecID, NodeID, Status, TmuxTarget }` —
    emitted on every node transition. Phase 3 sets `TmuxTarget`
    in the reused-session case so the frontend's View Terminal
    button picks up the right target.

### Frontend state flow for node status (already wired)

- `frontend/src/views/WorkflowBuilder.svelte`:
  - Listener on `bmad:node:status` updates `$nodes` via
    `nodes.update(...)`. Command nodes will appear here with
    `type: 'command'` after Phase 2, but the status-update path is
    already type-agnostic — no change needed in Phase 3.
