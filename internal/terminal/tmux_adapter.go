//go:build !windows

// Package terminal — TmuxAdapter bridges already-running tmux panes (typically
// spawned by BMAD's executor) to the Wails session model WITHOUT going through
// the PTY-based SessionManager. This allows a single long-running tmux pane to
// be attached, streamed to the UI, and written to from multiple sources.
//
// Concurrency model
// =================
//
// Per attachment, the adapter spawns up to three goroutines, all of which exit
// when att.ctx is cancelled:
//
//  1. FIFO reader  — blocks on os.File.Read of the named FIFO, pushes chunks
//     onto att.liveCh. Exits when the closer goroutine (#2) closes the file
//     out from under it.
//  2. FIFO closer  — blocks on <-att.ctx.Done(), then f.Close()s the FIFO to
//     unblock goroutine #1. Separated out because os.File.Read is not
//     context-aware on POSIX FIFOs.
//  3. Pane death watcher — ticks every watcherInterval and runs
//     `tmux list-panes -F #{pane_dead}`. On death or runner error it stores
//     io.EOF in att.readErr and cancels att.ctx.
//
// When setupFIFO (syscall.Mkfifo) fails, the adapter switches to a polling
// fallback (attachment.fallbackPolling = true). In this mode there is no FIFO
// reader / closer pair; instead a single polling goroutine runs
// `tmux capture-pane` at pollingInterval and emits the full output whenever it
// differs from the previous capture.
//
// Close cascade
// =============
//
// TmuxAttachment.Close is idempotent via sync.Once. It:
//  1. Cancels att.ctx (unblocks the pane death watcher / polling loop).
//  2. Closes att.fifoFile (unblocks the FIFO reader via its EOF).
//  3. Runs `tmux pipe-pane -t <target>` with NO inline command to stop the
//     pipeline tmux-side. Uses a fresh background context with a short
//     timeout because att.ctx is already cancelled.
//  4. Waits for all per-attachment goroutines via att.wg.
//  5. Removes the tmp dir (RemoveAll).
//  6. Deregisters from the adapter's active set.
//
// TmuxAdapter.Close iterates all active attachments and calls Close on each,
// which cascades via the per-attachment discipline above.
package terminal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// CommandRunner matches the signature used throughout the project for
// exec.CommandContext-style testability. Production wiring uses
// DefaultCommandRunner; tests inject a mock that records invocations.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// DefaultCommandRunner runs the given command and returns its combined output.
// Matches the bmad executor's runCmd signature so callers can share mocks.
func DefaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// Timing constants — tunable intervals for the per-attachment goroutines.
// watcherInterval is deliberately short (100ms) because the test suite
// operates under a 2-second context budget. At 100ms the death watcher fires
// ~10× within test budget while remaining cheap in production (one
// short-lived `tmux list-panes` per tick per attachment).
const (
	watcherInterval = 100 * time.Millisecond
	pollingInterval = 200 * time.Millisecond
	runCmdTimeout   = 5 * time.Second
	watcherRunCmdTO = 3 * time.Second
	stopRunCmdTO    = 2 * time.Second
	// fifoRetryInterval paces the reader goroutine when its non-blocking
	// FIFO Read returns EAGAIN. Short enough that live output feels
	// instant to the user; long enough that the retry loop never burns
	// noticeable CPU while waiting for tmux pipe-pane's cat to connect.
	fifoRetryInterval = 20 * time.Millisecond
	liveChBuffer      = 64
	fifoReadBuf       = 4096
	scrollbackLines   = 5000
	fifoFilePermBits  = 0o600

	// Tmux binary + subcommand names used when invoking the runner.
	// Names differ from the test-file constants (sub*) because tests and
	// production code both compile into the same package and must not
	// declare duplicate identifiers.
	cmdTmux        = "tmux"
	cmdListPanes   = "list-panes"
	cmdCapturePane = "capture-pane"
	cmdPipePane    = "pipe-pane"
	cmdSendKeys    = "send-keys"
	cmdResizeWin   = "resize-window"
	fifoFileName   = "pane.fifo"
)

