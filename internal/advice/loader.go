package advice

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed all:defaults
var defaultsFS embed.FS

// adviceFrontmatter holds the YAML frontmatter fields parsed from an advice file.
type adviceFrontmatter struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"displayName"`
	Icon        string `yaml:"icon"`
	Order       int    `yaml:"order"`
}

// LoadAdviceModes discovers and merges advice modes from 3 sources:
// bundled defaults (embedded), global (~/.mashed/advice/), and local ({repoPath}/.claude/advice/).
// Local overrides global, global overrides bundled. Sorted by Order asc, then DisplayName.
func LoadAdviceModes(repoPath string) ([]AdviceMode, error) {
	gDir, err := globalAdviceDir()
	if err != nil {
		return nil, fmt.Errorf("resolving global advice dir: %w", err)
	}

	var lDir string
	if repoPath != "" {
		lDir = localAdviceDir(repoPath)
	}

	return loadAdviceModesFromDirs(&defaultsFS, gDir, lDir)
}

// loadAdviceModesFromDirs is the internal implementation that accepts explicit directories
// for testability. bundledFS may be nil if no embedded defaults are available.
func loadAdviceModesFromDirs(bundledFS *embed.FS, globalDir, localDir string) ([]AdviceMode, error) {
	modeMap := make(map[string]AdviceMode)

	// 1. Bundled defaults (lowest priority)
	if bundledFS != nil {
		entries, err := bundledFS.ReadDir("defaults")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				mode, parseErr := parseAdviceFileFromFS(*bundledFS, filepath.Join("defaults", e.Name()))
				if parseErr != nil {
					continue // skip malformed
				}
				mode.Source = "bundled"
				modeMap[mode.Name] = mode
			}
		}
	}

	// 2. Global (medium priority)
	if globalDir != "" {
		modes, err := loadModesFromDir(globalDir, "global")
		if err == nil {
			for _, m := range modes {
				modeMap[m.Name] = m
			}
		}
	}

	// 3. Local (highest priority)
	if localDir != "" {
		modes, err := loadModesFromDir(localDir, "local")
		if err == nil {
			for _, m := range modes {
				modeMap[m.Name] = m
			}
		}
	}

	// Collect and sort
	result := make([]AdviceMode, 0, len(modeMap))
	for _, m := range modeMap {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Order != result[j].Order {
			return result[i].Order < result[j].Order
		}
		return result[i].DisplayName < result[j].DisplayName
	})

	return result, nil
}

// loadModesFromDir reads all .md files from a directory, parsing each as an advice file.
// Returns nil, nil if the directory doesn't exist.
func loadModesFromDir(dir, source string) ([]AdviceMode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var modes []AdviceMode
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		mode, parseErr := parseAdviceFile(filepath.Join(dir, e.Name()), source)
		if parseErr != nil {
			continue // skip malformed files
		}
		modes = append(modes, mode)
	}
	return modes, nil
}

// LoadAdviceBody finds the highest-priority file for the named mode and returns its body.
func LoadAdviceBody(repoPath, modeName string) (string, error) {
	gDir, err := globalAdviceDir()
	if err != nil {
		return "", fmt.Errorf("resolving global advice dir: %w", err)
	}

	var lDir string
	if repoPath != "" {
		lDir = localAdviceDir(repoPath)
	}

	return loadAdviceBodyFromDirs(modeName, &defaultsFS, gDir, lDir)
}

// loadAdviceBodyFromDirs is the internal implementation for LoadAdviceBody.
func loadAdviceBodyFromDirs(modeName string, bundledFS *embed.FS, globalDir, localDir string) (string, error) {
	// Check in priority order: local > global > bundled
	// Return the first match found.

	// Local (highest priority)
	if localDir != "" {
		mode, err := findModeInDir(localDir, modeName, "local")
		if err == nil {
			return mode.Body, nil
		}
	}

	// Global
	if globalDir != "" {
		mode, err := findModeInDir(globalDir, modeName, "global")
		if err == nil {
			return mode.Body, nil
		}
	}

	// Bundled (lowest priority)
	if bundledFS != nil {
		entries, err := bundledFS.ReadDir("defaults")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				mode, parseErr := parseAdviceFileFromFS(*bundledFS, filepath.Join("defaults", e.Name()))
				if parseErr != nil {
					continue
				}
				if mode.Name == modeName {
					return mode.Body, nil
				}
			}
		}
	}

	return "", fmt.Errorf("advice mode %q not found", modeName)
}

// findModeInDir searches a directory for an advice file matching the given name.
func findModeInDir(dir, modeName, source string) (AdviceMode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return AdviceMode{}, err
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		mode, parseErr := parseAdviceFile(filepath.Join(dir, e.Name()), source)
		if parseErr != nil {
			continue
		}
		if mode.Name == modeName {
			return mode, nil
		}
	}

	return AdviceMode{}, fmt.Errorf("advice mode %q not found in %s", modeName, dir)
}

// parseAdviceFile reads and parses a .md advice file from disk.
func parseAdviceFile(path string, source string) (AdviceMode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AdviceMode{}, fmt.Errorf("reading advice file %s: %w", path, err)
	}

	mode, err := parseAdviceContent(string(data), source)
	if err != nil {
		return AdviceMode{}, fmt.Errorf("parsing advice file %s: %w", path, err)
	}
	mode.FilePath = path
	return mode, nil
}

// parseAdviceFileFromFS reads and parses a .md advice file from an embed.FS.
func parseAdviceFileFromFS(fsys embed.FS, path string) (AdviceMode, error) {
	data, err := fsys.ReadFile(path)
	if err != nil {
		return AdviceMode{}, fmt.Errorf("reading embedded advice file %s: %w", path, err)
	}

	mode, err := parseAdviceContent(string(data), "bundled")
	if err != nil {
		return AdviceMode{}, fmt.Errorf("parsing embedded advice file %s: %w", path, err)
	}
	mode.FilePath = path
	return mode, nil
}

// parseAdviceContent parses YAML frontmatter and body from advice file content.
func parseAdviceContent(content, source string) (AdviceMode, error) {
	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return AdviceMode{}, err
	}

	var meta adviceFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return AdviceMode{}, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	if meta.Name == "" {
		return AdviceMode{}, fmt.Errorf("advice file missing required 'name' field")
	}

	displayName := meta.DisplayName
	if displayName == "" {
		displayName = meta.Name
	}

	return AdviceMode{
		Name:        meta.Name,
		DisplayName: displayName,
		Icon:        meta.Icon,
		Order:       meta.Order,
		Body:        body,
		Source:      source,
	}, nil
}

// splitFrontmatter splits content on the first two "---" delimiters.
// Returns the YAML frontmatter and the remaining body.
func splitFrontmatter(content string) (string, string, error) {
	trimmed := strings.TrimLeft(content, " \t")

	if !strings.HasPrefix(trimmed, "---") {
		return "", "", fmt.Errorf("no frontmatter found: file must start with '---'")
	}

	// Find the closing "---"
	rest := trimmed[3:] // skip opening "---"
	// Trim the newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", fmt.Errorf("no closing frontmatter delimiter '---' found")
	}

	frontmatter := rest[:idx]
	body := rest[idx+4:] // skip "\n---"

	// Trim the newline after closing ---
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	} else if len(body) > 1 && body[0] == '\r' && body[1] == '\n' {
		body = body[2:]
	}

	return frontmatter, body, nil
}

// globalAdviceDir returns the path to ~/.mashed/advice/.
func globalAdviceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".mashed", "advice"), nil
}

// localAdviceDir returns the path to {repoPath}/.claude/advice/.
func localAdviceDir(repoPath string) string {
	return filepath.Join(repoPath, ".claude", "advice")
}
