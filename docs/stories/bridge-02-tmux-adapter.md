# Story bridge-02: Tmux Adapter for Live Session Attachment

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** bridge-01
**Status:** ready

## Description

Introduce a new `TmuxAdapter` type that can attach to an existing, externally-created tmux session (such as those spawned by the BMAD executor), stream its live output, and forward input and resize events — without routing through the PTY-based `SessionManager`. This is delivered as a pure library in `internal/terminal/tmux_adapter.go` with no bridge wiring yet; bridge-03 will consume it. The adapter uses `tmux capture-pane` for scrollback replay and `tmux pipe-pane` (FIFO-based) for live streaming, falling back to `capture-pane` polling if FIFO creation fails.

## Developer Notes

### Architecture

- **New file:** `internal/terminal/tmux_adapter.go`
  - Struct `TmuxAdapter` with fields for command runner, temp FIFO dir, active attachments map
  - Struct `TmuxAttachment` representing one live attach to one pane: context, cancel, output reader, pane target string, close-once
  - Public methods on `*TmuxAdapter`:
    - `NewTmuxAdapter(runner CommandRunner) *TmuxAdapter`
    - `Attach(ctx context.Context, paneTarget string) (*TmuxAttachment, error)` — validates pane exists, replays scrollback, starts pipe-pane, returns a live attachment
    - `Close()` — shuts down all attachments
  - Public methods on `*TmuxAttachment`:
    - `Read(p []byte) (n int, err error)` — implements `io.Reader`; returns bytes from the FIFO (or polling fallback). Returns `io.EOF` when the pane dies.
    - `SendInput(data []byte) error` — escapes via `EscapeTmuxLiteral` and runs `tmux send-keys -l -t {target} {literal}`
    - `SendKey(key string) error` — runs `tmux send-keys -t {target} {key}` (for Enter, Tab, arrow keys, etc.)
    - `Resize(cols, rows uint16) error` — runs `tmux resize-window -t {target} -x {cols} -y {rows}`
    - `Close() error` — idempotent; stops pipe-pane, removes FIFO, cancels context

- **New file:** `internal/terminal/tmux_adapter_test.go` — table-driven with mock `CommandRunner`
- **Type alias for CommandRunner:** Import/re-declare the same `CommandRunner` signature used in `internal/bmad`. Define it locally in `internal/terminal` to avoid a circular import:
  ```go
  type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)
  ```
  A `DefaultCommandRunner` constant uses `exec.CommandContext`.

- **Move/share helper:** Export `EscapeTmuxLiteral(string) string` from a shared location.
  Recommended: create `internal/terminal/tmux_escape.go` with `func EscapeTmuxLiteral(s string) string` identical to the private one in `internal/bmad/question.go`, and update `internal/bmad/question.go` to import and delegate (or keep a local wrapper) so both packages agree. **Do not** create a circular import — `internal/bmad` may depend on `internal/terminal`, not the reverse.
  Alternative (preferred if it compiles cleanly): copy the function body verbatim into `internal/terminal` as `EscapeTmuxLiteral`, leaving the `bmad` version in place. Call out in the PR that this is intentional duplication until a shared `internal/tmuxutil` package is extracted.

### Streaming Design

**Preferred path — FIFO + pipe-pane:**
1. `Attach` creates a temp directory under `os.TempDir()` and a named pipe with `syscall.Mkfifo(path, 0600)`.
2. Opens the read end in a goroutine (non-blocking O_RDONLY | O_NONBLOCK then switch to blocking reads via `bufio.Reader`).
3. Runs `tmux pipe-pane -o -t {target} 'cat > /tmp/xxx/fifo'` to start piping the pane output into the FIFO.
4. First replays scrollback: `tmux capture-pane -p -J -e -S -5000 -t {target}` and emits those bytes through `Read` before live bytes.
5. Pane-death detection: background ticker every 2s runs `tmux list-panes -t {target} -F "#{pane_dead}"`; if result is `"1"` or the command errors, `Read` returns `io.EOF`.

