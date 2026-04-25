package uiadapter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitize_StripsANSI_Golden — Story v3-01 AC-1.1. Four scenarios.
func TestSanitize_StripsANSI_Golden(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "bare ANSI color",
			raw:  "\x1b[32mgreen text\x1b[0m",
			want: "green text",
		},
		{
			name: "cursor-position fragment",
			raw:  "before\x1b[2Kafter",
			want: "beforeafter",
		},
		{
			name: "OSC-8 hyperlink",
			raw:  "visit \x1b]8;;https://example.com\x07Example\x1b]8;;\x07 today",
			want: "visit Example today",
		},
		{
			name: "pathological mixed",
			raw:  "\x1b[2J\x1b[Hheader\x1b[32m highlight \x1b[0m\x1b[2K\nnext line",
			want: "header highlight \nnext line",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, delta := SanitizeCapture(tc.raw, nil)
			assert.Equal(t, tc.want, got, "golden sanitized output")
			assert.Equal(t, len(tc.raw)-len(got), delta, "deltaBytes = len(raw) - len(sanitized)")
		})
	}
}

// TestSanitize_PreservesFencedCodeBlocks — AC-1.3. Fenced triple-backtick
// blocks with brackets, braces, and backticks inside round-trip unchanged.
func TestSanitize_PreservesFencedCodeBlocks(t *testing.T) {
	t.Parallel()
	raw := "before prose\n" +
		"```go\nfunc foo() { return []int{1, 2, 3} }\n```\n" +
		"middle with \x1b[32mANSI\x1b[0m\n" +
		"```bash\necho \"`date`\" && printf '[x]\\n'\n```\n" +
		"trailing"
	got, _ := SanitizeCapture(raw, nil)
	// Inner code content must survive byte-for-byte.
	assert.Contains(t, got, "func foo() { return []int{1, 2, 3} }")
	assert.Contains(t, got, "echo \"`date`\" && printf '[x]\\n'")
	// ANSI in prose is stripped.
	assert.NotContains(t, got, "\x1b[32m")
	assert.Contains(t, got, "middle with ANSI")
}

// TestSanitize_Idempotent — running sanitize twice produces the same output.
// Property-style safeguard; no plan-level AC but desirable.
func TestSanitize_Idempotent(t *testing.T) {
	t.Parallel()
	raw := "\x1b[31merror:\x1b[0m failed at \x1b]8;;https://log\x07line 42\x1b]8;;\x07"
	once, _ := SanitizeCapture(raw, nil)
	twice, delta := SanitizeCapture(once, nil)
	assert.Equal(t, once, twice, "sanitize must be idempotent")
	assert.Equal(t, 0, delta, "second pass strips nothing")
}

// TestSanitize_EmptyInput — guard zero-case; deltaBytes is zero.
func TestSanitize_EmptyInput(t *testing.T) {
	t.Parallel()
	got, delta := SanitizeCapture("", nil)
	assert.Equal(t, "", got)
	assert.Equal(t, 0, delta)
}

// TestSanitize_WhitespaceTrimmed — leading/trailing whitespace removed per
// plan bullet 5.
func TestSanitize_WhitespaceTrimmed(t *testing.T) {
	t.Parallel()
	got, delta := SanitizeCapture("   \n\tprompt?\n   ", nil)
	assert.Equal(t, "prompt?", got)
	assert.Greater(t, delta, 0)
}

// TestSanitize_LargeInput — sanity perf: 64 KiB input processes quickly.
// No latency assertion; relies on race-runner timeout to catch pathology.
func TestSanitize_LargeInput(t *testing.T) {
	t.Parallel()
	raw := strings.Repeat("\x1b[32mA\x1b[0m", 8192) // 64 KiB
	got, delta := SanitizeCapture(raw, nil)
	assert.Equal(t, strings.Repeat("A", 8192), got)
	assert.Greater(t, delta, 0)
}
