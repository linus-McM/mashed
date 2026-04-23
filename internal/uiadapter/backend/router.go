package backend

import (
	"context"
	"errors"
	"regexp"
	"sync"

	"mashed/internal/uiadapter"
)

// Plan §3 Story 16 — Backend Router + Policy. Chooses the backend per
// request and orchestrates cross-backend fallback.

// RouterPolicy enumerates the supported policies. Values mirror the
// Config.RouterPolicy string enum.
type RouterPolicy string

const (
	PolicyLocalOnly     RouterPolicy = "local-only"
	PolicyClaudeOnly    RouterPolicy = "claude-only"
	PolicyClaudeFirst   RouterPolicy = "claude-first"
	PolicyOllamaFirst   RouterPolicy = "ollama-first"
	PolicyCostAware     RouterPolicy = "cost-aware"
	PolicyPrivacyStrict RouterPolicy = "privacy-strict"
)

// Router picks the primary + escalation + fallback chain for a given
// raw capture.
type Router struct {
	cfg             uiadapter.Config
	privacyPatterns []*regexp.Regexp
	backends        map[string]LLMBackend

	mu     sync.RWMutex
	health map[string]error // nil = healthy; non-nil = unreachable
}

// NewRouter builds a Router. `backends` is the map returned by the
// registry (name → concrete LLMBackend); `health` is injected by the
// lifecycle (Story v3-17) and consulted on every Decide call.
func NewRouter(cfg uiadapter.Config, backends map[string]LLMBackend) *Router {
	patterns := cfg.PrivacyPatterns
	if len(patterns) == 0 {
		patterns = defaultPrivacyPatterns()
	}
	return &Router{
		cfg:             cfg,
		privacyPatterns: patterns,
		backends:        backends,
		health:          map[string]error{},
	}
}

// SetHealth marks a backend's current health. nil means healthy. Safe
// for concurrent use — the lifecycle Health ticker writes while Decide
// reads.
func (r *Router) SetHealth(name string, state error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.health == nil {
		r.health = map[string]error{}
	}
	r.health[name] = state
}

// Decision is the routing outcome. Primary is the first-choice backend
// name; Fallbacks is the ordered alternatives; Escalation is the model
// used for hard cases (Claude Sonnet on escalation).
type Decision struct {
	Primary    string
	Fallbacks  []string
	Escalation string
	Reason     string
}

// Decide returns a Decision for the given raw capture. Privacy-strict
// short-circuits Claude when the raw matches a privacy pattern.
func (r *Router) Decide(raw string) Decision {
	policy := RouterPolicy(r.cfg.RouterPolicy)
	if policy == "" {
		if r.backends["claude-api"] != nil {
			policy = PolicyClaudeFirst
		} else {
			policy = PolicyLocalOnly
		}
	}

	// Privacy gate — overrides all policies.
	if r.matchesPrivacy(raw) && policy != PolicyLocalOnly {
		return Decision{
			Primary:   "ollama",
			Fallbacks: []string{},
			Reason:    "privacy-strict override: raw matches a privacy pattern",
		}
	}

	switch policy {
	case PolicyLocalOnly:
		return Decision{Primary: "ollama", Reason: string(policy)}
	case PolicyClaudeOnly:
		return Decision{Primary: "claude-api", Escalation: "claude-sonnet-4-6", Reason: string(policy)}
	case PolicyOllamaFirst:
		return Decision{Primary: "ollama", Fallbacks: []string{"claude-api"}, Reason: string(policy)}
	case PolicyCostAware:
		return r.decideCostAware()
	case PolicyPrivacyStrict:
		return Decision{Primary: "ollama", Reason: string(policy)}
	case PolicyClaudeFirst:
		fallthrough
	default:
		return Decision{Primary: "claude-api", Escalation: "claude-sonnet-4-6", Fallbacks: []string{"ollama"}, Reason: string(policy)}
	}
}

func (r *Router) decideCostAware() Decision {
	r.mu.RLock()
	err, ok := r.health["ollama"]
	r.mu.RUnlock()
	if ok && err == nil {
		return Decision{Primary: "ollama", Fallbacks: []string{"claude-api"}, Reason: "cost-aware: ollama healthy"}
	}
	return Decision{Primary: "claude-api", Fallbacks: []string{"ollama"}, Escalation: "claude-sonnet-4-6", Reason: "cost-aware: ollama down"}
}

func (r *Router) matchesPrivacy(raw string) bool {
	for _, p := range r.privacyPatterns {
		if p.MatchString(raw) {
			return true
		}
	}
	return false
}

// defaultPrivacyPatterns — plan §3 Story 16 "Privacy patterns (defaults)".
func defaultPrivacyPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		regexp.MustCompile(`ghp_[A-Za-z0-9]{36,}`),
		regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
		regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`),
		regexp.MustCompile(`eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`), // JWT
	}
}

// errNoBackend is surfaced when the decision names a backend that isn't
// registered.
var errNoBackend = errors.New("router: decided backend is not registered")

// Resolve returns the LLMBackend named by the decision. Returns errNoBackend
// if absent.
func (r *Router) Resolve(d Decision) (LLMBackend, error) {
	b, ok := r.backends[d.Primary]
	if !ok {
		return nil, errNoBackend
	}
	return b, nil
}

// Dispatch calls Decide then Resolve then invokes fn. Exposes the
// decision + escalation name for §6.1 `router_decision` slog attribute.
func (r *Router) Dispatch(ctx context.Context, raw string, fn func(ctx context.Context, b LLMBackend) (*uiadapter.UIAST, error)) (*uiadapter.UIAST, Decision, error) {
	d := r.Decide(raw)
	b, err := r.Resolve(d)
	if err != nil {
		return nil, d, err
	}
	ast, err := fn(ctx, b)
	return ast, d, err
}
