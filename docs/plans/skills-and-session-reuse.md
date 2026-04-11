# Skills, Commands, and workflow-chain execution

**Status**: proposed (revised after schema discussion)
**Owner**: this PR
**Last updated**: 2026-04-11

> **Revision note** — the original draft proposed one primitive ("skill node")
> that chains inside a workflow. After discussion, the right split is:
>
> - **Skills** are long-running, pin to a tmux session, are owned by an
>   **agent**, and are NOT directly chainable in a workflow. Think
>   `skill-creator`, `agent-validator`, `claude-md-improver`.
> - **Commands** are short, invocable, chainable, and **are** the thing that
>   becomes a node in a BMAD workflow chain. Think `/simplify`, `/commit`,
>   `/create-pr`, `/dev-story`.
>
> Both show up in the BMAD workspace, but only commands become draggable
> workflow nodes. A lightweight frontmatter schema distinguishes the two
> and marks each asset as "mashed-ready" or not.

## Goal

1. **Phase 1 — Listing**: The Skills tab in the BMAD workspace sidebar shows
   every **mashed-ready** skill AND command found under the repo's
   `.claude/skills/`, `.claude/commands/`, `~/.claude/skills/`, and
   `~/.claude/commands/` directories, grouped by {skill vs command} ×
   {local vs global}. **Assets whose frontmatter lacks a `mashedRole` are
   silently skipped** — the sidebar stays focused on things that actually
   work in the workspace, and the user's regular Claude Code assets stay
   undisturbed in the filesystem.
1b. **Phase 1b — `mashed-refactor-asset` skill** (the on-ramp): a Claude Code
   skill that takes an existing `SKILL.md` or `commands/<name>.md`, infers
   the right `mashed*` frontmatter fields from the body, and writes them
   back. Without this, the Skills tab is empty on day one and users would
   have to hand-annotate every file they wanted to use. This skill ships
   alongside Phase 1 so users can point it at their existing assets and
   populate the workspace immediately.
2. **Phase 2 — Command nodes**: Commands (not skills) become first-class
   canvas nodes. You drag a mashed-ready command from the sidebar onto the
   canvas, wire it into an edge chain, and the executor runs it as part of
   the workflow. Skills stay out of the canvas — they live with the agent
   that owns them.
3. **Phase 3 — Session reuse + idle completion**: When a command node is
   **not** the first node in a chain, it **reuses the tmux session its
   upstream parent already owns**. When it **is** the first node, it
   spawns a fresh session like today's process nodes. Completion signal
   for every node type shifts from "claude exited" to "output stable at
   idle prompt" (with a state-machine guard against false positives). At
   each chain-tail on workflow terminal state, the executor explicitly
   kills the tmux session.
4. **Phase 4 — Editor + filesystem watch + richer frontmatter**: quality-of-
   life polish on top of the working primitive.

## Why the current executor can't do this today

`internal/bmad/executor.go:executeNode` (lines 881–end) is the node runner. Its
completion detection is:

```go
out, _ := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
if err != nil { complete }
if strings.TrimSpace(string(out)) == "1" { complete }
```

Both signals mean **"the claude process exited"**. But the node's shell command is:

```go
claude --dangerously-skip-permissions --model <model> "use <skill><context>"
```

Real BMAD skills like `Brainstorming` and `Party Mode` are *interactive* — they
ask the user questions and wait at the `❯` prompt. Claude stays alive forever
in those cases and the node never completes. This is a latent bug that
session-reuse forces us to fix: the skill in a reused session can't rely on
pane death either, because the whole point is to keep the session alive for
the next node.

## The unified fix

Replace "pane death" with "**idle-prompt reached after output stabilised**"
as the primary completion signal. We already built 90% of this:

- `detectIdlePrompt` (question.go) — robustly matches the `❯` cursor line,
  ASCII or NBSP, with ANSI stripped. Real-fixture test pins it.
- `pollForIdle` (executor.go) — output-hash state machine that requires two
  consecutive stable polls before firing, so live claude turns don't
  false-positive.
- `EventIdle` / `EventIdleDismissed` — already flow through the bridge.

What's missing is wiring those into the `executeNode` completion path. Right
now `pollForIdle` only emits events to the frontend snackbar; it doesn't talk
to the node runner loop.

## Phase-by-phase scope

