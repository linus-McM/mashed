# Story 1: PTY Helper Protocol Types

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** ready

## Description

Define the shared protocol types and wire format for communication between the main Wails process and the `mashed-pty-helper` binary. This is the foundation layer that all other stories depend on -- it establishes the message types (SpawnRequest, SpawnResponse, KillRequest), the length-prefixed JSON wire format, and the SCM_RIGHTS fd-passing utilities. No networking or server logic -- just types, encode/decode, and sendFd/recvFd.

## Developer Notes

### Architecture
- **New package:** `internal/terminal/helper/`
- **New file:** `internal/terminal/helper/protocol.go`
- This package is imported by both the helper binary (`cmd/pty-helper/main.go`, Story 2) and the client (`internal/terminal/helper/client.go`, Story 4)
- The protocol is intentionally minimal: only `Spawn` and `Kill` operations. Resize is handled locally by the main process calling `pty.Setsize()` on the received fd.

### Types to Define

```go
// MessageType discriminates wire messages
type MessageType string
const (
    MsgSpawn MessageType = "spawn"
    MsgKill  MessageType = "kill"
)

// Envelope wraps all messages for type discrimination
type Envelope struct {
    Type MessageType     `json:"type"`
    Data json.RawMessage `json:"data"`
}

type SpawnRequest struct {
    ID    string   `json:"id"`
    Shell string   `json:"shell"`
    Args  []string `json:"args"`
    Env   []string `json:"env"`
    Cwd   string   `json:"cwd"`
    Cols  uint16   `json:"cols"`
    Rows  uint16   `json:"rows"`
}

type SpawnResponse struct {
    ID    string `json:"id"`
    PID   int    `json:"pid"`
    Error string `json:"error,omitempty"`
}

type KillRequest struct {
    ID     string `json:"id"`
    Signal int    `json:"signal"`
}
```

### Wire Format
- Length-prefixed: 4-byte big-endian `uint32` followed by JSON payload
- Functions: `WriteMessage(conn *net.UnixConn, msg interface{}) error` and `ReadMessage(conn *net.UnixConn) (Envelope, error)`
- Avoids newline-delimited parsing issues with embedded JSON strings

### SCM_RIGHTS Helpers
- `SendFd(conn *net.UnixConn, fd int) error` -- sends a single fd via `WriteMsgUnix` with `syscall.UnixRights`
- `RecvFd(conn *net.UnixConn) (int, error)` -- receives a single fd via `ReadMsgUnix` + `ParseSocketControlMessage` + `ParseUnixRights`
- These are low-level syscall wrappers; they are macOS/Linux only (no Windows support needed)

### Technical Considerations
- Use `encoding/binary.BigEndian` for the length prefix
- `json.RawMessage` in `Envelope.Data` allows lazy decoding -- the receiver reads the type field first, then decodes data into the correct struct
- SCM_RIGHTS uses `syscall.CmsgSpace(4)` for a single int32 fd
- All functions should return wrapped errors with context (e.g., `fmt.Errorf("protocol: write message: %w", err)`)

### Risks & Edge Cases
- `syscall.UnixRights` and `ParseUnixRights` are POSIX-only; guard with build tags if needed (currently macOS-only project, so `//go:build darwin` is fine)
- OOB buffer for `ReadMsgUnix` must be exactly `syscall.CmsgSpace(4)` -- too small silently drops the fd
- Zero-length read buffer in `ReadMsgUnix` (the `make([]byte, 1)` dummy byte) is required; macOS needs at least 1 byte of regular data alongside OOB

### Reference Files
- `internal/terminal/session.go` -- existing PTY session types, error patterns
- `internal/terminal/panes.go` -- `TerminalError` custom error type pattern
- `internal/terminal/manager.go` -- existing sentinel error pattern (`ErrSessionExists`, `ErrSessionNotFound`)

## Acceptance Criteria

AC-1: Protocol message types are defined
- Given the package `internal/terminal/helper` exists
- When a developer imports the package
- Then `SpawnRequest`, `SpawnResponse`, `KillRequest`, `Envelope`, and `MessageType` constants are available
- And all struct fields have correct JSON tags

AC-2: Length-prefixed wire encoding works
- Given a `SpawnRequest` with ID "test-123", Shell "/bin/zsh", Cols 80, Rows 24
- When `WriteMessage` encodes it and `ReadMessage` decodes the bytes
- Then the round-trip produces an identical struct
- And the first 4 bytes are a big-endian uint32 equal to the JSON payload length

