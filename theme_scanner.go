package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

	entries, err := os.ReadDir(extDir)
	if err != nil {
		return nil, fmt.Errorf("reading extensions directory: %w", err)
	}

	var themes []VSCodeThemeEntry

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pkgPath := filepath.Join(extDir, entry.Name(), "package.json")
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			continue // skip extensions without package.json
		}

		var pkg packageJSON
		if err := json.Unmarshal(data, &pkg); err != nil {
			continue // skip corrupt package.json
		}

		for _, t := range pkg.Contributes.Themes {
			// H-7 fix: skip non-JSON theme files (e.g., .tmTheme)
			if !strings.HasSuffix(strings.ToLower(t.Path), ".json") {
				continue
			}

			// Resolve the theme path relative to the extension directory
			themePath := filepath.Join(extDir, entry.Name(), t.Path)
			themePath = filepath.Clean(themePath)

			themes = append(themes, VSCodeThemeEntry{
				Label:       t.Label,
				ExtensionID: entry.Name(),
				ThemePath:   themePath,
				UITheme:     t.UITheme,
			})
		}
	}

	// Sort alphabetically by label
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Label < themes[j].Label
	})

	return themes, nil
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
			// Return the theme without includes on failure
		} else {
			var baseTheme rawTheme
			if err := json.Unmarshal([]byte(baseJSON), &baseTheme); err == nil {
				// Merge: base colors underneath current
				if baseTheme.Colors != nil {
					if theme.Colors == nil {
						theme.Colors = make(map[string]string)
					}
					for k, v := range baseTheme.Colors {
						if _, exists := theme.Colors[k]; !exists {
							theme.Colors[k] = v
						}
					}
				}

				// Prepend base tokenColors before current
				if len(baseTheme.TokenColors) > 0 {
					merged := make([]json.RawMessage, 0, len(baseTheme.TokenColors)+len(theme.TokenColors))
					merged = append(merged, baseTheme.TokenColors...)
					merged = append(merged, theme.TokenColors...)
					theme.TokenColors = merged
				}

				// Inherit name/type from base if not set
				if theme.Name == "" {
					theme.Name = baseTheme.Name
				}
				if theme.Type == "" {
					theme.Type = baseTheme.Type
				}
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

// SetImportedTheme persists the selected imported theme path to config.
func (a *App) SetImportedTheme(themePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := loadConfig()
	cfg.ImportedTheme = themePath
	return saveConfig(cfg)
}