### Phase 1 — Skills tab listing (small, standalone commit)

**Backend**:
- `internal/bmad/types.go`: `SkillInfo { Name, Path, Description, Source }`
  and `GroupedSkills { LocalSkills []SkillInfo, GlobalSkills []SkillInfo }`.
- `internal/bmad/skills.go` (new file): `LoadSkillsFromDir(dir string) ([]SkillInfo, error)`
  walks `dir`, looks for immediate subdirectories containing a `SKILL.md`,
  parses YAML frontmatter for `description`, falls back to the first non-empty
  body paragraph if frontmatter is missing or malformed. Sort by name.
- `app_bmad.go`: `ListAllSkills(repoPath string) (bmad.GroupedSkills, error)`
  — calls the loader twice (local, global), returns the pair. Nonexistent
  directories are not errors — they're empty groups.
- Unit tests for the loader: empty dir, nonexistent dir, dir with valid skill,
  dir with skill but no SKILL.md, skill with missing frontmatter, skill with
  malformed frontmatter, skill where description wraps across multiple YAML lines.

**Frontend**:
- `WorkflowBuilder.svelte`: fetch `groupedSkills` in `onMount` alongside the
  existing `Promise.all`, pass down as a prop.
- `ProcessSidebar.svelte`: replace the `Skills panel coming soon` placeholder
  with the same collapsible-group UI the Processes tab already uses
  (Analysis / Planning / etc.) — two groups: "Local" and "Global", each with
  a count badge, each row showing `skill name` + truncated description.
  Hover tooltip = full description.

**What this phase ships**: a populated, browsable Skills tab. No canvas
interaction, no execution. Lands independently, reviewable on its own.

### Phase 2 — Skill nodes on the canvas

**Backend**:
- `types.go`: extend `NodeType` with `NodeTypeSkill = "skill"` and add a
  `SkillName` field to `WorkflowNode` so the node knows which skill it
  represents.
- Storage: nothing — `WorkflowDef.Nodes[]` already carries arbitrary per-node
  fields.

**Frontend**:
- New `SkillNode.svelte` Svelte-Flow node type, or extend `ProcessNode.svelte`
  with a variant. I'd lean toward a distinct component so the visual and
  rendering logic stays simple — skill nodes look slightly different
  (different icon, no `in:`/`out:` artifact lines since skills don't declare
  I/O the way process definitions do).
- `WorkflowBuilder.svelte`:
  - Register the new node type in the `nodeTypes` map.
  - Accept `application/bmad-skill` drops from `ProcessSidebar` the same way
    `application/bmad-process` drops work today, but create a `skill`-type
    node instead of a `bmadProcess` one.
  - Teach `loadNodesEdges`, `saveWorkflow`, `snapshotCanvas`, and the restore
    overlay to treat skill nodes symmetrically with process nodes.

**What this phase ships**: you can drag skills onto the canvas, wire them up,
save the workflow, and re-load it with the skill nodes intact. The workflow
can't yet **run** a skill node — that's Phase 3.

### Phase 3 — Executor session-reuse + idle completion

This is the architecturally interesting one.

**New completion signal — `waitForIdleCompletion`**:

A new polling helper shared by both process nodes and skill nodes that
returns successfully when the pane's output has been stable AND
`detectIdlePrompt` has been true for N consecutive ticks. Returns on:
- idle detected (success → node complete)
- pane dead (success → node complete, current behaviour preserved)
- context cancelled (failure → node failed)
- timeout (configurable per node, default generous — 30 min — because skills
  can be long-running; configurable at workflow level later)

State machine that prevents false-positive "idle" when the skill hasn't
started yet:

1. On entry: capture baseline output hash.
2. Wait for first **change** in hash (claude has started responding).
3. From there, poll idle → stable → fire complete.