AC-3: SCM_RIGHTS fd passing works
- Given a connected Unix socket pair (from `net.ListenUnix` + `net.DialUnix`, or `syscall.Socketpair`)
- When `SendFd` sends a valid file descriptor from one end
- And `RecvFd` receives on the other end
- Then the received fd is a valid, readable file descriptor for the same underlying file

AC-4: Error handling follows project conventions
- Given invalid inputs (nil conn, closed conn, malformed JSON)
- When protocol functions are called
- Then they return wrapped errors with operation context
- And errors can be inspected with `errors.Is` / `errors.As`

## BDD Test Scenarios

### Scenario 1: Round-trip message encoding

```gherkin
Feature: PTY helper protocol wire format

  Scenario: SpawnRequest round-trip through length-prefixed encoding
    Given a SpawnRequest with ID "sess-001" and Shell "/bin/zsh" and Cols 120 and Rows 40
    When I encode it with WriteMessage into a buffer
    And I decode it with ReadMessage from the same buffer
    Then the decoded Envelope has Type "spawn"
    And unmarshaling Envelope.Data produces an identical SpawnRequest

  Scenario: KillRequest round-trip
    Given a KillRequest with ID "sess-001" and Signal 15
    When I encode and decode it through the wire format
    Then the decoded Envelope has Type "kill"
    And the KillRequest fields match the original

  Scenario: Large environment variable list
    Given a SpawnRequest with 200 environment variables each 1KB long
    When I encode and decode it
    Then the round-trip succeeds without truncation
```

### Scenario 2: SCM_RIGHTS fd passing

```gherkin
Feature: SCM_RIGHTS file descriptor passing

  Scenario: Pass a PTY fd between processes via Unix socket
    Given a Unix socket pair created with socketpair
    And a temporary file opened as a stand-in for a PTY master
    When SendFd sends the file's fd on one end of the socket
    And RecvFd receives on the other end
    Then os.NewFile on the received fd is readable
    And writing to the original file is visible via the received fd

  Scenario: RecvFd on closed connection
    Given a Unix socket pair where the sender has closed their end
    When RecvFd is called on the receiver end
    Then it returns a non-nil error
    And the error message contains "recvfd" or connection context
```

### Scenario 3: Error paths

```gherkin
Feature: Protocol error handling

  Scenario: ReadMessage with truncated length prefix
    Given a buffer containing only 2 bytes
    When ReadMessage attempts to decode
    Then it returns an error wrapping io.ErrUnexpectedEOF or io.EOF

  Scenario: ReadMessage with invalid JSON payload
    Given a buffer with valid 4-byte length prefix but garbage payload
    When ReadMessage decodes the envelope
    Then it returns an error wrapping a json.SyntaxError
```

## Tasks / Subtasks

- [ ] Task 1: Define protocol types (AC: AC-1)
  - [ ] Create `internal/terminal/helper/` directory
  - [ ] Create `protocol.go` with `MessageType`, `Envelope`, `SpawnRequest`, `SpawnResponse`, `KillRequest`
  - [ ] Add sentinel errors: `ErrHelperNotRunning`, `ErrSpawnFailed`, `ErrConnectionClosed`

- [ ] Task 2: Implement wire format (AC: AC-2, AC-4)
  - [ ] Implement `WriteMessage(w io.Writer, msgType MessageType, data interface{}) error`
  - [ ] Implement `ReadMessage(r io.Reader) (Envelope, error)`
  - [ ] Use `encoding/binary.BigEndian` for 4-byte length prefix
  - [ ] Wrap all errors with operation context

- [ ] Task 3: Implement SCM_RIGHTS helpers (AC: AC-3, AC-4)
  - [ ] Implement `SendFd(conn *net.UnixConn, fd int) error`
  - [ ] Implement `RecvFd(conn *net.UnixConn) (int, error)`
  - [ ] Add build tag `//go:build darwin` if needed

- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Create `protocol_test.go` with table-driven round-trip tests
  - [ ] Test SCM_RIGHTS with `syscall.Socketpair` or `net.ListenUnix`/`net.DialUnix`
  - [ ] Test error paths (truncated input, closed conn, invalid JSON)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/helper/protocol.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
