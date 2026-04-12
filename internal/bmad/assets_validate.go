package bmad

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationIssue describes a single frontmatter validation finding.
type ValidationIssue struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

const (
	SeverityWarn = "warn"
	SeverityInfo = "info"
)

// validCompletionValues are the recognised mashedCompletion values.
// Prefixed values ("marker:", "timeout:") are checked separately.
var validCompletionValues = map[string]bool{
	"idle": true,
	"exit": true,
}

// ValidateMashedAsset checks a parsed asset for completeness and
// correctness. Returns a (possibly empty) slice of issues. Validation
// is advisory — it never blocks loading, dragging, or running.
//
// When repoPath is empty, path-existence checks for mashedInputs and
// mashedOutputs are skipped silently.
func ValidateMashedAsset(asset MashedAssetInfo, repoPath string) []ValidationIssue {
	issues := []ValidationIssue{}

	if asset.Description == "" {
		issues = append(issues, ValidationIssue{
			Field:    "description",
			Severity: SeverityWarn,
			Message:  "description missing",
		})
	} else {
		if len(asset.Description) < 10 {
			issues = append(issues, ValidationIssue{
				Field:    "description",
				Severity: SeverityInfo,
				Message:  "description is very short (< 10 characters)",
			})
		}
		if len(asset.Description) > 200 {
			issues = append(issues, ValidationIssue{
				Field:    "description",
				Severity: SeverityInfo,
				Message:  "description is very long (> 200 characters)",
			})
		}
	}

	if repoPath != "" {
		for _, input := range asset.Inputs {
			if !pathExistsInRepo(repoPath, input) {
				issues = append(issues, ValidationIssue{
					Field:    "mashedInputs",
					Severity: SeverityWarn,
					Message:  fmt.Sprintf("input path does not exist: %s", input),
				})
			}
		}
		for _, output := range asset.Outputs {
			if !pathExistsInRepo(repoPath, output) {
				issues = append(issues, ValidationIssue{
					Field:    "mashedOutputs",
					Severity: SeverityInfo,
					Message:  fmt.Sprintf("output path does not exist: %s", output),
				})
			}
		}
	}

	if asset.Role == MashedRoleCommand {
		if !isValidCompletion(asset.Completion) {
			issues = append(issues, ValidationIssue{
				Field:    "mashedCompletion",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("unrecognised mashedCompletion value: %s", asset.Completion),
			})
		}
	}

	if asset.Role == MashedRoleCommand && asset.Chainable == "none" {
		issues = append(issues, ValidationIssue{
			Field:    "mashedChainable",
			Severity: SeverityInfo,
			Message:  "command has mashedChainable=none; commands usually chain",
		})
	}

	return issues
}

// isValidCompletion checks whether a completion value is in the
// recognised set: "idle", "exit", or prefixed with "marker:" / "timeout:".
func isValidCompletion(v string) bool {
	if validCompletionValues[v] {
		return true
	}
	if strings.HasPrefix(v, "marker:") || strings.HasPrefix(v, "timeout:") {
		return true
	}
	return false
}

// pathExistsInRepo checks whether a path (possibly a glob) resolves
// to at least one file under repoPath. Uses filepath.Glob for simple
// patterns. Note: ** recursive globs are not supported in v1 — they
// will fail to match and produce a warning, which is acceptable.
func pathExistsInRepo(repoPath, pattern string) bool {
	abs := filepath.Join(repoPath, pattern)

	// Try as a literal path first (fast path).
	if _, err := os.Stat(abs); err == nil {
		return true
	}

	// Try as a glob pattern.
	matches, err := filepath.Glob(abs)
	if err != nil {
		// Bad pattern — treat as non-matching.
		return false
	}
	return len(matches) > 0
}
