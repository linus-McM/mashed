package domain

// ModelInfo describes a Claude model available for use.
type ModelInfo struct {
	ID            string `json:"id"`            // full model ID, e.g. "claude-opus-4-6"
	Alias         string `json:"alias"`         // short alias, e.g. "opus"
	DisplayName   string `json:"displayName"`   // human label, e.g. "Opus 4.6"
	ContextWindow int    `json:"contextWindow"` // max context in tokens
	Tier          string `json:"tier"`          // "powerful" | "balanced" | "fast"
	IsDefault     bool   `json:"isDefault"`     // default for agent spawning
}

// FallbackModels is the minimal set returned when dynamic discovery fails.
// This ensures the app is always functional, even offline.
func FallbackModels() []ModelInfo {
	return []ModelInfo{
		{ID: "claude-opus-4-6", Alias: "opus", DisplayName: "Opus", ContextWindow: 1_000_000, Tier: "powerful", IsDefault: true},
		{ID: "claude-sonnet-4-6", Alias: "sonnet", DisplayName: "Sonnet", ContextWindow: 200_000, Tier: "balanced"},
		{ID: "claude-haiku-4-5-20251001", Alias: "haiku", DisplayName: "Haiku", ContextWindow: 200_000, Tier: "fast"},
	}
}

// DefaultAlias returns the alias of the default model from the given list.
func DefaultAlias(models []ModelInfo) string {
	for _, m := range models {
		if m.IsDefault {
			return m.Alias
		}
	}
	if len(models) > 0 {
		return models[0].Alias
	}
	return "opus"
}

// ModelByAlias finds a model by alias or ID. Returns the default if not found.
func ModelByAlias(models []ModelInfo, alias string) ModelInfo {
	for _, m := range models {
		if m.Alias == alias || m.ID == alias {
			return m
		}
	}
	if len(models) > 0 {
		for _, m := range models {
			if m.IsDefault {
				return m
			}
		}
		return models[0]
	}
	return FallbackModels()[0]
}
