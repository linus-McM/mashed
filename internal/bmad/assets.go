package bmad

// Mashed asset loader — scans `.claude/skills/` and `.claude/commands/`
// directories for Claude Code assets whose YAML frontmatter includes a
// `mashedRole` field, and exposes them to the frontend as typed
// MashedAssetInfo records grouped by kind × scope.
//
// The filesystem stays authoritative for "which assets exist"; this
// loader is the curated view of "which assets are mashed-ready". Assets
// that lack `mashedRole` are silently skipped so the sidebar never shows
// raw Claude Code files the user has not explicitly opted in to the
// Mashed workspace model.
//
// Two layouts are supported, matching the conventions Claude Code ships
// with today:
//
//   - **Skills** live in `{dir}/<skill-name>/SKILL.md` — one directory
//     per skill, with frontmatter at the top of SKILL.md.
//   - **Commands** live in `{dir}/<command-name>.md` — one flat markdown
//     file per command, with frontmatter at the top.
//
// LoadMashedAssetsFromDir handles both layouts via a `layout` parameter
// so callers need only one entry point per directory they want scanned.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// MashedAssetKind discriminates between the two filesystem layouts and
// the two node behaviours: commands chain inside workflows, skills pin
// to agent-owned sessions. The kind is inferred from the enclosing
// directory (`.claude/skills/` vs `.claude/commands/`), not from the
// frontmatter — the frontmatter's `mashedRole` describes the BMAD
// execution role, which may or may not match the filesystem layout.
type MashedAssetKind string

const (
	// MashedKindSkill — directory-per-skill layout (`<dir>/<name>/SKILL.md`).
	MashedKindSkill MashedAssetKind = "skill"
	// MashedKindCommand — flat-file layout (`<dir>/<name>.md`).
	MashedKindCommand MashedAssetKind = "command"
)

// MashedAssetSource records where the asset file lives on disk so the
// UI can label local vs global without re-parsing paths.
type MashedAssetSource string

const (
	MashedSourceLocal  MashedAssetSource = "local"  // {repoPath}/.claude/...
	MashedSourceGlobal MashedAssetSource = "global" // ~/.claude/...
)

// MashedAssetInfo is the typed record the frontend consumes. Every
// field mirrors the YAML schema described in docs/plans/skills-and-
// session-reuse.md, plus a few derived fields (Path, Kind, Source) so
// the sidebar does not need to re-walk the filesystem.
//
// All list-typed fields are non-nil slices when the asset is valid (so
// JavaScript callers can `.map` without null checks). They may be
// zero-length when the schema field was absent or empty.
type MashedAssetInfo struct {
	// Name is the skill/command identifier (directory name for skills,
	// filename-without-extension for commands). Not necessarily unique
	// across Source+Kind.
	Name string `json:"name"`
	// Path is the absolute filesystem path to the asset's markdown
	// file. Used by the refactor skill and by future editors.
	Path string `json:"path"`
	// Description is either the YAML `description:` field or, when
	// absent, the first non-empty body line after the frontmatter
	// block. May be empty if both are missing.
	Description string `json:"description"`
	// Kind is the filesystem layout (skill directory vs command file)
	// and determines the default execution behaviour.
	Kind MashedAssetKind `json:"kind"`
	// Source is the scope (local to a repo vs globally available).
	Source MashedAssetSource `json:"source"`

	// Schema fields — these mirror `mashedRole` etc. from the YAML
	// frontmatter. A present `mashedRole` is the admission gate for
	// this asset appearing in the listing at all; the rest inherit
	// per-role defaults when omitted.
	Role          MashedAssetRole `json:"role"`
	Completion    string          `json:"completion"`
	Inputs        []string        `json:"inputs"`
	Outputs       []string        `json:"outputs"`
	Chainable     string          `json:"chainable"`
	SessionPinned bool            `json:"sessionPinned"`
}

// MashedAssetRole is the workflow-execution role declared in the asset's
// frontmatter. Only three values are recognised; anything else causes
// the asset to be skipped with a warning rather than misclassified.
type MashedAssetRole string

