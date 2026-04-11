package bmad

// Unit tests for the mashed asset loader. Covers:
//
//   - Frontmatter extraction at the byte level (happy path, CRLF
//     line endings, unterminated frontmatter, no frontmatter at all).
//   - Per-file parsing (mashed-ready skip, role admission, unknown
//     role rejected, description fallback to first body line,
//     per-role defaults applied).
//   - Directory walking in both layouts (skill directory vs command
//     flat file), including the "nonexistent directory is not an
//     error" contract the sidebar relies on.
//
// All tests use t.TempDir() to build a realistic filesystem — no
// mocks — so the layout expectations and YAML parsing are exercised
// end to end.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── extractFrontmatter ────────────────────────────────────────────────

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   string
		wantBody string
		wantErr  bool
		// wantNoFrontmatter is true when we expect the specific
		// errNoFrontmatter sentinel (not a parse failure).
		wantNoFrontmatter bool
	}{
		{
			name:     "basic happy path",
			input:    "---\nname: foo\n---\nbody text\n",
			wantFM:   "name: foo",
			wantBody: "body text\n",
		},
		{
			name:     "multiple frontmatter fields",
			input:    "---\nname: foo\ndescription: bar\n---\n# Heading\n",
			wantFM:   "name: foo\ndescription: bar",
			wantBody: "# Heading\n",
		},
		{
			name:     "CRLF line endings",
			input:    "---\r\nname: foo\r\n---\r\nbody\r\n",
			wantFM:   "name: foo",
			wantBody: "body\r\n",
		},
		{
			name:     "empty frontmatter block",
			input:    "---\n---\nbody only\n",
			wantFM:   "",
			wantBody: "body only\n",
		},
		{
			name:              "no frontmatter — returns sentinel and full body",
			input:             "just a body, no delimiter\n",
			wantNoFrontmatter: true,
			wantBody:          "just a body, no delimiter\n",
		},
		{
			name:              "opening --- not at byte zero — not frontmatter",
			input:             "\n---\nname: foo\n---\nbody\n",
			wantNoFrontmatter: true,
			wantBody:          "\n---\nname: foo\n---\nbody\n",
		},
		{
			name:    "unterminated frontmatter — error",
			input:   "---\nname: foo\nno closing delim\n",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fm, body, err := extractFrontmatter([]byte(tc.input))

			if tc.wantNoFrontmatter {
				require.ErrorIs(t, err, errNoFrontmatter,
					"expected errNoFrontmatter sentinel")
				assert.Equal(t, tc.wantBody, string(body))
				return
			}
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantFM, string(fm))
			assert.Equal(t, tc.wantBody, string(body))
		})
	}
}

// ── normaliseRole ─────────────────────────────────────────────────────

func TestNormaliseRole(t *testing.T) {
	tests := []struct {
		in     string
		want   MashedAssetRole
		wantOk bool
	}{
		{"command", MashedRoleCommand, true},
		{"skill", MashedRoleSkill, true},
		{"agent", MashedRoleAgent, true},
		{"  COMMAND  ", MashedRoleCommand, true}, // trim + lowercase
		{"Skill", MashedRoleSkill, true},
		{"", "", false},
		{"process", "", false},
		{"Process", "", false},
		{"cmd", "", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%q", tc.in), func(t *testing.T) {
			got, ok := normaliseRole(tc.in)
			assert.Equal(t, tc.wantOk, ok)
			if ok {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// ── applyRoleDefaults ─────────────────────────────────────────────────

func TestApplyRoleDefaults_Command(t *testing.T) {
	info := &MashedAssetInfo{Role: MashedRoleCommand}
	applyRoleDefaults(info)
	assert.Equal(t, "idle", info.Completion)
	assert.Equal(t, "single", info.Chainable)
	assert.False(t, info.SessionPinned)
}

func TestApplyRoleDefaults_Skill(t *testing.T) {
	info := &MashedAssetInfo{Role: MashedRoleSkill}
	applyRoleDefaults(info)
	assert.Equal(t, "none", info.Chainable)
	assert.True(t, info.SessionPinned, "skill defaults to session-pinned")
}

func TestApplyRoleDefaults_AuthorOverridesPreserved(t *testing.T) {
	// An author-provided `mashedCompletion: exit` must not be
	// clobbered by the command-role default of `idle`.
	info := &MashedAssetInfo{
		Role:       MashedRoleCommand,
		Completion: "exit",
		Chainable:  "any",
	}
	applyRoleDefaults(info)
	assert.Equal(t, "exit", info.Completion)
	assert.Equal(t, "any", info.Chainable)
}

// ── firstBodyLine ─────────────────────────────────────────────────────

func TestFirstBodyLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple paragraph", "first line\nsecond line\n", "first line"},
		{"skips leading blanks", "\n\n  first real line  \nnext\n", "first real line"},
		{"skips H1 heading", "# Skill Name\n\nThe real description.\n", "The real description."},
		{"skips multi-level heading", "# Name\n## Section\n\nActual body.\n", "Actual body."},
		{"empty body returns empty string", "", ""},
		{"only whitespace", "   \n\n\t\n", ""},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := firstBodyLine([]byte(tc.in))
			assert.Equal(t, tc.want, got)
		})
	}
}

