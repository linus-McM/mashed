package uiadapter

import _ "embed"

//go:embed prompt.md
var embeddedSystemPrompt string

// promptVersion pins the telemetry "prompt_version" field (§4.7.7).
// Bump on any prompt.md edit that changes behaviour.
const promptVersion = "v1"

// SystemPrompt returns the embedded adapter system prompt.
func SystemPrompt() string { return embeddedSystemPrompt }
