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

// AvailableModels returns the canonical list of Claude models.
// This is the single source of truth — all model dropdowns and Go
// defaults should reference this list rather than hardcoding IDs.
func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			ID:            "claude-opus-4-6",
			Alias:         "opus",
			DisplayName:   "Opus 4.6",
			ContextWindow: 1_000_000,
			Tier:          "powerful",
			IsDefault:     true,
		},
		{
			ID:            "claude-sonnet-4-6",
			Alias:         "sonnet",
			DisplayName:   "Sonnet 4.6",
			ContextWindow: 200_000,
			Tier:          "balanced",
		},
		{
			ID:            "claude-haiku-4-5-20251001",
			Alias:         "haiku",
			DisplayName:   "Haiku 4.5",
			ContextWindow: 200_000,
			Tier:          "fast",
		},
	}
}

// DefaultModelID returns the ID of the default model.
func DefaultModelID() string {
	for _, m := range AvailableModels() {
		if m.IsDefault {
			return m.ID
		}
	}
	return "claude-opus-4-6"
}

// ModelByAlias finds a model by its alias. Returns the default if not found.
func ModelByAlias(alias string) ModelInfo {
	for _, m := range AvailableModels() {
		if m.Alias == alias || m.ID == alias {
			return m
		}
	}
	// Fallback: return default
	for _, m := range AvailableModels() {
		if m.IsDefault {
			return m
		}
	}
	return AvailableModels()[0]
}
