package bmad

import (
	"os"
	"path/filepath"
	"strings"
)

const bmadOutputDir = "_bmad-output"

// artifactPaths maps artifact names to their relative path under _bmad-output.
// An empty string means the artifact is not file-based (e.g. code, tests).
var artifactPaths = map[string]string{
	"brainstorm-notes":  "analysis-artifacts/brainstorm-notes.md",
	"product-brief":     "analysis-artifacts/product-brief.md",
	"domain-research":   "analysis-artifacts/domain-research.md",
	"market-research":   "analysis-artifacts/market-research.md",
	"tech-research":     "analysis-artifacts/tech-research.md",
	"PRD.md":            "planning-artifacts/PRD.md",
	"ux-spec.md":        "planning-artifacts/ux-spec.md",
	"prd-validation":    "planning-artifacts/prd-validation.md",
	"architecture.md":   "solutioning-artifacts/architecture.md",
	"readiness-report":  "solutioning-artifacts/readiness-report.md",
	"project-context.md": "solutioning-artifacts/project-context.md",
	"epics/":            "solutioning-artifacts/epics/",
	"story-*.md":        "implementation-artifacts/stories/",
	"sprint-status.yaml": "implementation-artifacts/sprint-status.yaml",
	"code":              "",
	"tests":             "",
	"file-path":         "",
	"review-report":     "implementation-artifacts/reviews/",
	"retro-notes":       "implementation-artifacts/retro-notes.md",
	"any-doc":           "",
	"reviewed-doc":      "support-artifacts/reviewed-docs/",
	"elicitation-notes": "support-artifacts/elicitation-notes.md",
	"edge-case-report":  "support-artifacts/edge-case-report.md",
	"adversarial-report": "support-artifacts/adversarial-report.md",
	"infra-config":      "implementation-artifacts/infra/",
	"project-docs":      "support-artifacts/project-docs/",
}

// ResolveArtifactPath returns the absolute path for a named artifact within a
// repository. Returns "" if the artifact is unmapped or unknown.
func ResolveArtifactPath(name, repoPath string) string {
	rel, ok := artifactPaths[name]
	if !ok || rel == "" {
		return ""
	}
	p := filepath.Join(repoPath, bmadOutputDir, rel)
	if strings.HasSuffix(rel, "/") {
		p += string(filepath.Separator)
	}
	return p
}

// GetArtifactStatus checks whether a single named artifact exists at its expected
// path. Returns (exists, resolvedPath, nil). For unmapped artifacts the path is
// empty and exists is false.
func GetArtifactStatus(repoPath, artifactName string) (bool, string, error) {
	resolved := ResolveArtifactPath(artifactName, repoPath)
	if resolved == "" {
		return false, "", nil
	}
	_, err := os.Stat(resolved)
	if err != nil {
		return false, resolved, nil
	}
	return true, resolved, nil
}

// VerifyArtifacts checks which named artifacts exist on disk under repoPath.
// Unmapped artifacts (empty path) are silently skipped. Non-existence errors
// and other stat failures (permission denied, symlink loops) are both treated
// as missing — appropriate for a desktop app where partial results are acceptable.
func VerifyArtifacts(repoPath string, outputNames []string) (found, missing []string) {
	found = []string{}
	missing = []string{}

	for _, name := range outputNames {
		p := ResolveArtifactPath(name, repoPath)
		if p == "" {
			continue
		}
		_, err := os.Stat(p)
		if err == nil {
			found = append(found, name)
		} else {
			// Intentionally treats all errors as missing — both os.IsNotExist
			// and other failures (permission denied, symlink loops). In a desktop
			// app context, partial results are acceptable.
			missing = append(missing, name)
		}
	}
	return found, missing
}
