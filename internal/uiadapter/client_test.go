package uiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withOllamaHost redirects the package-level ollamaHost var to url for the
// duration of a single test. Per §4.7.6, ollamaHost is a var (not a const)
// so tests can override it; production code never writes it.
func withOllamaHost(t *testing.T, url string) {
	t.Helper()
	old := ollamaHost
	ollamaHost = url
	t.Cleanup(func() { ollamaHost = old })
}

// newOllamaStub starts an httptest.Server running handler, rewires the
// package-global ollamaHost at it, and registers cleanup — the shared
// scaffolding every happy-path test needs.
func newOllamaStub(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	withOllamaHost(t, ts.URL)
	return ts
}

// closedSocketURL returns an http://127.0.0.1:<port> URL whose port has been
// bound and immediately released, so any connect attempt is guaranteed to
// fail with a transport error.
func closedSocketURL(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())
	return "http://" + addr
}

func TestU1_AC1_Client_Chat_RequestBodyShape(t *testing.T) {
	var gotBody map[string]any
	var gotContentType, gotMethod, gotPath string

	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotMethod = r.Method
		gotPath = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"}}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
	_, err := client.Chat(context.Background(), "gemma3:4b", "system text", "user text")
	require.NoError(t, err)

	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/api/chat", gotPath)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "json", gotBody["format"])
	assert.Equal(t, "gemma3:4b", gotBody["model"])
	assert.Equal(t, false, gotBody["stream"])

	msgs, ok := gotBody["messages"].([]any)
	require.True(t, ok, "messages must be a JSON array")
	require.Len(t, msgs, 2, "messages must contain system + user entries")

	first, _ := msgs[0].(map[string]any)
	second, _ := msgs[1].(map[string]any)
	assert.Equal(t, "system", first["role"])
	assert.Equal(t, "system text", first["content"])
	assert.Equal(t, "user", second["role"])
	assert.Equal(t, "user text", second["content"])
}

func TestU1_AC2_Client_LocalhostPinned(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("client.go")
	require.NoError(t, err, "internal/uiadapter/client.go must exist for the pin check")

	matches := regexp.MustCompile(`http://[^"]+`).FindAllString(string(data), -1)
	require.NotEmpty(t, matches, "expected at least one http:// literal in client.go (the localhost sentinel)")

	for _, m := range matches {
		assert.Equal(t, "http://localhost:11434", m,
			"§7.3: the only http:// literal in internal/uiadapter/*.go must be http://localhost:11434; found %q", m)
	}
}

func TestU1_AC3_Client_Chat_ContextDeadlineBeatsClientTimeout(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"}}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 100}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Chat(ctx, "gemma3:4b", "sys", "usr")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded),
		"context deadline must fire before http.Client.Timeout (§4.7.3); got %v", err)
	assert.NotContains(t, err.Error(), "Client.Timeout exceeded",
		"http.Client.Timeout leaked through — the 500ms buffer is missing or inverted")
}

func TestU1_AC5_Client_Chat_UnreachableSurfaceError(t *testing.T) {
	withOllamaHost(t, closedSocketURL(t))

	client := NewClient(ClientConfig{TimeoutMs: 100}, nil)
	_, err := client.Chat(context.Background(), "gemma3:4b", "sys", "usr")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrOllamaUnreachable),
		"transport failure must wrap ErrOllamaUnreachable; got %v", err)
}

func TestU1_AC6_Client_Chat_ExtractsMessageContent(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"version\":\"1\"}"}}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
	got, err := client.Chat(context.Background(), "gemma3:4b", "sys", "usr")
	require.NoError(t, err)
	assert.Equal(t, `{"version":"1"}`, got,
		"Chat must return message.content verbatim — no JSON parsing (that's U2's job)")
}

func TestU1_AC7_Client_ListModels_ParsesAndSorts(t *testing.T) {
	var gotMethod, gotPath string
	var gotContentLength int64

	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentLength = r.ContentLength
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen2.5:3b"},{"name":"gemma3:4b"},{"name":"llama3.2:3b"}]}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
	got, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"gemma3:4b", "llama3.2:3b", "qwen2.5:3b"}, got,
		"ListModels must return /api/tags names sorted lexicographically")

	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/api/tags", gotPath)
	assert.LessOrEqual(t, gotContentLength, int64(0),
		"GET /api/tags must not send a request body (ContentLength 0 or -1)")
}

func TestU1_AC8_Client_ListModels_UnreachableError(t *testing.T) {
	withOllamaHost(t, closedSocketURL(t))

	client := NewClient(ClientConfig{TimeoutMs: 100}, nil)
	got, err := client.ListModels(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrOllamaUnreachable),
		"transport failure must wrap ErrOllamaUnreachable; got %v", err)
	assert.Empty(t, got, "slice must be empty on transport failure (nil or zero-length both acceptable)")
}

