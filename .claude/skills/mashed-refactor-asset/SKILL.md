---
name: mashed-refactor-asset
description: >
  Upgrade existing Claude Code skills and commands so they appear in
  the Mashed workspace's Skills tab. Walks a target file OR directory,
  infers the right `mashedRole` / `mashedCompletion` / `mashedChainable`
  / `mashedSessionPinned` / `mashedInputs` / `mashedOutputs` frontmatter
  fields from the body content of each asset, and writes the merged
  frontmatter back atomically. Use this whenever the user wants to
  "populate the Skills tab", "make my commands available to the
  workspace", "convert my .claude/skills for BMAD", or asks about the
  mashed frontmatter schema. Idempotent — running twice produces the
  same result unless `--force-reinfer` is passed. Non-destructive:
  existing frontmatter keys are preserved, new keys append at the end
  of the YAML block, and the body bytes after the closing `---` are
  written back unchanged.

mashedRole: skill
mashedSessionPinned: true
---

# Mashed Refactor Asset

You are a frontmatter annotation specialist for Mashed, a Wails desktop
app that provides a visual workspace for composing Claude Code skills
and slash-commands into executable workflows. Your job is to upgrade
existing `.claude/skills/*/SKILL.md` and `.claude/commands/*.md` files
so they appear in the Mashed workspace's Skills tab.

## Context you should read first

Before touching any files, read `docs/plans/skills-and-session-reuse.md`
at the repo root (if it exists) so you understand the schema and the
reasoning behind each field. That document is the authoritative source
for what the `mashed*` frontmatter namespace means.

If the doc isn't present, proceed from the inline specification in
this skill — the rules below are a faithful excerpt.

## Inputs

- **target path**: a file or directory to process. Required. Resolve
  relative paths against the current working directory. Common targets:
  - `~/.claude/commands` — global commands
  - `~/.claude/skills` — global skills
  - `./.claude/commands` — repo-local commands
  - `./.claude/skills` — repo-local skills
  - a single `SKILL.md` or `<name>.md` file
- **`--force-reinfer`** (optional flag): re-write every `mashed*` field
  even if the file already has them. Default behaviour is to PRESERVE
  existing mashed frontmatter and only annotate files that lack it.

## What a mashed-ready asset looks like

Every asset is a markdown file with YAML frontmatter. The frontmatter
MUST contain `mashedRole` for the asset to appear in the Mashed
workspace. All other `mashed*` fields are optional and inherit
per-role defaults when omitted. Unknown YAML keys are ignored — the
Mashed parser is tolerant, so authors can layer extra metadata without
breaking things.

### Schema

```yaml
---
# ── Existing fields (already present; do not touch) ──
name: <slug>
description: "<one-line summary, shown in the sidebar tooltip>"

# ── Mashed fields (what this skill adds) ──
mashedRole: command | skill | agent       # required for mashed-ready
mashedCompletion: idle | exit | "marker: <text>" | "timeout: <duration>"
mashedInputs: []                          # artifact names / globs consumed
mashedOutputs: []                         # artifact names / globs produced
mashedChainable: single | none | any      # per-node chaining shape
mashedSessionPinned: true | false         # long-running agent ownership
---
```

### Role defaults (apply when the field is omitted)

| Field | `command` | `skill` | `agent` |
|---|---|---|---|
| `mashedCompletion` | `idle` | (N/A) | (N/A) |
| `mashedChainable` | `single` | `none` | `none` |
| `mashedSessionPinned` | `false` | `true` | `false` |

## Inference rules — how to pick each field

### `mashedRole`

- **Filesystem layout wins first**: if the file lives under a
  `commands/` directory, default to `command`. If it lives under a
  `skills/` directory (as `<name>/SKILL.md`), default to `skill`.
- **Body signals can override**: if the body describes an
  interactive, long-running workflow that "owns" a tmux session —
  keywords like "interactively", "waits for input", "pins a session",
  "hands control back and forth", "conversational", "multi-step
  dialogue" — upgrade to `skill` even in a commands directory.
- **Agent role is rare**: only set `agent` when the frontmatter
  already contains agent-specific fields OR the body explicitly says
  "this is an agent definition". Don't invent it.

### `mashedCompletion`

- Default to `idle` for commands with no other signal. That's the
  "pane output stabilises at the `❯` prompt" signal the Mashed
  executor polls for.
- Use `exit` when the body clearly says the command runs and
  terminates (e.g. a wrapper around `git commit`, a one-shot script
  runner, or any command that invokes an external tool and stops).
  Keywords: "runs and exits", "executes and terminates", "returns",
  "completes immediately".
- Use `"marker: <string>"` when the body describes a specific
  completion line the process prints, e.g. `DONE`, `✓ complete`,
  `===END===`. The string is what the marker literally says.
- Use `"timeout: <duration>"` only as a last resort — when there's no
  deterministic signal and the best we can do is wall-clock. Format
  the duration as Go's `time.ParseDuration` accepts: `5m`, `30s`,
  `2h`. Emit a warning in the report for every file you pick
  timeout for.

### `mashedInputs` and `mashedOutputs`

- Parse the body for concrete artifact names or globs. Signals:
  "reads from `<path>`", "expects `<file>`", "input: `<pattern>`",
  "writes to `<path>`", "produces `<name>`", "creates `<file>`".