**Fallback path — capture-pane polling:**
- Triggered if `syscall.Mkfifo` returns an error (platform-specific, rare on macOS/Linux).
- A goroutine runs `tmux capture-pane -p -J -t {target}` every 200ms, diffs against the last snapshot, and emits the delta as a byte stream.
- Loses real-time ANSI escapes but keeps the feature usable.
- Set a boolean `a.fallbackPolling` and log the degradation once.

### Technical Considerations

- **Concurrency:** `Attach` spawns up to three goroutines per attachment (FIFO reader, pane-death watcher, optional polling). All are tied to the attachment's `context.Context`; `Close` cancels the context and waits for them via `sync.WaitGroup`.
- **Close idempotence:** Wrap the cleanup block in `sync.Once` so double-`Close()` is safe.
- **Error wrapping:** Use custom error type `TerminalError` (already present in package `internal/terminal`) with `Op: "tmux_attach"`, `Op: "tmux_pipe"`, etc. Wrap underlying errors with `fmt.Errorf("...: %w", err)` so `errors.Is`/`errors.As` still work. See `/Users/linus/.claude/projects/-Users-linus-Development-mashed/memory/feedback_error_handling.md` for the project's preferred pattern.
- **Sentinels:** Add `ErrPaneDead = errors.New("terminal: pane is dead")` and `ErrTmuxUnavailable = errors.New("terminal: tmux not available")`.
- **FIFO cleanup on crash:** On `Close()`, always `os.Remove(fifoPath)` and `os.RemoveAll(tempDir)` even if other cleanup fails. Use `defer` chains.
- **`tmux pipe-pane -o` flag:** `-o` is critical — it stops any existing pipe before starting the new one so concurrent attachments don't fight. Document in a code comment.
- **Multi-client pane attachment:** Two clients attaching to the same pane creates two `TmuxAttachment` objects with two FIFOs and two `pipe-pane -o` invocations. The `-o` flag means the second attach replaces the first — by design, but worth noting. Multi-tab viewing of a single BMAD pane is considered out of scope for this story; it works for the dominant case of one viewer at a time.
- **Context cancellation:** All shell invocations use `ctx.WithTimeout(parent, 5*time.Second)` for control commands (pane-death check, send-keys). The FIFO reader uses the long-lived attachment context.

### Risks & Edge Cases

1. **FIFO creation fails** (read-only temp dir, SELinux) → automatic fallback to capture-pane polling. Logged once.
2. **Pane dies between attach and first read** → `Read` returns `io.EOF` on the first call.
3. **tmux binary missing** → `NewTmuxAdapter` does NOT probe; `Attach` surfaces the error on first exec. Add a helper `IsTmuxAvailable() bool` using `exec.LookPath("tmux")`. bridge-03 can call it before routing.
4. **Very chatty pane** (continuous claude output): FIFO handles backpressure via blocking writes. Buffer size 64 KiB in `bufio.Reader`.
5. **Resize on dead pane**: returns wrapped error; caller decides if it's fatal.
6. **Concurrent attachments to different panes**: independent FIFOs, no shared state beyond the adapter map (guarded by `sync.Mutex`).
7. **Windows**: FIFO via `syscall.Mkfifo` is POSIX only. This project is a Wails desktop app targeting macOS first; mark FIFO code with `//go:build !windows` and have the polling fallback cover Windows if it ever matters. For this story, target macOS and Linux only.

### Reference Files

- `internal/bmad/executor.go` lines 820-915 — pattern of using `runCmd` with tmux and pane-death detection
- `internal/bmad/question.go` lines 128-145 — `escapeTmuxLiteral` source to copy/share
- `internal/terminal/session.go` lines 216-260 — `AddClient` / `clientReader` goroutine pattern to mirror for adapter's read goroutine
- `internal/terminal/manager.go` — `TerminalError` custom error type already in the package (reuse for `Op` field)
- `.wolf/cerebrum.md` — check Do-Not-Repeat for any past tmux adapter pitfalls

Reference skills: `/golang-testing`, `/golang-error-handling`, `/simplify`.

## Acceptance Criteria

