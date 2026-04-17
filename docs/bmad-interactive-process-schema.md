# BMAD Interactive Process Schema

> **Date:** 2026-04-17
> **Scope:** First-class support for interactive BMAD processes (brainstorming, product-brief, party-mode, advanced-elicitation) with typed input/output contracts, declared suspension states, explicit downstream blocking, and iteration gates.
> **Status:** Design spec — not yet implemented.
> **Replaces:** Implicit tmux-heuristic gating in `internal/bmad/executor.go` (kept as fallback for autonomous nodes).
> **Related:** `docs/repomixer/bmad-method/bmad-method.xml` (upstream BMAD reference), `internal/bmad/question.go` (current gate implementation).

---

## 1. Problem Statement

The upstream BMAD Method (`bmad-code-org/BMAD-METHOD`) ships a family of **interactive** processes that iterate with the user until both agent and human are satisfied:

| Process | Interaction pattern |
|---|---|
| `bmad-brainstorming` | Pick technique → generate ideas → probe → organise → user says "done" |
| `bmad-product-brief` | 5 guided stages; each has a soft "anything else?" gate |
| `bmad-party-mode` | Multi-agent conversation until user calls it |
| `bmad-advanced-elicitation` | Present 5 methods → pick 1-5/r/a/x → apply → re-present; loop until `x` |

All four share three structural requirements:

1. **Typed inputs** mixing files, upstream outputs, user selections, and registry lookups.
2. **Suspension** — the process visibly pauses until a human responds.
3. **Downstream blocking** — no subsequent node starts while the current process is suspended.

Mashed's BMAD executor currently models these requirements implicitly:

- **Inputs.** `ProcessDef.Inputs []string` is a flat artifact-name list. `buildContextStringV3` (`internal/bmad/executor.go`) resolves file artifacts via `ResolveArtifactPath` and grabs direct-upstream `NodeOutputs` capped at 2000 chars. There is no declared shape for user inputs — every prompt is rendered inline by claude inside the tmux session.
- **Suspension.** Detected heuristically by `hasRecentQuestion` (last 15 lines of tmux capture) and `detectIdlePrompt` (`❯` chevron). The node's `WorkflowNodeStatus` stays `NodeRunning` while suspended — there is no `awaiting_input` state.
- **Downstream.** Held correctly as a side-effect: `runDynamic`'s ready-set only decrements a target's in-degree on `NodeComplete`, and a suspended node is not complete. Correct, but the guarantee is accidental rather than asserted.
- **Iteration.** Available only via `NodeTypeLoop` / `NodeTypeLoopUntil`, which treat each iteration as a fresh subgraph execution. There is no concept of "same tmux session, round N+1, user-gated continuation".

### Concrete gaps

1. No way for a node to declare "I need a choice from `methods.csv` before starting."
2. No way for the frontend to pre-render an input widget (radio, checkbox, textarea, approval button) because input shape is unknown until claude prints a question.
3. If the desktop app restarts mid-session, the pending user prompt is lost — the tmux pane is dead and the in-flight question hash is gone.
4. `RespondToQuestion` is a single-slot API: `(execID, nodeID, answer)` — it cannot distinguish "selection #3 in the method list" from "freeform answer to question 2 of 5".
5. No declared termination criterion for iterative processes; they rely on claude's own loop instructions plus tmux idle detection, which is brittle when the user wants a hard round cap.

### What this spec changes

- A typed input/output contract on `ProcessDef`.
- A new `NodeAwaitingInput` status that is the single source of truth for "suspended".
- A typed `PendingPrompt` record that survives snapshot/restore.
- A generalised `RespondToInput(execID, nodeID, inputID, value)` API.
- A declared `IterationGate` per process — accept tokens, round caps, artifact-exists, custom expressions.
- A dedicated `executeInteractiveNode` path in the executor, parallel to the existing `executeNode` for autonomous nodes.

Legacy tmux question/idle detection stays as a **fallback** for autonomous nodes where claude CLI still asks its own questions.

---

## 2. Design Principles

1. **Declared over heuristic.** Every interactive dependency is declared in `ProcessDef`. The executor resolves declarations; it does not guess from tmux output.
2. **Typed once, reused everywhere.** One `InputSpec` shape drives: executor resolution, frontend widget rendering, validation, persistence, and resume.
3. **Additive, not breaking.** Existing `ProcessDef.Inputs/Outputs` stay. New fields default to zero values; `Mode == ""` behaves as `InteractAutonomous` — identical to today.
4. **Downstream hold is a property, not a check.** The ready-set algorithm's existing invariant (in-degree decrements only on `NodeComplete`) already blocks downstream. The new `NodeAwaitingInput` state makes the property observable; it does not change ready-set logic.
5. **Survive restart.** All in-flight prompts, accumulated user inputs, and round counters are part of the `WorkflowExecution` snapshot persisted to `~/.mashed/workflows/{id}/execution.json`.
6. **One input, one suspension.** A node suspends once per unresolved `InputSpec`. It does not accumulate a queue of prompts the user must clear all at once.
7. **Validation at the boundary.** `RespondToInput` validates against declared `Shape` + `Options` + `Validation` regex before storing. Invalid input stays in the awaiting state and emits `:invalid`.

---

## 3. Core Types

All new types live in `internal/bmad/types.go`. Existing types gain optional fields (marked `NEW`).

### 3.1 Input declarations

