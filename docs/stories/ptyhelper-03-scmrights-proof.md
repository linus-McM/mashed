# Story 3: SCM_RIGHTS End-to-End Proof of Concept

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** Story 1 (ptyhelper-01-protocol-types), Story 2 (ptyhelper-02-helper-binary)
**Status:** done

## Description

Verify that the PTY helper architecture actually works end-to-end before wiring it into the Wails application. This story creates a standalone integration test that builds the helper binary, launches it as a subprocess, connects as a client, spawns a shell session, receives the PTY master fd via SCM_RIGHTS, and confirms that bytes flow bidirectionally (write to fd -> shell receives input; shell output -> read from fd). This is the highest-value early validation step identified in the plan -- if fd passing has macOS arm64 quirks, we catch them here.

## Developer Notes

### Architecture
- **New file:** `internal/terminal/helper/integration_test.go`
- This is a `go test` file, NOT a separate binary -- it uses `TestMain` or `exec.Command` to build and launch the helper binary, then exercises the full protocol
- It validates the entire data path: client -> Unix socket -> helper -> pty.StartWithSize -> SCM_RIGHTS -> client -> os.NewFile -> io.ReadWriter

### Test Strategy
The test should:
1. Build `cmd/pty-helper/main.go` into a temp directory using `go build`
2. Generate a unique socket path in `t.TempDir()`
3. Launch the helper binary as a subprocess with `MASHED_PTY_SOCK` and `MASHED_PARENT_PID` env vars
4. Wait for `"READY\n"` on the helper's stdout pipe
5. Connect to the socket via `net.DialUnix`
6. Send a `SpawnRequest` for `/bin/sh -c "echo PROOF_OF_CONCEPT"`
7. Receive `SpawnResponse` and verify PID > 0
8. Call `RecvFd` and wrap with `os.NewFile`
9. Read from the file and verify output contains "PROOF_OF_CONCEPT"
10. Send a `KillRequest` and verify the session is terminated
11. Signal the helper to exit and wait for clean shutdown

### Bidirectional Test
For a stronger proof, also test writing TO the PTY:
1. Spawn `/bin/cat` (echoes stdin to stdout)
2. Receive the fd
3. Write `"HELLO_FROM_MAIN\n"` to the fd
4. Read from the fd and verify it contains `"HELLO_FROM_MAIN"`

### Technical Considerations
- Use `testing.Short()` skip guard -- these tests launch real processes and are slow
- Use `t.TempDir()` for socket path to avoid cleanup issues
- Set a test timeout (`context.WithTimeout`) of 10 seconds to prevent hanging
- The helper binary build step can be done once in `TestMain` with a package-level variable for the binary path
- Use `bufio.Scanner` on the helper's stdout to read the "READY\n" line

### Risks & Edge Cases
- macOS arm64 may have different behavior with `SCM_RIGHTS` -- this test catches that
- The helper's stdout pipe must be read before it fills up (the "READY\n" message); use a goroutine reader
- Socket path length limit on macOS is 104 bytes -- `t.TempDir()` paths can be long; use a short socket filename
- The helper binary must be compiled for the current platform -- `go build` handles this automatically

### Reference Files
- `internal/terminal/helper/protocol.go` -- `WriteMessage`, `ReadMessage`, `SendFd`, `RecvFd` (Story 1)
- `internal/terminal/helper/server.go` -- server being tested (Story 2)
- `internal/terminal/manager_test.go` -- example of test helpers and PTY test patterns

## Acceptance Criteria

AC-1: Helper binary builds successfully
- Given the Go source at `cmd/pty-helper/main.go` and `internal/terminal/helper/`
- When `go build -o {tempdir}/mashed-pty-helper ./cmd/pty-helper` is run
- Then the build succeeds with exit code 0
- And the binary is executable

AC-2: End-to-end spawn and fd passing works
- Given the helper binary is running and a client is connected
- When a SpawnRequest for `/bin/sh -c "echo PROOF_OF_CONCEPT"` is sent
- Then a SpawnResponse with PID > 0 is received
- And `RecvFd` returns a valid file descriptor
- And reading from `os.NewFile(fd)` yields output containing "PROOF_OF_CONCEPT"

AC-3: Bidirectional I/O through received fd
- Given a session spawned with `/bin/cat`
- When the client writes "HELLO\n" to the received PTY fd
- Then reading from the same fd yields output containing "HELLO"
- And the bytes were not proxied through the helper (direct fd I/O)

AC-4: Helper shutdown is clean
- Given active sessions exist
- When the helper receives SIGTERM
- Then all child processes are terminated
- And the socket file is removed
- And the helper exits with code 0

## BDD Test Scenarios

### Scenario 1: Full integration test

```gherkin
Feature: SCM_RIGHTS end-to-end proof of concept

  Scenario: Spawn echo command and read output via received fd
    Given the helper binary is built to a temp directory
    And the helper is launched with a temp socket path
    And the client has connected after receiving READY signal
    When a SpawnRequest is sent for "/bin/sh" with Args ["-c", "echo PROOF_OF_CONCEPT"]
    Then SpawnResponse.PID is greater than 0
    And SpawnResponse.Error is empty
    And RecvFd returns a non-negative fd
    And os.NewFile(fd).Read() eventually contains "PROOF_OF_CONCEPT"

  Scenario: Bidirectional PTY I/O
    Given a session is spawned for "/bin/cat"
    And the PTY fd has been received
    When "TEST_INPUT\n" is written to the fd
    Then reading from the fd yields output containing "TEST_INPUT"

  Scenario: Multiple sequential spawns on same connection
    Given a connected client
    When 3 sequential SpawnRequests are sent (each with unique IDs)
    Then each SpawnResponse has the correct matching ID
    And each RecvFd returns a distinct valid fd
    And all 3 fds are independently readable
```

### Scenario 2: Error propagation

```gherkin
Feature: Error cases in integration

  Scenario: Spawn with non-existent shell
    Given a connected client
    When a SpawnRequest is sent with Shell "/no/such/binary"
    Then SpawnResponse.Error is non-empty
    And no fd is delivered (RecvFd is not called)

  Scenario: Kill then re-use session ID
    Given a session "reuse-test" was spawned with /bin/cat
    When a KillRequest terminates "reuse-test"
    And a new SpawnRequest is sent with the same ID "reuse-test"
    Then the new spawn succeeds with a fresh PID and fd
```

## Tasks / Subtasks

- [ ] Task 1: Create integration test file (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Create `internal/terminal/helper/integration_test.go`
  - [ ] Add `TestMain` that builds the helper binary to a temp dir
  - [ ] Add `testing.Short()` skip guard
  - [ ] Implement helper launch utility (start subprocess, wait for READY, return conn)

- [ ] Task 2: Test spawn + fd read (AC: AC-2)
  - [ ] Test spawning `/bin/sh -c "echo PROOF_OF_CONCEPT"` and reading output via received fd
  - [ ] Verify PID, error field, and fd validity

- [ ] Task 3: Test bidirectional I/O (AC: AC-3)
  - [ ] Test spawning `/bin/cat` and writing/reading through the received fd
  - [ ] Verify the helper is NOT proxying bytes (the fd is direct)

- [ ] Task 4: Test multiple spawns and clean shutdown (AC: AC-2, AC-4)
  - [ ] Test 3 sequential spawns on one connection
  - [ ] Test SIGTERM causes clean exit
  - [ ] Verify socket file removal on shutdown

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] Integration tests pass on macOS arm64
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes (with `-short` to skip integration tests in CI if needed)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