**AC-1: Attach validates the pane and replays scrollback**
- Given a mock `CommandRunner` where `tmux list-panes -t "bmad-xx:0.0" -F "#{pane_dead}"` returns `"0"`
- And `tmux capture-pane -p -J -e -S -5000 -t "bmad-xx:0.0"` returns `"hello history\n"`
- When `adapter.Attach(ctx, "bmad-xx:0.0")` is called
- Then it returns a non-nil `*TmuxAttachment` and nil error
- And the attachment's `Read` yields `"hello history\n"` before any live bytes

**AC-2: Attach rejects a dead pane**
- Given the mock runner returns `"1"` for the `pane_dead` check
- When `adapter.Attach(ctx, "bmad-dead:0.0")` is called
- Then the returned error wraps `ErrPaneDead`
- And no `pipe-pane` command is invoked

**AC-3: SendInput uses send-keys -l with escaped literal**
- Given an active attachment to `"bmad-xx:0.0"`
- When `attachment.SendInput([]byte("it's a test"))` is called
- Then the runner is invoked with `tmux send-keys -l -t bmad-xx:0.0 {escaped("it's a test")}`
- And the returned error is nil

**AC-4: SendKey forwards special keys without -l**
- Given an active attachment
- When `attachment.SendKey("Enter")` is called
- Then the runner is invoked with `tmux send-keys -t {target} Enter`

**AC-5: Resize invokes tmux resize-window**
- Given an active attachment
- When `attachment.Resize(120, 40)` is called
- Then the runner is invoked with `tmux resize-window -t {target} -x 120 -y 40`

**AC-6: Read returns io.EOF when the pane dies**
- Given an active attachment to a pane that transitions from alive to dead mid-stream
- When the background pane-death watcher detects `pane_dead="1"`
- Then a subsequent `Read` call returns `io.EOF`
- And the attachment's internal context is cancelled

**AC-7: Close is idempotent and cleans up the FIFO**
- Given an active attachment with a FIFO at a known temp path
- When `attachment.Close()` is called twice
- Then the second call does not panic and returns nil
- And the FIFO file no longer exists on disk
- And no `pipe-pane` reinvocation happens on the second Close

**AC-8: Fallback to capture-pane polling when FIFO creation fails**
- Given a test that forces `syscall.Mkfifo` to return an error (via injectable function)
- When `adapter.Attach` is called
- Then the attachment still succeeds
- And the adapter enters polling mode
- And `Read` yields bytes captured from subsequent `tmux capture-pane -p -J -t {target}` invocations

## BDD Test Scenarios

### Scenario 1: Basic attach

```gherkin
Feature: TmuxAdapter attach

  Scenario: Attach to live pane
    Given a mock CommandRunner with pane_dead="0"
    And capture-pane returns "previous output\n"
    When Attach is called with target "bmad-foo-main-node-deadbeef:0.0"
    Then an attachment is returned
    And Read yields "previous output\n" as the first bytes

  Scenario: Attach to dead pane
    Given a mock CommandRunner with pane_dead="1"
    When Attach is called
    Then the error wraps ErrPaneDead
```

### Scenario 2: Input and resize

```gherkin
Feature: TmuxAdapter input forwarding

  Scenario: SendInput with literal text
    Given an active attachment to "bmad-xx:0.0"
    When SendInput("ls -la") is called
    Then runCmd receives "tmux send-keys -l -t bmad-xx:0.0 ls -la"

  Scenario: SendInput with single quotes
    Given an active attachment
    When SendInput("it's fine") is called
    Then the literal passed to send-keys -l preserves the apostrophe per EscapeTmuxLiteral

  Scenario: SendKey Enter
    Given an active attachment
    When SendKey("Enter") is called
    Then runCmd receives "tmux send-keys -t bmad-xx:0.0 Enter"

  Scenario: Resize
    Given an active attachment
    When Resize(120, 40) is called
    Then runCmd receives "tmux resize-window -t bmad-xx:0.0 -x 120 -y 40"
```

### Scenario 3: Pane death

```gherkin
Feature: Pane death detection

  Scenario: Pane dies after attach
    Given an attached pane that initially reports pane_dead="0"
    When pane_dead flips to "1" during the attachment
    Then a subsequent Read call returns io.EOF
    And the attachment context is cancelled
```

### Scenario 4: Close idempotence

