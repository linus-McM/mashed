// NOTE: §11 Q6 resolved 2026-04-21 — adapter enabled by default; unreachable Ollama degrades to fallback:unreachable, not fallback:disabled

package uiadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// ollamaChatPayload wraps an AST-JSON blob inside an Ollama chat response —
// message.content is the string returned verbatim by Client.Chat.
func ollamaChatPayload(astJSON string) string {
	escaped, _ := json.Marshal(astJSON)
	return `{"message":{"role":"assistant","content":` + string(escaped) + `}}`
}

const validASTJSON = `{"version":"1","generated_by":"ollama:test","nodes":[{"type":"markdown","content":"ok"}],"fallback_answer_shape":"free"}`

// enabledConfig is the per-test baseline: all failure-path tests tweak only
// TimeoutMs, so centralising the rest avoids Config literal sprawl.
func enabledConfig(timeoutMs int) Config {
	return Config{Enabled: true, Model: "g", TimeoutMs: timeoutMs, MaxInflight: 1}
}

// newSaturatedAdapter returns an adapter whose single-slot semaphore is held
// by a background Translate call blocked inside the Ollama stub handler. The
// caller supplies a short ctx to observe the fallback:saturated path.
//
// Cleanup ordering matters: t.Cleanup runs registrations in LIFO order, so
// close(release) must be registered AFTER newOllamaStub (newOllamaStub
// registers ts.Close internally). Otherwise httptest.Server.Close would run
// first and block forever on the handler stuck at <-release.
func newSaturatedAdapter(t *testing.T) Adapter {
	t.Helper()
	started := make(chan struct{})
	release := make(chan struct{})
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
	})
	t.Cleanup(func() { close(release) })
	a := NewDefault(enabledConfig(2000), discardLogger)
	go func() { _ = a.Translate(context.Background(), "raw", "proc") }()
	<-started
	return a
}

// TestU2_AC1_Adapter_Translate_NeverNil — every failure path returns non-nil
// with Version "1". Table covers §4.8 modes end to end.
func TestU2_AC1_Adapter_Translate_NeverNil(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T) (Adapter, context.Context)
	}{
		{
			name: "dead socket",
			setup: func(t *testing.T) (Adapter, context.Context) {
				withOllamaHost(t, closedSocketURL(t))
				return NewDefault(enabledConfig(200), discardLogger), context.Background()
			},
		},
		{
			name: "timeout (server slower than TimeoutMs)",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(400 * time.Millisecond)
					_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
				})
				return NewDefault(enabledConfig(100), discardLogger), context.Background()
			},
		},
		{
			name: "http 500",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "nope", http.StatusInternalServerError)
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
		},
		{
			name: "malformed JSON",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload("not json")))
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
		},
		{
			name: "ctx canceled",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
				})
				a := NewDefault(enabledConfig(500), discardLogger)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return a, ctx
			},
		},
		{
			name: "semaphore saturated",
			setup: func(t *testing.T) (Adapter, context.Context) {
				a := newSaturatedAdapter(t)
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				t.Cleanup(cancel)
				return a, ctx
			},
		},
		{
			name: "disabled adapter",
			setup: func(t *testing.T) (Adapter, context.Context) {
				return NewDefault(Config{Enabled: false}, discardLogger), context.Background()
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, ctx := tc.setup(t)
			ast := a.Translate(ctx, "raw text", "proc-id")
			require.NotNil(t, ast, "%s: Translate must never return nil (§4.3)", tc.name)
			assert.Equal(t, "1", ast.Version, "%s: Version must always be \"1\"", tc.name)
		})
	}
}

// TestU2_AC2_Adapter_FallbackReasonTable — every §4.8 row maps to the exact
// GeneratedBy sentinel.
func TestU2_AC2_Adapter_FallbackReasonTable(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T) (Adapter, context.Context)
		want  string
	}{
		{
			name: "unreachable",
			setup: func(t *testing.T) (Adapter, context.Context) {
				withOllamaHost(t, closedSocketURL(t))
				return NewDefault(enabledConfig(100), discardLogger), context.Background()
			},
			want: "fallback:unreachable",
		},
		{
			name: "timeout",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(300 * time.Millisecond)
					_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
				})
				return NewDefault(enabledConfig(50), discardLogger), context.Background()
			},
			want: "fallback:timeout",
		},
		{
			name: "server:500",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "boom", http.StatusInternalServerError)
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
			want: "fallback:server:500",
		},
		{
			name: "validation:malformed",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload("not json")))
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
			want: "fallback:validation:malformed",
		},
		{
			name: "validation:required_dropped",
			setup: func(t *testing.T) (Adapter, context.Context) {
				var nodes []UINode
				for i := 0; i < 9; i++ {
					nodes = append(nodes, UINode{
						Type:        "decision_group",
						ResponseKey: fmt.Sprintf("k%d", i),
						Required:    i == 8,
						Widget:      &WidgetNode{Type: "free"},
					})
				}
				astBlob, _ := json.Marshal(&UIAST{
					Version:             "1",
					Nodes:               nodes,
					FallbackAnswerShape: "free",
				})
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload(string(astBlob))))
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
			want: "fallback:validation:required_dropped",
		},
		{
			name: "validation:oversize",
			setup: func(t *testing.T) (Adapter, context.Context) {
				huge := strings.Repeat("A", 7*1024)
				astBlob, _ := json.Marshal(&UIAST{
					Version:             "1",
					Nodes:               []UINode{{Type: "markdown", Content: huge}},
					FallbackAnswerShape: "free",
				})
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload(string(astBlob))))
				})
				return NewDefault(enabledConfig(500), discardLogger), context.Background()
			},
			want: "fallback:validation:oversize",
		},
		{
			name: "disabled",
			setup: func(t *testing.T) (Adapter, context.Context) {
				return NewDefault(Config{Enabled: false}, discardLogger), context.Background()
			},
			want: "fallback:disabled",
		},
		{
			name: "saturated",
			setup: func(t *testing.T) (Adapter, context.Context) {
				a := newSaturatedAdapter(t)
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				t.Cleanup(cancel)
				return a, ctx
			},
			want: "fallback:saturated",
		},
		{
			name: "canceled",
			setup: func(t *testing.T) (Adapter, context.Context) {
				newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
				})
				a := NewDefault(enabledConfig(500), discardLogger)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return a, ctx
			},
			want: "fallback:canceled",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, ctx := tc.setup(t)
			ast := a.Translate(ctx, "raw", "proc")
			require.NotNil(t, ast)
			assert.Equal(t, "1", ast.Version)
			assert.Equal(t, tc.want, ast.GeneratedBy,
				"row %q: §4.8 failure must map to the exact sentinel", tc.name)
		})
	}
}

