package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// claudeCommand creates an exec.Cmd for the Claude CLI with ANTHROPIC_API_KEY
// removed from the environment. This ensures the CLI uses OAuth/keychain auth
// rather than a potentially stale API key inherited from the parent process.
func claudeCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Env = envWithoutKey(os.Environ(), "ANTHROPIC_API_KEY")
	return cmd
}

// envWithoutKey returns a copy of the environment with the named key removed.
func envWithoutKey(env []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
