package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── ValidateMashedAsset — table-driven tests ────────────────────────────

func TestValidateMashedAsset_AC1_MissingDescription(t *testing.T) {
	asset := MashedAssetInfo{
		Name:       "test-cmd",
		Role:       MashedRoleCommand,
		Completion: "idle",
		Chainable:  "single",
		Inputs:     []string{},
		Outputs:    []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	require.NotEmpty(t, issues, "missing description must produce at least one issue")

	found := findIssue(issues, "description", "warn")
	require.NotNil(t, found, "expected a warn issue for field 'description'")
	assert.Contains(t, found.Message, "missing")
}

func TestValidateMashedAsset_DescriptionTooShort(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "Short",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	found := findIssue(issues, "description", "info")
	require.NotNil(t, found, "description < 10 chars should produce an info issue")
	assert.Contains(t, found.Message, "short")
}

func TestValidateMashedAsset_DescriptionTooLong(t *testing.T) {
	longDesc := string(make([]byte, 201))
	for i := range longDesc {
		longDesc = longDesc[:i] + "a" + longDesc[i+1:]
	}
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: longDesc,
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	found := findIssue(issues, "description", "info")
	require.NotNil(t, found, "description > 200 chars should produce an info issue")
	assert.Contains(t, found.Message, "long")
}

func TestValidateMashedAsset_AC2_NonexistentInputPath(t *testing.T) {
	repoPath := t.TempDir()
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{"docs/nonexistent.md"},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, repoPath)
	found := findIssue(issues, "mashedInputs", "warn")
	require.NotNil(t, found, "nonexistent input path should produce a warn issue")
	assert.Contains(t, found.Message, "docs/nonexistent.md")
}

func TestValidateMashedAsset_NonexistentOutputPath(t *testing.T) {
	repoPath := t.TempDir()
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{},
		Outputs:     []string{"out/missing.txt"},
	}
	issues := ValidateMashedAsset(asset, repoPath)
	found := findIssue(issues, "mashedOutputs", "info")
	require.NotNil(t, found, "nonexistent output path should produce an info issue")
	assert.Contains(t, found.Message, "out/missing.txt")
}

func TestValidateMashedAsset_ExistingInputPath_NoIssue(t *testing.T) {
	repoPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "docs", "exists.md"), []byte("hi"), 0o644))

	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{"docs/exists.md"},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, repoPath)
	found := findIssue(issues, "mashedInputs", "warn")
	assert.Nil(t, found, "existing input path should NOT produce an issue")
}

func TestValidateMashedAsset_GlobInputPath(t *testing.T) {
	repoPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, "src"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "src", "main.go"), []byte("package main"), 0o644))

	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{"src/*.go"},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, repoPath)
	found := findIssue(issues, "mashedInputs", "warn")
	assert.Nil(t, found, "glob matching existing files should NOT produce an issue")
}

func TestValidateMashedAsset_AC3_EmptyRepoPathSkipsPathChecks(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{"docs/any.md"},
		Outputs:     []string{"out/any.txt"},
	}
	issues := ValidateMashedAsset(asset, "")
	inputIssue := findIssue(issues, "mashedInputs", "warn")
	outputIssue := findIssue(issues, "mashedOutputs", "info")
	assert.Nil(t, inputIssue, "empty repoPath must skip input path checks")
	assert.Nil(t, outputIssue, "empty repoPath must skip output path checks")
}

func TestValidateMashedAsset_InvalidCompletion(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "bogus",
		Chainable:   "single",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	found := findIssue(issues, "mashedCompletion", "warn")
	require.NotNil(t, found, "invalid completion value should produce a warn issue")
	assert.Contains(t, found.Message, "bogus")
}

