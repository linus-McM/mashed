---
type: Module
title: storage_test.go
description: "Graphify community 35: internal/bmad/storage_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: storage_test, resource: internal/bmad/storage_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: cbabd36eb06877ea }
---

# Files
- `internal/bmad/storage_test.go`

# Symbols
- storage_test.go (internal/bmad/storage_test.go:L1)
- TestDeleteWorkflow_NotFound() (internal/bmad/storage_test.go:L105)
- TestSaveWorkflow_PathTraversal() (internal/bmad/storage_test.go:L113)
- TestLoadWorkflow_InvalidID() (internal/bmad/storage_test.go:L135)
- newTestStorage() (internal/bmad/storage_test.go:L14)
- TestDeleteWorkflow_InvalidID() (internal/bmad/storage_test.go:L141)
- TestListWorkflows_SkipsMalformedJSON() (internal/bmad/storage_test.go:L149)
- TestListWorkflows_SkipsNonJSON() (internal/bmad/storage_test.go:L163)
- TestSaveAndListAgents() (internal/bmad/storage_test.go:L177)
- TestDeleteAgent() (internal/bmad/storage_test.go:L188)
- TestDeleteAgent_NotFound() (internal/bmad/storage_test.go:L199)
- TestSaveAgent_InvalidID() (internal/bmad/storage_test.go:L205)
- sampleWorkflow() (internal/bmad/storage_test.go:L21)
- TestDeleteAgent_InvalidID() (internal/bmad/storage_test.go:L213)
- sampleWorkflowWithRepo() (internal/bmad/storage_test.go:L239)
- TestListWorkflowsByRepo() (internal/bmad/storage_test.go:L247)
- TestListWorkflowsByRepo_ListWorkflowsStillReturnsAll() (internal/bmad/storage_test.go:L316)
- TestWorkflowDef_RepoPathSerialization() (internal/bmad/storage_test.go:L329)
- TestSaveWorkflow_Overwrite() (internal/bmad/storage_test.go:L378)
- sampleAgent() (internal/bmad/storage_test.go:L53)
- TestSaveAndLoadWorkflow() (internal/bmad/storage_test.go:L67)
- TestLoadWorkflow_NotFound() (internal/bmad/storage_test.go:L78)
- TestListWorkflows() (internal/bmad/storage_test.go:L84)
- TestDeleteWorkflow() (internal/bmad/storage_test.go:L95)

# Depends on
- [NewExecutor](/modules/newexecutor.md)
- [Storage](/modules/storage.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)

# Features
- no feature plan names these files