func TestU1_AC9_Client_ListModels_EmptyOK(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"models":[]}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
	got, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 0, "no pulled models is not an error — returns empty slice")
}

func TestU2_ClientHTTPStatusError_ErrorString(t *testing.T) {
	t.Parallel()
	err := &HTTPStatusError{StatusCode: 500, Status: "500 Internal Server Error"}
	msg := err.Error()
	assert.Contains(t, msg, "500")
	assert.Contains(t, msg, "uiadapter")
}

// ChatDeterministic pins options.temperature=0 to eliminate Gemma sampling
// noise for the eval harness. Regular Chat must not carry that override so
// production traffic keeps Ollama's default sampling.
func TestClient_ChatDeterministic_SetsTemperature(t *testing.T) {
	t.Run("deterministic_sets_temperature_zero", func(t *testing.T) {
		var gotBody map[string]any
		newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"}}`))
		})

		client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
		_, err := client.ChatDeterministic(context.Background(), "gemma3:4b", "sys", "usr")
		require.NoError(t, err)

		opts, ok := gotBody["options"].(map[string]any)
		require.True(t, ok, "ChatDeterministic must send an options object")
		assert.Equal(t, float64(0), opts["temperature"],
			"ChatDeterministic must pin options.temperature=0 to eliminate sampling noise")
	})

	t.Run("regular_chat_omits_options_temperature", func(t *testing.T) {
		var gotBody map[string]any
		newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"}}`))
		})

		client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
		_, err := client.Chat(context.Background(), "gemma3:4b", "sys", "usr")
		require.NoError(t, err)

		if opts, has := gotBody["options"].(map[string]any); has {
			_, hasTemp := opts["temperature"]
			assert.False(t, hasTemp,
				"regular Chat must not pin temperature — eval-only path is ChatDeterministic")
		}
	})
}

func TestClient_ChatDeterministic_ReturnsContent(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"version\":\"1\"}"}}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
	got, err := client.ChatDeterministic(context.Background(), "gemma3:4b", "sys", "usr")
	require.NoError(t, err)
	assert.Equal(t, `{"version":"1"}`, got,
		"ChatDeterministic must return message.content verbatim, like Chat")
}

// -----------------------------------------------------------------------------
// Story 3 — `client.go` debug instrumentation (uiadapter-logging-3).
//
// RED-phase tests: every assertion below targets log records that do not yet
// exist in production. They will fail until the go-engineer wires the
// `client.chat.start` / `client.chat.response` / `client.chat.transport_error`
// / `client.chat.http_error` Debug emissions described in the story file.
// -----------------------------------------------------------------------------

// TestStory3_AC1_ClientChatStartAndResponse_DebugRecords — Story 3, AC-3.1.
//
// A successful Chat round-trip must emit exactly one `client.chat.start`
// record (with op/model/bytes_in) and exactly one `client.chat.response`
// record (with bytes_out and a non-negative latency_ms). Neither record
// may contain the system or user payload bytes — only lengths.
func TestStory3_AC1_ClientChatStartAndResponse_DebugRecords(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		// message.content "{}" — bytes_out must equal 2.
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{}"}}`))
	})

	logger, buf := testLogBuffer(t, slog.LevelDebug)
	client := NewClient(ClientConfig{TimeoutMs: 1000}, logger)

	const userPayload = "raw input" // len = 9 → bytes_in
	const systemPayload = "system"  // appears verbatim if leaked

	_, err := client.Chat(context.Background(), "gemma3:4b", systemPayload, userPayload)
	require.NoError(t, err)

	records := decodeRecords(t, buf)

	starts := recordsByMsg(records, "client.chat.start")
	require.Len(t, starts, 1, "expected exactly one client.chat.start record; got %d in %v", len(starts), records)
	start := starts[0]
	assert.Equal(t, "client.chat", start["op"], "client.chat.start must carry op=\"client.chat\"")
	assert.Equal(t, "gemma3:4b", start["model"], "client.chat.start must carry model")
	assert.EqualValues(t, len(userPayload), start["bytes_in"],
		"client.chat.start must carry bytes_in=len(user)=%d", len(userPayload))

	resps := recordsByMsg(records, "client.chat.response")
	require.Len(t, resps, 1, "expected exactly one client.chat.response record; got %d in %v", len(resps), records)
	resp := resps[0]
	assert.Equal(t, "client.chat", resp["op"], "client.chat.response must carry op=\"client.chat\"")
	assert.Equal(t, "gemma3:4b", resp["model"], "client.chat.response must carry model")
	assert.EqualValues(t, 2, resp["bytes_out"],
		"client.chat.response must carry bytes_out=len(message.content)=2 for body \"{}\"")
	latency, ok := resp["latency_ms"].(float64)
	require.True(t, ok, "client.chat.response must carry numeric latency_ms")
	assert.GreaterOrEqual(t, latency, 0.0, "latency_ms must be non-negative")

	// Sanitize discipline (AC-3.7 precondition for client.go): no record may
	// carry the raw system or user payload bytes.
	for i, rec := range records {
		raw, _ := json.Marshal(rec)
		assert.NotContains(t, string(raw), userPayload,
			"record %d (%v) must not contain the raw user payload %q", i, rec["msg"], userPayload)
		// The literal "system" can appear as a key name in some tracing
		// libraries; here we assert no top-level attr has the string
		// value "system" — i.e., the system prompt content was logged.
		for k, v := range rec {
			if s, ok := v.(string); ok && s == systemPayload {
				t.Fatalf("record %d attr %q leaked system prompt content %q", i, k, s)
			}
		}
	}
}

