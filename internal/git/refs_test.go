package git

import (
	"errors"
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	valid := []string{
		"main", "dev", "feature/foo-1", "fix/issue_42", "release/v1.2.3", "a", "user/name/topic",
	}
	for _, name := range valid {
		if err := ValidateBranchName(name); err != nil {
			t.Errorf("ValidateBranchName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"", "-", "-b", "--help", "--output=/tmp/x",
		"a..b", "x@{1}", "@", "a b", "tab\tname", "new\nline",
		"foo.lock", "dir/foo.lock", "/x", "x/", "a//b", ".hidden", "dir/.hidden", "ends.",
		"q?", "star*", "br[acket", "back\\slash", "col:on", "til~de", "car^et",
		"HEAD", "del\x7f",
	}
	for _, name := range invalid {
		err := ValidateBranchName(name)
		if !errors.Is(err, ErrInvalidRef) {
			t.Errorf("ValidateBranchName(%q) = %v, want ErrInvalidRef", name, err)
		}
	}
}
