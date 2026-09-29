// Package bmad contains tests for validateInput and resolveFileInput.
// Story bmad-interactive-03: Suspension primitive + RespondToInput Wails binding.
//
// RED Phase: These tests define the contract for the new validate.go file.
// They MUST fail until the go-engineer implements validateInput and
// resolveFileInput (GREEN phase).
package bmad

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── TestValidateInput — table-driven for all InputShapes ──────────────────────

// TestValidateInput covers every InputShape declared in §8.4.
// validateInput does not exist yet — RED phase: will fail to compile.
func TestValidateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    InputSpec
		value   string
		wantErr bool
		wantMsg string // substring expected in error message when wantErr is true
	}{
		// ── ShapeFree ──────────────────────────────────────────────────────────
		{
			name:    "ShapeFree required empty → error",
			spec:    InputSpec{Shape: ShapeFree, Required: true},
			value:   "",
			wantErr: true,
			wantMsg: "required",
		},
		{
			name:    "ShapeFree required non-empty → nil",
			spec:    InputSpec{Shape: ShapeFree, Required: true},
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "ShapeFree MaxLength exceeded → error",
			spec:    InputSpec{Shape: ShapeFree, MaxLength: 5},
			value:   "toolong",
			wantErr: true,
			wantMsg: "max length",
		},
		{
			name:    "ShapeFree MaxLength exactly at limit → nil",
			spec:    InputSpec{Shape: ShapeFree, MaxLength: 5},
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "ShapeFree Validation regex pass → nil",
			spec:    InputSpec{Shape: ShapeFree, Validation: `^\d{4}$`},
			value:   "1234",
			wantErr: false,
		},
		{
			name:    "ShapeFree Validation regex fail → error",
			spec:    InputSpec{Shape: ShapeFree, Validation: `^\d{4}$`},
			value:   "abcd",
			wantErr: true,
			wantMsg: "validation",
		},

		// ── ShapeChoice ────────────────────────────────────────────────────────
		{
			name:    "ShapeChoice value in Options → nil",
			spec:    InputSpec{Shape: ShapeChoice, Options: []string{"a", "b", "c"}},
			value:   "b",
			wantErr: false,
		},
		{
			name:    "ShapeChoice value not in Options → error mentioning options",
			spec:    InputSpec{Shape: ShapeChoice, Options: []string{"a", "b", "c"}},
			value:   "z",
			wantErr: true,
			wantMsg: "[a b c]",
		},

		// ── ShapeMultiChoice ───────────────────────────────────────────────────
		{
			name:    "ShapeMultiChoice all values in Options → nil",
			spec:    InputSpec{Shape: ShapeMultiChoice, Options: []string{"a", "b", "c"}},
			value:   "a,b",
			wantErr: false,
		},
		{
			name:    "ShapeMultiChoice one value not in Options → error",
			spec:    InputSpec{Shape: ShapeMultiChoice, Options: []string{"a", "b", "c"}},
			value:   "a,z",
			wantErr: true,
		},
		{
			name:    "ShapeMultiChoice single trimmed value → nil",
			spec:    InputSpec{Shape: ShapeMultiChoice, Options: []string{"a", "b"}},
			value:   " a ",
			wantErr: false,
		},

		// ── ShapeApproval ──────────────────────────────────────────────────────
		{
			name:    "ShapeApproval yes → nil",
			spec:    InputSpec{Shape: ShapeApproval},
			value:   "yes",
			wantErr: false,
		},
		{
			name:    "ShapeApproval no → nil",
			spec:    InputSpec{Shape: ShapeApproval},
			value:   "no",
			wantErr: false,
		},
		{
			name:    "ShapeApproval true → error (must be yes/no exactly)",
			spec:    InputSpec{Shape: ShapeApproval},
			value:   "true",
			wantErr: true,
			wantMsg: "yes or no",
		},
		{
			name:    "ShapeApproval false → error",
			spec:    InputSpec{Shape: ShapeApproval},
			value:   "false",
			wantErr: true,
			wantMsg: "yes or no",
		},
		{
			name:    "ShapeApproval empty required → error",
			spec:    InputSpec{Shape: ShapeApproval, Required: true},
			value:   "",
			wantErr: true,
		},

		// ── ShapeFile ──────────────────────────────────────────────────────────
		// (Path traversal is tested in TestResolveFileInput — separate concern.)
		{
			name:    "ShapeFile required empty → error",
			spec:    InputSpec{Shape: ShapeFile, Required: true},
			value:   "",
			wantErr: true,
			wantMsg: "required",
		},
		{
			name:    "ShapeFile Validation regex pass → nil",
			spec:    InputSpec{Shape: ShapeFile, Validation: `\.md$`},
			value:   "/repo/docs/file.md",
			wantErr: false,
		},
		{
			name:    "ShapeFile MaxLength exceeded → error",
			spec:    InputSpec{Shape: ShapeFile, MaxLength: 10},
			value:   "/very/long/path/that/exceeds/limit.md",
			wantErr: true,
			wantMsg: "max length",
		},

		// ── ShapeJSON ──────────────────────────────────────────────────────────
		{
			name: "ShapeJSON default 64KiB cap: 65537-byte payload → error",
			spec: InputSpec{Shape: ShapeJSON, MaxLength: 0}, // 0 → default 65536
			// Build a value of exactly 65537 bytes.
			value:   strings.Repeat("x", 65537),
			wantErr: true,
			wantMsg: "max length",
		},
		{
			name:    "ShapeJSON default 64KiB cap: 65536-byte payload → nil",
			spec:    InputSpec{Shape: ShapeJSON, MaxLength: 0},
			value:   strings.Repeat("x", 65536),
			wantErr: false,
		},
		{
			name:    "ShapeJSON explicit MaxLength overrides default",
			spec:    InputSpec{Shape: ShapeJSON, MaxLength: 10},
			value:   strings.Repeat("x", 11),
			wantErr: true,
			wantMsg: "max length",
		},
		{
			name:    "ShapeJSON explicit MaxLength: value within limit → nil",
			spec:    InputSpec{Shape: ShapeJSON, MaxLength: 10},
			value:   strings.Repeat("x", 10),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// validateInput does not exist yet — RED phase.
			err := validateInput(tt.spec, tt.value)

			if tt.wantErr {
				require.Error(t, err, "validateInput must return error for %q", tt.name)
				if tt.wantMsg != "" {
					assert.Contains(t, err.Error(), tt.wantMsg,
						"error message must contain %q", tt.wantMsg)
				}
				return
			}

			assert.NoError(t, err, "validateInput must return nil for %q", tt.name)
		})
	}
}

