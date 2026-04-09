# Story 4: PTY Helper Client

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** Story 1 (ptyhelper-01-protocol-types)
**Status:** ready

## Description

Build the client library that the main Wails process uses to communicate with the PTY helper binary. The client dials the Unix socket, sends spawn/kill requests, receives responses, and handles fd delivery via SCM_RIGHTS. It includes a critical serialization mutex to prevent SCM_RIGHTS fd/response ordering mismatches when multiple spawns are in flight. The client is the main process's only interface to PTY creation.

## Developer Notes

### Architecture
- **New file:** `internal/terminal/helper/client.go`
- Consumed by `internal/terminal/manager.go` (Story 5) and `main.go` (Story 6)
- Uses protocol types from Story 1

### Client Struct
```go
type Client struct {
    conn    *net.UnixConn
    mu      sync.Mutex    // serializes Spawn to prevent SCM_RIGHTS ordering issues
    readMu  sync.Mutex    // protects concurrent reads
    closed  atomic.Bool
}
```

### Key Methods

```go
// Dial connects to the helper's Unix socket.
func Dial(sockPath string) (*Client, error)

// Spawn sends a SpawnRequest and receives the PTY fd.
// IMPORTANT: This method is serialized (only one in flight at a time)
// to prevent SCM_RIGHTS fd ordering mismatches.
func (c *Client) Spawn(ctx context.Context, req SpawnRequest) (ptmx *os.File, pid int, err error)

// Kill sends a KillRequest to terminate a session.
func (c *Client) Kill(id string, sig syscall.Signal) error

// Close shuts down the connection.
func (c *Client) Close() error
```

### Spawn Serialization (CRITICAL)
The plan explicitly calls out: "Serialize spawn requests (only one in flight at a time via a mutex on the client side)". The reason:
- SCM_RIGHTS fd delivery is positional -- fd #1 goes to whoever calls `RecvFd` first
- If two Spawn calls are concurrent, their `RecvFd` calls race, and fd #1 might end up paired with response #2
- Solution: `c.mu.Lock()` at the start of `Spawn`, unlock after both `SpawnResponse` and `RecvFd` are complete

### Spawn Implementation Flow
```
1. c.mu.Lock()  // serialize spawns
2. WriteMessage(conn, MsgSpawn, req)
3. ReadMessage(conn) -> SpawnResponse
4. if resp.Error != "" { return error }
5. RecvFd(conn) -> fd
6. c.mu.Unlock()
7. return os.NewFile(uintptr(fd), "ptmx"), resp.PID, nil
```

### Context Cancellation
- `Spawn` accepts `context.Context` for timeout/cancellation
- If context is cancelled while waiting for response, return `ctx.Err()`
- Use a goroutine + channel pattern: start read in goroutine, select on result channel vs ctx.Done()
- On cancellation, the connection may be in an inconsistent state -- consider marking the client as unhealthy

### Technical Considerations
- `Kill` does NOT need serialization -- it's fire-and-forget (no fd involved)
- The `Close` method should be idempotent (use `atomic.Bool` for closed flag)
- After `Close`, all methods should return `ErrConnectionClosed`
- The client does NOT start a background read goroutine -- reads are synchronous within the Spawn mutex
- This is simpler than the plan's `pending` map approach because we serialize instead of multiplex

### Risks & Edge Cases
- If the helper crashes mid-spawn, `ReadMessage` returns EOF -- wrap as `ErrHelperNotRunning`
- If the socket doesn't exist when `Dial` is called, return a clear error (not a raw syscall error)
- Network errors should be wrapped with context: `"client: spawn: %w"`
- After a failed Spawn (helper returned error), the mutex is still released so subsequent calls work

### Reference Files
- `internal/terminal/helper/protocol.go` -- `WriteMessage`, `ReadMessage`, `SendFd`, `RecvFd`
- `internal/terminal/manager.go` -- the manager that will call this client (Story 5)

## Acceptance Criteria

AC-1: Client dials Unix socket successfully
- Given the helper is listening on a Unix socket at a known path
- When `Dial(sockPath)` is called
- Then it returns a non-nil Client with no error
- And the connection is established

