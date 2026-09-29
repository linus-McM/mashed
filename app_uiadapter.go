package main

// Wails bindings for the UI AST adapter settings surface: adapter/Ollama
// setters plus the Ollama probe and model-listing endpoints.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"sync/atomic"
	"time"

	"mashed/internal/uiadapter"
)

// ErrInvalidConfig is the sentinel returned by validation setters whose input
// fails the contract (timeout out of range, malicious model name). Callers
// branch on it via errors.Is — the wrapper %w preserves the specific reason.
var ErrInvalidConfig = errors.New("invalid config")

// ollamaBaseURLForBindings is the package-level seam tests override via
// swapOllamaBaseURLForTest. Atomic.Pointer protects against string-tuple
// tearing when concurrent tests swap the URL while a binding reads it.
var ollamaBaseURLForBindings atomic.Pointer[string]

// defaultOllamaBaseURL is the local Ollama endpoint. Settings UI pins this
// host per spec §7.3 (no remote inference).
const defaultOllamaBaseURL = "http://localhost:11434"

func init() {
	u := defaultOllamaBaseURL
	ollamaBaseURLForBindings.Store(&u)
}

// loadOllamaBaseURL returns the current Ollama base URL (default or
// test-override). Safe for concurrent use.
func loadOllamaBaseURL() string {
	if p := ollamaBaseURLForBindings.Load(); p != nil {
		return *p
	}
	return defaultOllamaBaseURL
}

// swapOllamaBaseURLForTest atomically replaces the base URL and returns a
// restore func. Tests wrap this with t.Cleanup so the override is scoped.
func swapOllamaBaseURLForTest(url string) (restore func()) {
	prev := ollamaBaseURLForBindings.Load()
	next := url
	ollamaBaseURLForBindings.Store(&next)
	return func() { ollamaBaseURLForBindings.Store(prev) }
}

// validOllamaModelRe rejects path traversal, whitespace, shell metacharacters,
// and anything over 64 bytes. See story §"Risks / gotchas".
var validOllamaModelRe = regexp.MustCompile(`^[a-zA-Z0-9._:-]{1,64}$`)

// validOllamaModelName mirrors the frontend guard in uiAdapterSettings.ts.
func validOllamaModelName(s string) bool { return validOllamaModelRe.MatchString(s) }

const ollamaProbeTimeout = 2 * time.Second

// SetUIAdapterEnabled persists the UI AST adapter master toggle.
func (a *App) SetUIAdapterEnabled(enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.UIAdapterEnabled = enabled
	return saveConfig(cfg)
}

// SetOllamaEnabled persists the Ollama master toggle.
func (a *App) SetOllamaEnabled(enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.OllamaEnabled = enabled
	return saveConfig(cfg)
}

// SetUIAdapterUntrustedExpanded persists the Q7 "expand View raw by default
// when flagged as untrusted" preference.
func (a *App) SetUIAdapterUntrustedExpanded(v bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.UIAdapterUntrustedExpanded = v
	return saveConfig(cfg)
}

// SetUIAdapterTimeoutMs validates the bound and persists the adapter timeout.
// Bounds trace to story §4.4: 500ms min prevents pathological UX, 30000ms max
// keeps a runaway from freezing every interactive node.
func (a *App) SetUIAdapterTimeoutMs(ms int) error {
	if ms < 500 || ms > 30000 {
		return fmt.Errorf("%w: timeoutMs %d out of range 500-30000", ErrInvalidConfig, ms)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.UIAdapterTimeoutMs = ms
	return saveConfig(cfg)
}

// SetOllamaModel validates the model name and persists it. Rejecting junk at
// the binding layer stops a malicious config from smuggling path traversal or
// scheme strings into the HTTP client path.
func (a *App) SetOllamaModel(model string) error {
	if !validOllamaModelName(model) {
		return fmt.Errorf("%w: model name %q fails ^[a-zA-Z0-9._:-]{1,64}$", ErrInvalidConfig, model)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.OllamaModel = model
	return saveConfig(cfg)
}

// ProbeOllamaReachable is the Settings-view liveness check. 2s timeout,
// no retries, no body read — caller only cares whether Ollama answered 200.
func (a *App) ProbeOllamaReachable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), ollamaProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loadOllamaBaseURL()+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// tagsListResponse mirrors Ollama's /api/tags body. Kept unexported because
// the binding only exports sorted names, not the raw metadata.
type tagsListResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListOllamaModels powers the Settings dynamic dropdown. Returns sorted model
// names on success; wraps uiadapter.ErrOllamaUnreachable on transport or non-2xx
// response so the frontend can branch on it for the offline fallback.
//
// Implemented directly rather than via uiadapter.Client so the
// ollamaBaseURLForBindings seam covers both probe and listing without reaching
// into the uiadapter package's private host variable.
func (a *App) ListOllamaModels() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ollamaProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loadOllamaBaseURL()+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("uiadapter: build tags request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", uiadapter.ErrOllamaUnreachable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: ollama returned %s", uiadapter.ErrOllamaUnreachable, resp.Status)
	}

	var decoded tagsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("uiadapter: decode tags response: %w", err)
	}

	names := make([]string, 0, len(decoded.Models))
	for _, m := range decoded.Models {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names, nil
}
