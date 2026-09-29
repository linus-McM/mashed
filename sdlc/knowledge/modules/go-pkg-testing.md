---
type: Module
title: go_pkg_testing
description: "Graphify community 13: app_bmad_question_test.go, app_shutdown_test.go, app_uiadapter.go, app_uiadapter_claudecli.go, app_uiadapter_v3.go, internal/bmad/executor.go, internal/bmad/executor_anyuseransw"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: app_bmad_question_test, resource: app_bmad_question_test.go, last_modified: "2026-04-10T14:01:16+10:00", digest: 9386c29f8c1272f8 }
  - { id: app_shutdown_test, resource: app_shutdown_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: bb2b11e788c7ac0b }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-04-23T11:43:31+10:00", digest: c6ede9f0c3335d2e }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_flatten_test, resource: internal/bmad/executor_flatten_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 620dfc16866d8fff }
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
  - { id: registry_interactive_phase2_helper_test, resource: internal/bmad/registry_interactive_phase2_helper_test.go, last_modified: "2026-04-28T11:16:27+10:00", digest: b442030b66b1e92a }
  - { id: shell_quote_test, resource: internal/bmad/shell_quote_test.go, last_modified: "2026-04-12T12:32:31+10:00", digest: a1f5b92816b5a5c6 }
  - { id: validate_test, resource: internal/bmad/validate_test.go, last_modified: "2026-04-21T09:23:33+10:00", digest: 78fb1bf63e206f93 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 417a5e0974dc194b }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: logging_handler_test, resource: internal/uiadapter/logging_handler_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 998f6bb512a268ee }
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: bce0a1e9685e2603 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: schema, resource: internal/uiadapter/schema.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 31cb03cd04da1f49 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
  - { id: semaphore_test, resource: internal/uiadapter/semaphore_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8d59aeba4a59bab6 }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 82f02eccfdc52095 }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `app_bmad_question_test.go`
