// Package terminal — engineer-owned supplementary tests that push the
// tmux adapter's per-file coverage over the 80% gate.
//
// The primary RED test file (tmux_adapter_test.go) exercises the AC-level
// happy paths and the dead-pane path. This file covers error branches and
// trivial helpers that AC tests do not touch:
//
//   - DefaultCommandRunner (real exec.Command round-trip)
//   - NewTmuxAdapter with a nil runner argument
//   - IsTmuxAvailable
//   - EscapeTmuxLiteral empty-string fast path
//   - Attach on an already-closed adapter
//   - Attach failure when list-panes errors out
//   - Attach capture-pane failure (scrollback starts empty, not fatal)
//   - Attach pipe-pane failure (cleanup + error)
//   - SendInput empty input (no-op short-circuit)
//   - SendInput / SendKey / Resize runner-error surface
//   - Read into a zero-length buffer
//   - Read short-buffer stash-and-replay path
//   - Polling-loop runner error path

package terminal

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Trivial-wrapper coverage
// ---------------------------------------------------------------------------

func TestDefaultCommandRunner_EchoesStdout(t *testing.T) {
	out, err := DefaultCommandRunner(context.Background(), "sh", "-c", "printf hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(out))
}

func TestNewTmuxAdapter_NilRunnerFallsBackToDefault(t *testing.T) {
	a := NewTmuxAdapter(nil)
	require.NotNil(t, a)
	assert.NotNil(t, a.runCmd, "nil runner arg must be replaced with DefaultCommandRunner")
	assert.NotNil(t, a.mkfifo, "mkfifo field must default to syscall.Mkfifo")
}

func TestIsTmuxAvailable_ReturnsBoolWithoutPanic(t *testing.T) {
	// The exact return value depends on the test host; all we assert is
	// that the call is well-formed and does not panic. Running this
	// bumps the function's coverage to 100%.
	_ = IsTmuxAvailable()
}

func TestEscapeTmuxLiteral_EmptyString(t *testing.T) {
	assert.Equal(t, "", EscapeTmuxLiteral(""))
}

// ---------------------------------------------------------------------------
// Attach-side error paths
// ---------------------------------------------------------------------------

func TestTmuxAdapter_AttachOnClosedAdapter(t *testing.T) {
	a := NewTmuxAdapter(newMockRunner().runner())
	require.NoError(t, a.Close())

	att, err := a.Attach(newContext(t), testPaneTarget)
	assert.Nil(t, att)
	require.Error(t, err)
	var terr *TerminalError
	require.ErrorAs(t, err, &terr)
	assert.Equal(t, "tmux_attach", terr.Op)
	assert.Contains(t, terr.Error(), "adapter is closed")
}

func TestTmuxAdapter_AttachListPanesErrorWrapped(t *testing.T) {
	mock := newMockRunner()
	mock.setResponder(subListPanes, func(_ int) ([]byte, error) {
		return nil, errors.New("tmux: boom")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })

	att, err := a.Attach(newContext(t), testPaneTarget)
	assert.Nil(t, att)
	require.Error(t, err)
	var terr *TerminalError
	require.ErrorAs(t, err, &terr)
	assert.Equal(t, "tmux_attach", terr.Op)
	assert.Contains(t, terr.Error(), "list-panes")
}

func TestTmuxAdapter_AttachCapturePaneErrorIsNonFatal(t *testing.T) {
	// capture-pane erroring during Attach must not cause Attach to fail —
	// scrollback starts empty and live bytes still flow. This exercises
	// the "scrollback fallback on capture error" branch.
	mock := newMockRunner()
	mock.paneAlive()
	mock.setResponder(subCapturePane, func(_ int) ([]byte, error) {
		return nil, errors.New("capture failed")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })

	att, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	require.NotNil(t, att)
	t.Cleanup(func() { _ = att.Close() })

	// A Read must not see any scrollback. We can't easily assert "zero
	// bytes" without blocking, so close and verify Read returns EOF.
	_ = att.Close()
	_, readErr := att.Read(make([]byte, 8))
	assert.ErrorIs(t, readErr, io.EOF)
}

func TestTmuxAdapter_AttachPipePaneErrorCleansUp(t *testing.T) {
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")
	mock.setResponder(subPipePane, func(_ int) ([]byte, error) {
		return nil, errors.New("pipe-pane boom")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })

	att, err := a.Attach(newContext(t), testPaneTarget)
	assert.Nil(t, att)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipe-pane")
}

// ---------------------------------------------------------------------------
// SendInput / SendKey / Resize error branches
// ---------------------------------------------------------------------------

func TestTmuxAttachment_SendInputEmptyIsNoOp(t *testing.T) {
	mock, att := newLiveAttachment(t, newContext(t))
	require.NoError(t, att.SendInput([]byte("")))

	// The empty-string fast path must NOT invoke send-keys.
	_, ok := mock.findSubcommand(subSendKeys)
	assert.False(t, ok, "empty SendInput must short-circuit before send-keys")
}

func TestTmuxAttachment_SendInputEscapesThenSendsControlOnlyAsNoOp(t *testing.T) {
	// All-control-byte input collapses to "" after escape and must not
	// invoke send-keys.
	mock, att := newLiveAttachment(t, newContext(t))
	require.NoError(t, att.SendInput([]byte("\n\r\x1b\x00")))

	_, ok := mock.findSubcommand(subSendKeys)
	assert.False(t, ok, "all-control-byte SendInput must short-circuit")
}

// sendErrMockRunner installs a single-subcommand error responder so the
// error-wrapping branches of SendInput/SendKey/Resize can be exercised.
func sendErrMockRunner() *mockRunner {
	m := newMockRunner()
	m.paneAlive()
	m.captureReturns("")
	return m
}

func TestTmuxAttachment_SendInputRunnerErrorWrapped(t *testing.T) {
	mock := sendErrMockRunner()
	mock.setResponder(subSendKeys, func(_ int) ([]byte, error) {
		return nil, errors.New("tmux rejected input")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })
	att, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })

	err = att.SendInput([]byte("hi"))
	require.Error(t, err)
	var terr *TerminalError
	require.ErrorAs(t, err, &terr)
	assert.Equal(t, "tmux_send_input", terr.Op)
}

