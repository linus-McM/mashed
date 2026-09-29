---
type: Module
title: tmux_adapter_coverage_test.go
description: "Graphify community 18: internal/terminal/tmux_adapter.go, internal/terminal/tmux_adapter_coverage_test.go, internal/terminal/tmux_adapter_test.go, internal/terminal/tmux_adapter_testhelpers_test.go, i"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: tmux_adapter, resource: internal/terminal/tmux_adapter.go, last_modified: "2026-04-12T15:24:29+10:00", digest: 1a5bf00ab35bc698 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-04-11T19:58:57+10:00", digest: 24b353b6c1480a46 }
  - { id: tmux_adapter_test, resource: internal/terminal/tmux_adapter_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: 86447d1134043b08 }
  - { id: tmux_adapter_testhelpers_test, resource: internal/terminal/tmux_adapter_testhelpers_test.go, last_modified: "2026-04-10T17:15:15+10:00", digest: faad2d4cabafa634 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-04-10T16:37:58+10:00", digest: d61b1108f3251770 }
---

# Files
- `internal/terminal/tmux_adapter.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_adapter_test.go`
- `internal/terminal/tmux_adapter_testhelpers_test.go`
- `internal/terminal/tmux_escape.go`

# Symbols
- TmuxAdapter (internal/terminal/tmux_adapter.go:L110)
- NewTmuxAdapter() (internal/terminal/tmux_adapter.go:L127)
- TmuxAttachment (internal/terminal/tmux_adapter.go:L140)
- .Attach() (internal/terminal/tmux_adapter.go:L189)
- .setupFIFO() (internal/terminal/tmux_adapter.go:L257)
- .startPipePane() (internal/terminal/tmux_adapter.go:L281)
- .startFIFOReader() (internal/terminal/tmux_adapter.go:L318)
- .startPaneDeathWatcher() (internal/terminal/tmux_adapter.go:L396)
- .startPollingLoop() (internal/terminal/tmux_adapter.go:L425)
- .Read() (internal/terminal/tmux_adapter.go:L466)
- .loadReadErr() (internal/terminal/tmux_adapter.go:L502)
- .runTmux() (internal/terminal/tmux_adapter.go:L515)
- .SendInput() (internal/terminal/tmux_adapter.go:L555)
- formatSendKeysHex() (internal/terminal/tmux_adapter.go:L569)
- .SendInputToTarget() (internal/terminal/tmux_adapter.go:L580)
- .SendKey() (internal/terminal/tmux_adapter.go:L593)
- .Resize() (internal/terminal/tmux_adapter.go:L598)
- .Close() (internal/terminal/tmux_adapter.go:L610)
- .cleanupFIFO() (internal/terminal/tmux_adapter.go:L637)
- .Close() (internal/terminal/tmux_adapter.go:L650)
- CommandRunner (internal/terminal/tmux_adapter.go:L67)
- DefaultCommandRunner() (internal/terminal/tmux_adapter.go:L71)
- tmux_adapter_coverage_test.go (internal/terminal/tmux_adapter_coverage_test.go:L1)
- TestTmuxAdapter_AttachCapturePaneErrorIsNonFatal() (internal/terminal/tmux_adapter_coverage_test.go:L101)
- TestTmuxAdapter_AttachPipePaneErrorCleansUp() (internal/terminal/tmux_adapter_coverage_test.go:L126)
- TestTmuxAttachment_SendInputEmptyIsNoOp() (internal/terminal/tmux_adapter_coverage_test.go:L147)
- TestTmuxAttachment_SendInputControlBytesAreSentAsHex() (internal/terminal/tmux_adapter_coverage_test.go:L156)
- sendErrMockRunner() (internal/terminal/tmux_adapter_coverage_test.go:L180)
- TestTmuxAttachment_SendInputRunnerErrorWrapped() (internal/terminal/tmux_adapter_coverage_test.go:L187)
- TestTmuxAttachment_SendKeyRunnerErrorWrapped() (internal/terminal/tmux_adapter_coverage_test.go:L206)
- TestTmuxAttachment_ResizeRunnerErrorWrapped() (internal/terminal/tmux_adapter_coverage_test.go:L225)
- TestTmuxAttachment_ReadZeroLengthBuffer() (internal/terminal/tmux_adapter_coverage_test.go:L248)
- TestTmuxAttachment_ReadShortBufferStashesRemainder() (internal/terminal/tmux_adapter_coverage_test.go:L256)
- TestTmuxAdapter_PollingLoopRunnerError() (internal/terminal/tmux_adapter_coverage_test.go:L298)
- TestTmuxAdapter_CloseCascadesToActiveAttachments() (internal/terminal/tmux_adapter_coverage_test.go:L357)
- TestTmuxAttachment_ClosePollingModeNoPipePaneStop() (internal/terminal/tmux_adapter_coverage_test.go:L392)
- TestTmuxAttachment_CloseRacesReaderAndWriter() (internal/terminal/tmux_adapter_coverage_test.go:L416)
- TestDefaultCommandRunner_EchoesStdout() (internal/terminal/tmux_adapter_coverage_test.go:L42)
- TestTmuxAdapter_FIFOPathWithSpaceInTempDir() (internal/terminal/tmux_adapter_coverage_test.go:L476)
- TestNewTmuxAdapter_NilRunnerFallsBackToDefault() (internal/terminal/tmux_adapter_coverage_test.go:L48)
- TestEscapeTmuxLiteral_TabPreserved() (internal/terminal/tmux_adapter_coverage_test.go:L516)
- TestEscapeTmuxLiteral_EmptyString() (internal/terminal/tmux_adapter_coverage_test.go:L62)
- TestTmuxAdapter_AttachOnClosedAdapter() (internal/terminal/tmux_adapter_coverage_test.go:L70)
- TestTmuxAdapter_AttachListPanesErrorWrapped() (internal/terminal/tmux_adapter_coverage_test.go:L83)
- tmux_adapter_test.go (internal/terminal/tmux_adapter_test.go:L1)
- .runner() (internal/terminal/tmux_adapter_test.go:L101)
- .invocations() (internal/terminal/tmux_adapter_test.go:L123)
- .countSubcommand() (internal/terminal/tmux_adapter_test.go:L132)
- .findSubcommand() (internal/terminal/tmux_adapter_test.go:L139)
- .paneAlive() (internal/terminal/tmux_adapter_test.go:L151)
- .paneDead() (internal/terminal/tmux_adapter_test.go:L158)
- .paneDiesAfter() (internal/terminal/tmux_adapter_test.go:L167)
- .captureReturns() (internal/terminal/tmux_adapter_test.go:L180)
- newContext() (internal/terminal/tmux_adapter_test.go:L187)
- newLiveAttachment() (internal/terminal/tmux_adapter_test.go:L197)
- readPrefix() (internal/terminal/tmux_adapter_test.go:L218)
- TestTmuxAdapter_AC1_AttachReplaysScrollback() (internal/terminal/tmux_adapter_test.go:L241)
- TestTmuxAdapter_AC2_AttachRejectsDeadPane() (internal/terminal/tmux_adapter_test.go:L280)
- TestTmuxAttachment_AC3_SendInputUsesSendKeysHex() (internal/terminal/tmux_adapter_test.go:L314)
- TestTmuxAttachment_AC4_SendKeyUsesSendKeysNoLiteral() (internal/terminal/tmux_adapter_test.go:L416)
- TestTmuxAttachment_AC5_ResizeInvokesResizeWindow() (internal/terminal/tmux_adapter_test.go:L450)
- TestTmuxAttachment_AC6_ReadReturnsEOFWhenPaneDies() (internal/terminal/tmux_adapter_test.go:L497)
- invocation (internal/terminal/tmux_adapter_test.go:L54)
- TestTmuxAttachment_AC7_CloseIsIdempotent() (internal/terminal/tmux_adapter_test.go:L556)
- .joined() (internal/terminal/tmux_adapter_test.go:L59)
- TestTmuxAdapter_AC8_MkfifoFailureFallsBackToPolling() (internal/terminal/tmux_adapter_test.go:L603)
- TestTmuxAttachment_ConcurrentReadSendInput() (internal/terminal/tmux_adapter_test.go:L647)
- responderFunc (internal/terminal/tmux_adapter_test.go:L69)
- TestTmuxAdapter_SendInputToTarget_HexArgv() (internal/terminal/tmux_adapter_test.go:L730)
- mockRunner (internal/terminal/tmux_adapter_test.go:L76)
- TestTmuxAdapter_SendInputToTarget_EmptyIsNoOp() (internal/terminal/tmux_adapter_test.go:L797)
- newMockRunner() (internal/terminal/tmux_adapter_test.go:L84)
- .setResponder() (internal/terminal/tmux_adapter_test.go:L94)
- tmux_adapter_testhelpers_test.go (internal/terminal/tmux_adapter_testhelpers_test.go:L1)
- overrideMkfifo() (internal/terminal/tmux_adapter_testhelpers_test.go:L21)
- attachmentIsPolling() (internal/terminal/tmux_adapter_testhelpers_test.go:L30)
- EscapeTmuxLiteral() (internal/terminal/tmux_escape.go:L48)

# Depends on
- [app_terminal_registry_test.go](/modules/app-terminal-registry-test-go.md)
- [mockTmuxAttacher](/modules/mocktmuxattacher.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