// TmuxAdapter manages bridges to already-running tmux panes.
type TmuxAdapter struct {
	runCmd CommandRunner
	// mkfifo is an injectable wrapper around syscall.Mkfifo so tests can
	// force the polling-fallback code path. Do NOT rename — the test
	// helper in tmux_adapter_testhelpers_test.go pokes this field by
	// name (overrideMkfifo).
	mkfifo  func(path string, mode uint32) error
	tempDir string

	mu     sync.Mutex
	closed bool
	active map[*TmuxAttachment]struct{}
}

// NewTmuxAdapter builds an adapter. If runner is nil, DefaultCommandRunner is
// used. The mkfifo field defaults to syscall.Mkfifo; tests override it via
// the internal overrideMkfifo helper.
func NewTmuxAdapter(runner CommandRunner) *TmuxAdapter {
	if runner == nil {
		runner = DefaultCommandRunner
	}
	return &TmuxAdapter{
		runCmd:  runner,
		mkfifo:  syscall.Mkfifo,
		tempDir: os.TempDir(),
		active:  make(map[*TmuxAttachment]struct{}),
	}
}

// TmuxAttachment represents a single live bridge to a tmux pane.
type TmuxAttachment struct {
	adapter *TmuxAdapter
	target  string

	ctx    context.Context
	cancel context.CancelFunc

	// fifoPath is empty in polling-fallback mode. The containing tmp dir
	// is derived as filepath.Dir(fifoPath) at cleanup time.
	fifoPath string
	fifoFile *os.File // read side (O_RDONLY|O_NONBLOCK); nil in polling-fallback mode
	// fifoKeepalive is a second handle to the FIFO opened O_WRONLY so the
	// Go process itself is always a writer on the pipe. Without this,
	// POSIX defines a FIFO with no writers as end-of-file, which races
	// against the asynchronous `tmux pipe-pane "cat > fifo"` spawn: if
	// the read side opens before cat connects, the very first Read on
	// the read fd returns (0, io.EOF) and the attachment tears itself
	// down before any live output can stream. Holding a WRONLY fd here
	// makes the "no external writer" state invisible to the read fd, so
	// the stream only terminates when Close() closes both handles.
	// nil in polling-fallback mode.
	fifoKeepalive *os.File

	// fallbackPolling — do NOT rename; tmux_adapter_testhelpers_test.go
	// pokes this field by name (attachmentIsPolling).
	fallbackPolling bool

	// scrollback is drained first by Read and also holds any stash from
	// a live chunk that exceeded the caller's buffer. Accessed only on
	// the Read caller's goroutine, so no mutex is needed.
	scrollback *bytes.Buffer

	liveCh chan []byte

	// readErr — atomic.Value so Read can poll without locks while the
	// watcher / FIFO reader stores the terminal error.
	readErr atomic.Value

	closeOnce sync.Once
	wg        sync.WaitGroup
}

