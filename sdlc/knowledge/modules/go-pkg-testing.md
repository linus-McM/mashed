---
type: Module
title: go_pkg_testing
description: "Graphify community 14: app_bmad_question_test.go, app_git.go, app_shutdown_test.go, internal/bmad/executor.go, internal/bmad/executor_anyuseranswer_test.go, internal/bmad/executor_cleanup_test.go, int"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_bmad_question_test, resource: app_bmad_question_test.go, last_modified: "2026-04-10T14:01:16+10:00", digest: 9386c29f8c1272f8 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T01:11:52+10:00", digest: f61e71fb9e1ebdc6 }
  - { id: app_shutdown_test, resource: app_shutdown_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: bb2b11e788c7ac0b }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_cleanup_test, resource: internal/bmad/executor_cleanup_test.go, last_modified: "2026-04-12T15:59:00+10:00", digest: b58becad5281bc85 }
  - { id: executor_flatten_test, resource: internal/bmad/executor_flatten_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 620dfc16866d8fff }
  - { id: fixture_verify_test, resource: internal/bmad/fixture_verify_test.go, last_modified: "2026-04-11T19:45:53+10:00", digest: 6c143c3418c90221 }
  - { id: registry_interactive_phase2_helper_test, resource: internal/bmad/registry_interactive_phase2_helper_test.go, last_modified: "2026-04-28T11:16:27+10:00", digest: b442030b66b1e92a }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
  - { id: shell_quote_test, resource: internal/bmad/shell_quote_test.go, last_modified: "2026-04-12T12:32:31+10:00", digest: a1f5b92816b5a5c6 }
  - { id: validate, resource: internal/bmad/validate.go, last_modified: "2026-04-21T09:23:33+10:00", digest: fddd00a4023adcb7 }
  - { id: validate_test, resource: internal/bmad/validate_test.go, last_modified: "2026-04-21T09:23:33+10:00", digest: 78fb1bf63e206f93 }
  - { id: refs_test, resource: internal/git/refs_test.go, last_modified: "2026-09-30T00:58:21+10:00", digest: 557cff720e95f689 }
  - { id: bridge_auth_dev_test, resource: internal/terminal/bridge_auth_dev_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 55bf715a2b15fa81 }
  - { id: bridge_origin_prod_test, resource: internal/terminal/bridge_origin_prod_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 8eb122b662553df1 }
  - { id: protocol_test, resource: internal/terminal/helper/protocol_test.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 500bab47812315f3 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: logging_handler_test, resource: internal/uiadapter/logging_handler_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 998f6bb512a268ee }
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: bce0a1e9685e2603 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: semaphore_test, resource: internal/uiadapter/semaphore_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8d59aeba4a59bab6 }
  - { id: testhelper_eval, resource: internal/uiadapter/testhelper_eval.go, last_modified: "2026-04-22T14:13:03+10:00", digest: d5eeb86f7290c4bf }
  - { id: readfilebase64_test, resource: readfilebase64_test.go, last_modified: "2026-09-30T00:48:40+10:00", digest: bf478ed246b9cbbd }
  - { id: screenshot_test, resource: screenshot_test.go, last_modified: "2026-04-21T10:01:10+10:00", digest: ba0791de42883540 }
---

# Files
- `app_bmad_question_test.go`
- `app_git.go`
- `app_shutdown_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_cleanup_test.go`
- `internal/bmad/executor_flatten_test.go`
- `internal/bmad/fixture_verify_test.go`
- `internal/bmad/registry_interactive_phase2_helper_test.go`
- `internal/bmad/registry_interactive_test.go`
- `internal/bmad/shell_quote_test.go`
- `internal/bmad/validate.go`
- `internal/bmad/validate_test.go`
- `internal/git/refs_test.go`
- `internal/terminal/bridge_auth_dev_test.go`
- `internal/terminal/bridge_origin_prod_test.go`
- `internal/terminal/helper/protocol_test.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/logging_handler_test.go`
- `internal/uiadapter/logging_plumbing_mock_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/semaphore_test.go`
- `internal/uiadapter/testhelper_eval.go`
- `readfilebase64_test.go`
- `screenshot_test.go`