If the pane never changes from baseline (claude doesn't start), treat
timeout as failure.

**`executeSkillNode`**:

Looks at the skill node's incoming edges. For each predecessor:
- If predecessor has a live `TmuxTarget` (verified via `list-panes`), collect it.

Then:
- **If exactly one live predecessor**: reuse its session. Inject the skill
  invocation via `tmux send-keys -H` (I already built this, it handles raw
  byte fidelity). The injection is `/<skill-name>` followed by `Enter`.
- **If multiple live predecessors**: pick the one with the latest
  `StartedAt`. Log a warning. Document this as "sequential-only" in the
  initial version; merge-after-skills is a future problem.
- **If zero live predecessors** (skill is the first node, or all parents
  failed/ended-session): spawn a new session like a process node, but keep
  the session alive after the skill is done so the next chained node can
  reuse it.

**`executeProcessNode` change**:

Current behaviour: spawn new session, wait for pane death, node complete,
pane dies forever. New behaviour: spawn new session, wait for
`waitForIdleCompletion`, node complete, **leave pane alive**. The downstream
node decides what to do with it. Cleanup happens at workflow-end via
`CleanupStaleSessions` (already exists) or via an explicit quit-claude at the
very last node of each chain (we can compute chain-tails from the DAG).

**Keeping claude alive after a skill finishes**:

Claude's REPL already does this — it sits at `❯` after any task. We just
stop killing the session ourselves. The subtle thing is that the first call
to `tmux new-session` runs a single command; when that command exits, the
pane dies. We need to either:
- (a) Wrap the claude invocation in `bash -c "claude ...; exec bash"` so
  after claude exits a shell takes over and keeps the pane alive, OR
- (b) Never let claude exit by running it with `-c "<initial prompt>"` and
  not piping stdin to EOF.

(a) is simpler and doesn't depend on claude CLI internals. I'd go with (a).

**Tracking which session to reuse**:

`execState.exec.Nodes[i].TmuxTarget` is already populated. Skill-node path
just reads the predecessor's `TmuxTarget` field. No new state needed.

**Completion order**:

After a skill completes in a reused session, the skill's `TmuxTarget` is set
to the **same** value as its parent. Both nodes point at the same session.
The downstream node does the same lookup. A chain of N skill nodes and M
process nodes can share the same session as long as each is connected
sequentially.

### Phase 4 — Polish

- **Skill editor**: modal that lets you create / edit a SKILL.md with name,
  description, and body text. Small component, mostly form handling.
- **FS watch**: use Wails' existing file-watcher pattern (already used for
  repo scanning) to re-fetch `ListAllSkills` when any file under
  `.claude/skills/` changes. Debounced so a `git checkout` doesn't melt
  anything.
- **Frontmatter validation**: warn in the sidebar when a skill has no
  `description` or a `description` that's too short / too long. Inline
  badge, not blocking.

## Decisions from schema discussion

| # | Question | Answer |
|---|---|---|
| 1 | How is a command invoked inside a live claude session? | Slash-command syntax: inject `/<command-name>\n` via `tmux send-keys -H` (the hex-input path we already built). Verify empirically before Phase 3 by opening a live claude session and trying it. |
| 2 | What's the "done" contract? | **Default is idle-prompt stable**, but the schema lets each asset override. See `mashedCompletion` below. |
| 3 | Chain-tail cleanup | **Kill each chain-tail's tmux session when the workflow reaches terminal state** (complete OR failed). Explicit, deterministic, no accumulation. `CleanupStaleSessions` at startup stays as a safety net. |
| 4 | Multi-parent nodes | Unresolved at the design layer — **the schema itself will carry the answer**. A command's frontmatter will declare whether it's chainable from multiple predecessors, and if so, how it picks. For v1 we only support single-predecessor chaining; multi-parent is a merge node, not a command node. |

## The BMAD asset schema

Every skill and command lives in a markdown file with YAML frontmatter
(`SKILL.md` or `<command-name>.md`). The schema adds a small optional
namespace. **Presence of any `mashed*` field is the "mashed-ready" signal** —
there is no separate `bmadReady: true` flag to drift out of sync.

### Minimum viable schema