// TestU2_AC5_Adapter_Semaphore_BlockedCallSaturates — §4.7.5: MaxInflight 1,
// slow server, second caller with short ctx → fallback:saturated via the
// ctx.Done() branch of the sem select.
func TestU2_AC5_Adapter_Semaphore_BlockedCallSaturates(t *testing.T) {
	a := newSaturatedAdapter(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	ast := a.Translate(ctx, "raw", "proc")
	require.NotNil(t, ast)
	assert.Equal(t, "fallback:saturated", ast.GeneratedBy,
		"second call contending for MaxInflight=1 with a short ctx must return fallback:saturated")
}

// TestU2_AC7_Adapter_NetworkPinned_NoNonLocalhostReach — §7.3 extended to
// the whole uiadapter package (U1 AC-2 only covered client.go).
func TestU2_AC7_Adapter_NetworkPinned_NoNonLocalhostReach(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	require.NotEmpty(t, files, "expected at least one .go file in internal/uiadapter")

	re := regexp.MustCompile(`http://[^"]+`)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		require.NoError(t, err)
		for _, m := range re.FindAllString(string(data), -1) {
			assert.Equal(t, "http://localhost:11434", m,
				"§7.3: every http:// literal across internal/uiadapter/*.go must be http://localhost:11434; found %q in %s",
				m, f)
		}
	}
}

// TestU2_AC9_Adapter_DisabledProducesFallback — §11 Q6 resolution: explicit
// opt-out produces a deterministic fallback:disabled AST with the raw as
// the single markdown node. Config.Enabled is the combined UIAdapterEnabled
// && OllamaEnabled gate — app.go AND-composes before passing, so either
// user-facing toggle OFF lands here.
func TestU2_AC9_Adapter_DisabledProducesFallback(t *testing.T) {
	t.Parallel()
	a := NewDefault(Config{Enabled: false, Model: "g", TimeoutMs: 500}, discardLogger)
	raw := "hello raw text"
	ast := a.Translate(context.Background(), raw, "proc")

	require.NotNil(t, ast)
	assert.Equal(t, "1", ast.Version)
	assert.Equal(t, "fallback:disabled", ast.GeneratedBy)
	require.Len(t, ast.Nodes, 1)
	assert.Equal(t, "markdown", ast.Nodes[0].Type)
	assert.Equal(t, raw, ast.Nodes[0].Content)
}

// TestU2_AC9_Adapter_DisabledNoHTTPCall — disabled adapter must never hit
// the HTTP client. Proof: a panic-on-call stub server is never touched.
func TestU2_AC9_Adapter_DisabledNoHTTPCall(t *testing.T) {
	var hits int32
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		panic("HTTP called despite adapter disabled")
	})

	a := NewDefault(Config{Enabled: false}, discardLogger)
	ast := a.Translate(context.Background(), "hi", "proc")

	require.NotNil(t, ast)
	assert.Equal(t, "fallback:disabled", ast.GeneratedBy)
	assert.Equal(t, int32(0), atomic.LoadInt32(&hits),
		"disabled adapter must NOT reach the HTTP client even once")
}

// TestU2_AC9_Adapter_UnreachableNotAutoDisabled — §11 Q6: enabled flags TRUE
// with a closed socket yields fallback:unreachable, NOT fallback:disabled.
// Unreachable Ollama never silently auto-flips the config.
func TestU2_AC9_Adapter_UnreachableNotAutoDisabled(t *testing.T) {
	t.Parallel()
	withOllamaHost(t, closedSocketURL(t))
	a := NewDefault(
		Config{Enabled: true, Model: "gemma3:4b", TimeoutMs: 100, MaxInflight: 1},
		discardLogger,
	)
	ast := a.Translate(context.Background(), "raw", "proc")

	require.NotNil(t, ast)
	assert.Equal(t, "fallback:unreachable", ast.GeneratedBy,
		"Q6 (2026-04-21): enabled + socket-closed must surface unreachable, not disabled")
	assert.NotEqual(t, "fallback:disabled", ast.GeneratedBy,
		"unreachable must never degrade into the disabled sentinel")
}