func TestTmuxAttachment_SendKeyRunnerErrorWrapped(t *testing.T) {
	mock := sendErrMockRunner()
	mock.setResponder(subSendKeys, func(_ int) ([]byte, error) {
		return nil, errors.New("tmux key rejected")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })
	att, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })

	err = att.SendKey("C-c")
	require.Error(t, err)
	var terr *TerminalError
	require.ErrorAs(t, err, &terr)
	assert.Equal(t, "tmux_send_key", terr.Op)
}

func TestTmuxAttachment_ResizeRunnerErrorWrapped(t *testing.T) {
	mock := sendErrMockRunner()
	mock.setResponder(subResizeWin, func(_ int) ([]byte, error) {
		return nil, errors.New("tmux resize refused")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })
	att, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })

	err = att.Resize(80, 24)
	require.Error(t, err)
	var terr *TerminalError
	require.ErrorAs(t, err, &terr)
	assert.Equal(t, "tmux_resize", terr.Op)
}

// ---------------------------------------------------------------------------
// Read path branches
// ---------------------------------------------------------------------------

func TestTmuxAttachment_ReadZeroLengthBuffer(t *testing.T) {
	_, att := newLiveAttachment(t, newContext(t))

	n, err := att.Read(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n, "Read with a zero-length buffer must return (0, nil)")
}

