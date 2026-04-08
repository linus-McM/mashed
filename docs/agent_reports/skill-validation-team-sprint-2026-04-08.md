# Skill Validation Report: team-sprint

**Date:** 2026-04-08
**Skill Path:** .claude/skills/team-sprint/
**Validator Version:** 1.1
**Focus:** Targeted investigation of tmux exit bug after Phase 0 story decomposition

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 9 | 0 | 0 | 0 |
| Functional | 0 | 0 | 0 | 1 |
| Efficiency | 3 | 0 | 0 | 0 |
| Instruction Compliance | 3 | 0 | 3 | 0 |
| Agent Simulation | 0 | 0 | 0 | 1 |
| **Total** | **15** | **0** | **3** | **2** |

**Overall Grade:** D (3 failures found and fixed)

## Root Cause Analysis: tmux Exit Bug

Three compounding bugs caused tmux to exit after Phase 0 (story decomposition):

### Bug 1: `--session-id` requires UUID (FAIL - FIXED)
The tmux spawning section used `tmux split-window -v "claude --session-id sprint-{id}-{role} --agent {agent-type}"`. Per `claude --help`, `--session-id` requires a valid UUID. The descriptive string `sprint-story1-test-writer` caused every `claude` process to error immediately on start, silently killing all tmux panes.

### Bug 2: tmux spawning incompatible with Agent tool coordination (FAIL - FIXED)
The skill defined two conflicting spawning mechanisms:
- **tmux section:** `tmux split-window -v "claude ..."` — spawns separate OS processes
- **Phase 0b:** `Task({subagent_type: "scrum-master", ...})` — meant for the Agent tool (in-process subagent)

Agents spawned via tmux are separate processes that **cannot** coordinate via Claude Code's `SendMessage`/`TaskList`/`TaskUpdate`. The skill's entire coordination model (Phases 3-5) depends on these tools, making tmux spawning architecturally incompatible.

### Bug 3: Phase 0b used `team_name` without `TeamCreate` (FAIL - FIXED)
Phase 0b specified `team_name: "sprint-plan"` in its `Task()` call, but no team was ever created with `TeamCreate` (first called in Phase 2a with a different name `sprint-{storyId}`). This caused the Agent tool to reference a non-existent team context.

### Cascade sequence:
1. Phase 0: Scrum master succeeds (via Agent tool subagent) -- stories are written
2. Phase 2e: Lead attempts to spawn engineers via tmux per rule #5
3. Every `claude --session-id <non-uuid>` fails instantly -- panes close
4. Lead retries, all fail, eventually exits
5. If user started tmux with `tmux new-session "claude"`, tmux exits

## Fixes Applied

| File | Change | Lines |
|------|--------|-------|
| `team-sprint.md` | Replaced `## tmux Spawning` section with `## Agent Spawning` using `Agent({ run_in_background: true })` | 39-50 |
| `team-sprint.md` | Removed tmux prerequisite check | 35 |
| `team-sprint.md` | Changed rule #5 from "tmux panes for all agents" to "Background agents with Agent tool" | 17 |
| `team-sprint.md` | Replaced `Task()` pseudo-tool with proper `Agent()` syntax in Phase 0b, removed `team_name`, fixed "message the lead" to "produce as final output" | 85-120 |
| `team-sprint.md` | Removed `kill tmux window` from Story Lifecycle step 4 | 69 |
| `team-sprint.md` | Removed `kill tmux window` from DONE shutdown task | 248-252 |
| `team-sprint.md` | Rewrote Phase 5d shutdown: removed `SendMessage` shutdown + `tmux kill-window`, replaced with `TeamDelete()` | 366-379 |
| `SKILL.md` | Updated rule #5 to use Agent tool | 27 |
| `SKILL.md` | Removed tmux prerequisite | 33 |
| `references/scaling-errors-antipatterns.md` | Changed "Spawn replacement in new tmux pane" to Agent tool | 19 |

## Structural Validation

- PASS: SKILL.md exists
- PASS: Frontmatter present with delimiters
- PASS: `name` field present (team-sprint)
- PASS: `description` field present (74 words)
- PASS: SKILL.md body non-empty (90 lines)
- PASS: SKILL.md under 500 lines (101 lines)
- PASS: No sensitive files detected
- PASS: 5 heavy directives (under threshold)
- PASS: All referenced files exist on disk

## Efficiency Analysis

| Metric | Value | Status |
|--------|-------|--------|
| SKILL.md lines | 101 | PASS |
| Estimated tokens | ~1070 | PASS |
| Heavy directives | 5 | PASS |

## Recommendations

All 3 failures have been fixed. The skill now uses Claude Code's native Agent tool for all agent spawning, which:
1. Keeps coordination within the same process (SendMessage/TaskList work)
2. Avoids invalid `--session-id` CLI flag issues
3. Provides automatic completion notifications via `run_in_background: true`
4. Eliminates tmux as a hard dependency
