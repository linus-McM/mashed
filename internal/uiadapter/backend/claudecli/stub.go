// Package claudecli wraps `claude -p` as a backend. Story C ships a stub
// that registers via init(); Story v3-15 wires the real subprocess.
package claudecli

import (
	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

const Name = "claude-cli"

func init() {
	backend.Register(Name, func(cfg uiadapter.Config) (backend.LLMBackend, error) {
		return backend.NewStub(Name, backend.Capabilities{
			Provider:            "claude-cli",
			Model:               cfg.ClaudeModelPrimary,
			MaxContextTokens:    200_000,
			SupportsSingleShot:  true,
			SupportsPromptCache: true, // via session --resume
			SupportsSeed:        false,
			TokenCostUSDPerMil:  1.0,
			IsLocal:             false,
		}), nil
	})
}