// ── parseMashedAsset ──────────────────────────────────────────────────

// writeAsset is a test helper that creates a file at `dir/relpath`
// with the given content and returns its absolute path. Parent
// directories are created as needed so tests can express layouts
// without worrying about MkdirAll boilerplate.
func writeAsset(t *testing.T, dir, relpath, content string) string {
	t.Helper()
	path := filepath.Join(dir, relpath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestParseMashedAsset_MashedReady(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "simplify.md", `---
name: simplify
description: "Review recent changes"
mashedRole: command
mashedCompletion: idle
mashedInputs: [foo.md]
mashedOutputs: []
mashedChainable: single
---

# Simplify
`)
	info, err := parseMashedAsset(path, "simplify", MashedKindCommand, MashedSourceGlobal)
	require.NoError(t, err)
	require.NotNil(t, info, "mashed-ready file must parse into a non-nil info")
	assert.Equal(t, "simplify", info.Name)
	assert.Equal(t, path, info.Path)
	assert.Equal(t, "Review recent changes", info.Description)
	assert.Equal(t, MashedKindCommand, info.Kind)
	assert.Equal(t, MashedSourceGlobal, info.Source)
	assert.Equal(t, MashedRoleCommand, info.Role)
	assert.Equal(t, "idle", info.Completion)
	assert.Equal(t, []string{"foo.md"}, info.Inputs)
	assert.Equal(t, []string{}, info.Outputs) // non-nil even when empty
	assert.Equal(t, "single", info.Chainable)
	assert.False(t, info.SessionPinned)
}

func TestParseMashedAsset_MissingMashedRole_SkipsSilently(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "other.md", `---
name: other
description: "Just a plain claude command"
---

Body goes here.
`)
	info, err := parseMashedAsset(path, "other", MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err, "missing mashedRole is not an error")
	assert.Nil(t, info, "missing mashedRole must produce a nil (skipped) info")
}

func TestParseMashedAsset_NoFrontmatter_SkipsSilently(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "no-fm.md", "Just a markdown body.\n")
	info, err := parseMashedAsset(path, "no-fm", MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err, "missing frontmatter is not an error")
	assert.Nil(t, info)
}

func TestParseMashedAsset_UnknownRole_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "weird.md", `---
name: weird
mashedRole: wizard
---
body
`)
	info, err := parseMashedAsset(path, "weird", MashedKindCommand, MashedSourceLocal)
	require.Error(t, err, "unknown role must surface as an error so the caller can log")
	assert.Contains(t, err.Error(), "wizard")
	assert.Nil(t, info)
}

func TestParseMashedAsset_MalformedYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "bad.md", `---
name: bad
mashedRole: command
mashedInputs: [unterminated
---
body
`)
	_, err := parseMashedAsset(path, "bad", MashedKindCommand, MashedSourceLocal)
	require.Error(t, err)
}

func TestParseMashedAsset_DescriptionFallbackToFirstBodyLine(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "dev-story.md", `---
name: dev-story
mashedRole: command
---

# Dev Story

Run the dev-story pipeline against the current branch.
`)
	info, err := parseMashedAsset(path, "dev-story", MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "Run the dev-story pipeline against the current branch.", info.Description)
}

func TestParseMashedAsset_SkillRoleDefaultsSessionPinned(t *testing.T) {
	dir := t.TempDir()
	path := writeAsset(t, dir, "SKILL.md", `---
name: skill-creator
description: "Create, improve and measure skills"
mashedRole: skill
---
body
`)
	info, err := parseMashedAsset(path, "skill-creator", MashedKindSkill, MashedSourceGlobal)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, MashedRoleSkill, info.Role)
	assert.True(t, info.SessionPinned, "skill role defaults to SessionPinned=true")
	assert.Equal(t, "none", info.Chainable, "skill role defaults to Chainable=none")
}

