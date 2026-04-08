# Sprint 4 Backlog: Standardized BMAD Artifact Schema & Process Registry Expansion

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | Artifact Path Convention & Resolution System | P0 | backend | M | none |
| 2 | File-Aware Context Passing in Executor | P0 | backend | M | Story 1 |
| 3 | Registry Expansion, Module Population & New Templates | P1 | backend | M | none |
| 4 | Programmatic Skill Definition Generation | P1 | backend | S | Story 1, Story 3 |
| 5 | Post-Completion Artifact Verification & Frontend Status | P1 | fullstack | L | Story 1, Story 2 |

**Total Stories:** 5
**Ready for Sprint:** Stories 1, 2, 3, 4, 5 (all status: ready)
**Recommended Sprint Order:** 1, 3, 2, 4, 5

## Dependency Graph

```
Story 1 (Artifact Path System)
  |         \
  v          v
Story 2      Story 3 (Registry Expansion) -- no dep on Story 1
  |            |
  v            v
Story 5      Story 4 (Skill Generation)
```

## Parallelization Notes

- **Stories 1 and 3** can be executed in parallel (no dependencies between them).
- **Story 2** requires Story 1 to be complete.
- **Story 4** requires both Story 1 and Story 3 to be complete.
- **Story 5** requires Stories 1 and 2 to be complete.

## Sprint Goal

Enable the BMAD workflow builder to pass actual file paths between process steps, verify artifact outputs, and provide 32 functional skill definitions -- making workflows executable end-to-end for the first time.

## Story Files

- `docs/stories/sprint4-01-artifact-path-system.md`
- `docs/stories/sprint4-02-file-aware-context.md`
- `docs/stories/sprint4-03-registry-expansion.md`
- `docs/stories/sprint4-04-skill-generation.md`
- `docs/stories/sprint4-05-artifact-verification-ui.md`
the