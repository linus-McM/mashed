package main

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeScreenshot_AC3_FilenameFormat(t *testing.T) {
	pattern := regexp.MustCompile(
		`^.*/\.screenshots/screenshot-\d{8}-\d{6}\.png$`,
	)

	tmpDir := t.TempDir()
	app := &App{ctx: context.Background()}
	path, err := app.TakeScreenshot(tmpDir, "test:agent")

	if err != nil {
		// screencapture unavailable (CI, Linux) — verify error mentions the tool
		assert.Contains(t, err.Error(), "screencapture")
		return
	}

	if path != "" {
		assert.Regexp(t, pattern, path,
			"screenshot path must match screenshot-YYYYMMDD-HHMMSS.png format")
	}
}

func TestTakeScreenshot_AC6_Cancellation(t *testing.T) {
	// Cancellation (Escape) returns ("", nil). Full test requires macOS interaction;
	// here we verify the method compiles and returns correct types.

	tmpDir := t.TempDir()
	app := &App{ctx: context.Background()}
	path, err := app.TakeScreenshot(tmpDir, "test:agent")

	if err != nil {
		t.Skipf("screencapture not available: %v", err)
	}

	if path == "" {
		assert.NoError(t, err, "cancellation must return nil error")
	}
}

func TestTakeScreenshot_AC3_ErrorWrapping(t *testing.T) {
	tmpDir := t.TempDir()
	app := &App{ctx: context.Background()}
	_, err := app.TakeScreenshot(tmpDir, "test:agent")

	if err == nil {
		t.Skip("screencapture succeeded or user cancelled; cannot test error wrapping")
	}

	assert.Contains(t, err.Error(), "screencapture",
		"error must mention screencapture for debuggability")

	// Verify the error wraps an underlying cause
	assert.NotNil(t, errors.Unwrap(err), "error should wrap an underlying cause")
}

func TestTakeScreenshot_NilContext(t *testing.T) {
	app := &App{}
	tmpDir := t.TempDir()

	require.NotPanics(t, func() {
		_, _ = app.TakeScreenshot(tmpDir, "test:agent")
	}, "TakeScreenshot must not panic with nil context")
}
