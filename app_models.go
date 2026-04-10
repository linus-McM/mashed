package main

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
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

// fetchModelsSchema is the JSON schema for structured output from the Claude CLI.
const fetchModelsSchema = `{
  "type": "object",
  "properties": {
    "models": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "alias":         { "type": "string" },
          "id":            { "type": "string" },
          "displayName":   { "type": "string" },
          "contextWindow": { "type": "integer" },
          "tier":          { "type": "string", "enum": ["powerful", "balanced", "fast"] },
          "isDefault":     { "type": "boolean" }
        },
        "required": ["alias", "id", "displayName", "contextWindow", "tier", "isDefault"]
      }
    }
  },
  "required": ["models"]
}`

const fetchModelsPrompt = `List every Claude model alias that the --model flag of the Claude Code CLI accepts.
For each alias, provide:
- alias: the short name (e.g. "opus", "sonnet", "haiku")
- id: the full model ID (e.g. "claude-opus-4-6")
- displayName: a human-readable label (e.g. "Opus 4.6")
- contextWindow: max context window in tokens
- tier: "powerful" for opus-class, "balanced" for sonnet-class, "fast" for haiku-class
- isDefault: true only for the most capable model (opus)

Include all current model aliases. Order from most powerful to least.`

// fetchModelsFromCLI spawns a Claude CLI session to discover available models.
// Returns the fallback list on any error.
func fetchModelsFromCLI(ctx context.Context) []domain.ModelInfo {
	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(fetchCtx, "claude",
		"--print",
		"--no-session-persistence",
		"--output-format", "json",
		"--json-schema", fetchModelsSchema,
		"--model", "haiku",
		"-p", fetchModelsPrompt,
	)

	out, err := cmd.Output()
	if err != nil {
		log.Printf("model discovery via CLI failed: %v — using fallback", err)
		return domain.FallbackModels()
	}

	// --output-format json wraps the response in a JSON envelope.
	// The "result" field contains the assistant's text (which is our structured JSON).
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

// cliJSONEnvelope is the top-level structure from --output-format json.
type cliJSONEnvelope struct {
	Result string `json:"result"`
}

// modelsResponse matches the --json-schema we defined.
type modelsResponse struct {
	Models []domain.ModelInfo `json:"models"`
}

// parseModelResponse extracts ModelInfo from the CLI JSON output.
func parseModelResponse(data []byte) ([]domain.ModelInfo, error) {
	// The output may contain multiple JSON lines (init, assistant, result).
	// We need the "result" line which has type:"result" and contains our data.
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as the result envelope
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		// Look for the result line
		if raw["type"] != "result" {
			continue
		}

		resultStr, ok := raw["result"].(string)
		if !ok || resultStr == "" {
			continue
		}

		// The result field contains the structured JSON from --json-schema
		var resp modelsResponse
		if err := json.Unmarshal([]byte(resultStr), &resp); err != nil {
			return nil, err
		}
		return resp.Models, nil
	}

	// If no result line found, try parsing the entire output as plain JSON
	var resp modelsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
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
