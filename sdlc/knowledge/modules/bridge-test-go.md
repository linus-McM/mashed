---
type: Module
title: bridge_test.go
description: "Graphify community 63: internal/terminal/bridge_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: b13529690aa8d966 }
---

# Files
- `internal/terminal/bridge_test.go`

# Symbols
- bridge_test.go (internal/terminal/bridge_test.go:L1)
- TestBridge_AC2_WSRoutesToSession() (internal/terminal/bridge_test.go:L112)
- TestBridge_AC3_ErrorResponses() (internal/terminal/bridge_test.go:L142)
- TestBridge_AC4_NoTmuxReferences() (internal/terminal/bridge_test.go:L194)
- TestBridge_AC5_MultipleClients() (internal/terminal/bridge_test.go:L231)
- mockResizeCall (internal/terminal/bridge_test.go:L372)
- mockTmuxSession (internal/terminal/bridge_test.go:L381)
- startBridgeWithManager() (internal/terminal/bridge_test.go:L39)
- newMockTmuxSession() (internal/terminal/bridge_test.go:L390)
- .Read() (internal/terminal/bridge_test.go:L402)
- .SendInput() (internal/terminal/bridge_test.go:L414)
- .SendKey() (internal/terminal/bridge_test.go:L425)
- .Resize() (internal/terminal/bridge_test.go:L433)
- .Close() (internal/terminal/bridge_test.go:L441)
- newMockTmuxAttacher() (internal/terminal/bridge_test.go:L466)
- .LastTarget() (internal/terminal/bridge_test.go:L491)
- startBridgeWithMockAdapter() (internal/terminal/bridge_test.go:L500)
- dialBridgeWSCtx() (internal/terminal/bridge_test.go:L515)
- TestBridge_AC1_NewBridgeAcceptsAdapter() (internal/terminal/bridge_test.go:L530)
- dialBridgeWS() (internal/terminal/bridge_test.go:L58)
- TestBridge_AC3_BMADPrefixRoutesToAdapter() (internal/terminal/bridge_test.go:L592)
- TestBridge_AC3_FullPaneTargetRoundTrip() (internal/terminal/bridge_test.go:L632)
- TestBridge_AC4_NonBMADMissReturns404() (internal/terminal/bridge_test.go:L665)
- TestBridge_AC5_BinaryInputFrameForwardsSendInput() (internal/terminal/bridge_test.go:L690)
- TestBridge_AC6_ResizeTextFrameForwardsResize() (internal/terminal/bridge_test.go:L716)
- readWSMessage() (internal/terminal/bridge_test.go:L74)
- TestBridge_AC7_AttachErrorClosesWebSocket() (internal/terminal/bridge_test.go:L744)
- TestBridge_ProxyTmuxSessionHandlesNilSessionFromAttach() (internal/terminal/bridge_test.go:L768)
- TestBridge_AC8_WebSocketCloseCancelsAttachment() (internal/terminal/bridge_test.go:L795)
- TestBridge_NilAdapterDegradesGracefully() (internal/terminal/bridge_test.go:L831)
- TestBridge_CompileTimeAssertBMADPrefixMatches() (internal/terminal/bridge_test.go:L854)

# Depends on
- [Bridge](/modules/bridge.md)
- [ManagedSession](/modules/managedsession.md)
- [mockTmuxAttacher](/modules/mocktmuxattacher.md)
- [session_test.go](/modules/session-test-go.md)

# Inferred
- [Bridge](/modules/bridge.md)
- [manager_test.go](/modules/manager-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
