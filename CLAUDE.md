# OpenWolf

@.wolf/OPENWOLF.md

This project uses OpenWolf for context management. Read and follow .wolf/OPENWOLF.md every session. Check .wolf/cerebrum.md before generating code. Check .wolf/anatomy.md before reading files.


# CLAUDE.md

## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.
The skill has specialized workflows that produce better results than ad-hoc answers.

## BMAD Artifact System

- `internal/bmad/artifacts.go` — canonical artifact path map (`artifactPaths`), `ResolveArtifactPath(name, repoPath)`, `VerifyArtifacts(repoPath, outputNames)`
- Artifacts resolve to `{repoPath}/_bmad-output/{category}/{artifact}` via the path map
- Unmapped artifacts (`"code"`, `"tests"`, `"any-doc"`) return `""` from ResolveArtifactPath — handled by callers
- `ArtifactType` and `ArtifactSpec` types in `internal/bmad/types.go`
- Shared constant `bmadOutputDir = "_bmad-output"` in `internal/bmad/sprint.go`

## BMAD Interactive Processes

Shipped in stories `bmad-interactive-01..07`. Full schema: `docs/bmad-interactive-process-schema.md`.

- **Types** — `internal/bmad/types.go`: `InputSpec`, `OutputSpec`, `IterationGate`, `PendingPrompt`, `NodeInputEntry`; `ProcessDef.{Mode, InputSpecs, OutputSpecs, Gate}`; `WorkflowExecution.{NodeRounds, PendingPrompts, NodeInputs, NodeInputHistory, Version}`.
- **Suspension state** — `NodeAwaitingInput` is the single source of truth for "paused". `activeOutEdges` runs only for `NodeComplete`, so downstream in-degree cannot decrement while upstream is awaiting (§5.4 invariant).
- **Events** (`internal/bmad/events.go`) — `bmad:node:awaiting_input`, `:input_resolved`, `:input_invalid`, `:round_complete`, `:gate_satisfied`, `:round_limit`, `:aborted`. Legacy `:question` / `:idle` remain as autonomous-node fallback.
- **Wails binding** — `(*App).RespondToInput(execID, nodeID, inputID, value)` (`app_bmad.go`). Legacy `RespondToQuestion` forwards to `RespondToQuestionLegacy` for autonomous nodes.
- **Registry lookups** — `registry:<csv>#<column>` and `registry:<csv>?random=N` resolved from embedded `internal/bmad/testdata/*.csv` via `registryLookup`; non-`registry:` schemes rejected.
- **Persistence** — `persistSnapshot` writes `~/.mashed/workflows/{execID}/execution.json` (serialized by `state.snapshotMu`, atomic rename with per-call unique tempfile); `Version` tagged 2 when interactive fields populated.
- **Security** — `ShapeFile` values rejected outside repo root via `Clean+Separator` prefix guard (§14.2); `:input_resolved` carries `valueHash` (SHA-256 prefix), never raw (§14.3); `OptionsRef` resolver accepts only `registry:*` scheme (§14.4).