const (
	MashedRoleCommand MashedAssetRole = "command"
	MashedRoleSkill   MashedAssetRole = "skill"
	MashedRoleAgent   MashedAssetRole = "agent"
)

// GroupedMashedAssets is the shape the WorkflowBuilder sidebar renders:
// four collapsible groups (Local/Global × Commands/Skills). Each group
// is a non-nil slice so the Svelte `{#each}` blocks don't need null
// guards.
type GroupedMashedAssets struct {
	LocalCommands  []MashedAssetInfo `json:"localCommands"`
	GlobalCommands []MashedAssetInfo `json:"globalCommands"`
	LocalSkills    []MashedAssetInfo `json:"localSkills"`
	GlobalSkills   []MashedAssetInfo `json:"globalSkills"`
}

// mashedAssetFrontmatter is the intermediate struct yaml.Unmarshal
// populates from the YAML block. Using a dedicated struct instead of
// unmarshalling into MashedAssetInfo keeps the file-format concerns
// decoupled from the in-memory shape and lets us handle the
// "any mashed* field present = promoted" rule explicitly in the caller.
//
// Unknown YAML keys are silently ignored (the yaml.v3 default), so
// authors can add their own frontmatter without tripping the parser.
type mashedAssetFrontmatter struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	MashedRole    string   `yaml:"mashedRole"`
	Completion    string   `yaml:"mashedCompletion"`
	Inputs        []string `yaml:"mashedInputs"`
	Outputs       []string `yaml:"mashedOutputs"`
	Chainable     string   `yaml:"mashedChainable"`
	SessionPinned bool     `yaml:"mashedSessionPinned"`
}

// errNoFrontmatter is returned by extractFrontmatter when the file does
// not begin with a `---\n` delimiter. It is a signal, not a failure —
// callers use it to distinguish "no frontmatter" from "malformed YAML".
var errNoFrontmatter = errors.New("mashed: no YAML frontmatter block")

// extractFrontmatter reads the YAML frontmatter block from a markdown
// file's bytes. Returns the raw YAML bytes, the body bytes (everything
// after the second `---` delimiter), and an error.
//
// Contract:
//   - A valid frontmatter block MUST start at byte 0 with `---\n` (or
//     `---\r\n`), then contain any number of YAML lines, then a closing
//     `---` on its own line.
//   - No frontmatter → returns (nil, full-content, errNoFrontmatter).
//   - Unterminated frontmatter → returns an error.
//
// Implemented as a byte-level search (not bufio.Scanner) so that the
// body's original line-ending encoding is preserved exactly — a scanner
// loses the distinction between LF and CRLF, which would corrupt the
// body offset calculation on files with mixed or CRLF line endings.
// The delimiters themselves are NOT valid YAML and must be stripped
// before the content reaches yaml.Unmarshal.
func extractFrontmatter(content []byte) (fm, body []byte, err error) {
	// Accept `---` only at the very start of the file, optionally
	// followed by CR and then LF. Anything else = no frontmatter.
	var openLen int
	switch {
	case bytes.HasPrefix(content, []byte("---\n")):
		openLen = 4
	case bytes.HasPrefix(content, []byte("---\r\n")):
		openLen = 5
	default:
		return nil, content, errNoFrontmatter
	}

	// Everything after the opening delimiter line — including potential
	// empty-frontmatter case where the closing delimiter is the first
	// thing we see.
	rest := content[openLen:]

	// Empty-frontmatter fast path: `---\n---\n...` or `---\n---\r\n...`
	// produce an empty yaml block and a body starting after the closing
	// delimiter.
	if bytes.HasPrefix(rest, []byte("---\n")) {
		return []byte{}, rest[4:], nil
	}
	if bytes.HasPrefix(rest, []byte("---\r\n")) {
		return []byte{}, rest[5:], nil
	}

	// Find the closing delimiter. It must appear AFTER a newline and
	// be followed by its own newline (so we don't match a `---` inside
	// a YAML value or a body line that coincidentally starts with
	// three dashes). Search for both LF and CRLF variants and pick the
	// earliest match.
	//
	// The CRLF needle includes the FULL preceding `\r\n` so the fm
	// slice stops cleanly at `foo` instead of `foo\r`. Same for LF:
	// the needle starts with `\n` so fm stops at `foo` without the
	// terminating newline. This keeps the returned YAML bytes free
	// of trailing whitespace artifacts.
	needleLF := []byte("\n---\n")
	needleCRLF := []byte("\r\n---\r\n")
	idxLF := bytes.Index(rest, needleLF)
	idxCRLF := bytes.Index(rest, needleCRLF)

	var endIdx, needleLen int
	switch {
	case idxLF < 0 && idxCRLF < 0:
		return nil, nil, errors.New("mashed: unterminated frontmatter block")
	case idxLF < 0:
		endIdx, needleLen = idxCRLF, len(needleCRLF)
	case idxCRLF < 0:
		endIdx, needleLen = idxLF, len(needleLF)
	case idxLF <= idxCRLF:
		endIdx, needleLen = idxLF, len(needleLF)
	default:
		endIdx, needleLen = idxCRLF, len(needleCRLF)
	}

	// fm is everything from the start of `rest` up to (but not
	// including) the `\n` that introduces the closing delimiter.
	// body is everything after the closing delimiter line.
	fm = rest[:endIdx]
	body = rest[endIdx+needleLen:]
	return fm, body, nil
}

