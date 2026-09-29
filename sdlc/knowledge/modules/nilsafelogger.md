---
type: Module
title: nilSafeLogger
description: "Graphify community 20: internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapt"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_comprehensive_test, resource: internal/uiadapter/logging_comprehensive_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 446816f0c9fa2be3 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
---

# Files
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_comprehensive_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`

# Symbols
- .Classify() (internal/uiadapter/backend/claudeapi/client.go:L93)
- .runOneShot() (internal/uiadapter/backend/claudecli/client.go:L234)
- .Classify() (internal/uiadapter/backend/claudecli/client.go:L72)
- .Generate() (internal/uiadapter/backend/claudecli/client.go:L93)
- schemaBytes() (internal/uiadapter/encode.go:L112)
- OllamaFormatPayload() (internal/uiadapter/encode.go:L35)
- emitOllamaFormat() (internal/uiadapter/encode.go:L51)
- ClaudeToolInputSchema() (internal/uiadapter/encode.go:L71)
- TestOllamaFormatPayload_SchemaObject() (internal/uiadapter/encode_test.go:L15)
- TestOllamaFormatPayload_LooseRollback() (internal/uiadapter/encode_test.go:L28)
- TestClaudeToolInputSchema_SchemaObject() (internal/uiadapter/encode_test.go:L37)
- TestSchemas_IdenticalAcrossBackends() (internal/uiadapter/encode_test.go:L49)
- nilSafeLogger() (internal/uiadapter/logging.go:L169)
- TestStory6_AC1_NilSafeLoggerComprehensive() (internal/uiadapter/logging_comprehensive_test.go:L305)
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
- stages.go (internal/uiadapter/stages.go:L1)
- AssembleStage1() (internal/uiadapter/stages.go:L100)
- assembleStage1String() (internal/uiadapter/stages.go:L116)
- AssembleStage2() (internal/uiadapter/stages.go:L133)
- assembleStage2String() (internal/uiadapter/stages.go:L152)
- RunTwoStage() (internal/uiadapter/stages.go:L180)
- StageKind (internal/uiadapter/stages.go:L35)
- ParseStageKind() (internal/uiadapter/stages.go:L53)
- ClassifyPrompt() (internal/uiadapter/stages.go:L73)
- GeneratePromptFor() (internal/uiadapter/stages.go:L76)
- stages_test.go (internal/uiadapter/stages_test.go:L1)
- TestRunTwoStage_ClassifyErrorAborts() (internal/uiadapter/stages_test.go:L109)
- TestRunTwoStage_ContextCancelled() (internal/uiadapter/stages_test.go:L124)
- TestStageKind_ParseRoundtrip() (internal/uiadapter/stages_test.go:L17)
- TestAssembleStage1_ByteStable() (internal/uiadapter/stages_test.go:L36)
- TestAssembleStage2_ContainsKindDirective() (internal/uiadapter/stages_test.go:L50)
- TestAssembleStage2_UnknownKindErrors() (internal/uiadapter/stages_test.go:L61)
- TestPromptsBudget() (internal/uiadapter/stages_test.go:L69)
- TestRunTwoStage_OllamaPolicyTwoCalls() (internal/uiadapter/stages_test.go:L86)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