```gherkin
Feature: TmuxAttachment close

  Scenario: Double close
    Given an active attachment with a FIFO at /tmp/mashed-abc/fifo
    When Close is called twice
    Then both calls return nil
    And /tmp/mashed-abc/fifo does not exist
    And pipe-pane stop command is only invoked once
```

### Scenario 5: FIFO fallback

```gherkin
Feature: Polling fallback

  Scenario: Mkfifo fails
    Given the injectable mkfifo returns an error
    When Attach is called
    Then the attachment is returned successfully
    And a.fallbackPolling is true
    And Read yields bytes from polled capture-pane invocations
```

## Tasks / Subtasks

- [ ] Task 1: Define `TmuxAdapter` and `TmuxAttachment` types with method stubs (AC: all)
  - [ ] Subtask 1a: Create `internal/terminal/tmux_adapter.go` with struct definitions
  - [ ] Subtask 1b: Declare local `CommandRunner` type alias and `DefaultCommandRunner`
  - [ ] Subtask 1c: Add `ErrPaneDead`, `ErrTmuxUnavailable` sentinels
  - [ ] Subtask 1d: Add `EscapeTmuxLiteral` in `internal/terminal/tmux_escape.go`
  - [ ] Subtask 1e: Add `IsTmuxAvailable() bool` helper

- [ ] Task 2: Implement `Attach` with scrollback replay and FIFO setup (AC: AC-1, AC-2, AC-8)
  - [ ] Subtask 2a: Pane liveness check via `tmux list-panes -F "#{pane_dead}"`
  - [ ] Subtask 2b: Scrollback capture via `tmux capture-pane -p -J -e -S -5000`
  - [ ] Subtask 2c: FIFO creation with injectable `mkfifo` function; fallback flag
  - [ ] Subtask 2d: Launch `tmux pipe-pane -o -t {target} 'cat > {fifo}'`
  - [ ] Subtask 2e: Start background FIFO reader goroutine and pane-death watcher

- [ ] Task 3: Implement `Read` as an `io.Reader` that merges scrollback + live stream (AC: AC-1, AC-6)
  - [ ] Subtask 3a: Buffer scrollback bytes in a `bytes.Buffer`; drain before live reads
  - [ ] Subtask 3b: Block on FIFO reader channel; return on EOF
  - [ ] Subtask 3c: Signal `io.EOF` when pane-death watcher cancels context

- [ ] Task 4: Implement `SendInput`, `SendKey`, `Resize` (AC: AC-3, AC-4, AC-5)
  - [ ] Subtask 4a: `SendInput` uses `send-keys -l` with `EscapeTmuxLiteral`
  - [ ] Subtask 4b: `SendKey` uses `send-keys` without `-l`
  - [ ] Subtask 4c: `Resize` uses `resize-window -x -y`

- [ ] Task 5: Implement `Close` and idempotent cleanup (AC: AC-7)
  - [ ] Subtask 5a: `sync.Once`-guarded cleanup
  - [ ] Subtask 5b: Stop pipe-pane via `tmux pipe-pane -t {target}` (no command argument stops)
  - [ ] Subtask 5c: Remove FIFO path and temp dir
  - [ ] Subtask 5d: Cancel attachment context; wait for goroutines

- [ ] Task 6: Implement polling fallback (AC: AC-8)
  - [ ] Subtask 6a: 200ms ticker driving `capture-pane -p -J`
  - [ ] Subtask 6b: Diff against last snapshot; emit delta bytes

- [ ] Task 7: Write table-driven tests in `tmux_adapter_test.go` (AC: AC-1 through AC-8)
  - [ ] Subtask 7a: Mock CommandRunner that records all invocations with args
  - [ ] Subtask 7b: Happy-path attach with scrollback
  - [ ] Subtask 7c: Dead-pane rejection
  - [ ] Subtask 7d: SendInput/SendKey/Resize verification
  - [ ] Subtask 7e: Pane death mid-stream
  - [ ] Subtask 7f: Close idempotence
  - [ ] Subtask 7g: Mkfifo failure → polling fallback
  - [ ] Subtask 7h: `-race` test exercising concurrent SendInput + Read

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/tmux_adapter.go` and `internal/terminal/tmux_escape.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Story status updated to `done`
