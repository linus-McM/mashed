package main

import (
	"archive/zip"
	"encoding/base64"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// NerdFontEntry represents a detected Nerd Font family on the system.
type NerdFontEntry struct {
	Family   string `json:"family"`
	FilePath string `json:"filePath"`
}

// knownNerdFonts maps filename prefixes (no spaces, no weight suffix) to the
// correct CSS font-family name. Nerd Fonts v3 uses this naming convention.
var knownNerdFonts = map[string]string{
	"JetBrainsMonoNerdFontMono":    "JetBrainsMono Nerd Font Mono",
	"JetBrainsMonoNerdFont":        "JetBrainsMono Nerd Font",
	"FiraCodeNerdFontMono":         "FiraCode Nerd Font Mono",
	"FiraCodeNerdFont":             "FiraCode Nerd Font",
	"HackNerdFontMono":             "Hack Nerd Font Mono",
	"HackNerdFont":                 "Hack Nerd Font",
	"MesloLGSNerdFontMono":         "MesloLGS Nerd Font Mono",
	"MesloLGSNerdFont":             "MesloLGS Nerd Font",
	"MesloLGMNerdFontMono":         "MesloLGM Nerd Font Mono",
	"MesloLGMNerdFont":             "MesloLGM Nerd Font",
	"MesloLGLNerdFontMono":         "MesloLGL Nerd Font Mono",
	"MesloLGLNerdFont":             "MesloLGL Nerd Font",
	"CaskaydiaCoveNerdFontMono":    "CaskaydiaCove Nerd Font Mono",
	"CaskaydiaCoveNerdFont":        "CaskaydiaCove Nerd Font",
	"CaskaydiaMono NerdFontMono":   "CaskaydiaMono Nerd Font Mono",
	"CaskaydiaMono NerdFont":       "CaskaydiaMono Nerd Font",
	"SauceCodeProNerdFontMono":     "SauceCodePro Nerd Font Mono",
	"SauceCodeProNerdFont":         "SauceCodePro Nerd Font",
	"UbuntuMonoNerdFontMono":       "UbuntuMono Nerd Font Mono",
	"UbuntuMonoNerdFont":           "UbuntuMono Nerd Font",
	"UbuntuSansMonoNerdFontMono":   "UbuntuSansMono Nerd Font Mono",
	"UbuntuSansMonoNerdFont":       "UbuntuSansMono Nerd Font",
	"RobotoMonoNerdFontMono":       "RobotoMono Nerd Font Mono",
	"RobotoMonoNerdFont":           "RobotoMono Nerd Font",
	"InconsolataNerdFontMono":      "Inconsolata Nerd Font Mono",
	"InconsolataNerdFont":          "Inconsolata Nerd Font",
	"VictorMonoNerdFontMono":       "VictorMono Nerd Font Mono",
	"VictorMonoNerdFont":           "VictorMono Nerd Font",
	"IosekvaNerdFontMono":          "Iosevka Nerd Font Mono",
	"IosekvaNerdFont":              "Iosevka Nerd Font",
	"IosevkaTermNerdFontMono":      "IosevkaTerm Nerd Font Mono",
	"IosevkaTermNerdFont":          "IosevkaTerm Nerd Font",
	"ComicShannsMonoNerdFontMono":  "ComicShannsMono Nerd Font Mono",
	"ComicShannsMonoNerdFont":      "ComicShannsMono Nerd Font",
	"GeistMonoNerdFontMono":        "GeistMono Nerd Font Mono",
	"GeistMonoNerdFont":            "GeistMono Nerd Font",
	"MonaspaceNeonNerdFontMono":    "MonaspaceNeon Nerd Font Mono",
	"MonaspaceNeonNerdFont":        "MonaspaceNeon Nerd Font",
	"MonaspaceArgonNerdFontMono":   "MonaspaceArgon Nerd Font Mono",
	"MonaspaceArgonNerdFont":       "MonaspaceArgon Nerd Font",
	"ZedMonoNerdFontMono":          "ZedMono Nerd Font Mono",
	"ZedMonoNerdFont":              "ZedMono Nerd Font",
	"MapleMonoNerdFontMono":        "MapleMono Nerd Font Mono",
	"MapleMonoNerdFont":            "MapleMono Nerd Font",
	"CommitMonoNerdFontMono":       "CommitMono Nerd Font Mono",
	"CommitMonoNerdFont":           "CommitMono Nerd Font",
	"DaddyTimeMonoNerdFontMono":    "DaddyTimeMono Nerd Font Mono",
	"DaddyTimeMonoNerdFont":        "DaddyTimeMono Nerd Font",
	"AnonymiceProNerdFontMono":     "AnonymicePro Nerd Font Mono",
	"AnonymiceProNerdFont":         "AnonymicePro Nerd Font",
	"BitstreamVeraSansMonoNerdFontMono": "BitstromWera Nerd Font Mono",
	"BitstreamVeraSansMonoNerdFont":     "BitstromWera Nerd Font",
	"DejaVuSansMonoNerdFontMono":   "DejaVuSansM Nerd Font Mono",
	"DejaVuSansMonoNerdFont":       "DejaVuSansM Nerd Font",
	"LiterationMonoNerdFontMono":   "LiterationMono Nerd Font Mono",
	"LiterationMonoNerdFont":       "LiterationMono Nerd Font",
	"SpaceMonoNerdFontMono":        "SpaceMono Nerd Font Mono",
	"SpaceMonoNerdFont":            "SpaceMono Nerd Font",
}

// fontDirs returns macOS font directories to scan.
func fontDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, "Library", "Fonts"),
		"/Library/Fonts",
	}
}