// ── LoadMashedAssetsFromDir (integration with the filesystem walk) ────

func TestLoadMashedAssetsFromDir_NonexistentDir_NotAnError(t *testing.T) {
	// A user who has never created any local skills should see an
	// empty group, not a crash. This is the contract the sidebar
	// depends on for the "empty state" render.
	got, err := LoadMashedAssetsFromDir("/nonexistent/path/that/does/not/exist", MashedKindSkill, MashedSourceLocal)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestLoadMashedAssetsFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadMashedAssetsFromDir(dir, MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestLoadMashedAssetsFromDir_EmptyDirString(t *testing.T) {
	got, err := LoadMashedAssetsFromDir("", MashedKindCommand, MashedSourceLocal)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestLoadMashedAssetsFromDir_CommandLayout_MixedAssets(t *testing.T) {
	dir := t.TempDir()
	// Three mashed-ready commands...
	writeAsset(t, dir, "zzz-last.md", "---\nmashedRole: command\n---\nlast\n")
	writeAsset(t, dir, "aaa-first.md", "---\nmashedRole: command\n---\nfirst\n")
	writeAsset(t, dir, "mmm-middle.md", "---\nmashedRole: command\n---\nmiddle\n")
	// ...plus one command without mashedRole (should be skipped)...
	writeAsset(t, dir, "plain.md", "---\nname: plain\n---\nno mashed fields\n")
	// ...plus one bare markdown file (no frontmatter at all)...
	writeAsset(t, dir, "bare.md", "just text\n")
	// ...plus a stray non-md file that must be ignored entirely.
	writeAsset(t, dir, "README.txt", "not a command")

	got, err := LoadMashedAssetsFromDir(dir, MashedKindCommand, MashedSourceGlobal)
	require.NoError(t, err)
	require.Len(t, got, 3, "only the 3 mashed-ready commands must appear")

	// Sorted alphabetically by Name.
	assert.Equal(t, "aaa-first", got[0].Name)
	assert.Equal(t, "mmm-middle", got[1].Name)
	assert.Equal(t, "zzz-last", got[2].Name)

	// Source is propagated onto every item.
	for _, info := range got {
		assert.Equal(t, MashedSourceGlobal, info.Source)
		assert.Equal(t, MashedKindCommand, info.Kind)
	}
}

func TestLoadMashedAssetsFromDir_SkillLayout_DirectoryPerSkill(t *testing.T) {
	dir := t.TempDir()
	// Mashed-ready skill at skills/alpha/SKILL.md
	writeAsset(t, dir, "alpha/SKILL.md", "---\nname: alpha\nmashedRole: skill\n---\nalpha body\n")
	// Mashed-ready skill at skills/beta/skill.md (lowercase fallback)
	writeAsset(t, dir, "beta/skill.md", "---\nname: beta\nmashedRole: skill\n---\nbeta body\n")
	// Subdirectory WITHOUT a SKILL.md → silently skipped
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "gamma"), 0o755))
	// Subdirectory with SKILL.md but no mashedRole → skipped
	writeAsset(t, dir, "delta/SKILL.md", "---\nname: delta\n---\nno mashed role\n")
	// Stray flat file at the top level must NOT be treated as a skill.
	writeAsset(t, dir, "README.md", "---\nmashedRole: command\n---\ntop-level readme\n")

	got, err := LoadMashedAssetsFromDir(dir, MashedKindSkill, MashedSourceLocal)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "alpha", got[0].Name)
	assert.Equal(t, "beta", got[1].Name)
	// Source + kind threaded through.
	for _, info := range got {
		assert.Equal(t, MashedKindSkill, info.Kind)
		assert.Equal(t, MashedSourceLocal, info.Source)
	}
}

func TestLoadMashedAssetsFromDir_MalformedFile_SkippedNotFatal(t *testing.T) {
	// A broken SKILL.md in one subdirectory must not hide valid
	// siblings. This is the "one bad file does not break the sidebar"
	// contract.
	dir := t.TempDir()
	writeAsset(t, dir, "good/SKILL.md", "---\nname: good\nmashedRole: skill\n---\nok\n")
	writeAsset(t, dir, "bad/SKILL.md", "---\nname: bad\nmashedRole: command\nmashedInputs: [unterminated\n---\n")

	got, err := LoadMashedAssetsFromDir(dir, MashedKindSkill, MashedSourceLocal)
	require.NoError(t, err, "malformed sibling must not propagate as a top-level error")
	require.Len(t, got, 1)
	assert.Equal(t, "good", got[0].Name)
}
