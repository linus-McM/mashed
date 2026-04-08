# Sprint 2 Backlog: Repo-Scoped BMAD Workflow Builder Driven by sprint-status.yaml

## Sprint Backlog

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | Sprint Status YAML Parser | P0 | backend | M | none |
| 2 | Wails Bindings for Sprint Status | P0 | backend | S | Story 1 |
| 3 | Add RepoPath to WorkflowDef + Storage Filtering | P0 | backend | M | none |
| 4 | Repo Context Flow (App.svelte to WorkflowBuilder) | P0 | fullstack | M | Story 3 |
| 5 | Sprint Panel in Sidebar | P1 | frontend | L | Story 2, Story 4 |
| 6 | ExecutionBar Simplification | P1 | frontend | S | Story 4 |
| 7 | Sprint-Aware Canvas Nodes | P2 | fullstack | L | Story 2, Story 5 |
| 8 | Repo Context Header Bar | P3 | frontend | S | Story 4, Story 5 |

**Total Stories:** 8
**Ready for Sprint:** Stories 1, 2, 3, 4, 5, 6, 7, 8 (all status: ready)
**Recommended Sprint Order:** 1, 3, 2, 4, 6, 5, 7, 8

## Dependency Graph

```
Story 1 (Parser) ──> Story 2 (Bindings) ──> Story 5 (Sprint Panel) ──> Story 7 (Sprint Nodes)
                                                    │                          │
Story 3 (RepoPath) ──> Story 4 (Repo Flow) ────────┘                          │
                              │                                                │
                              ├──> Story 6 (ExecBar Simplify)                  │
                              │                                                │
                              └──> Story 8 (Context Header) <──────────────────┘
```

## Parallelization Opportunities

- **Stories 1 and 3** can be developed in parallel (no dependencies between them)
- **Story 2** depends on Story 1 only
- **Story 4** depends on Story 3 only
- **Stories 5 and 6** can be developed in parallel once Story 4 is complete (Story 5 also needs Story 2)
- **Story 7** requires Stories 2 and 5
- **Story 8** requires Stories 4 and 5

## Recommended Sprint Execution

**Wave 1 (parallel):** Stories 1 + 3 -- pure backend, no dependencies
**Wave 2 (parallel):** Stories 2 + 4 -- bindings and frontend plumbing
**Wave 3 (parallel):** Stories 5 + 6 -- sidebar panel and exec bar cleanup
**Wave 4 (sequential):** Story 7 -- sprint-aware nodes (needs 5)
**Wave 5 (sequential):** Story 8 -- polish header bar

Story files written to: docs/stories/sprint2-{01..08}-*.md
