package main

// Story ui-ast-U5: contract tests for the UI-adapter Wails bindings.

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// withBindingsBaseURL overrides the package-level seam that routes binding
// HTTP calls (probe + ListOllamaModels) at a test stub and restores it.
func withBindingsBaseURL(t *testing.T, url string) {
	t.Helper()
	t.Cleanup(swapOllamaBaseURLForTest(url))
}

// closedSocketURL binds a localhost port, closes it, and returns a URL
// pointing at the now-dead address — guarantees a transport-level refusal.
func closedSocketURL(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())
	return "http://" + addr
}

// AC-1: SetUIAdapterEnabled persists the toggle.

func TestU5_AC1_App_SetUIAdapterEnabled_Persists(t *testing.T) {
	app := setupTestConfig(t, "")

	require.NoError(t, app.SetUIAdapterEnabled(false))

	got := loadConfig()
	assert.False(t, got.UIAdapterEnabled, "explicit opt-out must persist to disk")

	raw := readRawConfig(t)
	assert.Contains(t, raw, "uiAdapterEnabled", "key must be present on disk")
	var val bool
	require.NoError(t, json.Unmarshal(raw["uiAdapterEnabled"], &val))
	assert.False(t, val, "on-disk uiAdapterEnabled must be false after opt-out")
}

func TestU5_AC1_App_SetUIAdapterEnabled_RoundTripBothDirections(t *testing.T) {
	app := setupTestConfig(t, "")

	require.NoError(t, app.SetUIAdapterEnabled(false))
	assert.False(t, loadConfig().UIAdapterEnabled, "after opt-out")

	require.NoError(t, app.SetUIAdapterEnabled(true))
	assert.True(t, loadConfig().UIAdapterEnabled, "after opt-in restore (default TRUE per 2026-04-21)")

	require.NoError(t, app.SetUIAdapterEnabled(false))
	assert.False(t, loadConfig().UIAdapterEnabled, "after second opt-out (round-trip twice)")
}

// AC-2: SetUIAdapterTimeoutMs rejects out-of-range values.

func TestU5_AC2_App_SetUIAdapterTimeoutMs_BoundsCheck(t *testing.T) {
	defaultMs := defaultConfig().UIAdapterTimeoutMs
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{name: "below min rejects", input: 250, wantErr: true},
		{name: "above max rejects", input: 50000, wantErr: true},
		{name: "exact min accepts", input: 500},
		{name: "exact max accepts", input: 30000},
		{name: "mid-range accepts", input: 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, "")
			// Prime disk with the default so wantOnDisk=3000 is a real round-trip check.
			require.NoError(t, saveConfig(defaultConfig()))

			err := app.SetUIAdapterTimeoutMs(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidConfig),
					"out-of-range error must wrap ErrInvalidConfig; got %v", err)
				assert.Equal(t, defaultMs, loadConfig().UIAdapterTimeoutMs,
					"rejected value must NOT touch disk")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, loadConfig().UIAdapterTimeoutMs,
				"accepted value must persist to disk")
		})
	}
}

// AC-3: SetOllamaModel validates against ^[a-zA-Z0-9._:-]{1,64}$.

func TestU5_AC3_App_SetOllamaModel_ValidatesName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid gemma3", input: "gemma3:4b"},
		{name: "valid qwen", input: "qwen2.5:3b"},
		{name: "valid llama", input: "llama3.2:3b"},
		{name: "empty string rejects", input: "", wantErr: true},
		{name: "path traversal rejects", input: "../../etc/passwd", wantErr: true},
		{name: "forward slash rejects", input: "foo/bar", wantErr: true},
		{name: "whitespace rejects", input: "gemma 3:4b", wantErr: true},
		{name: "over 64 chars rejects", input: "a23456789012345678901234567890123456789012345678901234567890123456", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestConfig(t, "")
			require.NoError(t, saveConfig(defaultConfig()))
			baseline := loadConfig().OllamaModel

			err := app.SetOllamaModel(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidConfig),
					"validation error must wrap ErrInvalidConfig; got %v", err)
				assert.Equal(t, baseline, loadConfig().OllamaModel,
					"rejected model name must NOT touch disk")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, loadConfig().OllamaModel,
				"accepted model name must persist to disk")
		})
	}
}

// AC-4: ProbeOllamaReachable returns false within 2s for a dead socket.

func TestU5_AC4_App_ProbeOllamaReachable_DeadSocket(t *testing.T) {
	app := setupTestConfig(t, "")
	withBindingsBaseURL(t, closedSocketURL(t))

	start := time.Now()
	reachable := app.ProbeOllamaReachable()
	elapsed := time.Since(start)

	assert.False(t, reachable, "dead socket must yield false")
	assert.Less(t, elapsed, 2500*time.Millisecond,
		"probe must complete inside 2s timeout + overhead; took %s", elapsed)
}

