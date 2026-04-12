package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleSkillMD is a realistic SKILL.md fixture with known frontmatter
// and a multi-line body. Tests reference it to assert round-trip
// symmetry and selective-edit correctness.
const sampleSkillMD = `---
name: simplify
description: Review changed code for reuse and quality
mashedRole: skill
mashedCompletion: idle
mashedChainable: single
mashedSessionPinned: true
mashedInputs:
  - repo
  - branch
mashedOutputs:
  - summary
customField: preserve-me
---
# Simplify

This is the body content.
It spans multiple lines.

` + "```go" + `
func main() {
    fmt.Println("hello")
}
` + "```" + `
`

func TestWriteMashedAssetFrontmatter(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		updates   map[string]interface{}
		wantErr   error
		checkFile func(t *testing.T, got []byte)
	}{
		{
			name:    "AC-1: unchanged save is byte-identical",
			content: sampleSkillMD,
			updates: map[string]interface{}{
				"name":                "simplify",
				"description":         "Review changed code for reuse and quality",
				"mashedRole":          "skill",
				"mashedCompletion":    "idle",
				"mashedChainable":     "single",
				"mashedSessionPinned": true,
				"mashedInputs":        []interface{}{"repo", "branch"},
				"mashedOutputs":       []interface{}{"summary"},
			},
			checkFile: func(t *testing.T, got []byte) {
				assert.Equal(t, sampleSkillMD, string(got),
					"unchanged save must produce byte-identical file")
			},
		},
		{
			name:    "AC-2: edit mashedRole persists, body unchanged",
			content: sampleSkillMD,
			updates: map[string]interface{}{
				"mashedRole": "command",
			},
			checkFile: func(t *testing.T, got []byte) {
				// The frontmatter should now say mashedRole: command
				assert.Contains(t, string(got), "mashedRole: command")
				assert.NotContains(t, string(got), "mashedRole: skill")

				// Body must be byte-identical
				_, origBody, _ := extractFrontmatter([]byte(sampleSkillMD))
				_, newBody, _ := extractFrontmatter(got)
				assert.Equal(t, string(origBody), string(newBody),
					"body must be byte-identical after frontmatter edit")

				// Other frontmatter fields unchanged
				assert.Contains(t, string(got), "customField: preserve-me")
				assert.Contains(t, string(got), "mashedCompletion: idle")
			},
		},
		{
			name:    "preserves non-mashed frontmatter fields",
			content: sampleSkillMD,
			updates: map[string]interface{}{
				"mashedChainable": "any",
				"unknownField":   "should-be-ignored", // not in mashedWriteableFields
			},
			checkFile: func(t *testing.T, got []byte) {
				assert.Contains(t, string(got), "mashedChainable: any")
				assert.Contains(t, string(got), "customField: preserve-me")
				assert.NotContains(t, string(got), "unknownField")
			},
		},
		{
			name: "AC-6: file without frontmatter returns ErrNoFrontmatterBlock",
			content: `# Legacy Skill

This file has no frontmatter block.
`,
			updates: map[string]interface{}{"mashedRole": "skill"},
			wantErr: ErrNoFrontmatterBlock,
		},
		{
			name:    "updates list fields (mashedInputs)",
			content: sampleSkillMD,
			updates: map[string]interface{}{
				"mashedInputs": []interface{}{"file", "context", "prompt"},
			},
			checkFile: func(t *testing.T, got []byte) {
				assert.Contains(t, string(got), "- file")
				assert.Contains(t, string(got), "- context")
				assert.Contains(t, string(got), "- prompt")
				// Old values should be gone
				assert.NotContains(t, string(got), "- repo")
			},
		},
		{
			name:    "updates boolean field (mashedSessionPinned to false)",
			content: sampleSkillMD,
			updates: map[string]interface{}{
				"mashedSessionPinned": false,
			},
			checkFile: func(t *testing.T, got []byte) {
				assert.Contains(t, string(got), "mashedSessionPinned: false")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write the fixture to a temp .claude/skills/ directory so
			// path validation passes.
			dir := t.TempDir()
			assetDir := filepath.Join(dir, ".claude", "skills", "test-skill")
			require.NoError(t, os.MkdirAll(assetDir, 0o755))
			path := filepath.Join(assetDir, "SKILL.md")
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o644))

			err := WriteMashedAssetFrontmatter(path, tt.updates)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				// Verify file was not modified
				got, rerr := os.ReadFile(path)
				require.NoError(t, rerr)
				assert.Equal(t, tt.content, string(got),
					"file must not be modified on error")
				return
			}

			require.NoError(t, err)
			got, err := os.ReadFile(path)
			require.NoError(t, err)
			if tt.checkFile != nil {
				tt.checkFile(t, got)
			}
		})
	}
}

