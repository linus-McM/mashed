# Skill Validation Report: team-sprint

**Date:** 2026-04-12
**Skill Path:** `/Users/linus/Development/mashed/.claude/skills/team-sprint`
**Validator Version:** 1.1
**Change under review:** integrate `use-repo-code` skill as the canonical codebase-reference path for the lead and all spawned agents.

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 9 | 0 | 0 | 0 |
| Functional | 7 | 0 | 0 | 0 |
| Efficiency | 4 | 0 | 0 | 0 |
| Instruction Compliance | 5 | 0 | 0 | 0 |
| Agent Simulation | 0 | 0 | 0 | 1 |
| **Total** | **25** | **0** | **0** | **1** |

**Overall Grade: A** — production-ready. No rectifier loop needed.

## Changes Applied

1. **`SKILL.md`** — added Non-Negotiable Rule #7 requiring the lead and every spawned agent to route codebase *discovery* / *reference* reads through `use-repo-code`; added the skill to the Prerequisites list with a note on re-running Repomix if stale.
2. **`team-sprint.md`** — Phase 0a plan-decomposition note, Phase 0b scrum-master prompt (replaced "Scan codebase structure" with explicit use-repo-code guidance), Phase 0d ui-architect prompt (grep `frontend/src/style.css` via use-repo-code instead of raw Read), Phase 4c ui-architect critique prompt (Svelte component inspection via use-repo-code with `-A` tuned from `project-structure.md` line counts).
3. **`references/agent-prompts.md`** — added a top-of-file "Codebase Lookup Rule" block that applies to all engineer (go / ui / fullstack) sign-offs, with an explicit carve-out: normal Read+Edit still applies to files the task is actively modifying.

## Validator Output

```
PASS  skill_md_exists
PASS  frontmatter_present
PASS  name_field         (team-sprint)
PASS  description_length (73 words, >30 threshold)
PASS  body_nonempty      (72 content lines)
PASS  size_limit         (SKILL.md 83 lines / ~1000 tokens)
PASS  no_sensitive_files
PASS  directive_count    (6, well under 10 threshold)
PASS  all referenced files exist (6 references/*.md + team-sprint.md)
```

## Instruction Compliance

| Instruction | Ambiguity | Buried | Conflict | Missing Why | Implicit | Drift |
|---|---|---|---|---|---|---|
| Use `use-repo-code` for discovery (Rule #7) | OK | Line ~26 top section | None | OK — reason stated (266K tokens) | OK — carve-out explicit | OK |
| Scrum master uses use-repo-code in Phase 0b | OK | Line ~52 | None | OK | OK | OK |
| UI architect uses use-repo-code in 0d/4c | OK | In agent prompt | None | OK | OK | OK |
| Engineer sign-off codebase rule | OK | Top of agent-prompts.md | None | OK | OK | OK |
| Prerequisite lists use-repo-code as required | OK | Top of SKILL.md | None | OK | OK | OK |

No conflicts detected. The new rule coexists cleanly with existing rules 1–6 because it governs *reference reads*, not the write path (tests, build, coverage, /simplify, AC validation are untouched).

## Agent Simulation — SKIPPED

No `evals/` directory present; deterministic text changes don't require simulation to verify grade.

## Recommendations

None. Skill is at grade A after a single-pass edit — no rectifier loop triggered.