// isFontFile checks if a filename has a font extension.
func isFontFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".ttf" || ext == ".otf" || ext == ".woff2"
}

// matchNerdFont tries to match a filename against known Nerd Font prefixes.
// Returns the CSS family name and true if matched.
func matchNerdFont(filename string) (string, bool) {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Strip weight suffixes like -Regular, -Bold, -Italic, -BoldItalic
	for _, suffix := range []string{"-Regular", "-Bold", "-Italic", "-BoldItalic",
		"-Light", "-LightItalic", "-Medium", "-MediumItalic",
		"-SemiBold", "-SemiBoldItalic", "-ExtraBold", "-ExtraBoldItalic",
		"-Thin", "-ThinItalic", "-ExtraLight", "-ExtraLightItalic"} {
		base = strings.TrimSuffix(base, suffix)
	}

	if family, ok := knownNerdFonts[base]; ok {
		return family, true
	}

	// Heuristic fallback: if the filename contains "NerdFont", reconstruct the family name.
	if strings.Contains(base, "NerdFont") {
		// Insert spaces before "Nerd" to create "Foo Nerd Font Mono" or "Foo Nerd Font"
		family := strings.Replace(base, "NerdFontMono", " Nerd Font Mono", 1)
		family = strings.Replace(family, "NerdFont", " Nerd Font", 1)
		return family, true
	}

	return "", false
}

// LocalFontFile represents a single font file variant (e.g. Bold, Italic).
type LocalFontFile struct {
	FileName string `json:"fileName"`
	Weight   string `json:"weight"`
	Style    string `json:"style"`
	Format   string `json:"format"`
	Base64   string `json:"base64"`
}

// LocalFontFamily groups font files into a CSS font-family with variants.
type LocalFontFamily struct {
	Family string          `json:"family"`
	Files  []LocalFontFile `json:"files"`
}

// fontsDir returns the local fonts directory path.
func fontsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".conductor", "fonts")
}

// ensureFontsDir creates ~/.conductor/fonts/ if it doesn't exist.
func ensureFontsDir() {
	os.MkdirAll(fontsDir(), 0755)
}

// GetFontsDir returns the local fonts directory path for the frontend.
func (a *App) GetFontsDir() string {
	return fontsDir()
}

// OpenFontsDir opens the fonts directory in Finder.
func (a *App) OpenFontsDir() error {
	dir := fontsDir()
	ensureFontsDir()
	return exec.Command("open", dir).Start()
}

