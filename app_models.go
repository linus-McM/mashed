package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"mashed/internal/domain"
)

// modelCache stores models discovered at startup via the Claude CLI.
var (
	modelCache     []domain.ModelInfo
	modelCacheOnce sync.Once
)

const fetchModelsPrompt = `Output ONLY a JSON object (no markdown, no explanation, no code fences) listing every Claude model alias the --model flag accepts.

Format: {"models":[{"alias":"opus","id":"claude-opus-4-6","displayName":"Opus 4.6","contextWindow":1000000,"tier":"powerful","isDefault":true},...]}

tier values: "powerful" for opus, "balanced" for sonnet, "fast" for haiku.
isDefault: true only for opus.
Include all current aliases. Order: most powerful first.
RESPOND WITH ONLY THE JSON OBJECT. NO OTHER TEXT.`

// fetchModelsFromCLI spawns a Claude CLI session to discover available models.
// Returns the fallback list on any error.
func fetchModelsFromCLI(ctx context.Context) []domain.ModelInfo {
	fetchCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := claudeCommand(fetchCtx,
		"--print",
		"--no-session-persistence",
		"--disable-slash-commands",
		"--no-chrome",
		"--tools", "",
		"--model", "haiku",
		"-p", fetchModelsPrompt,
	)

	out, err := cmd.Output()
	if err != nil {
		log.Printf("model discovery via CLI failed: %v (stdout: %.800s) — using fallback", err, string(out))
		return domain.FallbackModels()
	}

	// Plain text output — parse as JSON (handles code fences and surrounding prose).
	// Parse directly as our models response.
	models, err := parseModelResponse(out)
	if err != nil {
		log.Printf("model discovery parse failed: %v — using fallback", err)
		return domain.FallbackModels()
	}

	if len(models) == 0 {
		log.Printf("model discovery returned 0 models — using fallback")
		return domain.FallbackModels()
	}

	log.Printf("discovered %d models from Claude CLI", len(models))
	return models
}

// modelsResponse matches the --json-schema we defined.
type modelsResponse struct {
	Models []domain.ModelInfo `json:"models"`
}

// parseModelResponse extracts ModelInfo from Claude CLI output.
// Handles both clean JSON and JSON wrapped in markdown/prose.
func parseModelResponse(data []byte) ([]domain.ModelInfo, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, fmt.Errorf("empty response")
	}

	// Try direct parse first.
	var resp modelsResponse
	if err := json.Unmarshal([]byte(text), &resp); err == nil {
		return resp.Models, nil
	}

	// Extract JSON object from surrounding text (Claude sometimes adds prose).
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		extracted := text[start : end+1]
		if err := json.Unmarshal([]byte(extracted), &resp); err == nil {
			return resp.Models, nil
		}
	}

	return nil, fmt.Errorf("no valid JSON found in response")
}

// initModelCache populates the model cache. Called once at startup.
func (a *App) initModelCache() {
	modelCacheOnce.Do(func() {
		modelCache = fetchModelsFromCLI(a.ctx)
	})
}

// ListModels returns the available Claude models for UI dropdowns.
func (a *App) ListModels() []domain.ModelInfo {
	// Ensure cache is populated (idempotent via sync.Once).
	a.initModelCache()
	return modelCache
}