```go
// InputSource classifies the origin of a process input.
type InputSource string

const (
    InputFromFile     InputSource = "file"     // artifact on disk; resolved via ResolveArtifactPath
    InputFromUpstream InputSource = "upstream" // captured NodeOutput of a prior node
    InputFromUser     InputSource = "user"     // user-provided text, selection, approval
    InputFromEnv      InputSource = "env"      // repo context (branch, recent commits, HEAD sha)
    InputFromRegistry InputSource = "registry" // auxiliary lookup (methods.csv, techniques.csv)
)

// InputShape describes the widget shape for user-provided inputs.
// Unused when Source != InputFromUser.
type InputShape string

const (
    ShapeFree        InputShape = "free"        // freeform text
    ShapeChoice      InputShape = "choice"      // one-of Options
    ShapeMultiChoice InputShape = "multi"       // subset of Options
    ShapeApproval    InputShape = "approval"    // yes/no
    ShapeFile        InputShape = "file"        // file path (absolute or repo-relative)
    ShapeJSON        InputShape = "json"        // structured payload, validated against schema
)

// InputSpec declares a single input to a process node.
type InputSpec struct {
    ID             string      `json:"id"`                       // stable within a ProcessDef, e.g. "method", "topic", "approval"
    Source         InputSource `json:"source"`
    Shape          InputShape  `json:"shape,omitempty"`          // required iff Source == InputFromUser
    Required       bool        `json:"required"`
    ArtifactName   string      `json:"artifactName,omitempty"`   // Source=file; key into artifactPaths
    UpstreamNodeID string      `json:"upstreamNodeId,omitempty"` // Source=upstream; empty = first direct predecessor
    Prompt         string      `json:"prompt,omitempty"`         // user-facing prompt for Source=user
    Options        []string    `json:"options,omitempty"`        // ShapeChoice / ShapeMultiChoice
    OptionsRef     string      `json:"optionsRef,omitempty"`     // dynamic options lookup, e.g. "registry:methods.csv#method_name"
    Default        string      `json:"default,omitempty"`
    Validation     string      `json:"validation,omitempty"`     // regex; Shape=free, json, file
    MaxLength      int         `json:"maxLength,omitempty"`      // Shape=free
    HelpText       string      `json:"helpText,omitempty"`       // tooltip / subtitle in frontend
}
```

Rationale for `OptionsRef`: `bmad-advanced-elicitation` shuffles 5 methods from a 50-row CSV each round. Hard-coding `Options` in the registry would force a registry edit per reshuffle; `OptionsRef` lets the executor lookup live at suspension time.

### 3.2 Output declarations

```go
// OutputTarget classifies where a process output lives.
type OutputTarget string

const (
    OutputToFile   OutputTarget = "file"   // written to artifactPaths[ArtifactName]
    OutputToMemory OutputTarget = "memory" // lives only in NodeOutputs
    OutputToBoth   OutputTarget = "both"   // both
)

// OutputSpec declares a single output of a process node.
type OutputSpec struct {
    ID           string       `json:"id"`
    Target       OutputTarget `json:"target"`
    ArtifactName string       `json:"artifactName,omitempty"`   // required when Target ∈ {file, both}
    Description  string       `json:"description,omitempty"`
    Optional     bool         `json:"optional,omitempty"`       // if true, missing output does not fail the node
}
```

### 3.3 Interaction mode and iteration gate

```go
// InteractionMode classifies how a process engages the user.
type InteractionMode string

const (
    InteractAutonomous InteractionMode = "autonomous" // no pauses; current default behaviour
    InteractGuided     InteractionMode = "guided"     // sequential prompts, single pass
    InteractIterative  InteractionMode = "iterative"  // rounds until gate satisfied
    InteractParty      InteractionMode = "party"      // perpetual conversation with multi-agent rotation
)

// GateKind classifies how an iterative process terminates.
type GateKind string

const (
    GateUserConfirm    GateKind = "userConfirm"  // user sends one of AcceptTokens
    GateArtifactExists GateKind = "artifact"     // OutputSpec[0].ArtifactName file exists on disk
    GateExpression     GateKind = "expression"   // CustomExpr evaluates true over NodeOutputs
    GateRoundLimit     GateKind = "rounds"       // hard cap, no user confirm
)

// IterationGate declares termination criteria.
type IterationGate struct {
    Kind         GateKind `json:"kind"`
    MaxRounds    int      `json:"maxRounds,omitempty"`    // hard safety cap for all kinds; 0 = no cap
    AcceptTokens []string `json:"acceptTokens,omitempty"` // GateUserConfirm; e.g. ["x", "proceed", "done"]
    RejectTokens []string `json:"rejectTokens,omitempty"` // emits :aborted instead of :gate_satisfied
    CustomExpr   string   `json:"customExpr,omitempty"`   // GateExpression; same syntax as Condition
}
```

Rationale for `MaxRounds` on every kind: every gate needs a safety ceiling. A user might walk away from a `GateUserConfirm` loop indefinitely; a `GateExpression` might never evaluate true due to a bug. 0 means "no cap", but registry entries SHOULD set one (30 for brainstorm, 100 for party, 10 for elicitation).

### 3.4 Extended `ProcessDef`

```go
type ProcessDef struct {
    // existing fields (unchanged)
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Phase       BmadPhase     `json:"phase"`
    AgentRole   BmadAgentRole `json:"agentRole"`
    SkillName   string        `json:"skillName"`
    Description string        `json:"description"`
    Inputs      []string      `json:"inputs"`   // legacy; kept for back-compat with registry.go
    Outputs     []string      `json:"outputs"`  // legacy
    ModuleID    string        `json:"moduleId"`
    Version     string        `json:"version"`

    // NEW: interactive protocol
    Mode        InteractionMode `json:"mode,omitempty"`
    InputSpecs  []InputSpec     `json:"inputSpecs,omitempty"`
    OutputSpecs []OutputSpec    `json:"outputSpecs,omitempty"`
    Gate        *IterationGate  `json:"gate,omitempty"`
}
```

When `Mode == ""` or `InteractAutonomous` and `InputSpecs == nil`, the executor falls through to the existing `executeNode` path — byte-for-byte compatible.

### 3.5 Extended node status and execution