```yaml
---
# Existing fields (already present today)
name: simplify
description: "Review recent changes for reuse, quality, and efficiency, then fix any issues found."

# ── Mashed fields — presence of mashedRole promotes this asset to a
# ── first-class workflow primitive. All other mashed* fields are optional
# ── and inherit sensible defaults when omitted.
mashedRole: command               # required when mashed-ready. enum:
                                #   command — chainable node in a workflow
                                #   skill   — agent-owned, pinned to a session
                                #             (listed in sidebar, NOT draggable)
                                #   agent   — metadata only, no execution

mashedCompletion: idle            # optional. enum:
                                #   idle        — default. stable ❯ prompt = done
                                #   exit        — claude must exit for node to complete
                                #   marker: <s> — a line containing <s> means done
                                #   timeout: 5m — done after wall-clock N (last resort)

mashedInputs: []                  # optional. List of artifact globs or type
                                # names the command consumes. Used by the
                                # canvas to validate edges at save time
                                # (i.e. "can I connect this source to this
                                # target?"). Empty = "takes whatever the
                                # upstream node leaves in the session".

mashedOutputs: []                 # optional. List of artifact globs this
                                # command produces. Surfaced in the
                                # completed-node artifact panel.

mashedChainable: single           # optional. enum:
                                #   single   — only one upstream parent (default
                                #              for command)
                                #   none     — terminal node; nothing may
                                #              chain TO this
                                #   any      — may merge multiple parents
                                #              (advanced; reserved for future)

mashedSessionPinned: false        # optional. true for skills that own a
                                # long-running tmux session; irrelevant
                                # for commands (always false by default).
---

# Simplify

...body text, shown as hover tooltip / description card in the sidebar...
```

### Defaults by role

