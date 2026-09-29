---
type: Module
title: BuiltinTemplates
description: "Graphify community 179: internal/bmad/templates.go, internal/bmad/templates_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: templates, resource: internal/bmad/templates.go, last_modified: "2026-04-08T22:05:53+10:00", digest: 4a56b72dabf842f3 }
  - { id: templates_test, resource: internal/bmad/templates_test.go, last_modified: "2026-04-08T22:05:53+10:00", digest: f401386247c0f702 }
---

# Files
- `internal/bmad/templates.go`
- `internal/bmad/templates_test.go`

# Symbols
- templates.go (internal/bmad/templates.go:L1)
- architectureReview() (internal/bmad/templates.go:L107)
- storyDevelopment() (internal/bmad/templates.go:L123)
- prdPipeline() (internal/bmad/templates.go:L139)
- qaAndPolish() (internal/bmad/templates.go:L156)
- rapidPrototype() (internal/bmad/templates.go:L172)
- fullInfra() (internal/bmad/templates.go:L187)
- TemplateByName() (internal/bmad/templates.go:L20)
- node() (internal/bmad/templates.go:L29)
- edge() (internal/bmad/templates.go:L40)
- chain() (internal/bmad/templates.go:L44)
- itoa() (internal/bmad/templates.go:L56)
- BuiltinTemplates() (internal/bmad/templates.go:L6)
- fullProductLifecycle() (internal/bmad/templates.go:L64)
- quickSprint() (internal/bmad/templates.go:L90)
- templates_test.go (internal/bmad/templates_test.go:L1)
- TestBuiltinTemplates_Count() (internal/bmad/templates_test.go:L10)
- TestAC3_BuiltinTemplateCount() (internal/bmad/templates_test.go:L102)
- TestAC3_RapidPrototypeTemplate() (internal/bmad/templates_test.go:L107)
- TestAC3_FullInfraTemplate() (internal/bmad/templates_test.go:L130)
- TestBuiltinTemplates_AllAreTemplates() (internal/bmad/templates_test.go:L15)
- TestAC4_TemplateChainValidation() (internal/bmad/templates_test.go:L156)
- TestTemplateByName_Missing() (internal/bmad/templates_test.go:L200)
- TestBuiltinTemplates_HaveNodesAndEdges() (internal/bmad/templates_test.go:L21)
- TestBuiltinTemplates_NodeProcessIDsExistInRegistry() (internal/bmad/templates_test.go:L28)
- TestBuiltinTemplates_UniqueIDs() (internal/bmad/templates_test.go:L37)
- TestBuiltinTemplates_UniqueNodeIDs() (internal/bmad/templates_test.go:L45)
- TestBuiltinTemplates_EdgeSourceTargetExist() (internal/bmad/templates_test.go:L55)
- TestBuiltinTemplates_AllNodesPending() (internal/bmad/templates_test.go:L68)
- TestQuickSprint_Structure() (internal/bmad/templates_test.go:L76)
- TestFullProductLifecycle_Structure() (internal/bmad/templates_test.go:L93)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
