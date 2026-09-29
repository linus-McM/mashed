---
type: Module
title: prompt_test.go
description: "Graphify community 70: internal/uiadapter/prompt.go, internal/uiadapter/prompt_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
  - { id: prompt_test, resource: internal/uiadapter/prompt_test.go, last_modified: "2026-04-27T10:45:20+10:00", digest: c0e8d16fee76a549 }
---

# Files
- `internal/uiadapter/prompt.go`
- `internal/uiadapter/prompt_test.go`

# Symbols
- prompt.go (internal/uiadapter/prompt.go:L1)
- SystemPrompt() (internal/uiadapter/prompt.go:L13)
- PromptVersion() (internal/uiadapter/prompt.go:L17)
- prompt_test.go (internal/uiadapter/prompt_test.go:L1)
- TestPromptVersion_IsV2() (internal/uiadapter/prompt_test.go:L107)
- TestAdapter_SendsSystemPromptInRequest() (internal/uiadapter/prompt_test.go:L116)
- TestPrompt_Golden_Brainstorming() (internal/uiadapter/prompt_test.go:L139)
- TestPrompt_Golden_Elicitation() (internal/uiadapter/prompt_test.go:L159)
- TestPrompt_Golden_ProductBrief() (internal/uiadapter/prompt_test.go:L182)
- TestPrompt_Golden_PartyMode() (internal/uiadapter/prompt_test.go:L214)
- TestPrompt_Golden_PartyMode_CodeBlockPreserved() (internal/uiadapter/prompt_test.go:L232)
- TestPrompt_Golden_Freeform() (internal/uiadapter/prompt_test.go:L255)
- readFixture() (internal/uiadapter/prompt_test.go:L30)
- newStubbedOllama() (internal/uiadapter/prompt_test.go:L41)
- decisionGroups() (internal/uiadapter/prompt_test.go:L53)
- goldenAdapter() (internal/uiadapter/prompt_test.go:L66)
- runGolden() (internal/uiadapter/prompt_test.go:L77)
- TestSystemPrompt_ContainsAllSections() (internal/uiadapter/prompt_test.go:L88)

# Depends on
- [NewDefault](/modules/newdefault.md)

# Inferred
- [NewDefault](/modules/newdefault.md)

# Features
- no feature plan names these files