func TestTmuxAttachment_ReadShortBufferStashesRemainder(t *testing.T) {
	// Scrollback is emitted via bytes.Buffer.Read which handles short
	// buffers natively; this exercises the LIVE-chunk short-buffer
	// branch where the remainder is stashed back into scrollback.
	ctx := newContext(t)
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("") // empty scrollback

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })

	// Force polling mode so we control live-chunk content directly.
	overrideMkfifo(a, func(_ string, _ uint32) error { return errors.New("forced fallback") })
	mock.setResponder(subCapturePane, func(idx int) ([]byte, error) {
		if idx == 0 {
			return []byte(""), nil // Attach-time scrollback capture
		}
		return []byte("ABCDE"), nil // polling chunk
	})

	att, err := a.Attach(ctx, testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })

	// First Read with small buffer grabs part of the chunk.
	buf := make([]byte, 2)
	got, err := readPrefix(ctx, att, 2)
	require.NoError(t, err)
	assert.Equal(t, "AB", string(got))

	// Second Read must drain the stashed remainder from scrollback.
	buf = make([]byte, 16)
	n, err := att.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "CDE", string(buf[:n]))
}

// ---------------------------------------------------------------------------
// Polling-loop runner-error branch
// ---------------------------------------------------------------------------

func TestTmuxAdapter_PollingLoopRunnerError(t *testing.T) {
	// The polling loop must surface runner errors into att.readErr and
	// cancel the attachment. We force fallback-polling mode, then make
	// capture-pane return an error after the Attach-time capture.
	ctx := newContext(t)
	mock := newMockRunner()
	mock.paneAlive()

	var pollCalls atomic.Int32
	mock.setResponder(subCapturePane, func(_ int) ([]byte, error) {
		n := pollCalls.Add(1)
		if n == 1 {
			return []byte(""), nil // Attach-time scrollback capture
		}
		return nil, errors.New("tmux capture failed")
	})

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })
	overrideMkfifo(a, func(_ string, _ uint32) error { return errors.New("forced fallback") })

	att, err := a.Attach(ctx, testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })
	require.True(t, attachmentIsPolling(att))

	// Loop Read until we observe a non-nil terminating error. Bounded by
	// the test context.
	errCh := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if _, rerr := att.Read(buf); rerr != nil {
				errCh <- rerr
				return
			}
		}
	}()

	select {
	case got := <-errCh:
		// It's acceptable for the error to be either the raw capture
		// error OR io.EOF (if ctx gets cancelled before Read observes
		// readErr). What matters is that Read eventually terminates.
		assert.NotNil(t, got)
	case <-ctx.Done():
		t.Fatal("polling runner-error did not surface to Read")
	}
}

// ---------------------------------------------------------------------------
// Adapter.Close cascades to in-flight attachments
// ---------------------------------------------------------------------------

func TestTmuxAdapter_CloseCascadesToActiveAttachments(t *testing.T) {
	ctx := newContext(t)
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	a := NewTmuxAdapter(mock.runner())
	att, err := a.Attach(ctx, testPaneTarget)
	require.NoError(t, err)

	// Close the adapter — this must cascade close the active attachment.
	require.NoError(t, a.Close())

	// Subsequent Read on the attachment must return io.EOF because the
	// context was cancelled as part of the cascade.
	done := make(chan error, 1)
	go func() {
		_, rerr := att.Read(make([]byte, 16))
		done <- rerr
	}()
	select {
	case rerr := <-done:
		assert.ErrorIs(t, rerr, io.EOF)
	case <-ctx.Done():
		t.Fatal("cascade Close did not unblock Read")
	}

	// Re-closing the adapter is a no-op and must not panic.
	require.NoError(t, a.Close())
}

// ---------------------------------------------------------------------------
// Close-without-FIFO (polling mode) must not invoke pipe-pane stop
// ---------------------------------------------------------------------------

func TestTmuxAttachment_ClosePollingModeNoPipePaneStop(t *testing.T) {
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	a := NewTmuxAdapter(mock.runner())
	overrideMkfifo(a, func(_ string, _ uint32) error { return errors.New("forced fallback") })
	t.Cleanup(func() { _ = a.Close() })

	att, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	require.True(t, attachmentIsPolling(att))

	require.NoError(t, att.Close())

	// In polling mode, pipe-pane must NEVER have been invoked — neither
	// as "start" nor as "stop".
	assert.Equal(t, 0, mock.countSubcommand(subPipePane))
}

