---
type: Module
title: mashed/internal/uiadapter.UIAST
description: "Graphify community 38: internal/bmad/registry_fs.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudecli/client.go, internal/uia"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: corpus, resource: internal/uiadapter/eval/corpus.go, last_modified: "2026-04-22T14:13:03+10:00", digest: be80b9817e3d70f3 }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
---

# Files
- `internal/bmad/registry_fs.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/eval/corpus.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/prompt.go`

# Symbols
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- backend.go (internal/uiadapter/backend/backend.go:L1)
- Capabilities (internal/uiadapter/backend/backend.go:L45)
- .Capabilities() (internal/uiadapter/backend/claudeapi/client.go:L65)
- .Capabilities() (internal/uiadapter/backend/claudecli/client.go:L44)
- stubs.go (internal/uiadapter/backend/stubs.go:L1)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .Generate() (internal/uiadapter/backend/stubs.go:L39)
- .GenerateSingleShot() (internal/uiadapter/backend/stubs.go:L44)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- synthUIAST() (internal/uiadapter/backend/stubs.go:L77)
- corpus.go (internal/uiadapter/eval/corpus.go:L1)
- Fixture (internal/uiadapter/eval/corpus.go:L21)
- Expected (internal/uiadapter/eval/corpus.go:L30)
- LoadCorpus() (internal/uiadapter/eval/corpus.go:L55)
- collectFixtureIDs() (internal/uiadapter/eval/corpus.go:L69)
- loadFixture() (internal/uiadapter/eval/corpus.go:L89)
- TestEval_Corpus_MinimumCount() (internal/uiadapter/eval/corpus_test.go:L13)
- scorecard.go (internal/uiadapter/eval/scorecard.go:L1)
- .recordPreservation() (internal/uiadapter/eval/scorecard.go:L112)
- classifyGeneratedBy() (internal/uiadapter/eval/scorecard.go:L246)
- inferModel() (internal/uiadapter/eval/scorecard.go:L259)
- widgetTypesFrom() (internal/uiadapter/eval/scorecard.go:L267)
- renderedText() (internal/uiadapter/eval/scorecard.go:L280)
- toSet() (internal/uiadapter/eval/scorecard.go:L300)
- Score() (internal/uiadapter/eval/scorecard.go:L53)
- .record() (internal/uiadapter/eval/scorecard.go:L73)
- .recordWidgets() (internal/uiadapter/eval/scorecard.go:L95)
- scorecard_test.go (internal/uiadapter/eval/scorecard_test.go:L1)
- TestEval_PerWidgetPrecisionRecall() (internal/uiadapter/eval/scorecard_test.go:L104)
- TestEval_MeetsThresholds_Table() (internal/uiadapter/eval/scorecard_test.go:L178)
- sequentialMockAdapter (internal/uiadapter/eval/scorecard_test.go:L19)
- .Translate() (internal/uiadapter/eval/scorecard_test.go:L24)
- TestEval_Preservation_URLAndCodeBlockCounts() (internal/uiadapter/eval/scorecard_test.go:L249)
- newSynthFixture() (internal/uiadapter/eval/scorecard_test.go:L33)
- uniformLatencies() (internal/uiadapter/eval/scorecard_test.go:L331)
- astOllama() (internal/uiadapter/eval/scorecard_test.go:L44)
- astValidatorFailed() (internal/uiadapter/eval/scorecard_test.go:L53)
- astMalformed() (internal/uiadapter/eval/scorecard_test.go:L62)
- TestEval_Scorecard_Metrics() (internal/uiadapter/eval/scorecard_test.go:L72)
- TestEval_FullCorpus_MeetsThresholds() (internal/uiadapter/eval_test.go:L52)
- prompt.go (internal/uiadapter/prompt.go:L1)
- PromptVersion() (internal/uiadapter/prompt.go:L17)

# Depends on
- [backend/registry_test.go](/modules/backend-registry-test-go.md)
- [Config](/modules/config.md)
- [testing.T](/modules/testing-t.md)
- [time.Duration](/modules/time-duration.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
