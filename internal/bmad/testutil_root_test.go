package bmad

import (
	"os"
	"testing"
)

// skipIfRoot skips tests that rely on read-only directories: root bypasses
// file permissions, so they cannot fail as intended in root containers (R23).
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission test: running as root, which ignores read-only dirs")
	}
}