func TestValidateMashedAsset_ValidCompletions(t *testing.T) {
	validCompletions := []string{"idle", "exit", "marker:DONE", "timeout:30"}
	for _, comp := range validCompletions {
		t.Run(comp, func(t *testing.T) {
			asset := MashedAssetInfo{
				Name:        "test-cmd",
				Description: "A valid description for the test",
				Role:        MashedRoleCommand,
				Completion:  comp,
				Chainable:   "single",
				Inputs:      []string{},
				Outputs:     []string{},
			}
			issues := ValidateMashedAsset(asset, "")
			found := findIssue(issues, "mashedCompletion", "warn")
			assert.Nil(t, found, "valid completion %q should not produce an issue", comp)
		})
	}
}

func TestValidateMashedAsset_CommandChainableNone(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A valid description for the test",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "none",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	found := findIssue(issues, "mashedChainable", "info")
	require.NotNil(t, found, "command with chainable=none should produce an info issue")
}

func TestValidateMashedAsset_FullyValid(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-cmd",
		Description: "A perfectly valid description here",
		Role:        MashedRoleCommand,
		Completion:  "idle",
		Chainable:   "single",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	assert.Empty(t, issues, "fully valid asset should produce zero issues")
}

func TestValidateMashedAsset_SkillRoleSkipsCompletionCheck(t *testing.T) {
	asset := MashedAssetInfo{
		Name:        "test-skill",
		Description: "A valid skill description",
		Role:        MashedRoleSkill,
		Completion:  "bogus-but-ignored",
		Chainable:   "none",
		Inputs:      []string{},
		Outputs:     []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	// Skills are session-pinned; completion is meaningless for them.
	// The validator should not warn about completion for non-command roles.
	found := findIssue(issues, "mashedCompletion", "warn")
	assert.Nil(t, found, "skill role should not validate mashedCompletion")
}

func TestValidateMashedAsset_MultipleIssues(t *testing.T) {
	asset := MashedAssetInfo{
		Name:       "broken",
		Role:       MashedRoleCommand,
		Completion: "bogus",
		Chainable:  "none",
		Inputs:     []string{},
		Outputs:    []string{},
	}
	issues := ValidateMashedAsset(asset, "")
	// Should have: description missing (warn), invalid completion (warn),
	// command+chainable=none (info)
	assert.GreaterOrEqual(t, len(issues), 3, "multiple validation failures should all be reported")
}

// ── Integration: ValidateMashedAsset on loaded assets ───────────────────

func TestValidateMashedAsset_LoadedAsset_AttachesIssues(t *testing.T) {
	dir := t.TempDir()
	writeAsset(t, dir, "no-desc.md", "---\nmashedRole: command\nmashedCompletion: idle\nmashedChainable: single\n---\n")

	got, err := LoadMashedAssetsFromDir(dir, MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err)
	require.Len(t, got, 1)

	// Simulate what ListAllMashedAssets does: validate with repoPath.
	issues := ValidateMashedAsset(got[0], "")
	assert.NotEmpty(t, issues, "missing description should produce issues")
	found := findIssue(issues, "description", "warn")
	assert.NotNil(t, found)
}

func TestValidateMashedAsset_LoadedAsset_NoIssues(t *testing.T) {
	dir := t.TempDir()
	writeAsset(t, dir, "good.md", "---\nname: good\ndescription: \"A perfectly valid description here\"\nmashedRole: command\nmashedCompletion: idle\nmashedChainable: single\n---\nbody\n")

	got, err := LoadMashedAssetsFromDir(dir, MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err)
	require.Len(t, got, 1)

	issues := ValidateMashedAsset(got[0], "")
	assert.Empty(t, issues, "valid asset should produce zero issues")
}

// ── test helpers ────────────────────────────────────────────────────────

// findIssue locates the first ValidationIssue matching field+severity.
func findIssue(issues []ValidationIssue, field, severity string) *ValidationIssue {
	for i := range issues {
		if issues[i].Field == field && issues[i].Severity == severity {
			return &issues[i]
		}
	}
	return nil
}