func TestWriteMashedAssetFrontmatter_PathValidation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "AC-5: /etc/passwd rejected",
			path:    "/etc/passwd",
			wantErr: ErrPathOutsideAllowedRoots,
		},
		{
			name:    "traversal attack rejected",
			path:    "/home/user/.claude/skills/../../etc/passwd",
			wantErr: ErrPathOutsideAllowedRoots, // filepath.Clean resolves this
		},
		{
			name:    "valid global path accepted",
			path:    "/Users/test/.claude/skills/foo/SKILL.md",
			wantErr: nil, // path validation passes but file won't exist
		},
		{
			name:    "valid local path accepted",
			path:    "/some/repo/.claude/commands/test.md",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAssetPath(tt.path)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWriteMashedAssetFrontmatter_WriteFailure(t *testing.T) {
	// AC-3: simulate write failure via read-only directory.
	dir := t.TempDir()
	assetDir := filepath.Join(dir, ".claude", "skills", "test-skill")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	path := filepath.Join(assetDir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(sampleSkillMD), 0o644))

	// Make directory read+execute-only so temp file creation fails
	// but existing files remain readable (macOS needs +x to traverse).
	require.NoError(t, os.Chmod(assetDir, 0o555))
	defer os.Chmod(assetDir, 0o755) // restore for cleanup

	err := WriteMashedAssetFrontmatter(path, map[string]interface{}{
		"mashedRole": "command",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create temp")

	// Verify original file is unchanged (dir still 0o555 — readable).
	os.Chmod(assetDir, 0o755) // restore for cleanup to work
	got, rerr := os.ReadFile(path)
	require.NoError(t, rerr)
	assert.Equal(t, sampleSkillMD, string(got),
		"original file must be unchanged after write failure")
}

func TestWriteMashedAssetFrontmatter_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude", "skills", "missing", "SKILL.md")
	err := WriteMashedAssetFrontmatter(path, map[string]interface{}{
		"mashedRole": "skill",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading")
}

func TestWriteMashedAssetFrontmatter_RoundTrip(t *testing.T) {
	// Write → re-read via parseMashedAsset → assert fields match.
	dir := t.TempDir()
	assetDir := filepath.Join(dir, ".claude", "skills", "roundtrip")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	path := filepath.Join(assetDir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(sampleSkillMD), 0o644))

	// Update several fields.
	err := WriteMashedAssetFrontmatter(path, map[string]interface{}{
		"name":             "renamed-skill",
		"mashedRole":       "command",
		"mashedCompletion": "exit",
		"mashedInputs":     []interface{}{"alpha", "beta"},
		"mashedOutputs":    []interface{}{"result"},
	})
	require.NoError(t, err)

	// Re-read via the same parser the loader uses.
	info, err := parseMashedAsset(path, "roundtrip", MashedKindSkill, MashedSourceLocal)
	require.NoError(t, err)
	require.NotNil(t, info)

	// parseMashedAsset takes Name from the directory, not frontmatter.
	// Verify the frontmatter name was written by reading raw bytes.
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "name: renamed-skill")
	assert.Equal(t, MashedRoleCommand, info.Role)
	assert.Equal(t, "exit", info.Completion)
	assert.Equal(t, []string{"alpha", "beta"}, info.Inputs)
	assert.Equal(t, []string{"result"}, info.Outputs)
	// Description should survive from original.
	assert.Equal(t, "Review changed code for reuse and quality", info.Description)
}

func TestWriteMashedAssetFrontmatter_AddNewField(t *testing.T) {
	// Tests the "append new key" path in applyNodeUpdate when a mashed
	// field doesn't exist in the original frontmatter.
	const minimalFM = `---
name: bare
mashedRole: skill
---
Body text.
`
	dir := t.TempDir()
	assetDir := filepath.Join(dir, ".claude", "skills", "bare")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	path := filepath.Join(assetDir, "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte(minimalFM), 0o644))

	err := WriteMashedAssetFrontmatter(path, map[string]interface{}{
		"mashedCompletion":    "exit",
		"mashedSessionPinned": true,
		"mashedInputs":        []string{"a", "b"}, // exercises []string branch
	})
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "mashedCompletion: exit")
	assert.Contains(t, string(got), "mashedSessionPinned: true")
	assert.Contains(t, string(got), "- a")
	assert.Contains(t, string(got), "- b")
	// Body preserved.
	assert.Contains(t, string(got), "Body text.")
}

func TestWriteMashedAssetFrontmatter_EmptyFrontmatter(t *testing.T) {
	// File with empty frontmatter block (---\n---\n).
	const emptyFM = "---\n---\nBody only.\n"
	dir := t.TempDir()
	assetDir := filepath.Join(dir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	path := filepath.Join(assetDir, "test.md")
	require.NoError(t, os.WriteFile(path, []byte(emptyFM), 0o644))

	err := WriteMashedAssetFrontmatter(path, map[string]interface{}{
		"mashedRole": "command",
		"name":       "test-cmd",
	})
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "mashedRole: command")
	assert.Contains(t, string(got), "name: test-cmd")
	assert.Contains(t, string(got), "Body only.")
}
