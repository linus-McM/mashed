# Playbook: Deduplicate `use-repo-code` skill against `.wolf/anatomy.md`

**Audience:** an agent working in a repo that has BOTH OpenWolf (`.wolf/anatomy.md`)
AND a Repomix-generated `use-repo-code` skill at `.claude/skills/use-repo-code/`.

**Goal:** eliminate the overlap between the two indexes so `.wolf/anatomy.md`
is the single source for directory navigation and `use-repo-code` becomes a
pure bulk-grep target (file contents only). No duplicate directory trees, no
stale snapshot of file layout, clear arbitration between the two tools.

**Why this matters:** both indexes answer "what files exist and where" which
causes agents to pick one arbitrarily, cite stale data, or ignore project
conventions. `anatomy.md` is auto-maintained by `openwolf scan` on a fresh
filesystem walk; the Repomix `project-structure.md` is frozen at skill
generation time and drifts on every commit. Keep the live one, drop the
frozen one.

**Do not hand-edit `anatomy.md`.** Its header says
`Auto-maintained by OpenWolf. Last scanned: ...` and running `openwolf scan`
rewrites the file from scratch. Manual edits get wiped. Fix the overlap at
the skill side.

---

## Preconditions (verify before starting)

Run these checks. If any fail, STOP and report — this playbook does not apply.

1. `.wolf/anatomy.md` exists and its header contains `Auto-maintained by OpenWolf`
2. `.claude/skills/use-repo-code/SKILL.md` exists
3. `.claude/skills/use-repo-code/references/` directory exists
4. `openwolf` binary is on PATH (`which openwolf` returns a path)
5. `repomix.config.json` exists at repo root (if missing, skip Step 3 below
   but still do Steps 1, 2, 4)

Also skim `CLAUDE.md` and `.wolf/OPENWOLF.md` for any project-specific rules
that contradict this playbook.

---

## Steps

### 1. Refresh `anatomy.md`

```bash
openwolf scan
```

Expected: `Anatomy scan complete: N files indexed in Xms`. Confirms anatomy.md
is current before we start routing agents to it.

### 2. Delete the duplicate directory index from the skill

```bash
rm .claude/skills/use-repo-code/references/project-structure.md
```

If the file does not exist, that is fine — the skill may have been
regenerated already. Continue.

### 3. Prevent regeneration of the duplicate

Edit `repomix.config.json`. In the `output` section, set:

```json
"directoryStructure": false,
```

(It is almost always `true` by default.) This stops future `repomix` runs
from recreating `project-structure.md` in the skill bundle.

If the repo has no `repomix.config.json`, create one or document the
regeneration-risk in the skill's SKILL.md instead.

### 4. Rewrite `SKILL.md` to route directory questions to anatomy.md

Edit `.claude/skills/use-repo-code/SKILL.md`. Apply these changes:

**4a. Files table** — remove the `project-structure.md` row. Add a note
below the table:

> For the directory tree / per-file descriptions, use `.wolf/anatomy.md` —
> this skill no longer ships a duplicate. If anatomy is stale, run
> `openwolf scan`.

**4b. Arbitration table** — if SKILL.md has (or should have) a "Relationship
to `.wolf/anatomy.md`" section, ensure it contains:

| Question                                    | Use                          |
|---------------------------------------------|------------------------------|
| "Where does feature X live?"                | `.wolf/anatomy.md`           |
| "What does file Y do?"                      | `.wolf/anatomy.md`           |
| "What is the full directory tree?"          | `.wolf/anatomy.md`           |
| "Which files import symbol Z?"              | this skill (`files.md` grep) |
| "Show me every usage of `<symbol>`"         | this skill (`files.md` grep) |
| "What deps are installed?"                  | this skill (`tech-stack.md`) |

**4c. "How to Use" section** — find any step that says "locate a file" or
"find large files" by grepping `project-structure.md`. Replace with:
"Use `.wolf/anatomy.md` (canonical navigation index, auto-maintained by
OpenWolf)." The skill should no longer reference `project-structure.md`
anywhere.

**4d. Common Workflows** — any "find large or complex files" step should
point at anatomy.md's `~NNN tok` estimates, not at snapshot line counts.

### 5. Update `references/summary.md`

Edit `.claude/skills/use-repo-code/references/summary.md`. In the
"File Structure" table, remove the `project-structure.md` row and add this
explanatory note directly below:

> **For the directory tree, use `.wolf/anatomy.md`** — that is the canonical,
> auto-maintained navigation index for this project (per `CLAUDE.md` /
> `.wolf/OPENWOLF.md`). Repomix no longer emits a `project-structure.md` here
> (`directoryStructure: false` in `repomix.config.json`) to avoid duplicating it.

Leave the rest of summary.md alone (Repomix overwrites it on regeneration,
but the note survives until the next `repomix` run — which is fine, because
the repomix config change prevents the duplicate from coming back).

---

## Verification

Run these checks and confirm each:

1. **No dead references in the skill:**
   ```bash
   grep -rn 'project-structure' .claude/skills/use-repo-code/
   ```
   Expected: only hits are in explanatory prose (the note in summary.md and
   maybe SKILL.md explaining the intentional omission). NO grep examples,
   NO instructions telling the agent to read or search it.

2. **File is gone:**
   ```bash
   ls .claude/skills/use-repo-code/references/
   ```
   Expected: `files.md`, `summary.md`, `tech-stack.md`. No
   `project-structure.md`.

3. **Repomix config is set:**
   ```bash
   grep directoryStructure repomix.config.json
   ```
   Expected: `"directoryStructure": false,`

4. **Anatomy is fresh:**
   ```bash
   head -5 .wolf/anatomy.md
   ```
   Expected: recent `Last scanned:` timestamp from Step 1.

5. **Regeneration test (optional but recommended):** if the repo has a
   repomix regeneration script or CI job, run it and re-verify check 2.
   `project-structure.md` must NOT reappear.

All five checks pass → the dedup is complete.

---

## Rollback

If something goes wrong:

1. Restore `repomix.config.json`: set `directoryStructure` back to `true`
2. Regenerate the skill with `repomix` (this recreates `project-structure.md`)
3. Revert SKILL.md and summary.md via `git checkout`

No irreversible operations in this playbook. Anatomy.md is always
regenerable via `openwolf scan`.

---

## Notes for the agent doing this

- **Do not regenerate the whole skill** just because repomix config changed.
  The existing `files.md` and `tech-stack.md` are still correct. Only the
  directory-tree output is stale, and we are deleting that entirely.
- **Do not write into `anatomy.md`.** Any attempt to hand-edit it will be
  lost on the next `openwolf scan`.
- **Check project-specific conventions first.** If `CLAUDE.md` or
  `.wolf/OPENWOLF.md` in the target repo contradicts this playbook (e.g.,
  the repo has an explicit rule that use-repo-code owns directory navigation),
  STOP and ask the user instead of blindly applying.
- **Token budget:** `files.md` is typically 50k-200k tokens. Never `Read` it
  in full — always `Grep` with `-A`/`-B` context flags. SKILL.md should
  already enforce this; if it doesn't, add the warning while you are in there.
- **Commit strategy:** one commit per logical change is fine, or a single
  commit titled something like `chore(skill): dedup use-repo-code against
  anatomy.md`. Do not commit unless the user asks.

---

## Source incident

This playbook was extracted from a fix applied to the `esurfr` repo on
2026-04-12 after a skill-validator run flagged the overlap as a FAIL.
Report: `docs/agent_reports/skill-validation-use-repo-code-2026-04-12.md`.
