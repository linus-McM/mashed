// Package terminal tests for TmuxAdapter (Story bridge-02).
//
// RED Phase: These tests define expected behavior for ACs 1-8 and BDD
// scenarios from docs/stories/bridge-02-tmux-adapter.md. They must FAIL
// to compile until internal/terminal/tmux_adapter.go and
// internal/terminal/tmux_escape.go are implemented by the go-engineer.
//
// Concurrency rules:
//   - No time.Sleep — bounded context + channels + sync primitives only.
//   - Every test body is wrapped in a 2-second context timeout.
//   - Mock runner is goroutine-safe (sync.Mutex around invocation log and
//     state-machine counters) because tests run under -race.

package terminal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

const (
	testPaneTarget     = "bmad-xx:0.0"
	testDeadPaneTarget = "bmad-dead:0.0"
	testTimeout        = 2 * time.Second

	// tmux subcommands referenced by tests and mock responders.
	subListPanes   = "list-panes"
	subCapturePane = "capture-pane"
	subSendKeys    = "send-keys"
	subPipePane    = "pipe-pane"
	subResizeWin   = "resize-window"

	// canonical responses for the `#{pane_dead}` format.
	paneAliveResp = "0\n"
	paneDeadResp  = "1\n"
)

// invocation records one CommandRunner call for later assertion.
type invocation struct {
	name string
	args []string
}

func (i invocation) joined() string {
	return i.name + " " + strings.Join(i.args, " ")
}

// responderFunc returns (stdout, err) for one invocation. callIndex is
// the zero-based index within that subcommand, enabling state-machine
// responders without external counters.
//
// Responders must NOT re-enter mockRunner methods — they may be invoked
// concurrently from goroutines the adapter spawns.
type responderFunc func(callIndex int) ([]byte, error)

// mockRunner is a thread-safe stand-in for the tmux CommandRunner used by
// TmuxAdapter. It logs every invocation and dispatches to per-subcommand
// responders (keyed by the first tmux arg such as "list-panes",
// "capture-pane", "pipe-pane", "send-keys", "resize-window"). Unknown
// subcommands return ("ok", nil) so tests can focus on what matters.
type mockRunner struct {
	mu         sync.Mutex
	calls      []invocation
	responders map[string]responderFunc
	subCounts  map[string]int
	defaultFn  responderFunc
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		responders: map[string]responderFunc{},
		subCounts:  map[string]int{},
		defaultFn: func(_ int) ([]byte, error) {
			return []byte("ok"), nil
		},
	}
}

func (m *mockRunner) setResponder(subcommand string, fn responderFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responders[subcommand] = fn
}

// runner returns the CommandRunner closure to hand to NewTmuxAdapter.
func (m *mockRunner) runner() CommandRunner {
	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		m.mu.Lock()
		m.calls = append(m.calls, invocation{name: name, args: append([]string(nil), args...)})

		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}
		m.subCounts[sub]++
		subIdx := m.subCounts[sub] - 1
		fn, ok := m.responders[sub]
		m.mu.Unlock()

		if ok {
			return fn(subIdx)
		}
		return m.defaultFn(subIdx)
	}
}

// invocations returns a snapshot copy of the call log.
func (m *mockRunner) invocations() []invocation {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]invocation, len(m.calls))
	copy(out, m.calls)
	return out
}

// countSubcommand returns how many times a tmux subcommand was invoked.
func (m *mockRunner) countSubcommand(sub string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subCounts[sub]
}

// findSubcommand returns the first invocation whose first arg matches sub.
func (m *mockRunner) findSubcommand(sub string) (invocation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, inv := range m.calls {
		if len(inv.args) > 0 && inv.args[0] == sub {
			return inv, true
		}
	}
	return invocation{}, false
}

// paneAlive installs a list-panes responder that always reports alive.
func (m *mockRunner) paneAlive() {
	m.setResponder(subListPanes, func(_ int) ([]byte, error) {
		return []byte(paneAliveResp), nil
	})
}

