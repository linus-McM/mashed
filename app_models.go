package main

import "mashed/internal/domain"

// ListModels returns the available Claude models for UI dropdowns.
// This is the single source of truth for model selection across the app.
func (a *App) ListModels() []domain.ModelInfo {
	return domain.AvailableModels()
}