AC-2: Client.Spawn sends request and returns PTY fd
- Given a connected Client
- When `Spawn(ctx, req)` is called with a valid SpawnRequest
- Then it returns a non-nil `*os.File`, a PID > 0, and nil error
- And the file is the PTY master (readable/writable)

AC-3: Client.Spawn serializes concurrent requests
- Given a connected Client
- When two goroutines call `Spawn` concurrently
- Then only one Spawn is in flight at a time (the second blocks until the first completes)
- And both return correct, non-swapped fd/PID pairs

AC-4: Client.Kill sends termination signal
- Given a session "kill-test" was spawned
- When `Kill("kill-test", syscall.SIGTERM)` is called
- Then the helper receives a KillRequest and terminates the session

AC-5: Client handles helper errors
- Given a SpawnRequest for a non-existent shell
- When `Spawn` is called
- Then it returns an error containing the helper's error message
- And no fd is leaked

AC-6: Client handles disconnection
- Given the helper process has crashed
- When `Spawn` or `Kill` is called
- Then it returns an error wrapping `ErrHelperNotRunning` or `ErrConnectionClosed`
- And subsequent calls also return errors (client is in a terminal state)

## BDD Test Scenarios

### Scenario 1: Spawn flow

```gherkin
Feature: PTY helper client spawn

  Scenario: Successful spawn
    Given a test helper server is running
    And a Client is connected via Dial
    When Spawn is called with Shell "/bin/echo" Args ["test"] Cols 80 Rows 24
    Then the returned file is non-nil
    And the returned PID is greater than 0
    And reading from the file yields output containing "test"

  Scenario: Spawn with context timeout
    Given a Client is connected
    And the context has a 1ms timeout (already expired)
    When Spawn is called
    Then it returns context.DeadlineExceeded
    And no resources are leaked
```

### Scenario 2: Serialization

```gherkin
Feature: Spawn serialization

  Scenario: Concurrent spawns are serialized
    Given a Client is connected
    When 5 goroutines each call Spawn concurrently with unique session IDs
    Then all 5 complete successfully
    And each returned fd is distinct
    And each PID is distinct
    And no fd/PID pairs are swapped (verified by reading unique echo output from each)
```

### Scenario 3: Error handling

```gherkin
Feature: Client error handling

  Scenario: Spawn with invalid shell
    Given a Client is connected
    When Spawn is called with Shell "/no/such/binary"
    Then the error message contains "no such file" or the helper's error text
    And the Client is still usable for subsequent Spawn calls

  Scenario: Dial non-existent socket
    Given no helper is running
    When Dial("/tmp/nonexistent-pty.sock") is called
    Then it returns a non-nil error
    And the error message indicates connection failure

  Scenario: Operations after Close
    Given a Client that has been Close()d
    When Spawn is called
    Then it returns ErrConnectionClosed
    When Kill is called
    Then it returns ErrConnectionClosed
```

## Tasks / Subtasks

- [ ] Task 1: Implement Client struct and Dial (AC: AC-1, AC-6)
  - [ ] Create `internal/terminal/helper/client.go`
  - [ ] Implement `Dial(sockPath string) (*Client, error)` with `net.DialUnix`
  - [ ] Implement `Close() error` with idempotent guard via `atomic.Bool`

- [ ] Task 2: Implement Spawn with serialization (AC: AC-2, AC-3, AC-5)
  - [ ] Implement `Spawn(ctx context.Context, req SpawnRequest) (*os.File, int, error)`
  - [ ] Add mutex serialization (lock at entry, unlock after fd received)
  - [ ] Add context cancellation via goroutine + select
  - [ ] Handle error responses (non-empty SpawnResponse.Error)

- [ ] Task 3: Implement Kill (AC: AC-4)
  - [ ] Implement `Kill(id string, sig syscall.Signal) error`
  - [ ] No serialization needed (no fd exchange)

- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Create `internal/terminal/helper/client_test.go`
  - [ ] Test Dial with a real test server (start server in test, then dial)
  - [ ] Test Spawn round-trip
  - [ ] Test concurrent Spawn serialization (5 goroutines, verify no fd swaps)
  - [ ] Test Kill
  - [ ] Test error paths (bad shell, closed client, missing socket)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/helper/client.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
