package main

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFileBase64_AC1_Base64Encoding(t *testing.T) {
	// Known PNG header bytes (minimal valid PNG-like content).
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}

	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	require.NoError(t, os.WriteFile(pngPath, pngBytes, 0644))

	app := appWithDevDir(tmpDir) // tmpDir is an allowed root (R7)
	result, err := app.ReadFileBase64(pngPath)
	require.NoError(t, err)

	// Must start with the correct data URI prefix.
	prefix := "data:image/png;base64,"
	assert.True(t, strings.HasPrefix(result, prefix),
		"expected data URI prefix %q, got %q", prefix, result[:min(len(result), 50)])

	// Decode the base64 payload and verify it matches the original bytes.
	b64Payload := strings.TrimPrefix(result, prefix)
	decoded, err := base64.StdEncoding.DecodeString(b64Payload)
	require.NoError(t, err, "base64 payload must be valid")
	assert.Equal(t, pngBytes, decoded, "decoded bytes must match original file content")
}

func TestReadFileBase64_AC2_MimeTypes(t *testing.T) {
	tests := []struct {
		ext      string
		wantMime string
	}{
		{".png", "image/png"},
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".gif", "image/gif"},
		{".svg", "image/svg+xml"},
		{".webp", "image/webp"},
		{".bmp", "image/bmp"},
		{".ico", "image/x-icon"},
		{".raw", "application/octet-stream"},
	}

	t.Run("mimeForExt", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.ext, func(t *testing.T) {
				got := mimeForExt(tt.ext)
				assert.Equal(t, tt.wantMime, got)
			})
		}
	})

	t.Run("ReadFileBase64_prefix", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := []byte("test-content")
		app := appWithDevDir(tmpDir)

		for _, tt := range tests {
			t.Run(tt.ext, func(t *testing.T) {
				fpath := filepath.Join(tmpDir, "testfile"+tt.ext)
				require.NoError(t, os.WriteFile(fpath, content, 0644))

				result, err := app.ReadFileBase64(fpath)
				require.NoError(t, err)

				wantPrefix := "data:" + tt.wantMime + ";base64,"
				assert.True(t, strings.HasPrefix(result, wantPrefix),
					"ext %s: expected prefix %q, got %q", tt.ext, wantPrefix, result[:min(len(result), 50)])
			})
		}
	})
}

func TestReadFileBase64_AC3_FileSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	bigFile := filepath.Join(tmpDir, "huge.png")

	// Create a file just over 10 MB.
	tenMBPlus := 10*1024*1024 + 1
	require.NoError(t, os.WriteFile(bigFile, make([]byte, tenMBPlus), 0644))

	app := appWithDevDir(tmpDir)
	result, err := app.ReadFileBase64(bigFile)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file too large")
	assert.Empty(t, result, "oversized file must return empty string")
}

func TestReadFileBase64_AC4_MissingFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "nonexistent", "ghost.png")

	app := &App{}
	result, err := app.ReadFileBase64(missingPath)

	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist),
		"error must wrap os.ErrNotExist, got: %v", err)
	assert.Contains(t, err.Error(), missingPath,
		"error must include the file path")
	assert.Empty(t, result)
}

func TestReadFileBase64_AC5_EmptyPath(t *testing.T) {
	app := &App{}
	result, err := app.ReadFileBase64("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty file path")
	assert.Empty(t, result)
}