// parseMashedAsset parses a single markdown file into a MashedAssetInfo,
// applying the "mashedRole present = mashed-ready" filter. Returns
// (nil, nil) when the asset is valid but not mashed-ready (caller
// silently skips). Returns (nil, err) for actual parse errors the
// caller may want to log or surface.
//
// The Name, Kind, and Source fields are passed in by the caller because
// they derive from the filesystem walk, not from the frontmatter.
func parseMashedAsset(path string, name string, kind MashedAssetKind, source MashedAssetSource) (*MashedAssetInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("mashed: reading %s: %w", path, err)
	}

	fmBytes, body, err := extractFrontmatter(data)
	if errors.Is(err, errNoFrontmatter) {
		// No frontmatter at all → not mashed-ready, silently skip.
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var raw mashedAssetFrontmatter
	if err := yaml.Unmarshal(fmBytes, &raw); err != nil {
		return nil, fmt.Errorf("mashed: parsing frontmatter in %s: %w", path, err)
	}

	// The admission gate: no mashedRole → not mashed-ready.
	if raw.MashedRole == "" {
		return nil, nil
	}

	role, ok := normaliseRole(raw.MashedRole)
	if !ok {
		// Unknown role → skip with a warning. The caller logs.
		return nil, fmt.Errorf("mashed: unknown mashedRole %q in %s", raw.MashedRole, path)
	}

	info := &MashedAssetInfo{
		Name:        name,
		Path:        path,
		Description: strings.TrimSpace(raw.Description),
		Kind:        kind,
		Source:      source,
		Role:        role,
		Completion:  strings.TrimSpace(raw.Completion),
		Inputs:      nonNilStrings(raw.Inputs),
		Outputs:     nonNilStrings(raw.Outputs),
		Chainable:   strings.TrimSpace(raw.Chainable),
	}
	info.SessionPinned = raw.SessionPinned
	applyRoleDefaults(info)

	// Description fallback: first non-empty body line if frontmatter
	// omits `description:`.
	if info.Description == "" {
		info.Description = firstBodyLine(body)
	}

	return info, nil
}

// normaliseRole maps a raw YAML role string to the typed enum. Returns
// (_, false) for any value outside the three recognised roles so the
// caller can treat it as an unknown-role error.
func normaliseRole(raw string) (MashedAssetRole, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(MashedRoleCommand):
		return MashedRoleCommand, true
	case string(MashedRoleSkill):
		return MashedRoleSkill, true
	case string(MashedRoleAgent):
		return MashedRoleAgent, true
	}
	return "", false
}