// Attach validates the target pane, captures scrollback, sets up a FIFO +
// pipe-pane (or falls back to polling) and spawns the death watcher.
//
// Returns TmuxSession (interface defined in bridge.go) rather than the
// concrete *TmuxAttachment so that *TmuxAdapter implicitly satisfies the
// bridge's TmuxAttacher interface. The concrete type is still recoverable
// via type assertion in tests that need access to internal fields.
func (a *TmuxAdapter) Attach(parent context.Context, paneTarget string) (TmuxSession, error) {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil, &TerminalError{Op: "tmux_attach", Err: errors.New("adapter is closed")}
	}
	a.mu.Unlock()

	// 1. Pane liveness check.
	out, err := a.runCmd(parent, cmdTmux, cmdListPanes, "-t", paneTarget, "-F", "#{pane_dead}")
	if err != nil {
		return nil, &TerminalError{Op: "tmux_attach", Err: fmt.Errorf("list-panes %q: %w", paneTarget, err)}
	}
	if strings.TrimSpace(string(out)) != "0" {
		return nil, &TerminalError{Op: "tmux_attach", Err: fmt.Errorf("%w: %s", ErrPaneDead, paneTarget)}
	}

	// 2. Scrollback capture — non-fatal on error; Read still operates
	// with an empty buffer.
	scrollback := &bytes.Buffer{}
	if capOut, capErr := a.runCmd(
		parent, cmdTmux, cmdCapturePane,
		"-p", "-J", "-e", "-S", "-"+strconv.Itoa(scrollbackLines),
		"-t", paneTarget,
	); capErr == nil {
		scrollback.Write(capOut)
	}

	// 3. Build the attachment. Derive a cancelable context from parent
	// so adapter.Close can cascade via att.cancel.
	attCtx, cancel := context.WithCancel(parent)
	att := &TmuxAttachment{
		adapter:    a,
		target:     paneTarget,
		ctx:        attCtx,
		cancel:     cancel,
		scrollback: scrollback,
		liveCh:     make(chan []byte, liveChBuffer),
	}

	// 4. FIFO setup — on mkfifo failure, switch to polling fallback.
	if err := a.setupFIFO(att); err != nil {
		att.fallbackPolling = true
		a.startPollingLoop(att)
	} else {
		if err := a.startPipePane(att); err != nil {
			_ = att.cleanupFIFO()
			cancel()
			return nil, &TerminalError{Op: "tmux_attach", Err: fmt.Errorf("pipe-pane: %w", err)}
		}
		if err := a.startFIFOReader(att); err != nil {
			_ = att.cleanupFIFO()
			cancel()
			return nil, &TerminalError{Op: "tmux_attach", Err: fmt.Errorf("open fifo: %w", err)}
		}
	}

	// 5. Pane death watcher — always runs (even in polling mode).
	a.startPaneDeathWatcher(att)

	a.mu.Lock()
	a.active[att] = struct{}{}
	a.mu.Unlock()

	return att, nil
}

// setupFIFO creates a tmp dir + named FIFO via the injectable mkfifo field.
func (a *TmuxAdapter) setupFIFO(att *TmuxAttachment) error {
	tmpDir, err := os.MkdirTemp(a.tempDir, "mashed-tmux-")
	if err != nil {
		return err
	}
	fifo := filepath.Join(tmpDir, fifoFileName)
	if err := a.mkfifo(fifo, fifoFilePermBits); err != nil {
		_ = os.RemoveAll(tmpDir)
		return err
	}
	att.fifoPath = fifo
	return nil
}

// startPipePane issues `tmux pipe-pane -o -t <target> "cat > '<fifo>'"`.
// The `-o` flag stops any pre-existing pipe on the pane so re-attaches do
// not fork the stream between two FIFOs.
//
// The FIFO path is wrapped in single quotes because tmux executes the
// inline command via `/bin/sh -c`; without quoting, a `$TMPDIR` containing
// a space (e.g. "/home/me/my dir/tmp") word-splits the redirect target and
// silently breaks the pipe. Single quotes are safe: os.MkdirTemp's prefix
// is caller-supplied alphanumeric and its random suffix is base32, so the
// resulting path can never embed a single quote.
func (a *TmuxAdapter) startPipePane(att *TmuxAttachment) error {
	inline := fmt.Sprintf("cat > '%s'", att.fifoPath)
	ctx, cancel := context.WithTimeout(att.ctx, runCmdTimeout)
	defer cancel()
	_, err := a.runCmd(ctx, cmdTmux, cmdPipePane, "-o", "-t", att.target, inline)
	return err
}

