---
type: Module
title: bridge_test.go
description: "Graphify community 63: internal/terminal/bridge_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: a14cd774ef1482a3 }
---

# Files
- `internal/terminal/bridge_test.go`

# Symbols
- bridge_test.go (internal/terminal/bridge_test.go:L1)
- TestBridge_AC2_WSRoutesToSession() (internal/terminal/bridge_test.go:L112)
- TestBridge_AC3_ErrorResponses() (internal/terminal/bridge_test.go:L142)
- TestBridge_AC4_NoTmuxReferences() (internal/terminal/bridge_test.go:L197)
- TestBridge_AC5_MultipleClients() (internal/terminal/bridge_test.go:L234)
- TestBridge_StartStop() (internal/terminal/bridge_test.go:L298)
- startBridgeWithManager() (internal/terminal/bridge_test.go:L39)
- newMockTmuxSession() (internal/terminal/bridge_test.go:L393)
- .Close() (internal/terminal/bridge_test.go:L444)
- newMockTmuxAttacher() (internal/terminal/bridge_test.go:L469)
- .LastTarget() (internal/terminal/bridge_test.go:L494)
- startBridgeWithMockAdapter() (internal/terminal/bridge_test.go:L503)
- dialBridgeWSCtx() (internal/terminal/bridge_test.go:L518)
- TestBridge_AC1_NewBridgeAcceptsAdapter() (internal/terminal/bridge_test.go:L534)
- TestBridge_AC2_PTYSessionTakesPrecedence() (internal/terminal/bridge_test.go:L549)
- dialBridgeWS() (internal/terminal/bridge_test.go:L58)
- TestBridge_AC3_BMADPrefixRoutesToAdapter() (internal/terminal/bridge_test.go:L596)
- TestBridge_AC3_FullPaneTargetRoundTrip() (internal/terminal/bridge_test.go:L636)
- TestBridge_AC4_NonBMADMissReturns404() (internal/terminal/bridge_test.go:L669)
- TestBridge_AC5_BinaryInputFrameForwardsSendInput() (internal/terminal/bridge_test.go:L695)
- TestBridge_AC6_ResizeTextFrameForwardsResize() (internal/terminal/bridge_test.go:L721)
- readWSMessage() (internal/terminal/bridge_test.go:L74)
- TestBridge_AC7_AttachErrorClosesWebSocket() (internal/terminal/bridge_test.go:L749)
- TestBridge_ProxyTmuxSessionHandlesNilSessionFromAttach() (internal/terminal/bridge_test.go:L773)
- TestBridge_AC8_WebSocketCloseCancelsAttachment() (internal/terminal/bridge_test.go:L800)
- TestBridge_NilAdapterDegradesGracefully() (internal/terminal/bridge_test.go:L836)
- TestBridge_CompileTimeAssertBMADPrefixMatches() (internal/terminal/bridge_test.go:L860)

# Depends on
- [Bridge](/modules/bridge.md)
- [ManagedSession](/modules/managedsession.md)
- [mockTmuxSession](/modules/mocktmuxsession.md)

# Inferred
- [Bridge](/modules/bridge.md)
- [bridge_auth_test.go](/modules/bridge-auth-test-go.md)
- [NewSessionManager](/modules/newsessionmanager.md)
- [session_test.go](/modules/session-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
