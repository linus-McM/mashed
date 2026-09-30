---
type: Module
title: go_pkg_os
description: "Graphify community 7: app.go, app_bmad.go, app_claude.go, app_explain.go, app_git.go, app_models.go, app_review.go, app_review_scoped.go, app_scan.go, app_sessions.go, app_spawn.go, app_terminal_regis"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-29T07:07:25Z", digest: 295875db4f0bdc1e }
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-09-29T07:07:25Z", digest: fca7c61a439c150e }
  - { id: app_claude, resource: app_claude.go, last_modified: "2026-09-29T07:07:25Z", digest: aff648c4f900ca53 }
  - { id: app_explain, resource: app_explain.go, last_modified: "2026-09-29T07:07:25Z", digest: 4796af23d019c064 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-29T07:07:25Z", digest: 6abf9d08d507b1db }
  - { id: app_models, resource: app_models.go, last_modified: "2026-09-29T07:07:25Z", digest: cf387264112e9ff8 }
  - { id: app_review, resource: app_review.go, last_modified: "2026-09-29T07:07:25Z", digest: f186f322bd66e914 }
  - { id: app_review_scoped, resource: app_review_scoped.go, last_modified: "2026-09-29T07:07:25Z", digest: 83c692114f39a0c9 }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-09-29T07:07:25Z", digest: e5b2c4798c9f1cad }
  - { id: app_sessions, resource: app_sessions.go, last_modified: "2026-09-29T07:07:25Z", digest: 7b79f8178152c1fb }
  - { id: app_spawn, resource: app_spawn.go, last_modified: "2026-09-29T07:07:25Z", digest: cbe44d5e06097d50 }
  - { id: app_terminal_registry, resource: app_terminal_registry.go, last_modified: "2026-09-29T07:07:25Z", digest: 1c83e41a5c27d12e }
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-09-29T07:07:25Z", digest: 0fcf7b7f19ac0972 }
  - { id: assets_write, resource: internal/bmad/assets_write.go, last_modified: "2026-09-29T07:07:25Z", digest: 4f3137dec4993601 }
  - { id: condition, resource: internal/bmad/condition.go, last_modified: "2026-09-29T07:07:25Z", digest: 956ffd61999ceb65 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-09-29T07:07:25Z", digest: 244fdd7f469d5870 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 38021efc1dc00f72 }
  - { id: fixture_verify_test, resource: internal/bmad/fixture_verify_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 6c143c3418c90221 }
  - { id: shell_quote_test, resource: internal/bmad/shell_quote_test.go, last_modified: "2026-09-29T07:07:25Z", digest: a1f5b92816b5a5c6 }
  - { id: skillgen, resource: internal/bmad/skillgen.go, last_modified: "2026-09-29T07:07:25Z", digest: 925f4004ca59e699 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-09-29T07:07:25Z", digest: 360ebf80068d1480 }
  - { id: validate, resource: internal/bmad/validate.go, last_modified: "2026-09-29T07:07:25Z", digest: fddd00a4023adcb7 }
  - { id: validate_test, resource: internal/bmad/validate_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 78fb1bf63e206f93 }
  - { id: diff, resource: internal/git/diff.go, last_modified: "2026-09-29T07:07:25Z", digest: 6abe6e537093b3dc }
  - { id: claude, resource: internal/scanner/claude.go, last_modified: "2026-09-29T07:07:25Z", digest: 2c44ec40e17d728e }
  - { id: watcher, resource: internal/scanner/watcher.go, last_modified: "2026-09-29T07:07:25Z", digest: 8aeb13ebec4f12a0 }
  - { id: login_path, resource: internal/terminal/login_path.go, last_modified: "2026-09-29T07:07:25Z", digest: 7ad84d121b34092b }
  - { id: signal_unix, resource: internal/uiadapter/backend/claudecli/signal_unix.go, last_modified: "2026-09-29T07:07:25Z", digest: 60897ce7552e8510 }
  - { id: signal_windows, resource: internal/uiadapter/backend/claudecli/signal_windows.go, last_modified: "2026-09-29T07:07:25Z", digest: d0934456df766740 }
  - { id: readfilebase64_test, resource: readfilebase64_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 82165929a5efaa1e }
  - { id: screenshot_fullstack_test, resource: screenshot_fullstack_test.go, last_modified: "2026-09-29T07:07:25Z", digest: da72543a3716b145 }
  - { id: screenshot_test, resource: screenshot_test.go, last_modified: "2026-09-29T07:07:25Z", digest: ba0791de42883540 }
---

# Files
- `app.go`
- `app_bmad.go`
- `app_claude.go`
- `app_explain.go`
- `app_git.go`
- `app_models.go`
- `app_review.go`
- `app_review_scoped.go`
- `app_scan.go`
- `app_sessions.go`
- `app_spawn.go`
- `app_terminal_registry.go`
- `internal/bmad/artifacts.go`
- `internal/bmad/assets_write.go`
- `internal/bmad/condition.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/fixture_verify_test.go`
- `internal/bmad/shell_quote_test.go`
- `internal/bmad/skillgen.go`
- `internal/bmad/storage.go`
- `internal/bmad/validate.go`
- `internal/bmad/validate_test.go`
- `internal/git/diff.go`
- `internal/scanner/claude.go`
- `internal/scanner/watcher.go`
- `internal/terminal/login_path.go`
- `internal/uiadapter/backend/claudecli/signal_unix.go`
- `internal/uiadapter/backend/claudecli/signal_windows.go`
- `readfilebase64_test.go`
- `screenshot_fullstack_test.go`
- `screenshot_test.go`

# Symbols
- app.go (app.go:L1)
- .TakeScreenshot() (app.go:L431)
- ensureGitignoreEntry() (app.go:L470)
- VSCodeThemeEntry (app.go:L82)
- app_bmad.go (app_bmad.go:L1)
- app_claude.go (app_claude.go:L1)
- app_explain.go (app_explain.go:L1)
- app_git.go (app_git.go:L1)
- mimeForExt() (app_git.go:L783)
- .ReadFileBase64() (app_git.go:L805)
- app_models.go (app_models.go:L1)
- app_review.go (app_review.go:L1)
- FileSummary (app_review.go:L22)
- ReviewSummary (app_review.go:L31)
- app_review_scoped.go (app_review_scoped.go:L1)
- app_scan.go (app_scan.go:L1)
- app_sessions.go (app_sessions.go:L1)
- app_spawn.go (app_spawn.go:L1)
- app_terminal_registry.go (app_terminal_registry.go:L1)
- artifacts.go (internal/bmad/artifacts.go:L1)
- assets_write.go (internal/bmad/assets_write.go:L1)
- applyNodeUpdate() (internal/bmad/assets_write.go:L160)
- setNodeValue() (internal/bmad/assets_write.go:L178)
- condition.go (internal/bmad/condition.go:L1)
- ConditionType (internal/bmad/condition.go:L13)
- Condition (internal/bmad/condition.go:L35)
- .Evaluate() (internal/bmad/condition.go:L44)
- executor.go (internal/bmad/executor.go:L1)
- wrapInBashExec() (internal/bmad/executor.go:L1191)
- truncate() (internal/bmad/executor.go:L2764)
- appendUpstreamContext() (internal/bmad/executor.go:L2948)
- TestTruncateCap() (internal/bmad/executor_interactive_test.go:L752)
- fixture_verify_test.go (internal/bmad/fixture_verify_test.go:L1)
- TestWrapInBashExec() (internal/bmad/shell_quote_test.go:L9)
- skillgen.go (internal/bmad/skillgen.go:L1)
- skillTemplateData (internal/bmad/skillgen.go:L11)
- artifactPathSpec (internal/bmad/skillgen.go:L21)
- buildSkillData() (internal/bmad/skillgen.go:L97)
- storage.go (internal/bmad/storage.go:L1)
- validate.go (internal/bmad/validate.go:L1)
- validateInput() (internal/bmad/validate.go:L17)
- containsString() (internal/bmad/validate.go:L79)
- resolveFileInput() (internal/bmad/validate.go:L95)
- validate_test.go (internal/bmad/validate_test.go:L1)
- TestResolveFileInput() (internal/bmad/validate_test.go:L221)
- TestValidateInput() (internal/bmad/validate_test.go:L23)
- TestU0_AC4_ValidateInput_ShapeJSON_BareStringAccepted() (internal/bmad/validate_test.go:L308)
- diff.go (internal/git/diff.go:L1)
- claude.go (internal/scanner/claude.go:L1)
- watcher.go (internal/scanner/watcher.go:L1)
- login_path.go (internal/terminal/login_path.go:L1)
- signal_unix.go (internal/uiadapter/backend/claudecli/signal_unix.go:L1)
- signal_windows.go (internal/uiadapter/backend/claudecli/signal_windows.go:L1)
- TestReadFileBase64_AC2_MimeTypes() (readfilebase64_test.go:L40)
- screenshot_fullstack_test.go (screenshot_fullstack_test.go:L1)
- TestTakeScreenshot_AC1_NewSignature() (screenshot_fullstack_test.go:L115)
- TestTakeScreenshot_AC6_Cancellation_NewSig() (screenshot_fullstack_test.go:L145)
- TestEnsureGitignoreEntry_AC5() (screenshot_fullstack_test.go:L60)
- screenshot_test.go (screenshot_test.go:L1)
- skipUnlessInteractiveScreenshotsEnabled() (screenshot_test.go:L19)
- TestTakeScreenshot_AC3_FilenameFormat() (screenshot_test.go:L26)
- TestTakeScreenshot_AC6_Cancellation() (screenshot_test.go:L48)
- TestTakeScreenshot_AC3_ErrorWrapping() (screenshot_test.go:L66)
- TestTakeScreenshot_NilContext() (screenshot_test.go:L83)

# Depends on
- [App](/modules/app.md)
- [App](/modules/app-348.md)
- [App](/modules/app-349.md)
- [App](/modules/app-452.md)
- [App](/modules/app-63.md)
- [App](/modules/app-76.md)
- [app_review_test.go](/modules/app-review-test-go.md)
- [artifacts_test.go](/modules/artifacts-test-go.md)
- [assets_test.go](/modules/assets-test-go.md)
- [bmad/types.go](/modules/bmad-types-go.md)
- [BmadAgentConfig](/modules/bmadagentconfig.md)
- [context.Context](/modules/context-context.md)
- [.doScan](/modules/doscan.md)
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [loader_test.go](/modules/loader-test-go.md)
- [main.go](/modules/main-go.md)
- [ManagedSession](/modules/managedsession.md)
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)
- [ModelInfo](/modules/modelinfo.md)
- [NewExecutor](/modules/newexecutor.md)
- [newHarness](/modules/newharness.md)
- [ProcessByID](/modules/processbyid.md)
- [question_test.go](/modules/question-test-go.md)
- [.resolveInputs](/modules/resolveinputs.md)
- [ScopedDiff](/modules/scopeddiff.md)
- [skillgen_test.go](/modules/skillgen-test-go.md)
- [Storage](/modules/storage.md)
- [testing.T](/modules/testing-t.md)
- [time.Duration](/modules/time-duration.md)
- [WorkflowExecution](/modules/workflowexecution.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