# Symbols
- app_bmad_question_test.go (app_bmad_question_test.go:L1)
- mimeForExt() (app_git.go:L888)
- app_shutdown_test.go (app_shutdown_test.go:L1)
- wrapInBashExec() (internal/bmad/executor.go:L1191)
- executor_anyuseranswer_test.go (internal/bmad/executor_anyuseranswer_test.go:L1)
- executor_cleanup_test.go (internal/bmad/executor_cleanup_test.go:L1)
- executor_flatten_test.go (internal/bmad/executor_flatten_test.go:L1)
- fixture_verify_test.go (internal/bmad/fixture_verify_test.go:L1)
- registry_interactive_phase2_helper_test.go (internal/bmad/registry_interactive_phase2_helper_test.go:L1)
- registry_interactive_test.go (internal/bmad/registry_interactive_test.go:L1)
- shell_quote_test.go (internal/bmad/shell_quote_test.go:L1)
- TestWrapInBashExec() (internal/bmad/shell_quote_test.go:L9)
- validateInput() (internal/bmad/validate.go:L17)
- containsString() (internal/bmad/validate.go:L79)
- resolveFileInput() (internal/bmad/validate.go:L95)
- validate_test.go (internal/bmad/validate_test.go:L1)
- TestResolveFileInput() (internal/bmad/validate_test.go:L221)
- TestValidateInput() (internal/bmad/validate_test.go:L23)
- TestU0_AC4_ValidateInput_ShapeJSON_BareStringAccepted() (internal/bmad/validate_test.go:L308)
- refs_test.go (internal/git/refs_test.go:L1)
- bridge_auth_dev_test.go (internal/terminal/bridge_auth_dev_test.go:L1)
- bridge_origin_prod_test.go (internal/terminal/bridge_origin_prod_test.go:L1)
- protocol_test.go (internal/terminal/helper/protocol_test.go:L1)
- claudecli/client_test.go (internal/uiadapter/backend/claudecli/client_test.go:L1)
- config_test.go (internal/uiadapter/config_test.go:L1)
- encode_test.go (internal/uiadapter/encode_test.go:L1)
- corpus_test.go (internal/uiadapter/eval/corpus_test.go:L1)
- eval_test.go (internal/uiadapter/eval_test.go:L1)
- TestEval_SkipWhenUnreachable() (internal/uiadapter/eval_test.go:L29)
- logging_handler_test.go (internal/uiadapter/logging_handler_test.go:L1)
- logging_plumbing_mock_test.go (internal/uiadapter/logging_plumbing_mock_test.go:L1)
- logging_plumbing_test.go (internal/uiadapter/logging_plumbing_test.go:L1)
- logging_story3_sanitize_test.go (internal/uiadapter/logging_story3_sanitize_test.go:L1)
- captureWriter (internal/uiadapter/logging_story3_sanitize_test.go:L247)
- .Write() (internal/uiadapter/logging_story3_sanitize_test.go:L252)
- logging_story4_sanitize_test.go (internal/uiadapter/logging_story4_sanitize_test.go:L1)
- logging_story5_sanitize_test.go (internal/uiadapter/logging_story5_sanitize_test.go:L1)
- mock_test.go (internal/uiadapter/mock_test.go:L1)
- sampling_test.go (internal/uiadapter/sampling_test.go:L1)
- sanitize_adapter_test.go (internal/uiadapter/sanitize_adapter_test.go:L1)
- semaphore_test.go (internal/uiadapter/semaphore_test.go:L1)
- testhelper_eval.go (internal/uiadapter/testhelper_eval.go:L1)
- WithOllamaHost() (internal/uiadapter/testhelper_eval.go:L13)
- readfilebase64_test.go (readfilebase64_test.go:L1)
- TestReadFileBase64_AC2_MimeTypes() (readfilebase64_test.go:L40)
- screenshot_test.go (screenshot_test.go:L1)

# Depends on
- [App](/modules/app-73.md)
- [applyIterativeUpgrade](/modules/applyiterativeupgrade.md)
- [bridge_auth_test.go](/modules/bridge-auth-test-go.md)
- [cleanup_test.go](/modules/cleanup-test-go.md)
- [Config](/modules/config.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [Executor](/modules/executor.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [gate.go](/modules/gate-go.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewMock](/modules/newmock.md)
- [ProcessByID](/modules/processbyid.md)
- [.resolveInputs](/modules/resolveinputs.md)
- [server_test.go](/modules/server-test-go.md)
- [testing.T](/modules/testing-t.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
