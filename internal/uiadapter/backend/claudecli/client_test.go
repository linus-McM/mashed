package claudecli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// TestClaudeCLI_VersionCheck — AC-15.1. Missing `claude` binary surfaces
// ErrBackendUnreachable.
func TestClaudeCLI_VersionCheck(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.ClaudeCLIBinary = "/nonexistent/claude"
	c := NewClient(cfg)
	err := c.WarmUp(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

// TestClaudeCLI_ParsingHandlesShapes — AC-15.3. Three shapes:
// (a) clean fenced JSON, (b) JSON with trailing prose, (c) no JSON at all.
func TestClaudeCLI_ParsingHandlesShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		wantErr bool
		wantJSON string
	}{
		{"clean fenced", "```json\n{\"version\":\"1\"}\n```", false, `{"version":"1"}`},
		{"trailing prose", "sure, here:\n```\n{\"v\":1}\n```\ndone.", false, `{"v":1}`},
		{"bare json", `{"v":1}`, false, `{"v":1}`},
		{"no json", "I cannot produce that output.", true, ""},
		{"empty", "", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractFencedJSON(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantJSON, got)
		})
	}
}

// TestClaudeCLI_Capabilities — Capabilities reflects the CLI backend
// profile.
func TestClaudeCLI_Capabilities(t *testing.T) {
	t.Parallel()
	c := NewClient(uiadapter.DefaultConfig())
	caps := c.Capabilities()
	assert.Equal(t, "claude-cli", caps.Provider)
	assert.True(t, caps.SupportsSingleShot)
	assert.False(t, caps.IsLocal)
}