// paneDead installs a list-panes responder that always reports dead.
func (m *mockRunner) paneDead() {
	m.setResponder(subListPanes, func(_ int) ([]byte, error) {
		return []byte(paneDeadResp), nil
	})
}

// paneDiesAfter flips list-panes output from alive to dead once flip()
// is invoked. Uses atomic.Int32 so the watcher goroutine and the test
// can observe the flip without races.
func (m *mockRunner) paneDiesAfter() (flip func()) {
	var dead atomic.Int32
	m.setResponder(subListPanes, func(_ int) ([]byte, error) {
		if dead.Load() == 1 {
			return []byte(paneDeadResp), nil
		}
		return []byte(paneAliveResp), nil
	})
	return func() { dead.Store(1) }
}

// captureReturns installs a capture-pane responder that yields the given
// body on every call.
func (m *mockRunner) captureReturns(body string) {
	m.setResponder(subCapturePane, func(_ int) ([]byte, error) {
		return []byte(body), nil
	})
}

// newContext builds a bounded test context with automatic cancel cleanup.
func newContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)
	return ctx
}

// newLiveAttachment wires a mockRunner that reports the pane alive with
// an empty scrollback, attaches to testPaneTarget, and registers Close
// cleanup. Returns the mock so tests can assert on captured invocations.
func newLiveAttachment(t *testing.T, ctx context.Context) (*mockRunner, *TmuxAttachment) {
	t.Helper()
	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	sess, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err, "newLiveAttachment: Attach should succeed")
	t.Cleanup(func() { _ = sess.Close() })
	att, ok := sess.(*TmuxAttachment)
	require.True(t, ok, "newLiveAttachment: Attach must return *TmuxAttachment in tests")
	return mock, att
}

// readPrefix reads up to n bytes from r under a small deadline and
// returns whatever arrived. Implemented without time.Sleep: a helper
// goroutine pushes results to a channel so the main test can select
// against the caller context.
func readPrefix(ctx context.Context, r io.Reader, n int) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		buf := make([]byte, n)
		got, err := io.ReadFull(r, buf)
		ch <- result{data: buf[:got], err: err}
	}()
	select {
	case res := <-ch:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ---------------------------------------------------------------------------
// AC-1: Attach validates the pane and replays scrollback
// ---------------------------------------------------------------------------

func TestTmuxAdapter_AC1_AttachReplaysScrollback(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("hello history\n")

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	att, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err, "Attach should succeed on a live pane")
	require.NotNil(t, att, "Attach must return a non-nil *TmuxAttachment")
	t.Cleanup(func() { _ = att.Close() })

	got, err := readPrefix(ctx, att, len("hello history\n"))
	require.NoError(t, err, "Read should deliver scrollback bytes without error")
	assert.Equal(t, "hello history\n", string(got),
		"Read must yield the full scrollback body before any live bytes")

	// AC-1 also requires the pane-liveness check to run before capture.
	inv, ok := mock.findSubcommand(subListPanes)
	require.True(t, ok, "Attach must invoke tmux list-panes to check pane_dead")
	assert.Contains(t, strings.Join(inv.args, " "), "#{pane_dead}",
		"list-panes must use -F #{pane_dead} format")

	capInv, ok := mock.findSubcommand(subCapturePane)
	require.True(t, ok, "Attach must invoke tmux capture-pane for scrollback replay")
	joined := strings.Join(capInv.args, " ")
	assert.Contains(t, joined, "-S -5000", "capture-pane must replay up to 5000 lines")
	assert.Contains(t, joined, "-p", "capture-pane must use -p (print to stdout)")
	assert.Contains(t, joined, "-J", "capture-pane must use -J (join wrapped lines)")
	assert.Contains(t, joined, testPaneTarget, "capture-pane must target the requested pane")
}

// ---------------------------------------------------------------------------
// AC-2: Attach rejects a dead pane
// ---------------------------------------------------------------------------