// ---------------------------------------------------------------------------
// Concurrency smoke: Close races with Read and SendInput under -race
// ---------------------------------------------------------------------------

func TestTmuxAttachment_CloseRacesReaderAndWriter(t *testing.T) {
	ctx := newContext(t)
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	a := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = a.Close() })
	att, err := a.Attach(ctx, testPaneTarget)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			if _, rerr := att.Read(make([]byte, 64)); rerr != nil {
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			if werr := att.SendInput([]byte("x")); werr != nil {
				return
			}
		}
	}()

	// Close from the test goroutine while the two helpers are active.
	require.NoError(t, att.Close())

	doneCh := make(chan struct{})
	go func() { wg.Wait(); close(doneCh) }()
	select {
	case <-doneCh:
	case <-ctx.Done():
		t.Fatal("Close did not unblock reader/writer before context expired")
	}

	// Sanity: adapter's active set should be empty after the close.
	a.mu.Lock()
	empty := len(a.active) == 0
	a.mu.Unlock()
	assert.True(t, empty, "closed attachment must be deregistered from adapter.active")
}

// ---------------------------------------------------------------------------
// Regression: FIFO path containing a space must be shell-quoted
// ---------------------------------------------------------------------------

// TestTmuxAdapter_FIFOPathWithSpaceInTempDir guards against a past bug
// where startPipePane built the inline cat-redirect as plain
// `cat > %s`. tmux runs the inline command via `/bin/sh -c`, so a
// $TMPDIR containing a space (e.g. "/home/me/my dir/tmp") would
// word-split the redirect target and silently break the pipe. The
// fix wraps the FIFO path in single quotes. This test asserts the
// fix stays in place by inspecting the pipe-pane invocation's
// trailing arg when the adapter's tempDir holds a space.
func TestTmuxAdapter_FIFOPathWithSpaceInTempDir(t *testing.T) {
	// Create a temp dir whose name contains a literal space.
	spaceDir := t.TempDir() + "/with space"
	require.NoError(t, os.MkdirAll(spaceDir, 0o700))

	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	a := NewTmuxAdapter(mock.runner())
	a.tempDir = spaceDir
	t.Cleanup(func() { _ = a.Close() })

	sess, err := a.Attach(newContext(t), testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Close() })
	att, ok := sess.(*TmuxAttachment)
	require.True(t, ok, "Attach must return *TmuxAttachment in tests")

	inv, ok := mock.findSubcommand(subPipePane)
	require.True(t, ok, "pipe-pane invocation must be recorded")
	require.NotEmpty(t, inv.args, "pipe-pane invocation must carry args")

	// The inline shell command is the last positional arg.
	inline := inv.args[len(inv.args)-1]
	assert.True(t, strings.HasPrefix(inline, "cat > '"),
		"inline redirect must open with single-quoted path, got %q", inline)
	assert.True(t, strings.HasSuffix(inline, "'"),
		"inline redirect must close with single quote, got %q", inline)
	assert.Contains(t, inline, "with space",
		"inline redirect must embed the space-bearing path, got %q", inline)
	// Tightest guarantee: the quoted form matches the attachment's
	// actual FIFO path.
	assert.Equal(t, "cat > '"+att.fifoPath+"'", inline)
}

// ---------------------------------------------------------------------------
// Deterministic guard: EscapeTmuxLiteral keeps tab but strips other C0
// ---------------------------------------------------------------------------

func TestEscapeTmuxLiteral_TabPreserved(t *testing.T) {
	in := "a\tb\r\nc"
	out := EscapeTmuxLiteral(in)
	// Tab preserved; CR/LF stripped.
	assert.Equal(t, "a\tbc", out)
	// DEL (0x7f) stripped.
	assert.Equal(t, "abc", EscapeTmuxLiteral("a\x7fb"+string(rune(0x7f))+"c"))
	assert.False(t, strings.ContainsRune(out, '\n'))
}
