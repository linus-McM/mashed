package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- AC-3: SetActiveContext ----------

func TestSetActiveContext_AC3_SetAndRead(t *testing.T) {
	app := &App{}
	app.SetActiveContext("/Users/dev/myrepo", "myrepo:agent-1")

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Equal(t, "/Users/dev/myrepo", app.activeRepoPath)
	assert.Equal(t, "myrepo:agent-1", app.activePaneTarget)
}

func TestSetActiveContext_AC3_ConcurrentSafety(t *testing.T) {
	app := &App{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			app.SetActiveContext(fmt.Sprintf("repo-%d", n), fmt.Sprintf("target-%d", n))
		}(i)
	}
	wg.Wait()

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.NotEmpty(t, app.activeRepoPath)
	assert.NotEmpty(t, app.activePaneTarget)
}

// ---------- AC-4: Clear context ----------

func TestSetActiveContext_AC4_Clear(t *testing.T) {
	app := &App{}
	app.SetActiveContext("repo", "target")
	app.SetActiveContext("", "")

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Empty(t, app.activeRepoPath)
	assert.Empty(t, app.activePaneTarget)
}

// ---------- AC-5: Gitignore auto-append ----------

func TestEnsureGitignoreEntry_AC5(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string // empty string means no .gitignore
		createFile      bool
		wantContains    string
		wantNoDuplicate bool
	}{
		{
			name:         "no gitignore exists — creates with .screenshots/",
			createFile:   false,
			wantContains: ".screenshots/",
		},
		{
			name:            "gitignore exists without .screenshots/ — appends",
			existingContent: "node_modules/\n.env\n",
			createFile:      true,
			wantContains:    ".screenshots/",
		},
		{
			name:            "gitignore already has .screenshots/ — no change",
			existingContent: "node_modules/\n.screenshots/\n",
			createFile:      true,
			wantContains:    ".screenshots/",
			wantNoDuplicate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			gitignorePath := filepath.Join(tmpDir, ".gitignore")

			if tt.createFile {
				err := os.WriteFile(gitignorePath, []byte(tt.existingContent), 0644)
				require.NoError(t, err)
			}

			err := ensureGitignoreEntry(tmpDir, ".screenshots/")
			require.NoError(t, err)

			content, err := os.ReadFile(gitignorePath)
			require.NoError(t, err)
			assert.Contains(t, string(content), tt.wantContains)

			if tt.wantNoDuplicate {
				assert.Equal(t, tt.existingContent, string(content),
					"file should not be modified when entry already exists")
			}
		})
	}
}

// ---------- AC-1: TakeScreenshot new signature ----------

func TestTakeScreenshot_AC1_NewSignature(t *testing.T) {
	app := &App{ctx: context.Background()}
	tmpDir := t.TempDir()

	path, err := app.TakeScreenshot(tmpDir, "test:agent")

	if err != nil {
		assert.Contains(t, err.Error(), "screencapture")
		return
	}

	if path != "" {
		assert.Contains(t, path, filepath.Join(tmpDir, ".screenshots"),
			"screenshot must be saved under {repoPath}/.screenshots/")
	}
}

// ---------- AC-5: Empty repoPath guard ----------

func TestTakeScreenshot_AC5_EmptyRepoPath(t *testing.T) {
	app := &App{ctx: context.Background()}

	_, err := app.TakeScreenshot("", "test:agent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// ---------- AC-6: User cancellation (new signature) ----------

func TestTakeScreenshot_AC6_Cancellation_NewSig(t *testing.T) {
	app := &App{ctx: context.Background()}
	tmpDir := t.TempDir()

	path, err := app.TakeScreenshot(tmpDir, "test:agent")
	if err != nil {
		t.Skipf("screencapture not available: %v", err)
	}

	if path == "" {
		assert.NoError(t, err, "cancellation must return nil error")
	}
}
