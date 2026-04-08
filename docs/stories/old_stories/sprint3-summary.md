# Sprint 3 Summary: BMAD Control Flow Nodes

## Vision

Transform the BMAD workflow builder from a **flat sequential pipeline** into a **programmable orchestration engine** with conditional branching, loops, data extraction, and branch merging. When a workflow executes, the engine reads terminal output from completed Claude sessions, evaluates conditions, decides which branches to take, loops until criteria are met, and passes extracted data downstream -- all driven by a dynamic ready-set algorithm that replaces the static topological-tier approach.

---

## What Changes

### Before (Sprint 2)
- Every canvas node is a Claude Code session spawned in tmux
- Executor uses static topological sort -- fixed tiers, no branching, no cycles
- Executor only knows "done" or "failed" -- never reads terminal output
- No conditional logic, no loops, no data extraction between nodes

### After (Sprint 3)
- **5 new control flow node types**: Condition (if/else), Loop, LoopUntil, Transform, Merge
- **Terminal output capture**: Every completed node's tmux output is captured and stored
- **Condition evaluation engine**: Contains, regex, file-exists, exit code checks against captured output
- **Dynamic ready-set executor**: Replaces tier-based execution, supports branching and loops
- **Data passing**: Transform nodes extract data from output, inject it into downstream context
- **Frontend integration**: New node components, type-aware config panel, sidebar drag-and-drop, edge labels, output viewer

---

## Architecture Changes

### Go Backend

| File | Change | Story |
|------|--------|-------|
| `internal/bmad/types.go` | Add `NodeType` enum, extend `WorkflowNode/Edge/Execution`, add `ControlFlowNodeDef` | 01, 09 |
| `internal/bmad/executor.go` | Output capture, dynamic ready-set algorithm, control node dispatch, loop execution, transform dispatch, context string v2 | 02, 04, 05, 06, 10 |
| `internal/bmad/condition.go` | **NEW** -- condition types, evaluation engine, parser | 03 |
| `internal/bmad/condition_test.go` | **NEW** -- condition evaluation tests | 03 |
| `internal/bmad/types_nodetype_test.go` | **NEW** -- type system tests | 01 |
| `internal/bmad/executor_test.go` | Extended with output capture, branching, loop, transform tests | 02, 04, 05, 06 |
| `app.go` | `GetNodeOutput`, `GetControlFlowNodes` bindings | 09 |

### Svelte Frontend

| File | Change | Story |
|------|--------|-------|
| `frontend/src/components/bmad/ConditionNode.svelte` | **NEW** -- diamond condition node | 07 |
| `frontend/src/components/bmad/LoopNode.svelte` | **NEW** -- loop node with iteration display | 07 |
| `frontend/src/components/bmad/LoopUntilNode.svelte` | **NEW** -- loop-until node | 07 |
| `frontend/src/components/bmad/TransformNode.svelte` | **NEW** -- data transform pill | 07 |
| `frontend/src/components/bmad/MergeNode.svelte` | **NEW** -- branch merge diamond | 07 |
| `frontend/src/components/bmad/OutputViewerModal.svelte` | **NEW** -- terminal output viewer | 10 |
| `frontend/src/views/WorkflowBuilder.svelte` | Register node types, save/load handles, drop handler, output modal state | 08, 10 |
| `frontend/src/components/bmad/ProcessSidebar.svelte` | Control Flow sidebar group | 08 |
| `frontend/src/components/bmad/NodeConfigPanel.svelte` | Type-aware config switching | 08, 10 |
| `frontend/src/components/bmad/DeletableEdge.svelte` | Edge labels for conditional/loop edges | 10 |

---

## Stories at a Glance

| # | Story | Priority | Domain | Size | Depends On | Points |
|---|-------|----------|--------|------|------------|--------|
| 01 | [Node Type System & Edge Handles](sprint3-01-node-type-system.md) | P0 | backend | M | none | 3 |
| 02 | [Terminal Output Capture](sprint3-02-output-capture.md) | P0 | backend | M | 01 | 3 |
| 03 | [Condition Evaluation Engine](sprint3-03-condition-engine.md) | P0 | backend | M | 01 | 3 |
| 04 | [Dynamic Ready-Set Executor](sprint3-04-dynamic-executor.md) | P0 | backend | L | 01, 02, 03 | 5 |
| 05 | [Loop and LoopUntil Execution](sprint3-05-loop-execution.md) | P1 | backend | L | 04 | 5 |
| 06 | [Transform Node & Data Passing V2](sprint3-06-data-transform-merge.md) | P1 | backend | M | 04 | 3 |
| 07 | [Control Flow Svelte Components](sprint3-07-control-flow-components.md) | P1 | frontend | L | 01 | 4 |
| 08 | [Canvas Integration](sprint3-08-canvas-integration.md) | P1 | frontend | L | 07 | 4 |
| 09 | [Wails Bindings](sprint3-09-wails-bindings.md) | P2 | fullstack | S | 01, 02 | 2 |
| 10 | [Execution UX](sprint3-10-execution-ux.md) | P2 | frontend | M | 08, 09 | 3 |

