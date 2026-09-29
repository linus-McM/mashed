//go:build ollama_eval

package uiadapter

import "testing"

// WithOllamaHost overrides the package-level Ollama HTTP endpoint for the
// duration of a single test. Exposed only under the `ollama_eval` build tag
// so production callers cannot touch the host pin — the harness in
// eval_test.go (external test package) uses it to point
// TestEval_SkipWhenUnreachable at a closed port and prove the adapter
// surfaces `fallback:unreachable`.
func WithOllamaHost(t *testing.T, url string) {
	t.Helper()
	old := ollamaHost
	ollamaHost = url
	t.Cleanup(func() { ollamaHost = old })
}