// ── TestResolveFileInput ──────────────────────────────────────────────────────

// TestResolveFileInput covers the path-traversal guard from §14.2.
// resolveFileInput does not exist yet — RED phase: will fail to compile.
func TestResolveFileInput(t *testing.T) {
	t.Parallel()

	// Use a fixed-name parent so we can construct a "prefix-similar but outside" path.
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "repo")
	// Note: we don't create the directory — resolveFileInput only checks the path,
	// not whether the file exists. If the implementation does stat, adjust here.

	sep := string(filepath.Separator)

	tests := []struct {
		name        string
		value       string
		wantErr     bool
		wantErrIs   error
		wantClean   bool // if true, returned path must equal filepath.Clean(abs(value))
	}{
		{
			name:      "absolute path inside repo root → clean abs path returned",
			value:     filepath.Join(repoRoot, "subdir", "foo.md"),
			wantErr:   false,
			wantClean: true,
		},
		{
			name:      "/etc/passwd → ErrPathOutsideRepo",
			value:     "/etc/passwd",
			wantErr:   true,
			wantErrIs: ErrPathOutsideRepo,
		},
		{
			name:  "relative traversal resolves outside → error",
			// "../../etc/passwd" relative to cwd likely escapes t.TempDir.
			// We use an absolute path constructed to escape repoRoot.
			value:     filepath.Join(repoRoot, "..", "..", "etc", "passwd"),
			wantErr:   true,
			wantErrIs: ErrPathOutsideRepo,
		},
		{
			name:      "path with trailing separator handled correctly (if inside repo)",
			value:     repoRoot + sep + "file" + sep,
			wantErr:   false, // filepath.Clean removes trailing sep
			wantClean: true,
		},
		{
			name: "prefix-similar-but-outside: catches HasPrefix without sep bug",
			// e.g. repoRoot = /tmp/.../repo, outsider = /tmp/.../repo-other/foo
			// A naive strings.HasPrefix(abs, repoRoot) would incorrectly allow this.
			value:     filepath.Join(parent, "repo-other", "foo"),
			wantErr:   true,
			wantErrIs: ErrPathOutsideRepo,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// resolveFileInput does not exist yet — RED phase.
			got, err := resolveFileInput(tt.value, repoRoot)

			if tt.wantErr {
				require.Error(t, err, "resolveFileInput must error for %q", tt.name)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs),
						"error must wrap %v, got: %v", tt.wantErrIs, err)
				}
				return
			}

			require.NoError(t, err, "resolveFileInput must not error for %q", tt.name)
			if tt.wantClean {
				// The returned path must be a clean absolute path.
				assert.True(t, filepath.IsAbs(got),
					"returned path must be absolute, got: %q", got)
				assert.Equal(t, filepath.Clean(got), got,
					"returned path must be clean (no redundant elements)")
			}
		})
	}
}

// ── Story ui-ast-U0, AC-4: ShapeJSON accepts bare-string submissions ──────────
//
// Back-compat branch: until U7 lands (frontend submits structured JSON),
// validateInput must accept a plain string under ShapeJSON as a legacy
// submission. The 64 KiB cap from §13.4 still applies.
func TestU0_AC4_ValidateInput_ShapeJSON_BareStringAccepted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "bare non-JSON string is accepted (legacy submission)",
			value:   "pick option 2",
			wantErr: false,
		},
		{
			name:    "valid JSON map is accepted",
			value:   `{"confirm":"done"}`,
			wantErr: false,
		},
		{
			name:    "64 KiB + 1 bytes exceeds cap",
			value:   strings.Repeat("x", 64*1024+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateInput(InputSpec{Shape: ShapeJSON}, tt.value)
			if tt.wantErr {
				require.Error(t, err, "validateInput must return error for %q", tt.name)
				return
			}
			assert.NoError(t, err, "validateInput must return nil for %q", tt.name)
		})
	}
}
