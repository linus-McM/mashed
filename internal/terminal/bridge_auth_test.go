package terminal

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// R1-R4: the terminal WebSocket requires the per-launch token (sent as the
// `mashed.auth.<token>` subprotocol), an allowlisted Origin and a loopback
// Host before any session lookup.

// authHeader returns the headers a legitimate webview sends.
func authHeader(b *Bridge) http.Header {
	return http.Header{
		"Origin":                 {"wails://wails"},
		"Sec-WebSocket-Protocol": {SubprotocolV1 + ", " + authSubprotocolPrefix + b.Token()},
	}
}

// getWS issues a plain HTTP GET to /ws/<name> with the given headers and
// Host, returning the status code.
func getWS(t *testing.T, ctx context.Context, b *Bridge, name, host string, h http.Header) int {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%d/ws/%s", b.GetTerminalPort(), name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	for k, v := range h {
		req.Header[k] = v
	}
	if host != "" {
		req.Host = host
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	return resp.StatusCode
}

func authBridge(t *testing.T) (context.Context, *Bridge, *mockTmuxAttacher) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	adapter := newMockTmuxAttacher(newMockTmuxSession())
	_, b := startBridgeWithMockAdapter(t, ctx, adapter)
	return ctx, b, adapter
}

func TestBridge_RejectsMissingToken(t *testing.T) {
	ctx, b, adapter := authBridge(t)
	h := http.Header{"Origin": {"wails://wails"}}
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "bmad-x", "", h))
	assert.Equal(t, int32(0), adapter.attachCount.Load(), "no attach without token")
}

func TestBridge_RejectsWrongToken(t *testing.T) {
	ctx, b, adapter := authBridge(t)
	h := http.Header{
		"Origin":                 {"wails://wails"},
		"Sec-WebSocket-Protocol": {SubprotocolV1 + ", " + authSubprotocolPrefix + strings.Repeat("0", 64)},
	}
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "bmad-x", "", h))
	assert.Equal(t, int32(0), adapter.attachCount.Load())
}

func TestBridge_UnknownSessionWithoutToken403(t *testing.T) {
	ctx, b, _ := authBridge(t)
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "ghost", "", nil),
		"auth runs before session lookup, so unknown names do not leak 404")
	assert.Equal(t, http.StatusNotFound, getWS(t, ctx, b, "ghost", "", authHeader(b)),
		"with valid auth an unknown session is still 404")
}

func TestBridge_RejectsForeignOrigin(t *testing.T) {
	ctx, b, _ := authBridge(t)
	h := authHeader(b)
	h.Set("Origin", "https://evil.example")
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "bmad-x", "", h))
}

func TestBridge_RejectsEmptyOrigin(t *testing.T) {
	ctx, b, _ := authBridge(t)
	h := authHeader(b)
	h.Del("Origin")
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "bmad-x", "", h))
}

func TestBridge_RejectsForeignHost(t *testing.T) {
	ctx, b, _ := authBridge(t)
	assert.Equal(t, http.StatusForbidden,
		getWS(t, ctx, b, "bmad-x", fmt.Sprintf("evil.example:%d", b.GetTerminalPort()), authHeader(b)),
		"DNS-rebinding Host must be rejected")
}

func TestBridge_TokenUniquePerInstance(t *testing.T) {
	b1 := NewBridge(NewSessionManager(nil), nil)
	b2 := NewBridge(NewSessionManager(nil), nil)
	for _, b := range []*Bridge{b1, b2} {
		tok, err := hex.DecodeString(b.Token())
		require.NoError(t, err, "token must be hex")
		assert.Len(t, tok, 32, "token must be 32 random bytes")
	}
	assert.NotEqual(t, b1.Token(), b2.Token())
}

func TestBridge_TokenNotLogged(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })

	ctx, b, _ := authBridge(t)
	h := authHeader(b)
	h.Set("Origin", "https://evil.example")
	getWS(t, ctx, b, "bmad-x", "", h)
	getWS(t, ctx, b, "bmad-x", "", http.Header{"Origin": {"wails://wails"}})

	assert.Contains(t, buf.String(), "https://evil.example", "rejection reason and origin are logged")
	assert.NotContains(t, buf.String(), b.Token(), "token must never be logged")
}

func TestBridge_ValidTokenAndOriginUpgrades(t *testing.T) {
	ctx, b, adapter := authBridge(t)
	url := fmt.Sprintf("ws://127.0.0.1:%d/ws/bmad-x", b.GetTerminalPort())
	dialer := websocket.Dialer{Subprotocols: []string{SubprotocolV1, authSubprotocolPrefix + b.Token()}}
	ws, resp, err := dialer.DialContext(ctx, url, http.Header{"Origin": {"wails://wails"}})
	require.NoError(t, err)
	defer ws.Close()
	assert.Equal(t, SubprotocolV1, resp.Header.Get("Sec-WebSocket-Protocol"),
		"server echoes mashed.v1, never the auth subprotocol")
	require.Eventually(t, func() bool { return adapter.attachCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }

func TestNewBridge_RandFailureFailsClosed(t *testing.T) {
	b := newBridge(NewSessionManager(nil), nil, failingReader{})
	err := b.Start(context.Background())
	require.Error(t, err, "bridge must not listen without a token")
	assert.Equal(t, 0, b.GetTerminalPort())
}
