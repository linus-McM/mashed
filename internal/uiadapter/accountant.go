package uiadapter

import (
	"errors"
	"sync"
	"time"
)

// Plan §3 Story 11b — Cost + rate-limit accountant for Claude backends.
// Sliding-window counters per minute for RPM/TPM + running USD spend.
// Soft limits at 85% of configured budgets trip the breaker proactively.

// ErrAccountantLimit is surfaced when a soft/hard limit would be
// exceeded. Consumers should fall back via Story v3-11 tiered chain.
var ErrAccountantLimit = errors.New("uiadapter: accountant limit reached")

// Pricing is the per-model cost table. Values are USD per million tokens.
// Plan §3 Story 11b "Hardcoded for v3.0; externalizable later."
type Pricing struct {
	InputUSDPerMil      float64
	OutputUSDPerMil     float64
	CacheReadDiscount   float64 // ×(1 - CacheReadDiscount) on reads (typ. 0.9)
	CacheCreateMarkup   float64 // ×(1 + CacheCreateMarkup) on creates (typ. 0.25)
}

var claudeHaiku = Pricing{InputUSDPerMil: 0.80, OutputUSDPerMil: 4.00, CacheReadDiscount: 0.90, CacheCreateMarkup: 0.25}
var claudeSonnet = Pricing{InputUSDPerMil: 3.00, OutputUSDPerMil: 15.00, CacheReadDiscount: 0.90, CacheCreateMarkup: 0.25}

// PricingFor returns the Pricing for a model name. Unknown names return
// the zero Pricing (free) so non-Claude backends don't trip the
// accountant.
func PricingFor(model string) Pricing {
	switch model {
	case "claude-haiku-4-5":
		return claudeHaiku
	case "claude-sonnet-4-6", "claude-opus-4-6":
		return claudeSonnet
	default:
		return Pricing{}
	}
}

// Usage is a single-call token-use tuple the accountant adds to the
// running counters. Fields match Anthropic's Messages API usage block.
type Usage struct {
	Model                    string
	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
}

// CostUSD computes the USD cost of a Usage tuple.
func CostUSD(u Usage) float64 {
	p := PricingFor(u.Model)
	if p == (Pricing{}) {
		return 0
	}
	input := float64(u.InputTokens) / 1_000_000 * p.InputUSDPerMil
	output := float64(u.OutputTokens) / 1_000_000 * p.OutputUSDPerMil
	cacheRead := float64(u.CacheReadInputTokens) / 1_000_000 * p.InputUSDPerMil * (1 - p.CacheReadDiscount)
	cacheCreate := float64(u.CacheCreationInputTokens) / 1_000_000 * p.InputUSDPerMil * (1 + p.CacheCreateMarkup)
	return input + output + cacheRead + cacheCreate
}

// Accountant tracks RPM/TPM/USD across a single adapter session. Safe for
// concurrent use.
type Accountant struct {
	mu       sync.Mutex
	cfg      Config
	now      func() time.Time
	requests []time.Time // request timestamps
	tokens   []tokenTick
	spend    float64
}

type tokenTick struct {
	t      time.Time
	tokens int
}

// NewAccountant constructs the accountant from Config. The clock is
// injectable for deterministic tests.
func NewAccountant(cfg Config) *Accountant {
	return &Accountant{cfg: cfg, now: time.Now}
}

// CheckPrecall returns an error when the next call would exceed any soft
// limit at the 85% line (plan §3 Story 11b). Callers fall back via the
// breaker chain.
func (a *Accountant) CheckPrecall(estimatedTokens int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.expireLocked()

	if a.cfg.RPMSoftLimit > 0 {
		soft := int(float64(a.cfg.RPMSoftLimit) * 0.85)
		if len(a.requests) >= soft {
			return ErrAccountantLimit
		}
	}
	if a.cfg.TPMSoftLimit > 0 && estimatedTokens > 0 {
		soft := int(float64(a.cfg.TPMSoftLimit) * 0.85)
		sum := 0
		for _, t := range a.tokens {
			sum += t.tokens
		}
		if sum+estimatedTokens >= soft {
			return ErrAccountantLimit
		}
	}
	if a.cfg.UsdBudgetPerSession > 0 && a.spend >= a.cfg.UsdBudgetPerSession*0.85 {
		return ErrAccountantLimit
	}
	return nil
}

// Record ingests one Usage tuple plus its wall-clock timestamp. Updates
// running counters + USD spend.
func (a *Accountant) Record(u Usage) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	a.requests = append(a.requests, now)
	a.tokens = append(a.tokens, tokenTick{t: now, tokens: u.InputTokens + u.OutputTokens})
	a.spend += CostUSD(u)
}

// Snapshot returns the current running totals for telemetry.
func (a *Accountant) Snapshot() (requests int, tokens int, spendUSD float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.expireLocked()
	tok := 0
	for _, t := range a.tokens {
		tok += t.tokens
	}
	return len(a.requests), tok, a.spend
}

// expireLocked drops request/token entries older than 60 seconds so the
// sliding window stays bounded. Caller holds a.mu.
func (a *Accountant) expireLocked() {
	cutoff := a.now().Add(-time.Minute)
	i := 0
	for i < len(a.requests) && a.requests[i].Before(cutoff) {
		i++
	}
	a.requests = a.requests[i:]
	j := 0
	for j < len(a.tokens) && a.tokens[j].t.Before(cutoff) {
		j++
	}
	a.tokens = a.tokens[j:]
}