// TestStory3_AC2_ClientTransportError — Story 3, AC-3.2 (transport branch).
//
// A Chat call against a closed listener must emit `client.chat.transport_error`
// with a `reason` enum (one of the classifyChatErr enum strings) and must
// never carry an `error` attribute holding the raw err.Error() text.
func TestStory3_AC2_ClientTransportError(t *testing.T) {
	withOllamaHost(t, closedSocketURL(t))

	logger, buf := testLogBuffer(t, slog.LevelDebug)
	client := NewClient(ClientConfig{TimeoutMs: 200}, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	_, err := client.Chat(ctx, "gemma3:4b", "sys", "usr")
	require.Error(t, err, "closed-listener Chat must return an error")

	records := decodeRecords(t, buf)
	rejs := recordsByMsg(records, "client.chat.transport_error")
	require.Len(t, rejs, 1,
		"expected exactly one client.chat.transport_error record; got %d in %v", len(rejs), records)
	rec := rejs[0]
	assert.Equal(t, "client.chat", rec["op"], "transport_error must carry op=\"client.chat\"")
	assert.Equal(t, "gemma3:4b", rec["model"], "transport_error must carry model")

	reason, ok := rec["reason"].(string)
	require.True(t, ok, "transport_error must carry a string reason attr; got %v", rec["reason"])
	assert.NotEmpty(t, reason, "reason must be a non-empty enum string (classifyChatErr)")
	// classifyChatErr enum domain: "canceled" | "timeout" | "unreachable" | "server:NNN"
	allowed := map[string]struct{}{
		"canceled": {}, "timeout": {}, "unreachable": {}, "transport": {}, "saturated": {},
	}
	if _, isServer := allowed[reason]; !isServer && !strings.HasPrefix(reason, "server:") {
		t.Fatalf("reason %q is not a recognised classifyChatErr enum string", reason)
	}

	// Sanitize: no `error` attr carrying the raw err.Error() may be present.
	if errAttr, has := rec["error"]; has {
		t.Fatalf("transport_error must not carry an `error` attr; got %v", errAttr)
	}
	rawErr := err.Error()
	if rawErr != "" {
		raw, _ := json.Marshal(rec)
		assert.NotContains(t, string(raw), rawErr,
			"transport_error record must not contain the raw err.Error() text %q", rawErr)
	}
}

// TestStory3_AC2_ClientHttpError — Story 3, AC-3.2 (HTTP branch).
//
// A Chat call hitting a 500 must emit `client.chat.http_error` with the
// status_code attr; no payload bytes leak.
func TestStory3_AC2_ClientHttpError(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	})

	logger, buf := testLogBuffer(t, slog.LevelDebug)
	client := NewClient(ClientConfig{TimeoutMs: 1000}, logger)

	_, err := client.Chat(context.Background(), "gemma3:4b", "sys", "usr")
	require.Error(t, err, "500 response must surface as an error")

	records := decodeRecords(t, buf)
	httpErrs := recordsByMsg(records, "client.chat.http_error")
	require.Len(t, httpErrs, 1,
		"expected exactly one client.chat.http_error record; got %d in %v", len(httpErrs), records)
	rec := httpErrs[0]
	assert.Equal(t, "client.chat", rec["op"], "http_error must carry op=\"client.chat\"")
	assert.Equal(t, "gemma3:4b", rec["model"], "http_error must carry model")
	assert.EqualValues(t, 500, rec["status_code"], "http_error must carry status_code=500")

	latency, ok := rec["latency_ms"].(float64)
	require.True(t, ok, "http_error must carry numeric latency_ms")
	assert.GreaterOrEqual(t, latency, 0.0, "latency_ms must be non-negative")
}
