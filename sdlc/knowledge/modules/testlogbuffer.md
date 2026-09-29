---
type: Module
title: testLogBuffer
description: "Graphify community 23: internal/uiadapter/allowlist_test.go, internal/uiadapter/breaker_test.go, internal/uiadapter/cache_test.go, internal/uiadapter/contextguard_test.go, internal/uiadapter/encode_te"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 865e6c967ac886df }
  - { id: breaker_test, resource: internal/uiadapter/breaker_test.go, last_modified: "2026-09-29T07:07:25Z", digest: c08c33460a000e4b }
  - { id: cache_test, resource: internal/uiadapter/cache_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e8cf326b1dcc1d7a }
  - { id: contextguard_test, resource: internal/uiadapter/contextguard_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 0cf07afee9915c31 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-09-29T07:07:25Z", digest: a0cd96e6b9f32bfb }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 2fa70a931cc96a6b }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 73bb2c32262274bd }
  - { id: fastpath_test, resource: internal/uiadapter/fastpath_test.go, last_modified: "2026-09-29T07:07:25Z", digest: a7f912e39ab3f4b0 }
  - { id: log_test_helper_test, resource: internal/uiadapter/log_test_helper_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e2a6968886cbeb48 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 49f3cb55ded4bd06 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e6b7941070a1bd3a }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 448153a81ce2894c }
  - { id: repair_test, resource: internal/uiadapter/repair_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 7d95d0c21f4ee4e3 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 10d6cf4c605c03f7 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-09-29T07:07:25Z", digest: c90f1e17fbe3e70f }
  - { id: schema_test, resource: internal/uiadapter/schema_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 706f4f8efc89d9d5 }
  - { id: semaphore_test, resource: internal/uiadapter/semaphore_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 8d59aeba4a59bab6 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-09-29T07:07:25Z", digest: f9f81a13aa4a0344 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 74ccbf42bffba71e }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 45f6254d2ea3d0bc }
  - { id: validator_test, resource: internal/uiadapter/validator_test.go, last_modified: "2026-09-29T07:07:25Z", digest: ad90be2e31d6a4a0 }
---

# Files
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/breaker_test.go`
- `internal/uiadapter/cache_test.go`
- `internal/uiadapter/contextguard_test.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/fastpath_test.go`
- `internal/uiadapter/log_test_helper_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/repair_test.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/sanitize_test.go`
- `internal/uiadapter/schema_test.go`
- `internal/uiadapter/semaphore_test.go`
- `internal/uiadapter/spotlight_test.go`
- `internal/uiadapter/stages_test.go`
- `internal/uiadapter/translate_e2e_test.go`
- `internal/uiadapter/validator_test.go`

# Symbols
- TestStory5_AC6_AllowlistOK() (internal/uiadapter/allowlist_test.go:L71)
- TestStory3_AC4_BreakerTransitionAndReject() (internal/uiadapter/breaker_test.go:L99)
- TestStory3_AC3_CacheLifecycle() (internal/uiadapter/cache_test.go:L174)
- TestStory5_AC3_ApplyOllamaUnderBudget() (internal/uiadapter/contextguard_test.go:L135)
- TestStory5_AC3_ApplyClaudeTruncated() (internal/uiadapter/contextguard_test.go:L163)
- TestStory5_AC3_ApplyOllamaTruncated() (internal/uiadapter/contextguard_test.go:L94)
- TestStory5_AC5_EncodeClaudeToolName() (internal/uiadapter/encode_test.go:L114)
- TestStory5_AC5_EncodeSchemaSelect() (internal/uiadapter/encode_test.go:L132)
- TestStory5_AC5_EncodeOllamaFormat() (internal/uiadapter/encode_test.go:L72)
- TestStory5_AC5_EncodeClaudeToolSchema() (internal/uiadapter/encode_test.go:L93)
- TestStory4_AC6_FallbackASTTruncate() (internal/uiadapter/fallback_test.go:L108)
- TestStory4_AC6_FallbackASTBuild() (internal/uiadapter/fallback_test.go:L76)
- TestStory4_AC6_TierThirdSuccess() (internal/uiadapter/fallback_tiers_test.go:L124)
- TestStory4_AC6_TierAllExhausted() (internal/uiadapter/fallback_tiers_test.go:L187)
- TestStory4_AC1_FastpathHit() (internal/uiadapter/fastpath_test.go:L154)
- TestStory4_AC1_FastpathSkip() (internal/uiadapter/fastpath_test.go:L194)
- TestStory4_AC1_FastpathDisabled() (internal/uiadapter/fastpath_test.go:L219)
- testLogBuffer() (internal/uiadapter/log_test_helper_test.go:L28)
- decodeRecords() (internal/uiadapter/log_test_helper_test.go:L41)
- recordsByMsg() (internal/uiadapter/log_test_helper_test.go:L64)
- recordMsgsWithPrefix() (internal/uiadapter/log_test_helper_test.go:L78)
- TestStory5_AC8_PerFileEmissionCoverage() (internal/uiadapter/logging_story5_sanitize_test.go:L102)
- TestStory5_AC5_MockInit() (internal/uiadapter/mock_test.go:L64)
- TestStory5_AC5_MockTranslate() (internal/uiadapter/mock_test.go:L99)
- TestStory3_AC6_OllamaKeepAlive() (internal/uiadapter/prefix_cache_test.go:L137)
- TestStory4_AC2_RepairExhausted() (internal/uiadapter/repair_test.go:L175)
- TestStory4_AC3_RepairSuccessOnFirstTry() (internal/uiadapter/repair_test.go:L237)
- TestStory4_AC2_BuildRepairPromptEmits() (internal/uiadapter/repair_test.go:L281)
- TestStory5_AC5_SamplingOllama() (internal/uiadapter/sampling_test.go:L45)
- TestStory5_AC5_SamplingClaude() (internal/uiadapter/sampling_test.go:L63)
- TestStory5_AC1_SanitizeStartAndDone() (internal/uiadapter/sanitize_test.go:L110)
- TestStory5_AC7_WidgetNodeUnmarshalEmits() (internal/uiadapter/schema_test.go:L197)
- TestStory3_AC5_SemaphoreWaitAndAcquired() (internal/uiadapter/semaphore_test.go:L31)
- TestStory3_AC5_SemaphoreCancelled() (internal/uiadapter/semaphore_test.go:L77)
- TestStory5_AC2_SpotlightDisabled() (internal/uiadapter/spotlight_test.go:L111)
- TestStory5_AC2_UnspotlightRemoved() (internal/uiadapter/spotlight_test.go:L141)
- TestStory5_AC2_SpotlightAdded() (internal/uiadapter/spotlight_test.go:L75)
- TestStory4_AC4_TwoStageHappyPath() (internal/uiadapter/stages_test.go:L167)
- TestStory4_AC5_StagesParseError() (internal/uiadapter/stages_test.go:L226)
- TestStory4_AC4_AssembleEmits() (internal/uiadapter/stages_test.go:L259)
- TestStory4_AC4_ParseStageKindRejected() (internal/uiadapter/stages_test.go:L291)
- TestStory6_AC2_Translate_RepairTriggered() (internal/uiadapter/translate_e2e_test.go:L200)
- TestStory6_AC2_Translate_TierEscalation() (internal/uiadapter/translate_e2e_test.go:L253)
- TestStory5_AC4_ValidatorAggregateAndPerRule() (internal/uiadapter/validator_test.go:L375)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [FallbackAST](/modules/fallbackast.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [NewMock](/modules/newmock.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [ResponseCache](/modules/responsecache.md)
- [stages_test.go](/modules/stages-test-go.md)

# Features
- no feature plan names these files