func TestTmuxAdapter_AC2_AttachRejectsDeadPane(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	mock.paneDead()

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	att, err := adapter.Attach(ctx, testDeadPaneTarget)
	require.Error(t, err, "Attach on a dead pane must return an error")
	assert.Nil(t, att, "Attach must return nil attachment on dead pane")
	assert.True(t, errors.Is(err, ErrPaneDead),
		"returned error must wrap ErrPaneDead (errors.Is), got: %v", err)

	assert.Equal(t, 0, mock.countSubcommand(subPipePane),
		"no pipe-pane invocation should happen when the pane is dead")
	assert.Equal(t, 0, mock.countSubcommand(subCapturePane),
		"no capture-pane invocation should happen when the pane is dead")
}

// ---------------------------------------------------------------------------
// AC-3: SendInput uses send-keys -l with escaped literal
// ---------------------------------------------------------------------------

func TestTmuxAttachment_AC3_SendInputEscapesAndUsesSendKeysLiteral(t *testing.T) {
	ctx := newContext(t)

	type caseDef struct {
		name     string
		input    string
		expected string // expected literal arg (post-escape)
	}

	cases := []caseDef{
		{
			name:     "plain apostrophe preserved",
			input:    "it's a test",
			expected: "it's a test",
		},
		{
			name:     "shell metachars preserved (tmux -l is literal)",
			input:    "echo $HOME `pwd` | tee",
			expected: "echo $HOME `pwd` | tee",
		},
		{
			name:     "control bytes stripped (ESC removed)",
			input:    "hello\x1bworld",
			expected: "helloworld",
		},
		{
			name:     "newline stripped (submit guard)",
			input:    "foo\nbar",
			expected: "foobar",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock, att := newLiveAttachment(t, ctx)

			require.NoError(t, att.SendInput([]byte(tc.input)),
				"SendInput must succeed on live attachment")

			inv, ok := mock.findSubcommand(subSendKeys)
			require.True(t, ok, "send-keys invocation must be recorded")
			assert.Contains(t, inv.args, "-l",
				"SendInput must pass -l (literal) to send-keys")
			assert.Contains(t, inv.args, "-t",
				"send-keys must include -t target flag")
			assert.Contains(t, inv.args, testPaneTarget,
				"send-keys target must be the attached pane")

			// The last positional arg is the literal payload.
			literal := inv.args[len(inv.args)-1]
			assert.Equal(t, tc.expected, literal,
				"literal arg must equal EscapeTmuxLiteral(input)")

			// Cross-check: the escape helper itself must match.
			assert.Equal(t, tc.expected, EscapeTmuxLiteral(tc.input),
				"EscapeTmuxLiteral must produce the same output the adapter sent")
		})
	}
}

// ---------------------------------------------------------------------------
// AC-4: SendKey forwards special keys without -l
// ---------------------------------------------------------------------------

func TestTmuxAttachment_AC4_SendKeyUsesSendKeysNoLiteral(t *testing.T) {
	ctx := newContext(t)

	cases := []string{"Enter", "C-c", "Tab", "Up", "BSpace"}

	for _, key := range cases {
		key := key
		t.Run(key, func(t *testing.T) {
			mock, att := newLiveAttachment(t, ctx)

			require.NoError(t, att.SendKey(key))

			inv, ok := mock.findSubcommand(subSendKeys)
			require.True(t, ok, "send-keys invocation must be recorded")

			// Must NOT include the -l literal flag.
			for _, a := range inv.args {
				assert.NotEqual(t, "-l", a,
					"SendKey must not pass -l to send-keys (got %v)", inv.args)
			}
			assert.Contains(t, inv.args, "-t",
				"send-keys must include -t target flag")
			assert.Contains(t, inv.args, testPaneTarget,
				"send-keys target must be the attached pane")
			assert.Equal(t, key, inv.args[len(inv.args)-1],
				"last arg must be the key token verbatim")
		})
	}
}