| Field | `command` | `skill` | `agent` |
|---|---|---|---|
| `mashedCompletion` | `idle` | (N/A — skill doesn't "complete", agent does) | (N/A) |
| `mashedChainable` | `single` | `none` | `none` |
| `mashedSessionPinned` | `false` | `true` | `false` |
| Draggable onto canvas? | **yes** | no (listed, read-only) | no |

### Parser rules

- Assets whose frontmatter lacks `mashedRole` are **silently skipped**.
  They are NOT shown in the sidebar at all. The filesystem stays
  authoritative for "this asset exists"; the BMAD workspace is purely
  the view of "mashed-ready assets".
- Assets with `mashedRole` set but unrecognised → log a warning, skip the
  asset (fail-safe, don't crash the sidebar, don't mis-classify).
- Body-first-line fallback for missing `description:` still applies for
  assets that DO have `mashedRole`.
- Frontmatter parser is tolerant: any unknown keys are silently ignored
  so authors can extend without breaking older mashed builds.

### Example: `/simplify` as a chainable command

```yaml
---
name: simplify
description: "Review changed code for reuse, quality, and efficiency."
mashedRole: command
mashedCompletion: idle
mashedInputs: []
mashedOutputs: []
mashedChainable: single
---
```

### Example: `skill-creator` as an agent-pinned skill

```yaml
---
name: skill-creator
description: "Create, improve, and measure skills. Use interactively."
mashedRole: skill
mashedSessionPinned: true
---
```

Both appear in the Skills tab. Only the first is draggable.

## The `mashed-refactor-asset` skill (Phase 1b on-ramp)

The schema is useless if the Skills tab is empty on day one. Before anyone
hand-annotates a single file, we need a Claude Code skill that users can
invoke to bulk-upgrade their existing `SKILL.md` and `commands/<name>.md`
files into compliant mashed-ready assets.

### Shape

- **Location**: `~/.claude/skills/mashed-refactor-asset/SKILL.md` so it's
  globally available to any claude session in any repo. (Alternatively
  `mashed/.claude/skills/mashed-refactor-asset/SKILL.md` for a repo-scoped
  copy; we'll ship the latter and let users copy to global.)
- **Input**: a filesystem path to either a `SKILL.md`, a
  `commands/<name>.md`, or a directory. If a directory, the skill walks
  it and processes every qualifying file.
- **Output**: the target files are mutated in place with the `mashed*`
  fields written into their YAML frontmatter. A summary report is
  printed listing what was changed and what was skipped and why.

### Inference rules the refactor skill applies

Given a file with arbitrary existing frontmatter and body, the skill
infers each `mashed*` field from textual signals:

| Field | How the refactor skill decides |
|---|---|
| `mashedRole` | `command` by default, UNLESS: (a) the file is under a `skills/` directory (not `commands/`), (b) the body mentions "interactive", "waits for input", "pin", "owns a session", "long-running", in which case `skill`. |
| `mashedCompletion` | `idle` for commands with no further signal. `exit` if the body says "runs and exits" or there's a clear terminal-style command output. `marker: "<string>"` if the body prints a well-known completion line (e.g. `DONE`, `✓ complete`). |
| `mashedInputs` | Parse for file paths, artifact names, or "reads from X" phrases. Best-effort. Empty `[]` if nothing found — schema permits it. |
| `mashedOutputs` | Same heuristic for "writes to", "creates", "produces". Empty `[]` is fine. |
| `mashedChainable` | `single` for commands, `none` for skills. |
| `mashedSessionPinned` | `true` for skills, `false` for commands. |

The skill is explicit about provenance: when it writes a field, it
annotates the commit in stdout with "inferred from: <signal>". For
anything it can't confidently infer, it emits a warning and uses the
conservative default (empty list / `idle` / `single`). The user can
override any field manually afterwards — the schema is tolerant and the
hand-edit wins.

### Idempotency and safety

- **Idempotent**: running twice on the same file produces the same
  output. Existing `mashed*` fields are **preserved, not rewritten**,
  unless the user passes `--force-reinfer`.
- **Non-destructive to existing frontmatter**: the skill parses the full
  YAML frontmatter, merges the `mashed*` fields, and serializes back
  preserving key order for existing fields (new `mashed*` fields append at
  the end of the frontmatter block).
- **Preserves body verbatim**: the body bytes after the closing `---` are
  written back unchanged.
- **Atomic write**: temp file + rename so a crash mid-write doesn't
  corrupt the original.

### Invocation

Users run this from any repo:

```
claude --dangerously-skip-permissions "use mashed-refactor-asset on ~/.claude/commands"
```

The skill walks the directory, prints a report like:

```
mashed-refactor-asset: scanning 47 assets...
  [NEW]     /Users/linus/.claude/commands/simplify.md           → mashedRole: command, mashedCompletion: idle
  [NEW]     /Users/linus/.claude/commands/commit.md             → mashedRole: command, mashedCompletion: exit  (inferred: "runs git commit and exits")
  [KEEP]    /Users/linus/.claude/commands/loop.md               → already has mashedRole, leaving as-is
  [SKIP]    /Users/linus/.claude/skills/agent-validator/SKILL.md → mashedRole: skill, mashedSessionPinned: true (inferred: "interactively validates")
  [WARN]    /Users/linus/.claude/commands/schedule.md           → ambiguous role, defaulting to command. Review manually.
Updated: 42 files. Skipped: 4 files. Warnings: 1.
```

### When this ships

Phase 1b is authored as a SKILL.md file checked into the repo at
`.claude/skills/mashed-refactor-asset/SKILL.md` (local to mashed) AND as a
companion README that describes the schema and the inference rules so the
skill is reviewable as a markdown document, not just an opaque prompt.

It ships in the **same commit as Phase 1** so a user who pulls this
branch, opens the app, sees an empty Skills tab, and runs the refactor
skill once gets a populated tab on the next reload.

## Open questions that remain

None critical for Phase 1. For Phase 3 we still want to:

- **Verify empirically** that `tmux send-keys -H` injection of `/<cmd>\n`
  into a live claude pane actually fires the slash-command (vs. being
  treated as literal text). Simple manual test once Phase 2 lands.
- **Pick an author for the seed set** of `mashed*` frontmatter on existing
  commands in `~/.claude/commands/`. That's a doc/annotation task, not
  a coding task — can be done incrementally as people use the feature.

## Proposed delivery order

1. **Phase 1** (listing) — small, self-contained, ships immediately. **~1
   commit.**
2. **Phase 2** (canvas nodes) — frontend + minor backend. Can land without
   Phase 3 and simply be non-executable until then. **~1 commit.**
3. **Phase 3** (executor reuse + idle completion) — the architectural change.
   This is the risky one: it changes how every workflow node detects
   completion, not just skills. Deserves its own commit and its own careful
   testing pass. **~1 commit.**
4. **Phase 4** (editor, fs watch, frontmatter polish) — incremental, each
   item landable on its own. **~3 commits.**

Total: ~6 commits. Phase 1 lands today. Phases 2–3 land together as the
skill-execution feature. Phase 4 can drip.

## Tests

- **Go**: unit tests for loader + frontmatter parsing, executor tests for
  `waitForIdleCompletion` state machine (driven through the existing
  `CommandRunner` mock), `executeSkillNode` session-reuse decision tree,
  chain integration test.
- **Frontend**: vitest for `groupedSkills` partition rendering, Svelte
  component test for skill-node drag-drop.
- **Empirical**: run a two-node workflow (process → skill) end to end in
  the live app and confirm the skill inherits the session.
