package main

// Plan v3 Story 18 — Wails bindings for the title-bar Backend / Model /
// Router-policy selectors. Every setter validates against an enum + the
// frontend's client-side validator mirrors the same list.

import (
	"errors"
	"fmt"
	"slices"

	"mashed/internal/uiadapter"
)

// Sentinel errors the frontend branches on for precise error surfaces
// (AC-18.8).
var (
	ErrInvalidBackend      = errors.New("invalid backend")
	ErrInvalidClaudeModel  = errors.New("invalid claude model")
	ErrInvalidCLIModel     = errors.New("invalid cli model")
	ErrInvalidRouterPolicy = errors.New("invalid router policy")
)

// backendsEnum — Plan §3 Story 18 "Backend: ["ollama","claude-api","claude-cli"]".
var backendsEnum = []string{"ollama", "claude-api", "claude-cli"}

// routerPoliciesEnum — Plan §3 Story 16 six policies.
var routerPoliciesEnum = []string{
	"local-only",
	"claude-only",
	"claude-first",
	"ollama-first",
	"cost-aware",
	"privacy-strict",
}

// SetBackend persists the selected backend after validating against the
// enum. Returns ErrInvalidBackend wrapped on unknown names.
func (a *App) SetBackend(backend string) error {
	if !slices.Contains(backendsEnum, backend) {
		return fmt.Errorf("%w: %q not one of %v", ErrInvalidBackend, backend, backendsEnum)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.Backend = backend
	return saveConfig(cfg)
}

// SetClaudeModel validates against the ClaudeAllowlist + persists. The
// Allowlist lives in the uiadapter package (Story v3-12).
func (a *App) SetClaudeModel(model string) error {
	if !slices.Contains(uiadapter.ClaudeAllowlist, model) {
		return fmt.Errorf("%w: %q not in allowlist %v",
			ErrInvalidClaudeModel, model, uiadapter.ClaudeAllowlist)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.ClaudeModel = model
	return saveConfig(cfg)
}

// SetCLIModel mirrors SetClaudeModel for the claude-cli backend. Same
// allowlist — the CLI wraps the API and accepts the same model names.
func (a *App) SetCLIModel(model string) error {
	if !slices.Contains(uiadapter.ClaudeAllowlist, model) {
		return fmt.Errorf("%w: %q not in allowlist %v",
			ErrInvalidCLIModel, model, uiadapter.ClaudeAllowlist)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.CLIModel = model
	return saveConfig(cfg)
}

// SetRouterPolicy persists the policy after validation.
func (a *App) SetRouterPolicy(policy string) error {
	if !slices.Contains(routerPoliciesEnum, policy) {
		return fmt.Errorf("%w: %q not one of %v",
			ErrInvalidRouterPolicy, policy, routerPoliciesEnum)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.RouterPolicy = policy
	return saveConfig(cfg)
}

// ListBackendsAvailable returns the enum the frontend populates the
// Backend pulldown with.
func (a *App) ListBackendsAvailable() []string {
	out := make([]string, len(backendsEnum))
	copy(out, backendsEnum)
	return out
}

// ListClaudeModels returns the Story v3-12 Claude allowlist.
func (a *App) ListClaudeModels() []string {
	out := make([]string, len(uiadapter.ClaudeAllowlist))
	copy(out, uiadapter.ClaudeAllowlist)
	return out
}

// ListRouterPolicies returns the Plan §3 Story 16 policy enum.
func (a *App) ListRouterPolicies() []string {
	out := make([]string, len(routerPoliciesEnum))
	copy(out, routerPoliciesEnum)
	return out
}