```go
const (
    NodePending       WorkflowNodeStatus = "pending"
    NodeRunning       WorkflowNodeStatus = "running"
    NodeAwaitingInput WorkflowNodeStatus = "awaiting_input" // NEW
    NodeComplete      WorkflowNodeStatus = "complete"
    NodeFailed        WorkflowNodeStatus = "failed"
    NodeSkipped       WorkflowNodeStatus = "skipped"
)

// PendingPrompt tracks an open user input request.
type PendingPrompt struct {
    NodeID    string     `json:"nodeId"`
    InputID   string     `json:"inputId"`   // matches InputSpec.ID
    Prompt    string     `json:"prompt"`    // rendered from InputSpec.Prompt + context substitutions
    Shape     InputShape `json:"shape"`
    Options   []string   `json:"options,omitempty"` // resolved at suspension time (may differ from spec if OptionsRef)
    Round     int        `json:"round"`     // 1-indexed; always 1 for guided mode
    CreatedAt int64      `json:"createdAt"`
    PromptID  string     `json:"promptId"`  // stable hash(NodeID|InputID|Round); dedupe key for snackbar
}

// WorkflowExecution gains:
type WorkflowExecution struct {
    // existing fields (unchanged)
    ID          string             `json:"id"`
    WorkflowID  string             `json:"workflowId"`
    RepoPath    string             `json:"repoPath"`
    Status      WorkflowExecStatus `json:"status"`
    Nodes       []WorkflowNode     `json:"nodes"`
    StartedAt   string             `json:"startedAt"`
    CurrentNode string             `json:"currentNode"`
    NodeOutputs map[string]string  `json:"nodeOutputs,omitempty"`

    // NEW: interactive state
    NodeRounds     map[string]int                `json:"nodeRounds,omitempty"`     // rounds completed per node
    PendingPrompts []PendingPrompt               `json:"pendingPrompts,omitempty"` // at most one per (nodeID,inputID)
    NodeInputs     map[string]map[string]string  `json:"nodeInputs,omitempty"`     // nodeID → inputID → last-submitted value
    NodeInputHistory map[string][]NodeInputEntry `json:"nodeInputHistory,omitempty"` // audit trail
}

// NodeInputEntry records one user response for the audit trail.
type NodeInputEntry struct {
    InputID   string `json:"inputId"`
    Round     int    `json:"round"`
    Value     string `json:"value"`
    Timestamp int64  `json:"timestamp"`
}
```

`NodeInputHistory` is orthogonal to `NodeInputs` — the latter holds only the most recent answer (for context building), the former preserves every response for post-hoc review. A party-mode conversation of 50 messages lives entirely in `NodeInputHistory["party-node-1"]`.

---

## 4. Input Resolution Algorithm

`internal/bmad/executor.go` gains a helper that runs at the start of every interactive node and at every iteration round.

```go
type resolvedInputs map[string]string

func (e *Executor) resolveInputs(
    ctx context.Context,
    state *execState,
    nodeID string,
    round int,
) (resolved resolvedInputs, missing []InputSpec, err error) {
    proc, ok := processFor(state, nodeID)
    if !ok { return nil, nil, ErrProcessNotFound }

    resolved = resolvedInputs{}
    for _, spec := range proc.InputSpecs {
        switch spec.Source {
        case InputFromFile:
            path := ResolveArtifactPath(spec.ArtifactName, state.repoPath)
            if path == "" {
                if spec.Required { return nil, nil, fmt.Errorf("unmapped artifact %q", spec.ArtifactName) }
                continue
            }
            data, readErr := os.ReadFile(path)
            if readErr != nil {
                if spec.Required { return nil, nil, fmt.Errorf("artifact %s: %w", spec.ArtifactName, readErr) }
                continue
            }
            resolved[spec.ID] = string(data)

        case InputFromUpstream:
            srcID := spec.UpstreamNodeID
            if srcID == "" {
                srcID = firstDirectPredecessor(state, nodeID)
            }
            if v, ok := state.exec.NodeOutputs[srcID]; ok {
                resolved[spec.ID] = truncate(v, upstreamOutputCap)
            } else if spec.Required {
                return nil, nil, fmt.Errorf("upstream %s produced no output", srcID)
            }

        case InputFromUser:
            if v, ok := state.exec.NodeInputs[nodeID][spec.ID]; ok && v != "" {
                resolved[spec.ID] = v
                continue
            }
            if spec.Required {
                missing = append(missing, spec)
                continue
            }
            if spec.Default != "" {
                resolved[spec.ID] = spec.Default
            }

        case InputFromEnv:
            resolved[spec.ID] = envValue(state, spec.ID)

        case InputFromRegistry:
            v, lookupErr := registryLookup(spec.OptionsRef)
            if lookupErr != nil {
                if spec.Required { return nil, nil, lookupErr }
                continue
            }
            resolved[spec.ID] = v
        }
    }
    return resolved, missing, nil
}
```

Key properties:

- **Deterministic order.** Specs resolve in declaration order. User-spec missing means the node suspends immediately, *before* starting the tmux session — no claude tokens spent on a node that cannot run.
- **Idempotent.** Calling `resolveInputs` twice in a row yields the same result unless new user answers arrived. Iteration rounds use this.
- **Bounded memory.** Upstream outputs capped at `upstreamOutputCap` (2000 bytes, matching current `buildContextStringV3` behaviour). File inputs uncapped — the OS buffer is the limit. Frontend refuses to render file inputs > 1 MiB in the modal (render summary + "view full" button).
- **Registry extensibility.** `InputFromRegistry` + `OptionsRef` allows dynamic option sets without code changes. Supported refs: `registry:methods.csv#method_name` (column extract), `registry:techniques.csv?random=5` (5 random rows). Reserved for future: `mcp:server#tool`.

---

## 5. Executor Integration

### 5.1 Routing

`runDynamic`'s ready-set dispatch (`internal/bmad/executor.go:25481`) gains a branch on `Mode`:

```go
// Inside runDynamic's ready loop, per ready node:
proc, _ := ProcessByID(state.nodes[idx].ProcessID)
switch proc.Mode {
case InteractGuided, InteractIterative, InteractParty:
    go e.executeInteractiveNode(ctx, state, nodeIndex, nodeID, repoPath, model)
default:
    go e.executeNode(ctx, state, nodeIndex, nodeID, repoPath, model) // existing path
}
```

`executeControlNode` (condition/merge) and `executeLoopNode` (loop/loopUntil) are unchanged — they never become interactive in this spec.

### 5.2 `executeInteractiveNode` lifecycle