// applyRoleDefaults fills in per-role defaults for fields the author
// omitted. Called AFTER parseMashedAsset has copied raw values so
// explicit author choices always win.
func applyRoleDefaults(info *MashedAssetInfo) {
	switch info.Role {
	case MashedRoleCommand:
		if info.Completion == "" {
			info.Completion = "idle"
		}
		if info.Chainable == "" {
			info.Chainable = "single"
		}
		// SessionPinned defaults to false (Go zero value).
	case MashedRoleSkill:
		if info.Chainable == "" {
			info.Chainable = "none"
		}
		// Skills are pinned by default; only an explicit `false` in
		// YAML would override. yaml.v3 can't distinguish missing from
		// explicit-false for bool, so we promote-to-true here to match
		// the documented default.
		if !info.SessionPinned {
			info.SessionPinned = true
		}
	case MashedRoleAgent:
		if info.Chainable == "" {
			info.Chainable = "none"
		}
	}
}

// firstBodyLine returns the first non-empty, non-heading line of the
// markdown body, trimmed, as a description fallback when frontmatter
// omits `description:`. Skips heading lines (`# ...`) because they are
// almost always just the asset's own name repeated.
func firstBodyLine(body []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

// nonNilStrings returns an empty slice instead of nil so JSON-encoded
// output is `[]` instead of `null` — the frontend can then `.map` it
// without an extra guard.
func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// LoadMashedAssetsFromDir walks `dir` and returns every mashed-ready
// asset it finds, sorted by Name. The `kind` parameter selects the
// filesystem layout:
//
//   - MashedKindSkill expects one directory per skill with a SKILL.md
//     (or skill.md) file inside.
//   - MashedKindCommand expects flat `<name>.md` files directly inside
//     `dir`.
//
// The `source` parameter is propagated into every returned info so the
// caller (ListAllMashedAssets) can tell Local and Global apart.
//
// Nonexistent directories return (nil, nil) — NOT an error. A user who
// has never created any local skills should see an empty Local group,
// not a failure toast.
//
// Parse errors for individual files are logged via the standard logger
// and the bad file is skipped. This is intentional: one malformed
// SKILL.md must not hide every OTHER valid asset in the same directory.
func LoadMashedAssetsFromDir(dir string, kind MashedAssetKind, source MashedAssetSource) ([]MashedAssetInfo, error) {
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("mashed: reading %s: %w", dir, err)
	}

	var out []MashedAssetInfo
	for _, entry := range entries {
		info, perr := loadOneEntry(dir, entry, kind, source)
		if perr != nil {
			logMashedLoadWarning(dir, entry.Name(), perr)
			continue
		}
		if info == nil {
			continue
		}
		out = append(out, *info)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// loadOneEntry dispatches a single directory entry to the right parser
// based on the expected layout. Returns (nil, nil) when the entry is
// valid but not mashed-ready, (nil, err) for parse errors, and
// (info, nil) for successful loads.
func loadOneEntry(dir string, entry os.DirEntry, kind MashedAssetKind, source MashedAssetSource) (*MashedAssetInfo, error) {
	switch kind {
	case MashedKindSkill:
		if !entry.IsDir() {
			return nil, nil
		}
		// Accept SKILL.md or skill.md inside the subdirectory; prefer
		// SKILL.md because that is the Claude Code convention.
		subdir := filepath.Join(dir, entry.Name())
		for _, candidate := range []string{"SKILL.md", "skill.md"} {
			path := filepath.Join(subdir, candidate)
			if _, err := os.Stat(path); err == nil {
				return parseMashedAsset(path, entry.Name(), kind, source)
			}
		}
		return nil, nil // subdirectory without SKILL.md → silently skip

	case MashedKindCommand:
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil, nil
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		path := filepath.Join(dir, entry.Name())
		return parseMashedAsset(path, name, kind, source)
	}
	return nil, nil
}

// logMashedLoadWarning is a thin wrapper that emits a single log line
// for a skipped file. Isolated so tests can swap it with a capture
// helper via linker tricks or via the standard log package's own
// redirection hooks. For now it just uses the default logger.
func logMashedLoadWarning(dir, name string, err error) {
	fmt.Fprintf(os.Stderr, "mashed: skipping %s in %s: %v\n", name, dir, err)
}
