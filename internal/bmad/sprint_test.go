package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSprintYAML writes a sprint-status.yaml into a temp repo layout and
// returns the repo root directory.
func createSprintYAML(t *testing.T, content string) string {
	t.Helper()
	repoDir := t.TempDir()
	dir := filepath.Join(repoDir, "_bmad-output", "implementation-artifacts")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "sprint-status.yaml"),
		[]byte(content),
		0644,
	))
	return repoDir
}

const multiEpicYAML = `generated: "2026-04-07T10:00:00Z"
last_updated: "2026-04-07T12:00:00Z"
project: "mashed"
project_key: "MSHD"
tracking_system: "github"
story_location: "docs/stories"
development_status:
  epic-1: in-progress
  1-1-user-auth: done
  1-2-dashboard: in-progress
  1-3-settings: backlog
  epic-2: backlog
  2-1-notifications: ready-for-dev
  2-2-export-data: backlog
`

const emptyDevStatusYAML = `generated: "2026-04-07T10:00:00Z"
last_updated: "2026-04-07T12:00:00Z"
project: "mashed"
project_key: "MSHD"
tracking_system: "github"
story_location: "docs/stories"
development_status:
`

const malformedYAML = `generated: "ok"
development_status:
  - this is a list not a map
  - which is wrong
`

const retrospectiveYAML = `generated: "2026-04-08T10:00:00Z"
last_updated: "2026-04-08T12:00:00Z"
project: "mashed"
project_key: "MSHD"
tracking_system: "github"
story_location: "docs/stories"
development_status:
  epic-1: in-progress
  1-1-user-auth: done
  sprint1-retrospective: done
`

func TestParseSprintStatus_MultiEpic(t *testing.T) {
	repoDir := createSprintYAML(t, multiEpicYAML)

	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)

	// Top-level metadata.
	assert.Equal(t, "2026-04-07T10:00:00Z", result.Generated)
	assert.Equal(t, "2026-04-07T12:00:00Z", result.LastUpdated)
	assert.Equal(t, "mashed", result.Project)
	assert.Equal(t, "MSHD", result.ProjectKey)
	assert.Equal(t, "github", result.TrackingSystem)
	assert.Equal(t, "docs/stories", result.StoryLocation)

	// Two epics in order.
	require.Len(t, result.Epics, 2)

	// --- Epic 1 ---
	e1 := result.Epics[0]
	assert.Equal(t, "epic-1", e1.ID)
	assert.Equal(t, EpicInProgress, e1.Status)
	require.Len(t, e1.Stories, 3)

	assert.Equal(t, "1-1-user-auth", e1.Stories[0].ID)
	assert.Equal(t, "epic-1", e1.Stories[0].EpicID)
	assert.Equal(t, StoryDone, e1.Stories[0].Status)
	assert.Equal(t, 0, e1.Stories[0].Sequence)

	assert.Equal(t, "1-2-dashboard", e1.Stories[1].ID)
	assert.Equal(t, StoryInProgress, e1.Stories[1].Status)
	assert.Equal(t, 1, e1.Stories[1].Sequence)

	assert.Equal(t, "1-3-settings", e1.Stories[2].ID)
	assert.Equal(t, StoryBacklog, e1.Stories[2].Status)
	assert.Equal(t, 2, e1.Stories[2].Sequence)

	// --- Epic 2 ---
	e2 := result.Epics[1]
	assert.Equal(t, "epic-2", e2.ID)
	assert.Equal(t, EpicBacklog, e2.Status)
	require.Len(t, e2.Stories, 2)

	assert.Equal(t, "2-1-notifications", e2.Stories[0].ID)
	assert.Equal(t, StoryReadyForDev, e2.Stories[0].Status)
	assert.Equal(t, 0, e2.Stories[0].Sequence)

	assert.Equal(t, "2-2-export-data", e2.Stories[1].ID)
	assert.Equal(t, StoryBacklog, e2.Stories[1].Status)
	assert.Equal(t, 1, e2.Stories[1].Sequence)
}

func TestParseSprintStatus_MissingFile(t *testing.T) {
	repoDir := t.TempDir() // no yaml file created

	_, err := ParseSprintStatus(repoDir)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSprintFileNotFound)
}

func TestParseSprintStatus_EmptyDevStatus(t *testing.T) {
	repoDir := createSprintYAML(t, emptyDevStatusYAML)

	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)

	assert.Equal(t, "mashed", result.Project)
	assert.Empty(t, result.Epics)
}

func TestParseSprintStatus_Malformed(t *testing.T) {
	repoDir := createSprintYAML(t, malformedYAML)

	_, err := ParseSprintStatus(repoDir)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSprintFileMalformed)
}