- Extract the quoted / backticked identifier verbatim. Don't
  normalise case or expand globs.
- Empty `[]` is fine — schema permits it and the Mashed executor
  treats it as "takes whatever the upstream session already has in
  context".

### `mashedChainable`

- `single` for commands (the default for `mashedRole: command`).
- `none` for skills and agents — they don't chain inside workflows.
- `any` is RESERVED for future merge-node support. Do NOT assign it
  unless the body explicitly describes merging multiple upstream
  parents.

### `mashedSessionPinned`

- `true` for every `mashedRole: skill`. Skills own their tmux
  session by definition.
- `false` for commands — they run inside whatever session the
  upstream node established.

## Procedure

For each target file:

1. **Read** the full file content. Preserve byte-for-byte.
2. **Parse** the YAML frontmatter (the block between the opening `---`
   and the next `---` line). If no frontmatter exists, insert a new
   frontmatter block at the top with the inferred fields.
3. **Check for existing `mashed*` fields**. If any are present AND
   `--force-reinfer` is NOT set, keep them and move on — emit
   `[KEEP]` in the report.
4. **Infer** each field using the rules above. When you cannot
   confidently infer, use the conservative default (empty `[]` for
   lists, `idle` for completion, `single` for chainable) and emit a
   `[WARN]` line in the report identifying which field was
   defaulted.
5. **Merge** the inferred `mashed*` fields INTO the existing YAML
   frontmatter. Preserve the order of existing keys. Append the new
   `mashed*` keys at the END of the frontmatter block.
6. **Write atomically**: write to `<path>.tmp`, then rename to
   `<path>`. Never leave a partially-written file on disk.
7. **Preserve the body**: bytes after the closing `---` are written
   back unchanged.

## Report format

After processing every file, emit a summary report to stdout:

```
mashed-refactor-asset: scanning <target>...
  [NEW]  <absolute path>   → role: <X>, completion: <Y>  (inferred: "<signal>")
  [KEEP] <absolute path>   → already has mashedRole, left unchanged
  [SKIP] <absolute path>   → not a .md file / not a skill directory
  [WARN] <absolute path>   → <what was ambiguous and what you defaulted to>

Updated: <N> files. Skipped: <M> files. Warnings: <K>.
```

Per-line tags:
- `[NEW]`: file had no `mashed*` fields; you added them.
- `[KEEP]`: file already has `mashedRole`; you left it alone (or
  re-inferred under `--force-reinfer`).
- `[SKIP]`: entry at that path was not a valid asset (not a `.md`
  file, or a skill directory without a `SKILL.md` inside).
- `[WARN]`: file was annotated but at least one field was defaulted
  because you couldn't confidently infer it. The user should
  review.

## Safety invariants

- **Idempotent**: running twice on the same target produces the same
  output. `--force-reinfer` is the only escape hatch.
- **Non-destructive**: existing YAML keys outside the `mashed*`
  namespace are preserved byte-for-byte. Existing `mashed*` keys are
  preserved unless `--force-reinfer`.
- **Body-preserving**: bytes after the closing `---` are untouched.
- **Atomic writes**: always write to a temp file + rename. Never
  partial.
- **No network**: this skill never fetches anything. Purely local
  filesystem + inference.

## Examples

### Example input — `~/.claude/commands/simplify.md`

```markdown
---
name: simplify
description: "Review recent changes for reuse, quality, and efficiency, then fix issues."
---

# Simplify

Walk the diff against the base branch. For each changed file, look for
duplicated logic, missing abstractions, or overcomplicated
conditionals. Rewrite when it clearly improves the code. Don't touch
files outside the diff.
```

### Example output

```markdown
---
name: simplify
description: "Review recent changes for reuse, quality, and efficiency, then fix issues."
mashedRole: command
mashedCompletion: idle
mashedInputs: []
mashedOutputs: []
mashedChainable: single
mashedSessionPinned: false
---

# Simplify

(body unchanged)
```

Report line:

```
[NEW]  /Users/<user>/.claude/commands/simplify.md   → role: command, completion: idle  (inferred: command dir, no exit signal)
```

### Example input — `~/.claude/skills/skill-creator/SKILL.md`

```markdown
---
name: skill-creator
description: "Create new skills, modify existing skills, and measure skill performance."
---

# Skill Creator

Interactively helps the user design, implement, and validate a new
skill. Asks questions to gather context, iterates on the body, and
hands back control when the user says they're done.
```

### Example output

```markdown
---
name: skill-creator
description: "Create new skills, modify existing skills, and measure skill performance."
mashedRole: skill
mashedChainable: none
mashedSessionPinned: true
---

# Skill Creator

(body unchanged)
```

Report line:

```
[NEW]  /Users/<user>/.claude/skills/skill-creator/SKILL.md   → role: skill, session-pinned  (inferred: "interactively", "hands back control")
```

## Do not

- Do NOT touch files outside the target path.
- Do NOT modify the body bytes below the closing `---`.
- Do NOT remove existing YAML keys — even ones you don't recognise.
- Do NOT invent field values just to fill every slot. Empty `[]` and
  the per-role defaults are fine.
- Do NOT set `mashedChainable: any` unless the body explicitly
  describes merging multiple upstream parents.
- Do NOT overwrite existing `mashed*` fields unless the user passed
  `--force-reinfer`.
