package uiadapter

import _ "embed"

//go:embed prompt.md
var embeddedSystemPrompt string

// promptVersion pins the telemetry "prompt_version" field (§4.7.7).
// Bump on any prompt.md edit that changes behaviour.
const promptVersion = "v2"

// SystemPrompt returns the embedded adapter system prompt.
func SystemPrompt() string { return embeddedSystemPrompt }

// PromptVersion exposes the prompt-version tag for telemetry + eval scorecards
// (§4.7.7). Bumped in lockstep with prompt.md edits that change behaviour.
func PromptVersion() string { return promptVersion }