```go
func (e *Executor) executeInteractiveNode(
    ctx context.Context,
    state *execState,
    nodeIndex map[string]int,
    nodeID, repoPath, model string,
) {
    idx := nodeIndex[nodeID]
    proc, _ := ProcessByID(state.nodes[idx].ProcessID)

    // A. Mark running.
    e.setStatus(state, idx, NodeRunning)

    // B. Resolve declared inputs. Suspend for each missing user input.
    resolved, err := e.resolveWithSuspension(ctx, state, nodeID, 1, proc.InputSpecs)
    if err != nil {
        e.failNode(state, idx, nodeID); return
    }

    // C. Start tmux session with rendered context.
    prompt := buildInteractivePrompt(proc, resolved)
    if err := e.startSession(state, nodeID, prompt, model); err != nil {
        e.failNode(state, idx, nodeID); return
    }

    // D. Iteration loop (skipped for guided/party with single-round gate).
    round := 1
    for {
        // D1. Wait for claude's turn to settle (pane hash stable + idle prompt).
        if err := e.waitRoundStable(ctx, state, nodeID); err != nil {
            e.failNode(state, idx, nodeID); return
        }

        // D2. Capture round output to NodeOutputs.
        roundKey := fmt.Sprintf("%s-round-%d", nodeID, round)
        e.captureRoundOutput(state, nodeID, roundKey)
        state.exec.NodeRounds[nodeID] = round

        // D3. Gate check.
        if hit, reason := e.checkGate(state, proc.Gate, nodeID, round); hit {
            e.emit(EventGateSatisfied, gateSatisfiedPayload(nodeID, round, reason))
            break
        }
        if proc.Gate != nil && proc.Gate.MaxRounds > 0 && round >= proc.Gate.MaxRounds {
            e.emit(EventRoundLimit, roundLimitPayload(nodeID, round))
            break
        }

        // D4. Ask for next round's user input.
        nextSpec, ok := proc.iterationInput()
        if !ok { break } // no iteration input configured ⇒ single-round process
        if err := e.suspendForSpec(ctx, state, nodeID, round+1, nextSpec); err != nil {
            e.failNode(state, idx, nodeID); return
        }
        answer := state.exec.NodeInputs[nodeID][nextSpec.ID]

        // D5. Inject answer into the tmux pane.
        if err := e.sendToSession(ctx, state, nodeID, answer); err != nil {
            e.failNode(state, idx, nodeID); return
        }
        round++
    }

    // E. Verify declared outputs; fail-soft on optional.
    if err := e.verifyOutputs(state, nodeID, proc.OutputSpecs, repoPath); err != nil {
        e.failNode(state, idx, nodeID); return
    }

    // F. Complete. activeOutEdges runs → downstream in-degree decrements → ready-set advances.
    e.completeNode(state, idx, nodeID)
}
```

`iterationInput()` returns the `InputSpec` flagged as the per-round prompt. Convention: exactly one `InputSpec` in each `InteractIterative` / `InteractParty` process carries `Prompt != ""` and `Shape != ""` and represents the recurring user input. Pre-process inputs (the initial topic, method list, etc.) are resolved in step B and not repeated.

### 5.3 Suspension primitive

```go
func (e *Executor) suspendForSpec(
    ctx context.Context,
    state *execState,
    nodeID string,
    round int,
    spec InputSpec,
) error {
    prompt := PendingPrompt{
        NodeID:    nodeID,
        InputID:   spec.ID,
        Prompt:    renderPrompt(spec, state, nodeID),
        Shape:     spec.Shape,
        Options:   resolveOptions(spec, state),
        Round:     round,
        CreatedAt: time.Now().Unix(),
        PromptID:  hashPendingPrompt(nodeID, spec.ID, round),
    }

    state.mu.Lock()
    idx := state.nodeIndex[nodeID]
    state.nodes[idx].Status = NodeAwaitingInput
    state.exec.PendingPrompts = upsertPrompt(state.exec.PendingPrompts, prompt)
    state.mu.Unlock()

    e.persistSnapshot(state)
    e.emit(EventAwaitingInput, awaitingPayload(prompt))

    // Block on per-prompt channel, released by RespondToInput.
    waitCh := state.waiter(nodeID, spec.ID)
    select {
    case <-waitCh:
        // value stored, clear pending, restore status
    case <-ctx.Done():
        return ctx.Err()
    }

    state.mu.Lock()
    state.nodes[idx].Status = NodeRunning
    state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, nodeID, spec.ID)
    state.mu.Unlock()

    e.persistSnapshot(state)
    e.emit(EventInputResolved, inputResolvedPayload(nodeID, spec.ID, round))
    return nil
}
```

### 5.4 Ready-set invariants (unchanged, asserted)

`runDynamic`'s in-degree logic (`executor.go:25506-25511`) already respects suspended nodes transitively. The new status simply makes inspection easy.

**Invariant proof sketch:**

- Claim: A node N whose upstream U is `NodeAwaitingInput` remains `NodePending`.
- Proof: `activeOutEdges` runs only in `completeNode` → `runDynamic` processes its result. `NodeAwaitingInput` never calls `completeNode`. Therefore U's out-edges never fire their decrement. N's in-degree never reaches 0. N stays in `NodePending`. ∎

An assertion is added to `activeOutEdges` for defensive diagnostics:

```go
func (e *Executor) activeOutEdges(state *execState, nodeID, result string, effType NodeType) []WorkflowEdge {
    status := state.nodes[state.nodeIndex[nodeID]].Status
    if status != NodeComplete {
        // This is a programmer error: activeOutEdges should only be called for completed nodes.
        log.Printf("bmad: activeOutEdges called for node %s in status %s", nodeID, status)
    }
    // ... existing logic ...
}
```

---

## 6. Event Contract

Events fire on the Wails event bus, mirroring the existing `bmad:node:question` / `bmad:node:idle` pair.

| Event | Payload | When |
|---|---|---|
| `bmad:node:awaiting_input` | `PendingPrompt` | Node enters `NodeAwaitingInput` |
| `bmad:node:input_resolved` | `{execId, nodeId, inputId, round, valueHash}` | User answer accepted (hash, not raw value, to avoid logging PII) |
| `bmad:node:input_invalid` | `{execId, nodeId, inputId, reason}` | Validation failed; node stays awaiting |
| `bmad:node:round_complete` | `{execId, nodeId, round, outputKey}` | One iteration round finished (post-capture, pre-gate) |
| `bmad:node:gate_satisfied` | `{execId, nodeId, round, reason}` | Gate matched; loop exits |
| `bmad:node:round_limit` | `{execId, nodeId, round}` | `MaxRounds` reached without gate |
| `bmad:node:aborted` | `{execId, nodeId, reason}` | User sent `RejectToken` or called `StopBmadWorkflow` |

Legacy events kept:

- `bmad:node:question` / `bmad:node:question:dismissed` — still fire for autonomous nodes where claude CLI surprises us with its own question.
- `bmad:node:idle` / `bmad:node:idle:dismissed` — still fire for autonomous nodes that sit at the `❯` prompt.

These never fire for interactive nodes: the declared flow owns all pausing.

---

## 7. Persistence and Resume

