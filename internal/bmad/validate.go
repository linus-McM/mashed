package bmad

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// defaultJSONMaxLength caps ShapeJSON payloads when no explicit MaxLength is
// set on the InputSpec (§13.4). 64 KiB is enough for typical structured
// answers while preventing unbounded snapshot growth.
const defaultJSONMaxLength = 65536

// validateInput checks a user-supplied value against its declared InputSpec
// (§8.4). Returns an error wrapping ErrInvalidInput on failure.
func validateInput(spec InputSpec, value string) error {
	if spec.Required && value == "" {
		return fmt.Errorf("value is required: %w", ErrInvalidInput)
	}

	switch spec.Shape {
	case ShapeChoice:
		// When OptionsRef is set but static Options is empty, the option list
		// is resolved dynamically per-round (e.g. registry:methods.csv?random=5).
		// Strict list-membership validation is skipped because the resolved
		// snapshot lives in PendingPrompt.Options, not on the InputSpec.
		if len(spec.Options) > 0 {
			if !containsString(spec.Options, value) {
				return fmt.Errorf("value must be one of %v: %w", spec.Options, ErrInvalidInput)
			}
		}

	case ShapeMultiChoice:
		if len(spec.Options) == 0 {
			// Same dynamic-options relaxation as ShapeChoice above.
			break
		}
		for _, v := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(v)
			if !containsString(spec.Options, trimmed) {
				return fmt.Errorf("%q not in options %v: %w", trimmed, spec.Options, ErrInvalidInput)
			}
		}

	case ShapeApproval:
		if value != "yes" && value != "no" {
			return fmt.Errorf("approval must be yes or no: %w", ErrInvalidInput)
		}

	case ShapeFree, ShapeJSON, ShapeFile:
		if spec.Validation != "" {
			re, err := regexp.Compile(spec.Validation)
			if err != nil {
				return fmt.Errorf("invalid validation regex: %w", ErrInvalidInput)
			}
			if !re.MatchString(value) {
				return fmt.Errorf("value failed validation regex: %w", ErrInvalidInput)
			}
		}
		limit := spec.MaxLength
		if spec.Shape == ShapeJSON && limit == 0 {
			limit = defaultJSONMaxLength
		}
		if limit > 0 && len(value) > limit {
			return fmt.Errorf("value exceeds max length %d: %w", limit, ErrInvalidInput)
		}
	}
	return nil
}

// containsString returns true when s is an element of list.
func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// resolveFileInput enforces the repo-root containment guard for ShapeFile
// values (§14.2). On success returns the cleaned absolute path; on failure
// returns an error wrapping ErrPathOutsideRepo.
//
// The prefix check uses `filepath.Clean(repoRoot) + Separator` — the trailing
// separator is required to avoid false-positives like "/repo-other/foo"
// matching "/repo" via HasPrefix.
func resolveFileInput(value, repoRoot string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", ErrPathOutsideRepo)
	}
	clean := filepath.Clean(abs)
	prefix := filepath.Clean(repoRoot) + string(filepath.Separator)
	if !strings.HasPrefix(clean, prefix) {
		return "", fmt.Errorf("%q is outside %q: %w", clean, repoRoot, ErrPathOutsideRepo)
	}
	return clean, nil
}
