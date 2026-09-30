package git

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidRef is returned for a branch name that git would reject or that
// could be parsed as a command-line option.
var ErrInvalidRef = errors.New("invalid branch name")

// ValidateBranchName checks name against the rules of
// `git check-ref-format --branch`, plus a rejection of a leading '-', in pure
// Go so untrusted input never reaches a git process (spec R10).
func ValidateBranchName(name string) error {
	invalid := func(why string) error {
		return fmt.Errorf("%q: %s: %w", name, why, ErrInvalidRef)
	}
	switch {
	case name == "":
		return invalid("empty")
	case strings.HasPrefix(name, "-"):
		return invalid("starts with '-'")
	case name == "@" || name == "HEAD":
		return invalid("reserved name")
	case strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/"):
		return invalid("starts or ends with '/'")
	case strings.HasSuffix(name, "."):
		return invalid("ends with '.'")
	case strings.Contains(name, ".."):
		return invalid("contains '..'")
	case strings.Contains(name, "@{"):
		return invalid("contains '@{'")
	case strings.Contains(name, "//"):
		return invalid("contains '//'")
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(" ~^:?*[\\", r) {
			return invalid(fmt.Sprintf("contains %q", r))
		}
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return invalid("component starts with '.'")
		}
		if strings.HasSuffix(part, ".lock") {
			return invalid("component ends with '.lock'")
		}
	}
	return nil
}