### 7.1 Snapshot triggers

Snapshot is written to `~/.mashed/workflows/{id}/execution.json` on every:

1. Node status transition (running → awaiting → running → complete).
2. `PendingPrompts` mutation (add or remove).
3. `NodeInputs` update (new user answer accepted).
4. Round increment.

The existing snapshot cadence (which I haven't traced in detail) is extended — it already covers node transitions, so new writes are on prompt/input changes.

### 7.2 Restore algorithm

`GetBmadCurrentExecution(repoPath)` (`app_bmad.go`) already re-emits restore events. Extend it:

```go
func (a *App) GetBmadCurrentExecution(repoPath string) (*bmad.WorkflowExecution, error) {
    exec, err := a.bmadExecutor.CurrentExecution(repoPath)
    if err != nil || exec == nil { return exec, err }

    // Re-emit awaiting events for any pending prompts.
    for _, p := range exec.PendingPrompts {
        a.emit("bmad:node:awaiting_input", awaitingPayload(p))
    }
    return exec, nil
}
```

### 7.3 Resurrecting a dead tmux pane

On restore, some tmux panes may be dead (user rebooted, tmux server killed, etc.). The executor's existing behaviour is to treat a dead pane as a node failure. Interactive nodes need recovery:

```go
func (e *Executor) resumeInteractiveNode(ctx context.Context, state *execState, nodeID string) error {
    idx := state.nodeIndex[nodeID]
    node := state.nodes[idx]

    // If pane is alive, nothing to do — the existing goroutine resumed from its select.
    if e.paneAlive(node.TmuxTarget) { return nil }

    // Dead pane. Reconstruct context from accumulated history.
    proc, _ := ProcessByID(node.ProcessID)
    recap := renderRecap(proc, state.exec.NodeInputHistory[nodeID])
    resolved, _ := e.resolveInputs(ctx, state, nodeID, state.exec.NodeRounds[nodeID]+1)
    prompt := buildInteractivePrompt(proc, resolved) + "\n\n## Previous session recap\n" + recap

    if err := e.startSession(state, nodeID, prompt, state.model); err != nil {
        return err
    }
    // Resume from round N+1 where N = last round in NodeInputHistory.
    return nil
}
```

`renderRecap` formats the history as a markdown Q&A block. Claude reads it and continues; users see the snackbar re-appear with the next-round prompt already populated.

### 7.4 Edge cases

- **User edits sprint-status.yaml mid-awaiting.** Irrelevant; interactive nodes do not trigger story status advancement until `completeNode`.
- **Workflow paused while awaiting.** `PauseBmadWorkflow` flips exec status to `ExecPaused`. Interactive nodes hold their channel — when resumed, the channel is still waiting. No special handling needed.
- **Workflow stopped while awaiting.** `StopBmadWorkflow` cancels the exec context. `suspendForSpec`'s `ctx.Done` branch fires → returns error → node fails → snackbar dismissed via emit of `bmad:node:aborted`.

---

## 8. Response API

### 8.1 New binding

```go
// internal/bmad/executor.go
func (e *Executor) RespondToInput(execID, nodeID, inputID, value string) error {
    state, err := e.getState(execID)
    if err != nil { return err }

    // 1. Find matching PendingPrompt.
    state.mu.RLock()
    prompt, ok := findPendingPrompt(state.exec.PendingPrompts, nodeID, inputID)
    state.mu.RUnlock()
    if !ok { return fmt.Errorf("no pending prompt for %s/%s", nodeID, inputID) }

    // 2. Validate against InputSpec.
    proc, _ := ProcessByID(state.nodes[state.nodeIndex[nodeID]].ProcessID)
    spec, ok := findInputSpec(proc, inputID)
    if !ok { return fmt.Errorf("unknown input %s", inputID) }
    if err := validateInput(spec, value); err != nil {
        e.emit(EventInputInvalid, invalidPayload(nodeID, inputID, err.Error()))
        return err
    }

    // 3. Store in NodeInputs + append to NodeInputHistory.
    state.mu.Lock()
    if state.exec.NodeInputs == nil { state.exec.NodeInputs = map[string]map[string]string{} }
    if state.exec.NodeInputs[nodeID] == nil { state.exec.NodeInputs[nodeID] = map[string]string{} }
    state.exec.NodeInputs[nodeID][inputID] = value
    state.exec.NodeInputHistory[nodeID] = append(state.exec.NodeInputHistory[nodeID], NodeInputEntry{
        InputID: inputID, Round: prompt.Round, Value: value, Timestamp: time.Now().Unix(),
    })
    state.mu.Unlock()

    e.persistSnapshot(state)

    // 4. Signal the waiting goroutine.
    state.releaseWaiter(nodeID, inputID)
    return nil
}
```

### 8.2 Wails binding

```go
// app_bmad.go
func (a *App) RespondToInput(execID, nodeID, inputID, value string) error {
    if a.bmadExecutor == nil { return bmad.ErrExecNotInitialized }
    return a.bmadExecutor.RespondToInput(execID, nodeID, inputID, value)
}
```

### 8.3 Backward compat shim

Existing `RespondToQuestion(execID, nodeID, answer)` is retained:

```go
// Legacy: treats inputID as "" and paste+Enter into tmux like before.
// Used by autonomous nodes where claude CLI asked an unexpected question.
func (a *App) RespondToQuestion(execID, nodeID, answer string) error {
    if a.bmadExecutor == nil { return bmad.ErrExecNotInitialized }
    return a.bmadExecutor.RespondToQuestionLegacy(execID, nodeID, answer)
}
```

The frontend chooses based on current status: `NodeAwaitingInput` → `RespondToInput` (typed); `NodeRunning` with `bmad:node:question` fired → `RespondToQuestion` (legacy).

### 8.4 Validation rules

```go
func validateInput(spec InputSpec, value string) error {
    if spec.Required && value == "" { return errors.New("value is required") }
    switch spec.Shape {
    case ShapeChoice:
        if !contains(spec.Options, value) { return fmt.Errorf("value must be one of %v", spec.Options) }
    case ShapeMultiChoice:
        for _, v := range strings.Split(value, ",") {
            if !contains(spec.Options, strings.TrimSpace(v)) { return fmt.Errorf("%q not in options", v) }
        }
    case ShapeApproval:
        if value != "yes" && value != "no" { return errors.New("approval must be yes or no") }
    case ShapeFree, ShapeJSON, ShapeFile:
        if spec.Validation != "" {
            re, err := regexp.Compile(spec.Validation)
            if err != nil { return fmt.Errorf("invalid validation regex: %w", err) }
            if !re.MatchString(value) { return errors.New("value did not match validation regex") }
        }
        if spec.MaxLength > 0 && len(value) > spec.MaxLength {
            return fmt.Errorf("value exceeds max length %d", spec.MaxLength)
        }
    }
    return nil
}
```

---

## 9. Frontend Contract

### 9.1 Status display

`frontend/src/components/bmad/ProcessNode.svelte`:

- `awaiting_input` badge: amber, pulsing, icon `message-circle-question`.
- Round counter for iterative nodes: "3/30 rounds" rendered as a small subtitle.
- Click-through opens the input modal.

### 9.2 Input modal

Replace `QuestionResponseModal.svelte` with `InputResponseModal.svelte` (new component). The modal renders one of five widget shapes based on `PendingPrompt.Shape`:

| Shape | Widget | Submit condition |
|---|---|---|
| `free` | `<textarea>` | Submit button or Cmd+Enter |
| `choice` | `<div role="radiogroup">` with clickable options | Enter after selection |
| `multi` | `<div>` of checkboxes | Submit button when ≥ 1 checked |
| `approval` | Two buttons: "Yes" / "No" | Click |
| `file` | Drop zone + manual path input | Enter |
| `json` | Monaco editor mini-instance in JSON mode | Submit validates via `JSON.parse` before send |

Validation errors from `bmad:node:input_invalid` render inline with shake animation; the modal stays open.

### 9.3 Snackbar

`QuestionSnackbar.svelte` is renamed `NodeInputSnackbar.svelte` and dispatches the same event. It shows:

- Round indicator for iterative nodes.
- One-line preview of the prompt.
- "Respond" button → opens modal.
- "Skip" button when `spec.Required == false`.

### 9.4 Event subscriptions

`WorkflowBuilder.svelte` subscribes to:

- `bmad:node:awaiting_input` → shows snackbar for the node.
- `bmad:node:input_resolved` → dismisses snackbar, re-opens if another `PendingPrompt` is queued.
- `bmad:node:round_complete` → updates the round counter on the node card.
- `bmad:node:gate_satisfied` / `bmad:node:round_limit` → flashes the node card transitionally from `awaiting_input` → `complete`.

Legacy `bmad:node:question` subscription stays for autonomous fallback path.

---

## 10. Reference Process Specs

Four `ProcessDef` entries in `internal/bmad/registry.go` gain Mode + InputSpecs + Gate. Presented as JSON snippets for readability.

### 10.1 `bmad-brainstorming`

```json
{
  "id": "bmad-brainstorming",
  "name": "Brainstorming",
  "phase": "analysis",
  "skillName": "bmad-brainstorming",
  "mode": "iterative",
  "inputSpecs": [
    {
      "id": "topic",
      "source": "user",
      "shape": "free",
      "required": true,
      "prompt": "What topic do you want to brainstorm?",
      "maxLength": 500
    },
    {
      "id": "approach",
      "source": "user",
      "shape": "choice",
      "required": true,
      "prompt": "How should we pick techniques?",
      "options": ["user-pick", "ai-recommend", "random", "progressive"]
    },
    {
      "id": "technique",
      "source": "registry",
      "optionsRef": "registry:brain-methods.csv#technique_name",
      "required": false
    },
    {
      "id": "round-response",
      "source": "user",
      "shape": "free",
      "required": false,
      "prompt": "Add ideas, pivot, or type 'done' when satisfied.",
      "helpText": "Type 'done' to wrap up; 'skip' to move to the next technique."
    }
  ],
  "outputSpecs": [
    {
      "id": "brainstorm-notes",
      "target": "file",
      "artifactName": "brainstorm-notes",
      "description": "Organised brainstorm session notes"
    }
  ],
  "gate": {
    "kind": "userConfirm",
    "maxRounds": 30,
    "acceptTokens": ["done", "wrap up", "finish"],
    "rejectTokens": ["abort", "cancel"]
  }
}
```

`iterationInput()` returns `round-response` (the only `Shape != ""` spec with no initial value expected).

### 10.2 `bmad-product-brief`

```json
{
  "id": "bmad-product-brief",
  "name": "Product Brief",
  "phase": "analysis",
  "skillName": "bmad-product-brief",
  "mode": "guided",
  "inputSpecs": [
    {
      "id": "mode",
      "source": "user",
      "shape": "choice",
      "required": true,
      "prompt": "How do you want to work?",
      "options": ["guided", "yolo", "autonomous"],
      "default": "guided"
    },
    {
      "id": "existing-brief",
      "source": "file",
      "artifactName": "product-brief",
      "required": false
    },
    {
      "id": "brainstorm-input",
      "source": "upstream",
      "required": false
    },
    {
      "id": "stage-response",
      "source": "user",
      "shape": "free",
      "required": false,
      "prompt": "{{stage_prompt}}",
      "helpText": "Type 'skip' to move on without more detail."
    },
    {
      "id": "final-approval",
      "source": "user",
      "shape": "approval",
      "required": true,
      "prompt": "Approve this brief?"
    }
  ],
  "outputSpecs": [
    {
      "id": "product-brief",
      "target": "file",
      "artifactName": "product-brief"
    }
  ],
  "gate": {
    "kind": "userConfirm",
    "maxRounds": 10,
    "acceptTokens": ["yes"]
  }
}
```

Guided mode: the process loops through 5 stages, each re-rendering `stage-response` with a different `Prompt`. The claude side is responsible for recognising which stage it's in via the recap; the executor simply keeps sending `stage-response` values until `final-approval == "yes"`.

### 10.3 `bmad-party-mode`

```json
{
  "id": "bmad-party-mode",
  "name": "Party Mode",
  "phase": "utilities",
  "skillName": "bmad-party-mode",
  "mode": "party",
  "inputSpecs": [
    {
      "id": "topic",
      "source": "user",
      "shape": "free",
      "required": true,
      "prompt": "What do you want the team to discuss?",
      "maxLength": 1000
    },
    {
      "id": "message",
      "source": "user",
      "shape": "free",
      "required": false,
      "prompt": "Your turn. Type 'exit' to end the conversation."
    }
  ],
  "outputSpecs": [
    {
      "id": "transcript",
      "target": "file",
      "artifactName": "retro-notes",
      "optional": true
    }
  ],
  "gate": {
    "kind": "userConfirm",
    "maxRounds": 100,
    "acceptTokens": ["exit", "done", "wrap up"]
  }
}
```

Party mode's transcript is optional — casual conversations don't always need a saved artifact.

### 10.4 `bmad-advanced-elicitation`

```json
{
  "id": "bmad-advanced-elicitation",
  "name": "Advanced Elicitation",
  "phase": "support",
  "skillName": "bmad-advanced-elicitation",
  "mode": "iterative",
  "inputSpecs": [
    {
      "id": "target-content",
      "source": "upstream",
      "required": true
    },
    {
      "id": "method",
      "source": "user",
      "shape": "choice",
      "required": true,
      "prompt": "Pick a reasoning method:",
      "optionsRef": "registry:methods.csv?random=5",
      "helpText": "[r] reshuffle · [a] see all · [x] accept and proceed"
    },
    {
      "id": "apply-changes",
      "source": "user",
      "shape": "approval",
      "required": true,
      "prompt": "Apply these changes to the document?"
    }
  ],
  "outputSpecs": [
    {
      "id": "elicitation-notes",
      "target": "both",
      "artifactName": "elicitation-notes"
    }
  ],
  "gate": {
    "kind": "userConfirm",
    "maxRounds": 20,
    "acceptTokens": ["x", "proceed", "done"]
  }
}
```

Each round: `method` (with 5 fresh options) → apply → `apply-changes` (yes/no) → next round. The `x` accept token at the prompt level short-circuits the loop.

---

## 11. Migration Path

### 11.1 Additive changes only

- All new fields on `ProcessDef`, `WorkflowExecution`, `WorkflowNode` default to zero values.
- `Mode == ""` ≡ `InteractAutonomous` ≡ current behaviour.
- No existing registry entry is modified in the first PR.
- `RespondToQuestion` stays, wired to `RespondToQuestionLegacy`.

### 11.2 Staged rollout

| Sprint story | Scope | Verification |
|---|---|---|
| S1: Types | Add new types to `types.go` + JSON round-trip tests | Existing tests pass; new table-driven test in `types_test.go` |
| S2: Executor routing | Add `executeInteractiveNode` + routing switch; no-op for empty InputSpecs | All autonomous templates still run end-to-end |
| S3: Suspension + RespondToInput | Suspension primitive, waiter channels, snapshot hooks | Integration test: node with one user input completes after RespondToInput |
| S4: Iteration gate | Gate evaluator + round loop | Integration test: 3-round iterative process with `MaxRounds=3` auto-exits |
| S5: Persistence | Snapshot extension + `GetBmadCurrentExecution` re-emit | Integration test: kill executor mid-awaiting, restart, snackbar re-appears |
| S6: Frontend modal | `InputResponseModal.svelte` + shape widgets + snackbar rename | Playwright tests per shape (`tests/ac/bmad-input-*.spec.ts`) |
| S7: Registry entries | Populate brainstorm / brief / party / elicitation entries | Smoke tests invoking each via the UI |
| S8: Docs | Update `CLAUDE.md` BMAD section + this doc marked "implemented" | N/A |

### 11.3 Rollback plan

Every story is independently revertable:

- S1-S2: remove new fields and switch; zero behaviour change.
- S3-S5: revert `executeInteractiveNode` path; nodes fall back to autonomous (would break iterative registry entries, but S7 is last).
- S7: revert registry entries; executor stays on autonomous path for affected processes.

---

## 12. Test Strategy

### 12.1 Unit tests

- `types_test.go` — JSON round-trip for InputSpec, OutputSpec, IterationGate, PendingPrompt, NodeInputEntry.
- `executor_resolve_test.go` — `resolveInputs` table-driven: file present, file missing, upstream present, upstream missing, env, registry, user present, user missing.
- `executor_gate_test.go` — `checkGate` across all GateKinds, including `MaxRounds` cap precedence.
- `executor_validate_test.go` — `validateInput` per Shape, regex success/failure, MaxLength boundary.

### 12.2 Integration tests (executor-level, tmux-mocked)

- `executor_interactive_test.go`:
  - Happy path: 1 input → resolved → session starts → 1 round → gate hit → complete.
  - Suspension: missing user input → status flips to `NodeAwaitingInput` → `PendingPrompt` in snapshot.
  - Downstream hold: 3-node chain, middle node `NodeAwaitingInput`, assert last node stays `NodePending` indefinitely.
  - Round cap: `MaxRounds=3` gate never satisfied → node completes with `bmad:node:round_limit` event.
  - Rejection: user sends `RejectToken` → node fails → `bmad:node:aborted`.

### 12.3 Snapshot/resume tests

- `executor_resume_test.go`:
  - Write snapshot with `PendingPrompts` populated → restart executor → `GetBmadCurrentExecution` re-emits awaiting events.
  - Dead pane recovery: mark pane dead, verify recap prompt rendered with accumulated `NodeInputHistory`.

### 12.4 Frontend tests

- Vitest unit tests for `InputResponseModal.svelte` shape widgets.
- Playwright AC tests for end-to-end flow per reference process:
  - `tests/ac/bmad-brainstorm-iterate.spec.ts` — 2 rounds then `done`.
  - `tests/ac/bmad-brief-guided.spec.ts` — all 5 stages → approval.
  - `tests/ac/bmad-elicit-reshuffle.spec.ts` — method list shuffles on `r` without closing the modal.

---

## 13. Edge Cases and Failure Modes

### 13.1 Simultaneous `PendingPrompts` on the same node

**Not permitted.** The executor enforces one pending prompt per `(NodeID, InputID)` by upserting. Within a single node, `suspendForSpec` runs sequentially. The frontend therefore only ever shows one input widget per node at a time.

### 13.2 User answers a stale prompt after executor moved on

`PromptID` is `hash(NodeID | InputID | Round)`. `RespondToInput` rejects answers whose `PromptID` is not current (returns `ErrStalePrompt`). Frontend surfaces as a toast: "That prompt already expired." This happens if the user had the modal open during a round_limit event.

### 13.3 Registry lookup failure mid-session

`OptionsRef` resolution during `suspendForSpec` can fail (CSV file missing, malformed). Behaviour:

- Required spec → node fails with `bmad:node:failed`, reason logged.
- Optional spec → fall back to `Options` or `Default`; if both empty, skip suspension, resolve spec as empty.

### 13.4 User provides a very large `json` payload

`ShapeJSON` with `MaxLength == 0` defaults to 64 KiB. Exceeding → `bmad:node:input_invalid`. Explicit `MaxLength` overrides.

### 13.5 Tmux pane dies during awaiting

Detected by the pane liveness check inside `suspendForSpec` (runs every 5 s). If dead, emit `bmad:node:failed` (not aborted — the user did not reject). On restart, recovery logic in §7.3 reconstructs the session.

### 13.6 User closes snackbar without answering

Snackbar close is cosmetic only — `NodeAwaitingInput` persists. User can re-open from the node card. Design decision: never auto-dismiss awaiting prompts.

### 13.7 Condition/merge node downstream of awaiting node

Unchanged: condition and merge fire only after upstream complete. An awaiting node blocks them just like a running one. No new handling needed.

### 13.8 Loop nodes containing interactive body nodes

`executeLoopNode` runs a mini ready-set. If a body node is interactive and suspends, the mini ready-set blocks the same way the outer one does. No new handling — but this combination is untested in the initial rollout and should be covered in S5 if time permits.

### 13.9 Paused workflow mid-awaiting

`PauseBmadWorkflow` → exec status `ExecPaused`. Interactive nodes block in `select` on their channel; they don't care about exec status. Resume → nothing to do. Frontend shows paused indicator + awaiting indicator; both dismiss on `ResumeBmadWorkflow` + input submission respectively.

### 13.10 Stop while awaiting

`StopBmadWorkflow` cancels the context. `suspendForSpec` returns `context.Canceled`. Node transitions to failed. Emit `bmad:node:aborted` with reason `"workflow stopped"`.

---

## 14. Security Considerations

### 14.1 Input sanitisation

`RespondToInput` values flow into tmux via `tmux send-keys -l`. The existing `escapeTmuxLiteral` (`internal/bmad/question.go`) already strips C0 control bytes, keeping `$` `` ` `` `;` `|` backslash etc. safe because `-l` bypasses shell interpretation. Reuse unchanged.

### 14.2 Path traversal on `ShapeFile`

`ShapeFile` values must be validated against the repo root. Enforce via:

```go
func resolveFileInput(value, repoRoot string) (string, error) {
    abs, err := filepath.Abs(value)
    if err != nil { return "", err }
    if !strings.HasPrefix(abs, repoRoot + string(filepath.Separator)) {
        return "", errors.New("path outside repository root")
    }
    return abs, nil
}
```

### 14.3 PII in events

`bmad:node:input_resolved` emits `valueHash` (SHA-256), never raw value. Snapshot file on disk contains raw `NodeInputs` — acceptable because the file is under `~/.mashed/` (user-owned), same trust boundary as claude CLI session logs.

### 14.4 Injected options

`OptionsRef` points to CSV files inside the repo or bundled resources. Enforce resolver only accepts `registry:*` scheme; reject any path traversal.

---

## 15. Performance Considerations

### 15.1 Snapshot write amplification

Adding snapshot writes on every `PendingPrompts` mutation increases I/O. Mitigation: serialise snapshot writes through an existing debouncer (if present) or add a 100 ms coalescing window. Interactive nodes typically have one prompt every few seconds of human thinking — write amplification is negligible.

### 15.2 Channel bookkeeping

Per-waiter channel lives in `execState.waiters map[string]chan struct{}` keyed by `nodeID + "/" + inputID`. Cleaned up after resolve. O(pending) space.

### 15.3 Pane liveness polling

Existing poll interval (~1-2 s) covers this. No new pollers.

### 15.4 Event bus volume

Four new event types, all low-frequency (human-paced). No throttling needed.

---

## 16. Open Questions

1. **Should `InteractParty` support multi-agent rotation inside a single tmux pane, or spawn one pane per agent?** Current recommendation: single pane; `BMad Master` orchestrates as described in upstream docs. Multi-pane is a future extension.
2. **Should `iterationInput()` support multiple per-round inputs?** Not in v1. If needed, introduce `RoundInputSpecs []InputSpec`. Deferred.
3. **Should autonomous nodes migrate to `InputSpecs` too?** Long-term yes — would unify context resolution. Short-term no — scope creep for this sprint.
4. **What happens when an `OptionsRef` like `?random=5` is called twice in the same round (reshuffle)?** `RespondToInput` with value `"r"` could be a reserved reshuffle token that re-emits `bmad:node:awaiting_input` with new options. Cleaner than treating `r` as an accept/reject token. Decide in S3.
5. **Snapshot format versioning.** Existing `execution.json` has no version field. Add `"version": 2` when new fields appear; readers handle absence as v1.

---

## 17. Summary

Five new mechanisms, each in its own lane:

| Mechanism | Type | Where it lives |
|---|---|---|
| Input schema | `InputSpec[]` | `ProcessDef`, `types.go` |
| Output schema | `OutputSpec[]` | `ProcessDef`, `types.go` |
| Suspension state | `NodeAwaitingInput` + `PendingPrompt` | `WorkflowNodeStatus`, `WorkflowExecution` |
| Iteration gate | `IterationGate` | `ProcessDef.Gate` |
| Response API | `RespondToInput(execID, nodeID, inputID, value)` | `executor.go`, `app_bmad.go` |

Every existing invariant (ready-set correctness, tmux session lifecycle, artifact verification, condition/merge/loop control nodes) is preserved unchanged. The legacy question/idle detection path survives for autonomous nodes where claude CLI still surprises us.

The downstream-hold guarantee is strengthened from "accidental side-effect of in-degree math" to "asserted by node-status invariant": a node cannot be in both `NodeAwaitingInput` and `NodeComplete`; `activeOutEdges` fires only for the latter; therefore downstream in-degrees cannot decrement while upstream is awaiting.

The suspended-detection guarantee moves from heuristic tmux scanning ("did claude print a `?`") to declared state ("executor put the node into `NodeAwaitingInput` because a declared required input is unresolved"). Heuristic detection stays as fallback only for autonomous nodes.

Ready to break into a sprint of 8 stories (S1-S8 above); each is independently testable and revertable.
