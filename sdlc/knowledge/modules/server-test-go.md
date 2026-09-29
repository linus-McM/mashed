---
type: Module
title: server_test.go
description: "Graphify community 5: cmd/pty-helper/main.go, internal/terminal/helper/client.go, internal/terminal/helper/client_test.go, internal/terminal/helper/integration_test.go, internal/terminal/helper/protoc"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: main, resource: cmd/pty-helper/main.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 5b417efc875ee944 }
  - { id: client, resource: internal/terminal/helper/client.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 46fc572e21238bfb }
  - { id: client_test, resource: internal/terminal/helper/client_test.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 4004a3d885df1db3 }
  - { id: integration_test, resource: internal/terminal/helper/integration_test.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 9d9c306c952e3df8 }
  - { id: protocol, resource: internal/terminal/helper/protocol.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 07f801b52b8e41da }
  - { id: protocol_test, resource: internal/terminal/helper/protocol_test.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 500bab47812315f3 }
  - { id: server, resource: internal/terminal/helper/server.go, last_modified: "2026-04-09T21:03:43+10:00", digest: e54a6a9ea36ca59d }
  - { id: server_test, resource: internal/terminal/helper/server_test.go, last_modified: "2026-04-09T21:03:43+10:00", digest: d9814bf07caafe77 }
---

# Files
- `cmd/pty-helper/main.go`
- `internal/terminal/helper/client.go`
- `internal/terminal/helper/client_test.go`
- `internal/terminal/helper/integration_test.go`
- `internal/terminal/helper/protocol.go`
- `internal/terminal/helper/protocol_test.go`
- `internal/terminal/helper/server.go`
- `internal/terminal/helper/server_test.go`

