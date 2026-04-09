package bmad

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSkillFiles_CreatesAllDirectories(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	// Should have exactly 32 subdirectories
	assert.Len(t, entries, 32, "expected 32 skill directories")

	for _, entry := range entries {
		assert.True(t, entry.IsDir(), "expected %s to be a directory", entry.Name())
		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		_, err := os.Stat(skillPath)
		assert.NoError(t, err, "expected SKILL.md in %s", entry.Name())
	}
}

func TestGenerateSkillFiles_DirectoryNamesMatchSkillNames(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// Collect all expected skill names from AllProcesses
	procs := AllProcesses()
	expected := make(map[string]bool, len(procs))
	for _, p := range procs {
		if p.SkillName == "" {
			continue
		}
		expected[p.SkillName] = true
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	actual := make(map[string]bool, len(entries))
	for _, entry := range entries {
		actual[entry.Name()] = true
	}

	assert.Equal(t, expected, actual, "directory names should match SkillNames exactly")
}

func TestGenerateSkillFiles_FrontmatterCorrect(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	tests := []struct {
		skillName string
	}{
		{"bmad-brainstorming"},
		{"bmad-create-prd"},
	}

	for _, tc := range tests {
		t.Run(tc.skillName, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(dir, tc.skillName, "SKILL.md"))
			require.NoError(t, err)

			text := string(content)

			// Must start with frontmatter delimiter
			assert.True(t, strings.HasPrefix(text, "---\n"), "should start with ---")

			// Must contain name field
			assert.Contains(t, text, "name: "+tc.skillName)

			// Must contain description field
			assert.Contains(t, text, "description:")

			// Must have closing frontmatter delimiter (second ---)
			// Find second occurrence of "---"
			firstEnd := strings.Index(text, "---") + 3
			rest := text[firstEnd:]
			assert.Contains(t, rest, "---", "should have closing frontmatter delimiter")
		})
	}
}

func TestGenerateSkillFiles_InputPathsResolved(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// bmad-create-architecture has inputs: ["PRD.md"]
	content, err := os.ReadFile(filepath.Join(dir, "bmad-create-architecture", "SKILL.md"))
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "_bmad-output/planning-artifacts/PRD.md",
		"input PRD.md should resolve to its artifact path")
}

func TestGenerateSkillFiles_OutputPathsResolved(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// bmad-create-prd has outputs: ["PRD.md"]
	content, err := os.ReadFile(filepath.Join(dir, "bmad-create-prd", "SKILL.md"))
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "_bmad-output/planning-artifacts/PRD.md",
		"output PRD.md should resolve to its artifact path")
}

func TestGenerateSkillFiles_NoInputsFallback(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// bmad-brainstorming has inputs: []
	content, err := os.ReadFile(filepath.Join(dir, "bmad-brainstorming", "SKILL.md"))
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "No file inputs",
		"processes with no inputs should show fallback text")
}

func TestGenerateSkillFiles_UnmappedOutputFallback(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// bmad-dev-story has outputs: ["code", "tests"] — both unmapped
	content, err := os.ReadFile(filepath.Join(dir, "bmad-dev-story", "SKILL.md"))
	require.NoError(t, err)

	text := string(content)
	// Should NOT contain _bmad-output paths for unmapped outputs
	assert.NotContains(t, text, "_bmad-output/code")
	assert.NotContains(t, text, "_bmad-output/tests")
	// Should indicate it operates on the codebase directly
	assert.Contains(t, text, "No file outputs",
		"unmapped outputs should show codebase fallback text")
}

func TestGenerateSkillFiles_ErrorOnInvalidBaseDir(t *testing.T) {
	// Use a path that cannot be created (file exists where dir is expected)
	dir := t.TempDir()
	blockingFile := filepath.Join(dir, "bmad-brainstorming")
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), 0644))

	err := GenerateSkillFiles(dir)
	require.Error(t, err, "should fail when directory creation is blocked by a file")
	assert.Contains(t, err.Error(), "create skill dir")
}

func TestGenerateSkillFiles_ErrorOnReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	// Create the subdirectory but make it read-only so file creation fails
	subDir := filepath.Join(dir, "bmad-brainstorming")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.Chmod(subDir, 0555))
	t.Cleanup(func() { os.Chmod(subDir, 0755) })

	err := GenerateSkillFiles(dir)
	require.Error(t, err, "should fail when SKILL.md cannot be created")
	assert.Contains(t, err.Error(), "create SKILL.md")
}

func TestGenerateSkillFiles_MixedMappedAndUnmappedOutputs(t *testing.T) {
	dir := t.TempDir()
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// bmad-code-review has outputs: ["review-report", "code"]
	// review-report is mapped, code is unmapped
	content, err := os.ReadFile(filepath.Join(dir, "bmad-code-review", "SKILL.md"))
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "_bmad-output/implementation-artifacts/reviews/",
		"mapped output should have resolved path")
	assert.NotContains(t, text, "_bmad-output/code",
		"unmapped output 'code' should not appear as a path")
}

func TestGenerateSkillFiles_Idempotent(t *testing.T) {
	dir := t.TempDir()

	// First call
	err := GenerateSkillFiles(dir)
	require.NoError(t, err)

	// Second call — should not error
	err = GenerateSkillFiles(dir)
	require.NoError(t, err)

	// Still 32 directories
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 32, "idempotent call should still have 32 directories")

	// Each still has SKILL.md
	for _, entry := range entries {
		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		_, err := os.Stat(skillPath)
		assert.NoError(t, err, "SKILL.md should still exist in %s after second call", entry.Name())
	}
}
