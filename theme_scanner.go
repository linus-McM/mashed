package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const vsixSeparator = "::vsix::"

// isVSIXThemePath returns true if the theme path encodes a file inside a .vsix archive.
func isVSIXThemePath(p string) bool {
	return strings.Contains(p, vsixSeparator)
}

// parseVSIXThemePath splits a VSIX-encoded theme path into the .vsix file path
// and the zip-internal path.
func parseVSIXThemePath(p string) (vsixPath, internalPath string, ok bool) {
	parts := strings.SplitN(p, vsixSeparator, 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// makeVSIXThemePath encodes a .vsix file path and zip-internal path into a single string.
func makeVSIXThemePath(vsixPath, internalPath string) string {
	return vsixPath + vsixSeparator + internalPath
}

// readFileFromZip reads a named file from a zip archive, enforcing a 512KB size limit.
func readFileFromZip(zr *zip.ReadCloser, name string) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name == name {
			if f.UncompressedSize64 > 512*1024 {
				return nil, fmt.Errorf("file too large (%d bytes, max 512KB)", f.UncompressedSize64)
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file %q not found in archive", name)
}

// mergeThemes merges a base theme into a child theme. Base colors go underneath
// (child wins on conflict), base tokenColors are prepended, and name/type are
// inherited from base if the child's are empty.
func mergeThemes(child, base *rawTheme) {
	if base.Colors != nil {
		if child.Colors == nil {
			child.Colors = make(map[string]string)
		}
		for k, v := range base.Colors {
			if _, exists := child.Colors[k]; !exists {
				child.Colors[k] = v
			}
		}
	}
	if len(base.TokenColors) > 0 {
		merged := make([]json.RawMessage, 0, len(base.TokenColors)+len(child.TokenColors))
		merged = append(merged, base.TokenColors...)
		merged = append(merged, child.TokenColors...)
		child.TokenColors = merged
	}
	if child.Name == "" {
		child.Name = base.Name
	}
	if child.Type == "" {
		child.Type = base.Type
	}
}

// expandTilde expands ~ or ~/ prefix to the user's home directory.
// It does NOT expand ~otheruser paths.
func expandTilde(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// stripJSONC removes // line comments, /* */ block comments (not inside string
// literals), and trailing commas before } or ] from JSONC-formatted data.
func stripJSONC(data []byte) []byte {
	out := make([]byte, 0, len(data))
	i := 0
	n := len(data)

	for i < n {
		// Inside a string literal — copy verbatim until closing quote
		if data[i] == '"' {
			out = append(out, data[i])
			i++
			for i < n {
				if data[i] == '\\' && i+1 < n {
					out = append(out, data[i], data[i+1])
					i += 2
					continue
				}
				if data[i] == '"' {
					out = append(out, data[i])
					i++
					break
				}
				out = append(out, data[i])
				i++
			}
			continue
		}

		// Line comment: // to end of line
		if i+1 < n && data[i] == '/' && data[i+1] == '/' {
			// Skip until newline (don't consume the newline itself)
			i += 2
			for i < n && data[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment: /* ... */
		if i+1 < n && data[i] == '/' && data[i+1] == '*' {
			i += 2
			for i+1 < n {
				if data[i] == '*' && data[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}

		out = append(out, data[i])
		i++
	}

	// Strip trailing commas before } or ]
	out = stripTrailingCommas(out)
	return out
}

// stripTrailingCommas removes commas that appear before } or ] (with optional whitespace between).
func stripTrailingCommas(data []byte) []byte {
	result := make([]byte, 0, len(data))
	n := len(data)

	for i := 0; i < n; i++ {
		if data[i] == ',' {
			// Look ahead past whitespace to see if next non-whitespace is } or ]
			j := i + 1
			for j < n && (data[j] == ' ' || data[j] == '\t' || data[j] == '\n' || data[j] == '\r') {
				j++
			}
			if j < n && (data[j] == '}' || data[j] == ']') {
				// This is a trailing comma — strip it and any inline whitespace before it
				// (trim trailing spaces/tabs from the result buffer)
				for len(result) > 0 && (result[len(result)-1] == ' ' || result[len(result)-1] == '\t') {
					result = result[:len(result)-1]
				}
				continue
			}
		}
		result = append(result, data[i])
	}
	return result
}

// packageJSON is the minimal structure we parse from extension package.json files.
type packageJSON struct {
	Name        string `json:"name"`
	Contributes struct {
		Themes []struct {
			Label   string `json:"label"`
			UITheme string `json:"uiTheme"`
			Path    string `json:"path"`
		} `json:"themes"`
	} `json:"contributes"`
}

// scanVSIXDirectory scans a directory for .vsix files and returns all color
// theme entries found inside them, sorted alphabetically by label.
func scanVSIXDirectory(dir string) ([]VSCodeThemeEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var themes []VSCodeThemeEntry

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".vsix") {
			continue
		}

		vsixPath := filepath.Join(dir, entry.Name())
		extensionID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		zr, err := zip.OpenReader(vsixPath)
		if err != nil {
			log.Printf("skipping corrupt vsix %s: %v", entry.Name(), err)
			continue
		}

		pkgData, err := readFileFromZip(zr, "extension/package.json")
		zr.Close()
		if err != nil {
			continue // no package.json, skip
		}

		var pkg packageJSON
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			continue // corrupt package.json
		}

		for _, t := range pkg.Contributes.Themes {
			if !strings.HasSuffix(strings.ToLower(t.Path), ".json") {
				continue
			}

			// Normalize: "./themes/dark.json" -> "extension/themes/dark.json"
			internalPath := strings.TrimPrefix(t.Path, "./")
			internalPath = "extension/" + internalPath

			themePath := makeVSIXThemePath(vsixPath, internalPath)

			themes = append(themes, VSCodeThemeEntry{
				Label:       t.Label,
				ExtensionID: extensionID,
				ThemePath:   themePath,
				UITheme:     t.UITheme,
			})
		}
	}

	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Label < themes[j].Label
	})

	return themes, nil
}

// ListVSCodiumThemes scans the configured VSCodium extension directory for
// installed color themes and returns them sorted alphabetically by label.
func (a *App) ListVSCodiumThemes() ([]VSCodeThemeEntry, error) {
	cfg := loadConfig()
	extDir := cfg.VSCodiumExtPath
	if extDir == "" {
		return nil, fmt.Errorf("VSCodium extension path not configured")
	}

	// Expand tilde
	extDir = expandTilde(extDir)

	// Verify directory is accessible
	info, err := os.Stat(extDir)
	if err != nil {
		return nil, fmt.Errorf("extensions directory not accessible: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("extensions path is not a directory: %s", extDir)
	}

	return scanVSIXDirectory(extDir)
}

// bundledThemesDir returns the absolute path to the bundled themes directory.
// It checks next to the executable first (production), then falls back to CWD
// (dev mode). Returns empty string if neither location has a themes directory.
func bundledThemesDir() string {
	// Check relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "themes")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// Fall back to CWD
	if wd, err := os.Getwd(); err == nil {
		dir := filepath.Join(wd, "themes")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return ""
}

// ListBundledThemes scans the bundled themes directory for .vsix files and
// returns all color theme entries found. Returns empty slice (not error) if the
// themes directory doesn't exist.
func (a *App) ListBundledThemes() ([]VSCodeThemeEntry, error) {
	dir := bundledThemesDir()
	if dir == "" {
		return []VSCodeThemeEntry{}, nil
	}
	return scanVSIXDirectory(dir)
}

// ReadBundledThemeFile reads a theme file from a bundled VSIX archive.
// The themePath must be a VSIX-encoded path (e.g. /path/file.vsix::vsix::extension/themes/dark.json).
// Unlike ReadThemeFile, this does not require VSCodiumExtPath to be configured.
func (a *App) ReadBundledThemeFile(themePath string) (string, error) {
	if themePath == "" {
		return "", fmt.Errorf("empty theme path")
	}

	vsixPath, internalPath, ok := parseVSIXThemePath(themePath)
	if !ok {
		return "", fmt.Errorf("invalid vsix theme path: %s", themePath)
	}

	return readAndResolveVSIXTheme(vsixPath, internalPath, 0)
}

// readAndResolveVSIXTheme reads a theme file from inside a .vsix zip archive,
// strips JSONC comments, resolves include directives within the zip (up to
// depth 5), and returns clean JSON. This is the shared core used by both
// readThemeFromVSIX (with security check) and ReadBundledThemeFile (without).
func readAndResolveVSIXTheme(vsixPath, internalPath string, depth int) (string, error) {
	if depth > 5 {
		log.Printf("theme include depth limit reached (>5) for %s in %s", internalPath, vsixPath)
		return "", fmt.Errorf("include depth limit exceeded")
	}

	zr, err := zip.OpenReader(vsixPath)
	if err != nil {
		return "", fmt.Errorf("opening vsix: %w", err)
	}
	defer zr.Close()

	data, err := readFileFromZip(zr, internalPath)
	if err != nil {
		return "", fmt.Errorf("reading %s from vsix: %w", internalPath, err)
	}

	clean := stripJSONC(data)

	var theme rawTheme
	if err := json.Unmarshal(clean, &theme); err != nil {
		return string(clean), nil
	}

	if theme.Include != "" {
		// Use path (not filepath) since zip entries use forward slashes.
		includeDir := path.Dir(internalPath)
		includePath := path.Join(includeDir, theme.Include)
		includePath = path.Clean(includePath)

		baseJSON, err := readAndResolveVSIXTheme(vsixPath, includePath, depth+1)
		if err != nil {
			log.Printf("failed to resolve include %q in vsix: %v", theme.Include, err)
		} else {
			var baseTheme rawTheme
			if err := json.Unmarshal([]byte(baseJSON), &baseTheme); err == nil {
				mergeThemes(&theme, &baseTheme)
			}
		}
	}

	theme.Include = ""
	result, err := json.Marshal(theme)
	if err != nil {
		return string(clean), nil
	}
	return string(result), nil
}

// rawTheme is the intermediate representation used for include resolution.
type rawTheme struct {
	Include     string                 `json:"include"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Colors      map[string]string      `json:"colors"`
	TokenColors []json.RawMessage      `json:"tokenColors"`
}

// ReadThemeFile reads a theme file within the configured extensions directory,
// strips JSONC comments, resolves include directives (up to depth 5), and
// returns a clean JSON string.
func (a *App) ReadThemeFile(themePath string) (string, error) {
	cfg := loadConfig()
	extDir := cfg.VSCodiumExtPath
	if extDir == "" {
		return "", fmt.Errorf("VSCodium extension path not configured")
	}
	extDir = expandTilde(extDir)

	if isVSIXThemePath(themePath) {
		vsixPath, internalPath, ok := parseVSIXThemePath(themePath)
		if !ok {
			return "", fmt.Errorf("invalid vsix theme path: %s", themePath)
		}
		return a.readThemeFromVSIX(vsixPath, internalPath, extDir, 0)
	}

	return a.readThemeFileWithDepth(themePath, extDir, 0)
}

func (a *App) readThemeFileWithDepth(themePath, extDir string, depth int) (string, error) {
	if depth > 5 {
		log.Printf("theme include depth limit reached (>5) for %s", themePath)
		return "", fmt.Errorf("include depth limit exceeded")
	}

	// C-1 fix: resolve symlinks with EvalSymlinks
	absTheme, err := filepath.EvalSymlinks(themePath)
	if err != nil {
		return "", fmt.Errorf("resolving theme path: %w", err)
	}
	absExt, err := filepath.EvalSymlinks(extDir)
	if err != nil {
		return "", fmt.Errorf("resolving extensions path: %w", err)
	}

	// C-1 fix: append os.PathSeparator to prevent prefix confusion
	if !strings.HasPrefix(absTheme+string(os.PathSeparator), absExt+string(os.PathSeparator)) {
		return "", fmt.Errorf("theme path outside extensions directory")
	}

	// C-4 fix: stat before read to check size
	info, err := os.Stat(absTheme)
	if err != nil {
		return "", fmt.Errorf("stat theme file: %w", err)
	}
	if info.Size() > 512*1024 {
		return "", fmt.Errorf("theme file too large (%d bytes, max 512KB)", info.Size())
	}

	data, err := os.ReadFile(absTheme)
	if err != nil {
		return "", fmt.Errorf("reading theme file: %w", err)
	}

	// Strip JSONC comments
	clean := stripJSONC(data)

	// Parse as rawTheme to check for includes
	var theme rawTheme
	if err := json.Unmarshal(clean, &theme); err != nil {
		// If it doesn't parse as structured theme, just return the stripped JSON
		return string(clean), nil
	}

	// H-1 fix: resolve includes
	if theme.Include != "" {
		// Resolve relative to the current theme file's directory
		includeDir := filepath.Dir(absTheme)
		includePath := filepath.Join(includeDir, theme.Include)

		baseJSON, err := a.readThemeFileWithDepth(includePath, extDir, depth+1)
		if err != nil {
			log.Printf("failed to resolve include %q: %v", theme.Include, err)
		} else {
			var baseTheme rawTheme
			if err := json.Unmarshal([]byte(baseJSON), &baseTheme); err == nil {
				mergeThemes(&theme, &baseTheme)
			}
		}
	}

	// Clear the include field from output
	theme.Include = ""

	// Re-serialize as clean JSON
	result, err := json.Marshal(theme)
	if err != nil {
		return string(clean), nil
	}
	return string(result), nil
}

// readThemeFromVSIX validates that vsixPath is inside extDir, then delegates
// to readAndResolveVSIXTheme for the actual reading and include resolution.
func (a *App) readThemeFromVSIX(vsixPath, internalPath, extDir string, depth int) (string, error) {
	// Security: verify vsix is inside the extensions directory
	absVsix, err := filepath.EvalSymlinks(vsixPath)
	if err != nil {
		return "", fmt.Errorf("resolving vsix path: %w", err)
	}
	absExt, err := filepath.EvalSymlinks(extDir)
	if err != nil {
		return "", fmt.Errorf("resolving extensions path: %w", err)
	}
	if !strings.HasPrefix(absVsix, absExt+string(os.PathSeparator)) {
		return "", fmt.Errorf("vsix path outside extensions directory")
	}

	return readAndResolveVSIXTheme(absVsix, internalPath, depth)
}

// SetImportedTheme persists the selected imported theme path to config.
func (a *App) SetImportedTheme(themePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.ImportedTheme = themePath
	return saveConfig(cfg)
}