- `app_shutdown_test.go`
- `app_uiadapter.go`
- `app_uiadapter_claudecli.go`
- `app_uiadapter_v3.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_flatten_test.go`
- `internal/bmad/mock_helpers_test.go`
- `internal/bmad/registry_interactive_phase2_helper_test.go`
- `internal/bmad/shell_quote_test.go`
- `internal/bmad/validate_test.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/adapter_test.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/registry_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/logging_handler_test.go`
- `internal/uiadapter/logging_plumbing_mock_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/repair.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/schema.go`
- `internal/uiadapter/semaphore.go`
- `internal/uiadapter/semaphore_test.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- app_bmad_question_test.go (app_bmad_question_test.go:L1)
- TestStory2_AC6_AppRespondToQuestion_NilExecutor() (app_bmad_question_test.go:L59)
- app_shutdown_test.go (app_shutdown_test.go:L1)
- TestStory1_AC6_AppShutdownDrainsHooks() (app_shutdown_test.go:L28)
- app_uiadapter.go (app_uiadapter.go:L1)
- tagsListResponse (app_uiadapter.go:L141)
- init() (app_uiadapter.go:L34)
- app_uiadapter_claudecli.go (app_uiadapter_claudecli.go:L1)
- app_uiadapter_v3.go (app_uiadapter_v3.go:L1)
- wrapInBashExec() (internal/bmad/executor.go:L1191)
- executor_anyuseranswer_test.go (internal/bmad/executor_anyuseranswer_test.go:L1)
- executor_flatten_test.go (internal/bmad/executor_flatten_test.go:L1)
- mock_helpers_test.go (internal/bmad/mock_helpers_test.go:L1)
- registry_interactive_phase2_helper_test.go (internal/bmad/registry_interactive_phase2_helper_test.go:L1)
- shell_quote_test.go (internal/bmad/shell_quote_test.go:L1)
- TestWrapInBashExec() (internal/bmad/shell_quote_test.go:L9)
- validate_test.go (internal/bmad/validate_test.go:L1)
- adapter.go (internal/uiadapter/adapter.go:L1)
- adapter_test.go (internal/uiadapter/adapter_test.go:L1)
- backend.go (internal/uiadapter/backend/backend.go:L1)
- claudeapi/client.go (internal/uiadapter/backend/claudeapi/client.go:L1)
- init() (internal/uiadapter/backend/claudeapi/client.go:L302)
- claudecli/client_test.go (internal/uiadapter/backend/claudecli/client_test.go:L1)
- lifecycle.go (internal/uiadapter/backend/lifecycle.go:L1)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- stubs.go (internal/uiadapter/backend/stubs.go:L1)
- breaker.go (internal/uiadapter/breaker.go:L1)
- uiadapter/client.go (internal/uiadapter/client.go:L1)
- HTTPStatusError (internal/uiadapter/client.go:L29)
- .Error() (internal/uiadapter/client.go:L34)
- chatMessage (internal/uiadapter/client.go:L61)
- chatRequest (internal/uiadapter/client.go:L66)
- chatResponse (internal/uiadapter/client.go:L74)
- tagsResponse (internal/uiadapter/client.go:L80)
- config_test.go (internal/uiadapter/config_test.go:L1)
- contextguard.go (internal/uiadapter/contextguard.go:L1)
- corpus_test.go (internal/uiadapter/eval/corpus_test.go:L1)
- eval_test.go (internal/uiadapter/eval_test.go:L1)
- fallback.go (internal/uiadapter/fallback.go:L1)
- fallback_tiers.go (internal/uiadapter/fallback_tiers.go:L1)
- logging_handler_test.go (internal/uiadapter/logging_handler_test.go:L1)
- logging_plumbing_mock_test.go (internal/uiadapter/logging_plumbing_mock_test.go:L1)
- logging_plumbing_test.go (internal/uiadapter/logging_plumbing_test.go:L1)
- logging_story3_sanitize_test.go (internal/uiadapter/logging_story3_sanitize_test.go:L1)
- captureWriter (internal/uiadapter/logging_story3_sanitize_test.go:L247)
- .Write() (internal/uiadapter/logging_story3_sanitize_test.go:L252)
- logging_story4_sanitize_test.go (internal/uiadapter/logging_story4_sanitize_test.go:L1)
- mock_test.go (internal/uiadapter/mock_test.go:L1)
- repair.go (internal/uiadapter/repair.go:L1)
- sanitize.go (internal/uiadapter/sanitize.go:L1)
- sanitize_adapter_test.go (internal/uiadapter/sanitize_adapter_test.go:L1)
- schema.go (internal/uiadapter/schema.go:L1)
- Diagnostics (internal/uiadapter/schema.go:L25)
- UINode (internal/uiadapter/schema.go:L38)
- WidgetOption (internal/uiadapter/schema.go:L76)
- semaphore.go (internal/uiadapter/semaphore.go:L1)
- semaphore_test.go (internal/uiadapter/semaphore_test.go:L1)
- spotlight.go (internal/uiadapter/spotlight.go:L1)
- translate_e2e_test.go (internal/uiadapter/translate_e2e_test.go:L1)

# Depends on
- [allowlist_test.go](/modules/allowlist-test-go.md)
- [applyIterativeUpgrade](/modules/applyiterativeupgrade.md)
- [backend/registry.go](/modules/backend-registry-go.md)
- [BreakerSet](/modules/breakerset.md)
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [fastpath.go](/modules/fastpath-go.md)
- [gate.go](/modules/gate-go.md)
- [loadConfig](/modules/loadconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [newHarness](/modules/newharness.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [sessions.go](/modules/sessions-go.md)
- [setupTestConfig](/modules/setuptestconfig.md)
- [Spotlight](/modules/spotlight.md)
- [testing.T](/modules/testing-t.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [TestStory2_AC2_FreeFunctionsAcceptNilLogger](/modules/teststory2-ac2-freefunctionsacceptnillogger.md)
- [time.Duration](/modules/time-duration.md)
- [.Translate](/modules/translate.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)
- [validate.go](/modules/validate-go.md)
- [wait_idle_test.go](/modules/wait-idle-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
