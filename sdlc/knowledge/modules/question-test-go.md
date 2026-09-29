---
type: Module
title: question_test.go
description: "Graphify community 41: internal/bmad/fixture_verify_test.go, internal/bmad/question.go, internal/bmad/question_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: fixture_verify_test, resource: internal/bmad/fixture_verify_test.go, last_modified: "2026-04-11T19:45:53+10:00", digest: 6c143c3418c90221 }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: question_test, resource: internal/bmad/question_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 08a2a97c930cca46 }
---

# Files
- `internal/bmad/fixture_verify_test.go`
- `internal/bmad/question.go`
- `internal/bmad/question_test.go`

# Symbols
- fixture_verify_test.go (internal/bmad/fixture_verify_test.go:L1)
- TestDetectIdlePrompt_RealFixture() (internal/bmad/fixture_verify_test.go:L26)
- detectIdlePrompt() (internal/bmad/question.go:L191)
- hasRecentQuestion() (internal/bmad/question.go:L238)
- escapeTmuxLiteral() (internal/bmad/question.go:L276)
- stripANSI() (internal/bmad/question.go:L87)
- detectQuestion() (internal/bmad/question.go:L94)
- question_test.go (internal/bmad/question_test.go:L1)
- TestStory1_AC2_DetectFreeformQuestion() (internal/bmad/question_test.go:L109)
- TestStory1_AC3_DetectMenuQuestion() (internal/bmad/question_test.go:L172)
- questionTestWorkflow() (internal/bmad/question_test.go:L22)
- TestStory1_AC7_NoFalsePositives() (internal/bmad/question_test.go:L234)
- TestStory1_AC2_LastFiftyLinesOnly() (internal/bmad/question_test.go:L286)
- TestStory1_AC2_QuestionWithinLastFiftyLines() (internal/bmad/question_test.go:L302)
- TestStory1_AC2_MultipleQuestions_LastOneWins() (internal/bmad/question_test.go:L319)
- TestStory1_AC5_HashQuestion() (internal/bmad/question_test.go:L334)
- TestStory1_AC4_CaptureQuestionOutput_Success() (internal/bmad/question_test.go:L360)
- TestStory1_AC4_CaptureQuestionOutput_Error() (internal/bmad/question_test.go:L377)
- TestStory1_AC4_CaptureQuestionOutput_UsesShorterHistory() (internal/bmad/question_test.go:L391)
- TestStory1_AC6_DismissalOnComplete() (internal/bmad/question_test.go:L423)
- TestStory1_AC1_StripANSI() (internal/bmad/question_test.go:L46)
- TestStory1_AC6_DismissalOnFail() (internal/bmad/question_test.go:L473)
- TestStory1_AC5_DeduplicationSuppressesSameQuestion() (internal/bmad/question_test.go:L518)
- TestStory1_AC5_NewQuestionReplacesOld() (internal/bmad/question_test.go:L563)
- TestStory1_AC4_QuestionEventFields() (internal/bmad/question_test.go:L623)
- TestStory2_AC3_EscapeTmuxLiteral() (internal/bmad/question_test.go:L661)
- TestStory2_AC3_EscapeTmuxLiteral_StripsControlBytes() (internal/bmad/question_test.go:L704)
- TestDetectIdlePrompt() (internal/bmad/question_test.go:L740)
- TestHashCapturedOutput_StableAndDifferentiates() (internal/bmad/question_test.go:L828)
- newIdleTestState() (internal/bmad/question_test.go:L847)
- TestPollForIdle_EmitsAfterStableIdleFrame() (internal/bmad/question_test.go:L859)
- TestPollForIdle_DismissesOnOutputChange() (internal/bmad/question_test.go:L892)
- TestPollForIdle_DoesNotEmitWhenNotIdle() (internal/bmad/question_test.go:L930)
- TestPollForIdle_EmitsIdleEventPayload() (internal/bmad/question_test.go:L947)
- TestHasRecentQuestion() (internal/bmad/question_test.go:L971)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [Executor](/modules/executor.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