// startFIFOReader opens the FIFO and spawns two goroutines: a reader that
// blocks on os.File.Read (via the Go netpoll) and pushes chunks onto
// att.liveCh, and a closer that waits for att.ctx to cancel and then closes
// the file to unblock the reader.
//
// The FIFO is opened TWICE: first O_RDONLY|O_NONBLOCK for reading, then
// O_WRONLY|O_NONBLOCK as a keepalive writer. Both opens are needed to
// correctly handle the race against tmux's async `pipe-pane "cat > fifo"`:
//
//   - O_RDONLY alone would block in open(2) until a writer opens the far
//     end, stalling Attach indefinitely if cat never spawns.
//   - O_RDONLY|O_NONBLOCK open succeeds immediately without a writer, and
//     Go's netpoll integrates FIFO reads so Read blocks on the poller
//     waiting for data.
//   - BUT POSIX defines read() on a FIFO with no writers as returning 0
//     bytes ("end of file"). startPipePane only queues a command to the
//     tmux server and returns before `cat > fifo` has actually connected,
//     so on a slow schedule the read fd is live before any writer exists.
//     The first Read then returns (0, io.EOF) and tears the attachment
//     down before live output can stream — users see a blank viewport
//     that flips to "[disconnected]".
//   - The O_WRONLY keepalive fd makes the Go process itself a writer on
//     the FIFO. POSIX then guarantees no spurious EOF because there is
//     always at least one writer from the kernel's POV. The read-side
//     Read blocks on netpoll until actual data arrives (from cat, when
//     it connects) or the closer goroutine closes both handles.
//   - Opening WRONLY AFTER the RDONLY fd is safe: O_WRONLY|O_NONBLOCK on
//     a FIFO with no reader would fail with ENXIO, but we have a reader
//     (ourselves), so the open succeeds non-blockingly.
func (a *TmuxAdapter) startFIFOReader(att *TmuxAttachment) error {
	f, err := os.OpenFile(att.fifoPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	att.fifoFile = f

	// Keepalive writer — must be opened AFTER the reader fd is live, else
	// O_WRONLY|O_NONBLOCK returns ENXIO. Non-fatal: if the keepalive open
	// fails for any reason, the attachment still works, it is just
	// susceptible to the pipe-pane race. Log and continue so we never
	// regress from "works racily" to "does not attach at all".
	if kf, kerr := os.OpenFile(att.fifoPath, os.O_WRONLY|syscall.O_NONBLOCK, 0); kerr == nil {
		att.fifoKeepalive = kf
	}

	att.wg.Add(1)
	go func() {
		defer att.wg.Done()
		buf := make([]byte, fifoReadBuf)
		for {
			n, rerr := f.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				select {
				case att.liveCh <- chunk:
				case <-att.ctx.Done():
					return
				}
			}
			if rerr != nil {
				// EAGAIN ("resource temporarily unavailable") means the
				// FIFO has at least one writer (the keepalive fd) but no
				// data is currently available. Go's netpoll does not
				// integrate FIFO EAGAIN reliably on macOS kqueue, so the
				// error bubbles up to this goroutine. Treat it as "wait
				// briefly and retry" so the reader keeps polling the fd
				// until either data arrives (from tmux pipe-pane's cat
				// connecting, or from subsequent pane output) or att.ctx
				// is cancelled by the closer goroutine.
				if errors.Is(rerr, syscall.EAGAIN) {
					select {
					case <-att.ctx.Done():
						return
					case <-time.After(fifoRetryInterval):
						continue
					}
				}
				if !errors.Is(rerr, io.EOF) && !errors.Is(rerr, os.ErrClosed) {
					att.readErr.Store(rerr)
				}
				att.cancel()
				return
			}
		}
	}()

	att.wg.Add(1)
	go func() {
		defer att.wg.Done()
		<-att.ctx.Done()
		// Close the keepalive WRONLY fd first: this transitions the FIFO
		// to "no writers" which, combined with any pending buffered data
		// draining, lets the reader observe a clean EOF after the final
		// chunks are delivered. Then close the read fd to unblock the
		// reader goroutine via its netpoll wake.
		if att.fifoKeepalive != nil {
			_ = att.fifoKeepalive.Close()
		}
		_ = f.Close() // unblocks the reader goroutine's f.Read
	}()

	return nil
}

// startPaneDeathWatcher polls `tmux list-panes` at watcherInterval and flips
// the attachment to EOF when the pane dies (or the runner errors).
func (a *TmuxAdapter) startPaneDeathWatcher(att *TmuxAttachment) {
	att.wg.Add(1)
	go func() {
		defer att.wg.Done()
		ticker := time.NewTicker(watcherInterval)
		defer ticker.Stop()
		for {
			select {
			case <-att.ctx.Done():
				return
			case <-ticker.C:
			}

			checkCtx, cancelCheck := context.WithTimeout(att.ctx, watcherRunCmdTO)
			out, err := a.runCmd(checkCtx, cmdTmux, cmdListPanes, "-t", att.target, "-F", "#{pane_dead}")
			cancelCheck()

			if err != nil || strings.TrimSpace(string(out)) != "0" {
				att.readErr.Store(io.EOF)
				att.cancel()
				return
			}
		}
	}()
}

