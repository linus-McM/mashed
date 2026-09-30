package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AC-4 (ArtifactType/ArtifactSpec types) is tested in types_test.go.

const testRepoPath = "/tmp/test"

// ── AC-1: Registry completeness ──

func TestAC1_RegistryCompleteness(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range AllProcesses() {
		for _, name := range p.Inputs {
			seen[name] = true
		}
		for _, name := range p.Outputs {
			seen[name] = true
		}
	}

	for name := range seen {
		t.Run(name, func(t *testing.T) {
			_, exists := artifactPaths[name]
			assert.True(t, exists, "artifact %q used in registry but missing from artifactPaths", name)
		})
	}
}

// ── AC-2: ResolveArtifactPath ──

func TestAC2_ResolveArtifactPath(t *testing.T) {
	tests := []struct {
		name     string
		artifact string
		want     string
	}{
		{
			name:     "known markdown artifact",
			artifact: "PRD.md",
			want:     testRepoPath + "/_bmad-output/planning-artifacts/PRD.md",
		},
		{
			name:     "unmapped artifact code",
			artifact: "code",
			want:     "",
		},
		{
			name:     "unmapped artifact tests",
			artifact: "tests",
			want:     "",
		},
		{
			name:     "unknown artifact",
			artifact: "unknown-thing",
			want:     "",
		},
		{
			name:     "directory artifact epics",
			artifact: "epics/",
			want:     testRepoPath + "/_bmad-output/solutioning-artifacts/epics/",
		},
		{
			name:     "directory name without trailing slash is unknown",
			artifact: "epics",
			want:     "",
		},
		{
			name:     "architecture artifact",
			artifact: "architecture.md",
			want:     testRepoPath + "/_bmad-output/solutioning-artifacts/architecture.md",
		},
		{
			name:     "sprint status yaml",
			artifact: "sprint-status.yaml",
			want:     testRepoPath + "/_bmad-output/implementation-artifacts/sprint-status.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveArtifactPath(tt.artifact, testRepoPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ── AC-3: VerifyArtifacts ──

func TestAC3_VerifyArtifacts_MixedFoundMissing(t *testing.T) {
	repo := t.TempDir()

	prdDir := filepath.Join(repo, "_bmad-output", "planning-artifacts")
	require.NoError(t, os.MkdirAll(prdDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(prdDir, "PRD.md"), []byte("# PRD"), 0o644))

	found, missing := VerifyArtifacts(repo, []string{"PRD.md", "architecture.md"})

	assert.Equal(t, []string{"PRD.md"}, found)
	assert.Equal(t, []string{"architecture.md"}, missing)
}

func TestAC3_VerifyArtifacts_EdgeCases(t *testing.T) {
	t.Run("unmapped artifacts excluded from both slices", func(t *testing.T) {
		repo := t.TempDir()
		found, missing := VerifyArtifacts(repo, []string{"code", "tests"})
		assert.Empty(t, found)
		assert.Empty(t, missing)
	})

	t.Run("empty output list returns empty slices", func(t *testing.T) {
		repo := t.TempDir()
		found, missing := VerifyArtifacts(repo, []string{})
		assert.Empty(t, found)
		assert.Empty(t, missing)
	})
}

func TestAC3_VerifyArtifacts_NoBmadOutputDir(t *testing.T) {
	repo := t.TempDir()

	found, missing := VerifyArtifacts(repo, []string{"PRD.md", "architecture.md"})

	assert.Empty(t, found)
	assert.Equal(t, []string{"PRD.md", "architecture.md"}, missing)
}

// ── AC-5: Directory artifact verification ──

func TestAC5_VerifyArtifacts_DirectoryArtifact(t *testing.T) {
	repo := t.TempDir()

	epicsDir := filepath.Join(repo, "_bmad-output", "solutioning-artifacts", "epics")
	require.NoError(t, os.MkdirAll(epicsDir, 0o755))

	found, missing := VerifyArtifacts(repo, []string{"epics/"})

	assert.Equal(t, []string{"epics/"}, found)
	assert.Empty(t, missing)
}

func TestAC3_VerifyArtifacts_StatErrorNotNotExist(t *testing.T) {
	skipIfRoot(t)
	// Verify that non-NotExist os.Stat errors (e.g. permission denied) are
	// explicitly handled and still treated as missing.
	repo := t.TempDir()

	// Create a parent dir without execute permission so Stat on a child returns
	// permission denied rather than not-exist.
	parentDir := filepath.Join(repo, "_bmad-output", "planning-artifacts")
	require.NoError(t, os.MkdirAll(parentDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(parentDir, "PRD.md"), []byte("# PRD"), 0o644))
	require.NoError(t, os.Chmod(parentDir, 0o000))
	t.Cleanup(func() { os.Chmod(parentDir, 0o755) })

	found, missing := VerifyArtifacts(repo, []string{"PRD.md"})

	assert.Empty(t, found)
	assert.Equal(t, []string{"PRD.md"}, missing)
}

func TestAC5_VerifyArtifacts_DirectoryMissing(t *testing.T) {
	repo := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(repo, "_bmad-output"), 0o755))

	found, missing := VerifyArtifacts(repo, []string{"epics/"})

	assert.Empty(t, found)
	assert.Equal(t, []string{"epics/"}, missing)
}
