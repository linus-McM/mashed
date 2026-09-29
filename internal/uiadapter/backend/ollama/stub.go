// Package ollama hosts the Ollama LLMBackend implementation. Story C ships
// only a stub that registers via init(); the real client shipping on top
// of `internal/uiadapter/client.go` arrives as part of Phase 2/3 stories.
package ollama

import (
	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

const Name = "ollama"

func init() {
	backend.Register(Name, func(cfg uiadapter.Config) (backend.LLMBackend, error) {
		return backend.NewStub(Name, backend.Capabilities{
			Provider:            "ollama",
			Model:               cfg.Model,
			MaxContextTokens:    cfg.NumCtx,
			SupportsSingleShot:  false, // gemma3:4b is too weak for single-shot UIAST
			SupportsPromptCache: true,  // implicit via byte-stable prefix KV reuse
			SupportsSeed:        true,
			TokenCostUSDPerMil:  0,
			IsLocal:             true,
		}), nil
	})
}