**Total Stories:** 10
**Total Points:** 35
**Ready for Sprint:** All 10 stories (status: ready)

---

## Execution Plan

```
Wave 1 (parallel, no deps):
  Story 01: Node Type System          (backend, M)  ──┐
                                                       │
Wave 2 (parallel, dep on 01):                         │
  Story 02: Output Capture            (backend, M)  ──┤
  Story 03: Condition Engine           (backend, M)  ──┤
  Story 07: Control Flow Components    (frontend, L) ──┤
                                                       │
Wave 3 (parallel, deps on Wave 2):                    │
  Story 04: Dynamic Executor          (backend, L)  ──┤ (needs 01, 02, 03)
  Story 08: Canvas Integration         (frontend, L) ──┤ (needs 07)
  Story 09: Wails Bindings            (fullstack, S) ──┤ (needs 01, 02)
                                                       │
Wave 4 (parallel, deps on Wave 3):                    │
  Story 05: Loop Execution            (backend, L)  ──┤ (needs 04)
  Story 06: Transform & Data V2       (backend, M)  ──┤ (needs 04)
                                                       │
Wave 5 (final):                                        │
  Story 10: Execution UX              (frontend, M) ──┘ (needs 08, 09)
```

### Recommended Sprint Order (sequential for single-engineer)

1. sprint3-01 (foundation -- all others depend on it)
2. sprint3-03 (condition engine -- no executor dependency, unlocks 04)
3. sprint3-02 (output capture -- extends executor, needed by 04)
4. sprint3-04 (dynamic executor -- biggest story, core algorithm change)
5. sprint3-05 (loop execution -- extends the new executor)
6. sprint3-06 (transform + data passing -- extends the new executor)
7. sprint3-07 (frontend components -- can parallel with backend after 01)
8. sprint3-08 (canvas wiring -- needs components)
9. sprint3-09 (bindings -- small, bridges backend to frontend)
10. sprint3-10 (execution UX polish -- final integration)

---

## Key Design Decisions

1. **Dynamic ready-set replaces topoSort tiers** -- The tier-based algorithm cannot handle branching (where some edges are inactive) or loops (where nodes re-execute). The ready-set approach dynamically computes which nodes are ready after each completion, enabling both features. `topoSort()` is kept for cycle detection only.

2. **Output capture is best-effort** -- If `tmux capture-pane` fails, the node still completes. This prevents capture failures from blocking workflow execution while still providing output for conditions/transforms when available.

3. **Conditions stored as JSON in node.Config** -- Rather than adding new typed fields to `WorkflowNode`, conditions are serialized as JSON in the existing `Config map[string]string`. This avoids type system changes beyond what sprint3-01 adds.

4. **Loop body declared in config, not via back-edges** -- The DAG remains acyclic at the graph level. Loop iteration is managed by `resetLoopBody()` which resets body nodes to pending. This preserves `topoSort()` cycle detection and simplifies the graph model.

5. **Transform nodes are synchronous** -- No tmux session needed. They read from `NodeOutputs`, apply extraction, store result, and complete immediately. This makes them fast and testable without mocking tmux.

6. **Edge handles enable branch routing** -- xyflow already tracks `sourceHandle` on edges. By persisting these to the backend, the executor can filter which edges to activate based on condition results.

---

## Risk Register

| Risk | Mitigation | Story |
|------|------------|-------|
| Backward compat: existing workflows break | `omitempty` tags, `EffectiveType()` normalizer, default to `bmadProcess` | 01, 08 |
| Ready-set algorithm introduces regressions | All 13 existing tests must pass unchanged | 04 |
| Infinite loops | Max iterations cap (default 10, max 100) | 05 |
| Output capture OOM from large sessions | 100KB cap with tail preservation | 02 |
| Context string bloat from transform data | 2000-char cap per transform insertion | 06 |
| Invalid regex in conditions crashes executor | Safe compile with fallback to false | 03 |
| Diamond-after-condition deadlock | Skip propagation decrements in-degrees recursively | 04 |
| Concurrent NodeOutputs writes race | Writes under state.mu.Lock(), verified with -race | 02 |
| Nested loops complexity | Not supported in sprint 3; validated at startup | 05 |
