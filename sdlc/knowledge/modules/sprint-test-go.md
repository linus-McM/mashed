---
type: Module
title: sprint_test.go
description: "Graphify community 227: app_bmad.go, internal/bmad/sprint.go, internal/bmad/sprint_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: sprint, resource: internal/bmad/sprint.go, last_modified: "2026-05-07T21:14:42+10:00", digest: e3337099ff62a7c9 }
  - { id: sprint_test, resource: internal/bmad/sprint_test.go, last_modified: "2026-04-08T11:24:14+10:00", digest: 6c11bf041ab7601c }
---

# Files
- `app_bmad.go`
- `internal/bmad/sprint.go`
- `internal/bmad/sprint_test.go`

# Symbols
- .UpdateStoryStatus() (app_bmad.go:L141)
- UpdateStoryStatus() (internal/bmad/sprint.go:L247)
- sprintStatusPath() (internal/bmad/sprint.go:L91)
- ParseSprintStatus() (internal/bmad/sprint.go:L97)
- sprint_test.go (internal/bmad/sprint_test.go:L1)
- TestParseSprintStatus_MissingFile() (internal/bmad/sprint_test.go:L121)
- TestParseSprintStatus_EmptyDevStatus() (internal/bmad/sprint_test.go:L129)
- TestParseSprintStatus_Malformed() (internal/bmad/sprint_test.go:L139)
- createSprintYAML() (internal/bmad/sprint_test.go:L14)
- TestParseSprintStatus_Retrospective() (internal/bmad/sprint_test.go:L147)
- TestUpdateStoryStatus_Success() (internal/bmad/sprint_test.go:L180)
- TestUpdateStoryStatus_InvalidStatus() (internal/bmad/sprint_test.go:L208)
- TestUpdateStoryStatus_StoryNotFound() (internal/bmad/sprint_test.go:L226)
- TestUpdateStoryStatus_InvalidStoryID() (internal/bmad/sprint_test.go:L234)
- TestUpdateStoryStatus_PreservesFormatting() (internal/bmad/sprint_test.go:L255)
- TestUpdateStoryStatus_MissingFile() (internal/bmad/sprint_test.go:L281)
- TestParseSprintStatus_MultiEpic() (internal/bmad/sprint_test.go:L70)

# Depends on
- [sprint.go](/modules/sprint-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