# Symbols
- main() (cmd/pty-helper/main.go:L15)
- .Close() (internal/terminal/helper/client.go:L106)
- Client (internal/terminal/helper/client.go:L16)
- Dial() (internal/terminal/helper/client.go:L23)
- .Spawn() (internal/terminal/helper/client.go:L34)
- .Kill() (internal/terminal/helper/client.go:L98)
- helper/client_test.go (internal/terminal/helper/client_test.go:L1)
- TestClientSpawn_Serialization() (internal/terminal/helper/client_test.go:L111)
- TestClientSpawn_ContextCancelled() (internal/terminal/helper/client_test.go:L171)
- TestClientSpawn_AlreadyCancelledContext() (internal/terminal/helper/client_test.go:L190)
- TestClientKill() (internal/terminal/helper/client_test.go:L207)
- startMockServer() (internal/terminal/helper/client_test.go:L22)
- TestClientClose() (internal/terminal/helper/client_test.go:L236)
- TestClientSpawn_WriteError() (internal/terminal/helper/client_test.go:L257)
- TestClientSpawn_BadResponse() (internal/terminal/helper/client_test.go:L275)
- TestDial_Failure() (internal/terminal/helper/client_test.go:L296)
- TestDial_Success() (internal/terminal/helper/client_test.go:L302)
- TestClientSpawn_Success() (internal/terminal/helper/client_test.go:L44)
- TestClientSpawn_Error() (internal/terminal/helper/client_test.go:L89)
- TestIntegration_BidirectionalIO() (internal/terminal/helper/integration_test.go:L133)
- TestIntegration_ConcurrentSpawnsFdIsolation() (internal/terminal/helper/integration_test.go:L174)
- catSpawnReq() (internal/terminal/helper/integration_test.go:L22)
- TestIntegration_AC4_LatencyAndFdLeaks() (internal/terminal/helper/integration_test.go:L254)
- countOpenFds() (internal/terminal/helper/integration_test.go:L26)
- waitForPIDDeath() (internal/terminal/helper/integration_test.go:L40)
- TestIntegration_AC1_SpawnAndReadFd() (internal/terminal/helper/integration_test.go:L54)
- TestIntegration_AC2_KillTerminatesProcess() (internal/terminal/helper/integration_test.go:L93)
- SendFd() (internal/terminal/helper/protocol.go:L113)
- RecvFd() (internal/terminal/helper/protocol.go:L124)
- MessageType (internal/terminal/helper/protocol.go:L14)
- Envelope (internal/terminal/helper/protocol.go:L29)
- SpawnRequest (internal/terminal/helper/protocol.go:L35)
- SpawnResponse (internal/terminal/helper/protocol.go:L46)
- KillRequest (internal/terminal/helper/protocol.go:L53)
- WriteMessage() (internal/terminal/helper/protocol.go:L60)
- ReadMessage() (internal/terminal/helper/protocol.go:L91)
- protocol_test.go (internal/terminal/helper/protocol_test.go:L1)
- TestWriteReadMessage_SpawnResponse() (internal/terminal/helper/protocol_test.go:L110)
- TestWriteReadMessage_LargePayload() (internal/terminal/helper/protocol_test.go:L143)
- TestWriteReadMessage_SpawnRequest() (internal/terminal/helper/protocol_test.go:L17)
- TestReadMessage_TruncatedLength() (internal/terminal/helper/protocol_test.go:L174)
- TestReadMessage_TruncatedPayload() (internal/terminal/helper/protocol_test.go:L182)
- TestReadMessage_InvalidJSON() (internal/terminal/helper/protocol_test.go:L193)
- TestReadMessage_EmptyReader() (internal/terminal/helper/protocol_test.go:L209)
- TestWriteMessage_MultipleMessages() (internal/terminal/helper/protocol_test.go:L216)
- TestSendRecvFd() (internal/terminal/helper/protocol_test.go:L235)
- TestRecvFd_ClosedConn() (internal/terminal/helper/protocol_test.go:L286)
- TestSentinelErrors() (internal/terminal/helper/protocol_test.go:L307)
- TestMessageTypeConstants() (internal/terminal/helper/protocol_test.go:L313)
- TestEnvelopeJSONTags() (internal/terminal/helper/protocol_test.go:L318)
- TestSpawnResponseOmitEmpty() (internal/terminal/helper/protocol_test.go:L329)
- createSocketPair() (internal/terminal/helper/protocol_test.go:L342)
- TestWriteReadMessage_KillRequest() (internal/terminal/helper/protocol_test.go:L72)
- .handleKill() (internal/terminal/helper/server.go:L162)
- .Shutdown() (internal/terminal/helper/server.go:L189)
- .MonitorParent() (internal/terminal/helper/server.go:L204)
- Server (internal/terminal/helper/server.go:L27)
- NewServer() (internal/terminal/helper/server.go:L34)
- .Serve() (internal/terminal/helper/server.go:L42)
- .handleConn() (internal/terminal/helper/server.go:L53)
- .handleSpawn() (internal/terminal/helper/server.go:L86)
- server_test.go (internal/terminal/helper/server_test.go:L1)
- TestServerSpawnSuccess() (internal/terminal/helper/server_test.go:L103)
- TestServerSpawnBadShell() (internal/terminal/helper/server_test.go:L139)
- TestServerKillSuccess() (internal/terminal/helper/server_test.go:L155)
- TestServerKillNotFound() (internal/terminal/helper/server_test.go:L190)
- TestServerMultipleConcurrentSessions() (internal/terminal/helper/server_test.go:L201)
- testSockPath() (internal/terminal/helper/server_test.go:L22)
- TestServerShutdownCleansSessions() (internal/terminal/helper/server_test.go:L231)
- TestServerMonitorParentExitsOnDeadParent() (internal/terminal/helper/server_test.go:L272)
- TestServerSpawnDefaultShell() (internal/terminal/helper/server_test.go:L292)
- TestServerShutdownEmpty() (internal/terminal/helper/server_test.go:L319)
- TestServerNewServer() (internal/terminal/helper/server_test.go:L325)
- TestServerShutdownWithMockSessions() (internal/terminal/helper/server_test.go:L333)
- canPTYSpawn() (internal/terminal/helper/server_test.go:L35)
- TestServerKillWithMockSession() (internal/terminal/helper/server_test.go:L364)
- TestServerHandleConnUnknownType() (internal/terminal/helper/server_test.go:L390)
- TestServerHandleConnBadPayload() (internal/terminal/helper/server_test.go:L411)
- TestServerServeAcceptsMultipleConnections() (internal/terminal/helper/server_test.go:L449)
- TestServerMonitorParentAliveDoesNotExit() (internal/terminal/helper/server_test.go:L468)
- dialHelper() (internal/terminal/helper/server_test.go:L48)
- TestServerConcurrentSpawnRace() (internal/terminal/helper/server_test.go:L488)
- startTestServer() (internal/terminal/helper/server_test.go:L58)
- sendSpawn() (internal/terminal/helper/server_test.go:L78)
- sendKill() (internal/terminal/helper/server_test.go:L94)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