// ---------------------------------------------------------------------------
// AC-5: Resize invokes tmux resize-window
// ---------------------------------------------------------------------------

func TestTmuxAttachment_AC5_ResizeInvokesResizeWindow(t *testing.T) {
	ctx := newContext(t)

	type caseDef struct {
		name       string
		cols, rows uint16
		wantX      string
		wantY      string
	}

	cases := []caseDef{
		{"standard", 120, 40, "120", "40"},
		{"small", 80, 24, "80", "24"},
		{"large", 500, 200, "500", "200"},
		{"single digit", 1, 2, "1", "2"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock, att := newLiveAttachment(t, ctx)

			require.NoError(t, att.Resize(tc.cols, tc.rows))

			inv, ok := mock.findSubcommand(subResizeWin)
			require.True(t, ok, "Resize must invoke tmux resize-window")
			joined := strings.Join(inv.args, " ")

			assert.Contains(t, joined, "-t "+testPaneTarget,
				"resize-window must target attached pane")
			assert.Contains(t, joined, "-x "+tc.wantX,
				"resize-window -x must use plain decimal cols")
			assert.Contains(t, joined, "-y "+tc.wantY,
				"resize-window -y must use plain decimal rows")

			// Guard against leading zeros / scientific notation.
			assert.NotContains(t, joined, "-x 0"+tc.wantX,
				"resize-window must not emit leading-zero widths")
			assert.NotContains(t, joined, "e+", "resize-window must not use scientific notation")
		})
	}
}

// ---------------------------------------------------------------------------
// AC-6: Read returns io.EOF when the pane dies
// ---------------------------------------------------------------------------

func TestTmuxAttachment_AC6_ReadReturnsEOFWhenPaneDies(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	flip := mock.paneDiesAfter()
	mock.captureReturns("initial\n")

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	att, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err, "Attach succeeds while pane is alive")
	t.Cleanup(func() { _ = att.Close() })

	// Drain the scrollback that was queued at attach time.
	drain, err := readPrefix(ctx, att, len("initial\n"))
	require.NoError(t, err)
	assert.Equal(t, "initial\n", string(drain))

	// Flip the pane to dead and then keep reading until we observe EOF.
	// We must not rely on wall-clock timing — issue reads in a loop,
	// bounded by the test context, until EOF bubbles up.
	flip()

	eofCh := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := att.Read(buf)
			if err != nil {
				eofCh <- err
				return
			}
		}
	}()

	select {
	case err := <-eofCh:
		assert.ErrorIs(t, err, io.EOF,
			"Read must return io.EOF after pane death (got %v)", err)
	case <-ctx.Done():
		t.Fatalf("timed out waiting for Read to surface io.EOF after pane death")
	}

	// A follow-up Read must also return EOF — the context is cancelled.
	_, err2 := att.Read(make([]byte, 16))
	assert.ErrorIs(t, err2, io.EOF,
		"subsequent Read after death must also return io.EOF")
}

// ---------------------------------------------------------------------------
// AC-7: Close is idempotent and cleans up
// ---------------------------------------------------------------------------

func TestTmuxAttachment_AC7_CloseIsIdempotent(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("")

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	att, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err)

	// First Close must succeed.
	require.NoError(t, att.Close(), "first Close must return nil")
	// Second Close must also succeed (no panic, no error).
	require.NoError(t, att.Close(), "second Close must be idempotent")
	// Third Close for good measure.
	require.NoError(t, att.Close(), "third Close must still be idempotent")

	// pipe-pane "stop" (no inline shell command) must never fire twice,
	// even across repeated Close calls. Zero is acceptable — the polling
	// fallback path does not use pipe-pane at all.
	stopCalls := 0
	for _, inv := range mock.invocations() {
		if len(inv.args) < 1 || inv.args[0] != subPipePane {
			continue
		}
		hasInline := false
		for _, a := range inv.args {
			if strings.Contains(a, "cat >") {
				hasInline = true
				break
			}
		}
		if !hasInline {
			stopCalls++
		}
	}
	assert.LessOrEqual(t, stopCalls, 1,
		"pipe-pane stop must be invoked at most once across repeated Close calls")
}

