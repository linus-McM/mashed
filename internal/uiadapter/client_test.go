package uiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
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

	client := NewClient(ClientConfig{TimeoutMs: 1000})
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

	client := NewClient(ClientConfig{TimeoutMs: 100})
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

	client := NewClient(ClientConfig{TimeoutMs: 100})
	_, err := client.Chat(context.Background(), "gemma3:4b", "sys", "usr")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrOllamaUnreachable),
		"transport failure must wrap ErrOllamaUnreachable; got %v", err)
}

func TestU1_AC6_Client_Chat_ExtractsMessageContent(t *testing.T) {
	newOllamaStub(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"version\":\"1\"}"}}`))
	})

	client := NewClient(ClientConfig{TimeoutMs: 1000})
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

	client := NewClient(ClientConfig{TimeoutMs: 1000})
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

	client := NewClient(ClientConfig{TimeoutMs: 100})
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

	client := NewClient(ClientConfig{TimeoutMs: 1000})
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
