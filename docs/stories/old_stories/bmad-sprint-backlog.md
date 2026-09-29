# Sprint Backlog: BMAD Workflow Builder

## Stories

| # | Story | Priority | Domain | Size | Depends On |
|---|-------|----------|--------|------|------------|
| 1 | [Svelte 4 Upgrade & xyflow Installation](bmad-01-svelte-upgrade.md) | P0 | frontend | S | none |
| 2 | [Go Backend: BMAD Domain Types & Process Registry](bmad-02-go-types-registry.md) | P0 | backend | M | none |
| 3 | [Go Backend: Templates & File Storage](bmad-03-templates-storage.md) | P0 | backend | M | bmad-02 |
| 4 | [Go Backend: Workflow Executor (DAG Walker & tmux Spawner)](bmad-04-executor.md) | P0 | backend | L | bmad-02, bmad-03 |
| 5 | [Wails Bindings: BMAD API Surface](bmad-05-wails-bindings.md) | P0 | backend | M | bmad-02, bmad-03, bmad-04 |
| 6 | [Frontend: WorkflowBuilder View & SvelteFlow Canvas](bmad-06-workflow-view-canvas.md) | P1 | frontend | L | bmad-01, bmad-05 |
| 7 | [Frontend: Custom ProcessNode & Supporting Components](bmad-07-custom-nodes-components.md) | P1 | frontend | L | bmad-06 |
| 8 | [Execution Integration: Live Status, Terminal Access & Artifact Passing](bmad-08-execution-integration.md) | P1 | fullstack | M | bmad-05, bmad-07 |
| 9 | [Module Versioning & Custom Agent Management](bmad-09-modules-agents.md) | P2 | fullstack | M | bmad-05, bmad-07 |

## Dependency Graph

```
bmad-01 (Svelte Upgrade) ----\
                               \
bmad-02 (Types & Registry) -----> bmad-03 (Templates & Storage) --> bmad-04 (Executor)
                                                                         |
                                                                         v
bmad-01 + bmad-04 --------> bmad-05 (Wails Bindings) ---------> bmad-06 (View & Canvas)
                                      |                              |
                                      |                              v
                                      |                        bmad-07 (Custom Nodes)
                                      |                              |
                                      +----> bmad-08 (Execution Integration) <--+
                                      |                                         |
                                      +----> bmad-09 (Modules & Agents) --------+
```

## Parallel Execution Opportunities

**Wave 1** (no dependencies, start in parallel):
- bmad-01 (Svelte Upgrade) -- frontend only
- bmad-02 (Types & Registry) -- backend only

**Wave 2** (after Wave 1):
- bmad-03 (Templates & Storage) -- needs bmad-02

**Wave 3** (after Wave 2):
- bmad-04 (Executor) -- needs bmad-02 + bmad-03

**Wave 4** (after Wave 3):
- bmad-05 (Wails Bindings) -- needs bmad-02 + bmad-03 + bmad-04

**Wave 5** (after bmad-01 + bmad-05):
- bmad-06 (View & Canvas) -- needs bmad-01 + bmad-05

**Wave 6** (after bmad-06):
- bmad-07 (Custom Nodes) -- needs bmad-06

**Wave 7** (after bmad-07):
- bmad-08 (Execution Integration) -- needs bmad-05 + bmad-07
- bmad-09 (Modules & Agents) -- needs bmad-05 + bmad-07

## Recommended Sprint Order

1. **bmad-01** + **bmad-02** (parallel -- frontend upgrade + backend types)
2. **bmad-03** (backend templates & storage)
3. **bmad-04** (backend executor)
4. **bmad-05** (wails bindings -- integrates all backend)
5. **bmad-06** (frontend view & canvas)
6. **bmad-07** (frontend custom nodes & components)
7. **bmad-08** (execution integration -- end-to-end)
8. **bmad-09** (modules & agents -- polish)

## Summary

**Total Stories:** 9
**P0 (Critical Path):** 5 stories (bmad-01 through bmad-05)
**P1 (Core Feature):** 3 stories (bmad-06 through bmad-08)
**P2 (Polish):** 1 story (bmad-09)
**Ready for Sprint (no dependencies):** bmad-01, bmad-02

Story files written to: `/Users/dev/Development/mashed/docs/stories/`