// ---------------------------------------------------------------------------
// AC-8: Mkfifo failure falls back to polling
// ---------------------------------------------------------------------------

func TestTmuxAdapter_AC8_MkfifoFailureFallsBackToPolling(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	mock.paneAlive()

	// capture-pane must return a body on every poll so the fallback has
	// data to deliver. The callIndex arg from responderFunc gives each
	// invocation a unique sequence number without external state.
	mock.setResponder(subCapturePane, func(callIndex int) ([]byte, error) {
		return []byte(fmt.Sprintf("poll-%d\n", callIndex+1)), nil
	})

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	// Inject an mkfifo that always errors — force the polling fallback.
	mkfifoErr := errors.New("mkfifo: test-injected failure")
	overrideMkfifo(adapter, func(_ string, _ uint32) error { return mkfifoErr })

	att, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err, "Attach must succeed even when mkfifo fails")
	require.NotNil(t, att, "attachment must be returned in polling mode")
	t.Cleanup(func() { _ = att.Close() })

	assert.True(t, attachmentIsPolling(att),
		"attachment must expose fallbackPolling == true when mkfifo failed")

	// Read must yield bytes sourced from capture-pane polling.
	got, err := readPrefix(ctx, att, len("poll-1\n"))
	require.NoError(t, err,
		"Read must deliver polled bytes (got err=%v, bytes=%q)", err, string(got))
	assert.True(t, bytes.HasPrefix(got, []byte("poll-")),
		"polled output must originate from capture-pane responder (got %q)", string(got))
}

// ---------------------------------------------------------------------------
// Concurrency: SendInput + Read under -race
// ---------------------------------------------------------------------------

// TestTmuxAttachment_ConcurrentReadSendInput exercises parallel SendInput
// and Read on a single attachment to shake out data races under -race.
// The test does not assert any particular byte contents — its job is to
// run until completion without triggering the race detector.
func TestTmuxAttachment_ConcurrentReadSendInput(t *testing.T) {
	ctx := newContext(t)

	mock := newMockRunner()
	mock.paneAlive()
	mock.captureReturns("seed\n")

	adapter := NewTmuxAdapter(mock.runner())
	t.Cleanup(func() { _ = adapter.Close() })

	att, err := adapter.Attach(ctx, testPaneTarget)
	require.NoError(t, err)
	t.Cleanup(func() { _ = att.Close() })

	const writers = 8
	const writesEach = 25

	var writerWG sync.WaitGroup
	var readerWG sync.WaitGroup
	stop := make(chan struct{})

	// Reader goroutine: drain bytes until `stop` closes.
	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		buf := make([]byte, 256)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := att.Read(buf); err != nil {
				return
			}
		}
	}()

	// Writer goroutines: fire SendInput concurrently.
	for i := 0; i < writers; i++ {
		writerWG.Add(1)
		go func(id int) {
			defer writerWG.Done()
			for j := 0; j < writesEach; j++ {
				if err := att.SendInput([]byte(fmt.Sprintf("w%d-%d ", id, j))); err != nil {
					return
				}
			}
		}(i)
	}

	// Wait for writers, or abort on context deadline.
	writersDone := make(chan struct{})
	go func() {
		writerWG.Wait()
		close(writersDone)
	}()
	select {
	case <-writersDone:
	case <-ctx.Done():
		t.Fatal("concurrent writers did not finish within context deadline")
	}

	// Signal the reader to stop and wait for it.
	close(stop)
	readerDone := make(chan struct{})
	go func() {
		readerWG.Wait()
		close(readerDone)
	}()
	select {
	case <-readerDone:
	case <-ctx.Done():
		t.Fatal("reader did not exit within context deadline")
	}
}