// fontFormat returns the CSS font format string for a file extension.
func fontFormat(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".woff2":
		return "woff2"
	case ".otf":
		return "opentype"
	default:
		return "truetype"
	}
}

// weightSuffixes maps filename weight suffixes to CSS font-weight and font-style.
var weightSuffixes = []struct {
	suffix string
	weight string
	style  string
}{
	{"-ExtraLightItalic", "200", "italic"},
	{"-ExtraLight", "200", "normal"},
	{"-ThinItalic", "100", "italic"},
	{"-Thin", "100", "normal"},
	{"-LightItalic", "300", "italic"},
	{"-Light", "300", "normal"},
	{"-MediumItalic", "500", "italic"},
	{"-Medium", "500", "normal"},
	{"-SemiBoldItalic", "600", "italic"},
	{"-SemiBold", "600", "normal"},
	{"-ExtraBoldItalic", "800", "italic"},
	{"-ExtraBold", "800", "normal"},
	{"-BoldItalic", "700", "italic"},
	{"-Bold", "700", "normal"},
	{"-Italic", "400", "italic"},
	{"-Regular", "400", "normal"},
}

// detectWeightStyle extracts CSS font-weight and font-style from a filename.
func detectWeightStyle(filename string) (weight, style string) {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, ws := range weightSuffixes {
		if strings.HasSuffix(base, ws.suffix) {
			return ws.weight, ws.style
		}
	}
	return "400", "normal"
}

// familyFromFilename returns the CSS font-family name for a font filename.
func familyFromFilename(name string) string {
	family, ok := matchNerdFont(name)
	if ok {
		return family
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	for _, ws := range weightSuffixes {
		base = strings.TrimSuffix(base, ws.suffix)
	}
	return base
}

// addFontToFamilies indexes a font file (name + raw bytes) into the families map.
func addFontToFamilies(families map[string][]LocalFontFile, name string, data []byte) {
	family := familyFromFilename(name)
	weight, style := detectWeightStyle(name)
	families[family] = append(families[family], LocalFontFile{
		FileName: name,
		Weight:   weight,
		Style:    style,
		Format:   fontFormat(name),
		Base64:   base64.StdEncoding.EncodeToString(data),
	})
}

// scanZipForFonts reads font files from inside a zip archive.
func scanZipForFonts(zipPath string, families map[string][]LocalFontFile) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if f.FileInfo().IsDir() || !isFontFile(name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		addFontToFamilies(families, name, data)
	}
}

// ListLocalFonts scans ~/.conductor/fonts/ for loose font files and zip archives,
// returning font families with base64-encoded data for @font-face registration.
func (a *App) ListLocalFonts() []LocalFontFamily {
	dir := fontsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	families := make(map[string][]LocalFontFile)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())

		// Handle zip archives
		if strings.ToLower(filepath.Ext(entry.Name())) == ".zip" {
			scanZipForFonts(path, families)
			continue
		}

		// Handle loose font files
		if !isFontFile(entry.Name()) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		addFontToFamilies(families, entry.Name(), data)
	}

	result := make([]LocalFontFamily, 0, len(families))
	for family, files := range families {
		result = append(result, LocalFontFamily{Family: family, Files: files})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Family < result[j].Family
	})
	return result
}

// ListNerdFonts scans macOS font directories and returns detected Nerd Font families.
func (a *App) ListNerdFonts() []NerdFontEntry {
	seen := make(map[string]string) // family → first file path

	for _, dir := range fontDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !isFontFile(entry.Name()) {
				continue
			}
			family, ok := matchNerdFont(entry.Name())
			if !ok {
				continue
			}
			if _, exists := seen[family]; !exists {
				seen[family] = filepath.Join(dir, entry.Name())
			}
		}
	}

	result := make([]NerdFontEntry, 0, len(seen))
	for family, path := range seen {
		result = append(result, NerdFontEntry{Family: family, FilePath: path})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Family < result[j].Family
	})
	return result
}
