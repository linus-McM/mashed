---
type: Module
title: Validate
description: "Graphify community 51: internal/uiadapter/sanitize_adapter_test.go, internal/uiadapter/validator.go, internal/uiadapter/validator_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: validator, resource: internal/uiadapter/validator.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ea853a4a6aadb5f2 }
  - { id: validator_test, resource: internal/uiadapter/validator_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ad90be2e31d6a4a0 }
---

# Files
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/validator.go`
- `internal/uiadapter/validator_test.go`

# Symbols
- TestAdapter_Translate_ANSIWrappedURL_NotUntrusted() (internal/uiadapter/sanitize_adapter_test.go:L21)
- validator.go (internal/uiadapter/validator.go:L1)
- emitValidatorDone() (internal/uiadapter/validator.go:L107)
- emitRuleFail() (internal/uiadapter/validator.go:L123)
- dedupeAndSort() (internal/uiadapter/validator.go:L138)
- terminalReason() (internal/uiadapter/validator.go:L158)
- applyRule1() (internal/uiadapter/validator.go:L170)
- applyRules2And3() (internal/uiadapter/validator.go:L197)
- applyRule4() (internal/uiadapter/validator.go:L221)
- applyRule5() (internal/uiadapter/validator.go:L244)
- applyRule6() (internal/uiadapter/validator.go:L266)
- applyRule7() (internal/uiadapter/validator.go:L293)
- applyRule8() (internal/uiadapter/validator.go:L312)
- knownNodeType() (internal/uiadapter/validator.go:L327)
- contentPreserved() (internal/uiadapter/validator.go:L340)
- collectRenderedText() (internal/uiadapter/validator.go:L364)
- Validate() (internal/uiadapter/validator.go:L59)
- validator_test.go (internal/uiadapter/validator_test.go:L1)
- TestValidate_Rule5_LongKeyTruncated() (internal/uiadapter/validator_test.go:L111)
- TestValidate_Rule6_CountCapOptionalDrop() (internal/uiadapter/validator_test.go:L130)
- TestValidate_Rule1_UnknownTypeBecomesMarkdown() (internal/uiadapter/validator_test.go:L16)
- TestValidate_Rule6_CountCapRequiredDropFallback() (internal/uiadapter/validator_test.go:L160)
- TestValidate_Rule7_NodeCapRequiredDropFallback() (internal/uiadapter/validator_test.go:L179)
- TestValidate_Rule8_OversizeFullFallback() (internal/uiadapter/validator_test.go:L199)
- TestValidate_RuleOrdering_Deterministic() (internal/uiadapter/validator_test.go:L214)
- TestValidate_Security71_FileWidgetPreserved() (internal/uiadapter/validator_test.go:L247)
- TestContentPreservation_URLDropped_SetsUntrusted() (internal/uiadapter/validator_test.go:L278)
- TestContentPreservation_CodeBlockDropped_SetsUntrusted() (internal/uiadapter/validator_test.go:L293)
- TestContentPreservation_NumberedListToChoice_NoUntrusted() (internal/uiadapter/validator_test.go:L308)
- TestValidate_Rule2_EmptyOptionsDropsGroup() (internal/uiadapter/validator_test.go:L33)
- TestValidate_RuleOrdering_PerNodeDeclaration() (internal/uiadapter/validator_test.go:L334)
- TestValidate_WidgetUnknownField_Rejected() (internal/uiadapter/validator_test.go:L357)
- TestValidate_Rule3_NoWidgetDropsGroup() (internal/uiadapter/validator_test.go:L63)
- TestValidate_Rule4_DuplicateKeySuffix() (internal/uiadapter/validator_test.go:L76)

# Depends on
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [go_pkg_log_slog](/modules/go-pkg-log-slog.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files
