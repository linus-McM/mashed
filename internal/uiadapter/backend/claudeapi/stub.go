// Package claudeapi hosts the direct Anthropic Messages API backend. Story
// C ships a stub so the registry + router are exercisable end-to-end; the
// real implementation is Story v3-14.
package claudeapi

import (
	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

const Name = "claude-api"

func init() {
	backend.Register(Name, func(cfg uiadapter.Config) (backend.LLMBackend, error) {
		return backend.NewStub(Name, backend.Capabilities{
			Provider:            "anthropic-api",
			Model:               cfg.ClaudeModelPrimary,
			MaxContextTokens:    200_000,
			SupportsSingleShot:  true, // Haiku + Sonnet
			SupportsPromptCache: true, // explicit cache_control markers
			SupportsSeed:        false,
			TokenCostUSDPerMil:  1.0, // placeholder; replaced by v3-11b accountant
			IsLocal:             false,
		}), nil
	})
}