// startPollingLoop powers the fallback path when FIFO setup fails. It runs
// `tmux capture-pane` at pollingInterval and emits the full output whenever
// it differs from the previous capture.
func (a *TmuxAdapter) startPollingLoop(att *TmuxAttachment) {
	att.wg.Add(1)
	go func() {
		defer att.wg.Done()
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()

		var last []byte
		for {
			select {
			case <-att.ctx.Done():
				return
			case <-ticker.C:
			}

			capCtx, cancelCap := context.WithTimeout(att.ctx, runCmdTimeout)
			out, err := a.runCmd(capCtx, cmdTmux, cmdCapturePane, "-p", "-J", "-t", att.target)
			cancelCap()
			if err != nil {
				att.readErr.Store(err)
				att.cancel()
				return
			}
			if !bytes.Equal(out, last) {
				chunk := make([]byte, len(out))
				copy(chunk, out)
				select {
				case att.liveCh <- chunk:
				case <-att.ctx.Done():
					return
				}
				last = out
			}
		}
	}()
}

// Read implements io.Reader. It drains the scrollback buffer first, then
// blocks on the live channel until a chunk arrives or the attachment is
// closed. Short-buffer handling stashes the unread remainder in scrollback
// so the next Read serves it in order.
func (att *TmuxAttachment) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// 1. Drain scrollback (which also holds any stashed remainder from a
	// prior short-buffer live read).
	if att.scrollback.Len() > 0 {
		return att.scrollback.Read(p)
	}

	// 2. Check whether the stream has already terminated.
	if err := att.loadReadErr(); err != nil {
		return 0, err
	}

	// 3. Block on the live channel.
	select {
	case chunk, ok := <-att.liveCh:
		if !ok {
			return 0, io.EOF
		}
		n := copy(p, chunk)
		if n < len(chunk) {
			att.scrollback.Write(chunk[n:])
		}
		return n, nil
	case <-att.ctx.Done():
		if err := att.loadReadErr(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
}

// loadReadErr returns the stored terminal error (or nil).
func (att *TmuxAttachment) loadReadErr() error {
	if v := att.readErr.Load(); v != nil {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}

// runTmux runs `tmux <args>` under a short-timeout context derived from
// att.ctx, returning a *TerminalError tagged with the given op on failure.
// Shared by SendInput, SendKey, and Resize so the timeout + error-wrap
// boilerplate lives in one place.
func (att *TmuxAttachment) runTmux(op string, args ...string) error {
	ctx, cancel := context.WithTimeout(att.ctx, runCmdTimeout)
	defer cancel()
	if _, err := att.adapter.runCmd(ctx, cmdTmux, args...); err != nil {
		return &TerminalError{Op: op, Err: err}
	}
	return nil
}

// SendInput injects a raw byte sequence into the tmux pane's PTY, one
// byte at a time, via `tmux send-keys -H`. This is the live-terminal
// input path driven by the xterm.js client over the WebSocket bridge,
// so EVERY byte — including control codes, escape sequences, and UTF-8
// continuation bytes — must round-trip faithfully.
//
// The previous implementation routed this through EscapeTmuxLiteral +
// `send-keys -l`, which stripped every byte in [0x00, 0x1f] plus 0x7f
// (DEL) from the input. xterm sends 0x7f for Backspace, 0x0d for Enter,
// and 0x1b[A..D for the arrow keys — so the stripping made Backspace,
// Enter, arrows, Ctrl+C, and bracketed-paste markers (\x1b[200~ …
// \x1b[201~) silently unreachable. Users could type but not delete,
// submit, or navigate — which is how we discovered the bug.
//
// `send-keys -H` interprets each argument as a two-char hex byte value
// and injects the decoded byte directly into the pane; tmux does NOT
// parse it as a symbolic key name. Multi-byte sequences (arrows as
// ESC[A, UTF-8 as e2 9d af, etc.) are reconstructed at the pane side by
// the running process's own input decoder (claude's readline, the
// shell, etc.) — exactly as if the bytes had arrived over a real PTY.
//
// Batching: each WebSocket frame becomes one tmux invocation with up to
// len(data) hex arguments. Typical interactive typing is 1–20 bytes per
// frame, well under macOS ARG_MAX (~1 MB). For very large pastes we
// still emit a single process call; argv length is capped by the
// kernel, not by us, so the practical upper bound is roughly 300 KB of
// pasted text per frame. The frontend's clipboard paste path already
// arrives as one frame per paste so this is fine.
//
// Empty input is a no-op — not an error — to match the contract the
// bridge's writer goroutine expects on a closed peer.
func (att *TmuxAttachment) SendInput(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	args := make([]string, 0, 4+len(data))
	args = append(args, cmdSendKeys, "-H", "-t", att.target)
	args = append(args, formatSendKeysHex(data)...)
	return att.runTmux("tmux_send_input", args...)
}

// formatSendKeysHex converts a byte slice into the two-char lowercase hex
// arguments expected by `tmux send-keys -H`. Shared by both
// TmuxAttachment.SendInput (target from attachment) and
// TmuxAdapter.SendInputToTarget (explicit target string).
func formatSendKeysHex(data []byte) []string {
	hex := make([]string, len(data))
	for i, b := range data {
		hex[i] = fmt.Sprintf("%02x", b)
	}
	return hex
}

// SendInputToTarget injects a raw byte sequence into a tmux pane identified by
// target string, via `tmux send-keys -H`. Unlike TmuxAttachment.SendInput, this
// does not require a live attachment — it operates on any target string directly.
func (a *TmuxAdapter) SendInputToTarget(ctx context.Context, target string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	args := make([]string, 0, 5+len(data))
	args = append(args, cmdSendKeys, "-H", "-t", target)
	args = append(args, formatSendKeysHex(data)...)
	_, err := a.runCmd(ctx, cmdTmux, args...)
	return err
}

// SendKey forwards a symbolic key (Enter, C-c, Tab, Up, BSpace, …) through
// `tmux send-keys` WITHOUT the -l flag so tmux interprets the token.
func (att *TmuxAttachment) SendKey(key string) error {
	return att.runTmux("tmux_send_key", cmdSendKeys, "-t", att.target, key)
}

// Resize invokes `tmux resize-window -x <cols> -y <rows>`.
func (att *TmuxAttachment) Resize(cols, rows uint16) error {
	return att.runTmux(
		"tmux_resize", cmdResizeWin,
		"-t", att.target,
		"-x", strconv.Itoa(int(cols)),
		"-y", strconv.Itoa(int(rows)),
	)
}

// Close is idempotent via sync.Once. It cancels att.ctx, stops the tmux
// pipe, waits for all per-attachment goroutines to exit, cleans up the tmp
// dir, and deregisters from the adapter's active set.
func (att *TmuxAttachment) Close() error {
	att.closeOnce.Do(func() {
		att.cancel()

		// Only stop pipe-pane in FIFO mode. The fresh background context
		// is intentional: att.ctx is already cancelled above, so any
		// derived context would fail immediately.
		if att.fifoPath != "" {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), stopRunCmdTO)
			_, _ = att.adapter.runCmd(stopCtx, cmdTmux, cmdPipePane, "-t", att.target)
			stopCancel()
		}

		att.wg.Wait()
		_ = att.cleanupFIFO()

		att.adapter.mu.Lock()
		delete(att.adapter.active, att)
		att.adapter.mu.Unlock()
	})
	return nil
}

// cleanupFIFO removes the tmp dir (and therefore the FIFO node inside it).
// The containing directory is derived from fifoPath so we do not have to
// carry a redundant tmpDir field. Safe to call multiple times; Close
// serialises via sync.Once.
func (att *TmuxAttachment) cleanupFIFO() error {
	if att.fifoPath == "" {
		return nil
	}
	err := os.RemoveAll(filepath.Dir(att.fifoPath))
	att.fifoPath = ""
	att.fifoFile = nil
	att.fifoKeepalive = nil
	return err
}

// Close cascades — marks the adapter closed (rejecting new Attach calls)
// and closes every active attachment in turn. Safe to call multiple times.
func (a *TmuxAdapter) Close() error {
	a.mu.Lock()
	a.closed = true
	active := make([]*TmuxAttachment, 0, len(a.active))
	for att := range a.active {
		active = append(active, att)
	}
	a.mu.Unlock()

	for _, att := range active {
		_ = att.Close()
	}
	return nil
}