func TestParseSprintStatus_Retrospective(t *testing.T) {
	repoDir := createSprintYAML(t, retrospectiveYAML)

	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)

	require.Len(t, result.Epics, 1)
	e := result.Epics[0]
	require.Len(t, e.Stories, 2)

	assert.Equal(t, "1-1-user-auth", e.Stories[0].ID)
	assert.Equal(t, "sprint1-retrospective", e.Stories[1].ID)
	assert.Equal(t, "epic-1", e.Stories[1].EpicID)
	assert.Equal(t, StoryDone, e.Stories[1].Status)
}

// --- UpdateStoryStatus tests ---

const commentedYAML = `# Sprint status file with comments
generated: "2026-04-07T10:00:00Z"
last_updated: "2026-04-07T12:00:00Z"
project: "mashed"  # project name
project_key: "MSHD"
tracking_system: "github"
story_location: "docs/stories"
# Development status section
development_status:
  epic-1: in-progress  # first epic
  1-1-user-auth: done
  1-2-dashboard: in-progress  # WIP
  1-3-settings: backlog
`

func TestUpdateStoryStatus_Success(t *testing.T) {
	repoDir := createSprintYAML(t, multiEpicYAML)

	err := UpdateStoryStatus(repoDir, "1-2-dashboard", "done")
	require.NoError(t, err)

	// Verify via ParseSprintStatus round-trip.
	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)

	require.Len(t, result.Epics, 2)
	e1 := result.Epics[0]
	require.Len(t, e1.Stories, 3)

	// The updated story.
	assert.Equal(t, "1-2-dashboard", e1.Stories[1].ID)
	assert.Equal(t, StoryDone, e1.Stories[1].Status)

	// Other stories unchanged.
	assert.Equal(t, StoryDone, e1.Stories[0].Status)
	assert.Equal(t, StoryBacklog, e1.Stories[2].Status)

	// Epic 2 unchanged.
	e2 := result.Epics[1]
	assert.Equal(t, StoryReadyForDev, e2.Stories[0].Status)
	assert.Equal(t, StoryBacklog, e2.Stories[1].Status)
}

func TestUpdateStoryStatus_InvalidStatus(t *testing.T) {
	repoDir := createSprintYAML(t, multiEpicYAML)

	// Read original content for comparison.
	path := sprintStatusPath(repoDir)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	err = UpdateStoryStatus(repoDir, "1-2-dashboard", "invalid-status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// File must be unchanged.
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestUpdateStoryStatus_StoryNotFound(t *testing.T) {
	repoDir := createSprintYAML(t, multiEpicYAML)

	err := UpdateStoryStatus(repoDir, "99-99-nope", "done")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrStoryNotFound)
}

func TestUpdateStoryStatus_InvalidStoryID(t *testing.T) {
	repoDir := createSprintYAML(t, multiEpicYAML)

	tests := []struct {
		name    string
		storyID string
	}{
		{"no digits prefix", "not-a-valid-id"},
		{"empty string", ""},
		{"just dashes", "---"},
		{"epic key", "epic-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateStoryStatus(repoDir, tt.storyID, "done")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid story ID")
		})
	}
}

func TestUpdateStoryStatus_PreservesFormatting(t *testing.T) {
	repoDir := createSprintYAML(t, commentedYAML)

	err := UpdateStoryStatus(repoDir, "1-2-dashboard", "done")
	require.NoError(t, err)

	// Read the raw file back.
	path := sprintStatusPath(repoDir)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	// Comments must be preserved.
	assert.Contains(t, content, "# Sprint status file with comments")
	assert.Contains(t, content, "# project name")
	assert.Contains(t, content, "# Development status section")
	assert.Contains(t, content, "# first epic")

	// The updated value should be "done".
	assert.Contains(t, content, "1-2-dashboard: done")

	// Other values unchanged.
	assert.Contains(t, content, "1-1-user-auth: done")
	assert.Contains(t, content, "1-3-settings: backlog")
}

func TestUpdateStoryStatus_MissingFile(t *testing.T) {
	repoDir := t.TempDir() // no yaml file

	err := UpdateStoryStatus(repoDir, "1-2-dashboard", "done")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSprintFileNotFound)
}

func TestValidateStoryStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StoryStatus
		wantOK  bool
	}{
		{"backlog", "backlog", StoryBacklog, true},
		{"ready-for-dev", "ready-for-dev", StoryReadyForDev, true},
		{"in-progress", "in-progress", StoryInProgress, true},
		{"review", "review", StoryReview, true},
		{"done", "done", StoryDone, true},
		{"invalid", "cancelled", "", false},
		{"empty", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ValidateStoryStatus(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestValidateEpicStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EpicStatus
		wantOK  bool
	}{
		{"backlog", "backlog", EpicBacklog, true},
		{"in-progress", "in-progress", EpicInProgress, true},
		{"done", "done", EpicDone, true},
		{"invalid", "cancelled", "", false},
		{"empty", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ValidateEpicStatus(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