func TestU5_AC4_App_ProbeOllamaReachable_LiveServer(t *testing.T) {
	app := setupTestConfig(t, "")

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/tags", r.URL.Path, "probe must hit /api/tags")
		assert.Equal(t, http.MethodGet, r.Method, "probe must use GET")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(stub.Close)
	withBindingsBaseURL(t, stub.URL)

	assert.True(t, app.ProbeOllamaReachable(), "200 OK must yield true")
}

func TestU5_AC4_App_ProbeOllamaReachable_Non2xx(t *testing.T) {
	app := setupTestConfig(t, "")

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(stub.Close)
	withBindingsBaseURL(t, stub.URL)

	assert.False(t, app.ProbeOllamaReachable(), "5xx must yield false (probe is a liveness check, not a tolerant one)")
}

// AC-8: UIAdapterUntrustedExpanded default is false + setter persists.

func TestU5_AC8_MashedConfig_UntrustedExpanded_DefaultFalse(t *testing.T) {
	setupTestConfig(t, "")

	cfg := loadConfig()
	assert.False(t, cfg.UIAdapterUntrustedExpanded,
		"fresh config must have UIAdapterUntrustedExpanded == false (§11 Q7)")
}

func TestU5_AC8_MashedConfig_UntrustedExpanded_ExplicitTrue_Honored(t *testing.T) {
	setupTestConfig(t, `{"devDir":"/x","uiAdapterUntrustedExpanded":true}`)

	cfg := loadConfig()
	assert.True(t, cfg.UIAdapterUntrustedExpanded,
		"explicit uiAdapterUntrustedExpanded:true must round-trip")
}

func TestU5_AC8_App_SetUIAdapterUntrustedExpanded_Persists(t *testing.T) {
	app := setupTestConfig(t, "")

	require.NoError(t, app.SetUIAdapterUntrustedExpanded(true))
	assert.True(t, loadConfig().UIAdapterUntrustedExpanded, "setter must persist true")

	require.NoError(t, app.SetUIAdapterUntrustedExpanded(false))
	assert.False(t, loadConfig().UIAdapterUntrustedExpanded, "setter must persist false")

	raw := readRawConfig(t)
	assert.Contains(t, raw, "uiAdapterUntrustedExpanded",
		"key must appear on disk so explicit false survives reload")
}

// Q6 coverage: SetOllamaEnabled persists.

func TestU5_App_SetOllamaEnabled_Persists(t *testing.T) {
	app := setupTestConfig(t, "")

	require.NoError(t, app.SetOllamaEnabled(false))
	assert.False(t, loadConfig().OllamaEnabled, "opt-out must persist")

	require.NoError(t, app.SetOllamaEnabled(true))
	assert.True(t, loadConfig().OllamaEnabled, "opt-in restore")

	raw := readRawConfig(t)
	assert.Contains(t, raw, "ollamaEnabled")
}

// AC-10/AC-11 binding layer: ListOllamaModels delegates to uiadapter.Client.

func TestU5_App_ListOllamaModels_ReturnsSortedList(t *testing.T) {
	app := setupTestConfig(t, "")

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/tags", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen2.5:3b"},{"name":"gemma3:4b"},{"name":"llama3.2:3b"}]}`))
	}))
	t.Cleanup(stub.Close)
	withBindingsBaseURL(t, stub.URL)

	got, err := app.ListOllamaModels()
	require.NoError(t, err)
	assert.Equal(t, []string{"gemma3:4b", "llama3.2:3b", "qwen2.5:3b"}, got,
		"binding must return models sorted lexicographically")
}

func TestU5_App_ListOllamaModels_EmptyOK(t *testing.T) {
	app := setupTestConfig(t, "")

	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	t.Cleanup(stub.Close)
	withBindingsBaseURL(t, stub.URL)

	got, err := app.ListOllamaModels()
	require.NoError(t, err)
	assert.Empty(t, got, "zero pulled models is not an error")
}

func TestU5_App_ListOllamaModels_WrapsErrOllamaUnreachable(t *testing.T) {
	app := setupTestConfig(t, "")
	withBindingsBaseURL(t, closedSocketURL(t))

	_, err := app.ListOllamaModels()
	require.Error(t, err)
	assert.True(t, errors.Is(err, uiadapter.ErrOllamaUnreachable),
		"transport failure must wrap uiadapter.ErrOllamaUnreachable so the frontend can branch on it; got %v", err)
}

// ErrInvalidConfig sentinel — AC-2/AC-3 wrap it, callers branch on it.

func TestU5_ErrInvalidConfig_IsDistinctSentinel(t *testing.T) {
	require.NotNil(t, ErrInvalidConfig, "ErrInvalidConfig must be a defined error value")
	assert.NotEqual(t, ErrInvalidConfig, os.ErrInvalid,
		"ErrInvalidConfig must be its own sentinel, not a re-export of os.ErrInvalid")
	assert.False(t, errors.Is(errors.New("unrelated"), ErrInvalidConfig))
}
