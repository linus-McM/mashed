// Package eval tests cover the offline Gemma eval harness (Story ui-ast-U9).
package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEval_Corpus_MinimumCount — AC-1: LoadCorpus returns ≥ 30 fixtures
// across 6 categories with ≥ 4 per category.
func TestEval_Corpus_MinimumCount(t *testing.T) {
	t.Parallel()

	corpus := LoadCorpus()
	require.GreaterOrEqual(t, len(corpus), 30,
		"AC-1 §4.5: LoadCorpus must return ≥ 30 fixtures; got %d", len(corpus))

	counts := map[string]int{}
	for _, f := range corpus {
		counts[f.Expected.Category]++
	}

	required := []string{
		"brainstorming",
		"elicitation",
		"product-brief",
		"party",
		"freeform",
		"adversarial",
	}
	for _, cat := range required {
		assert.GreaterOrEqual(t, counts[cat], 4,
			"AC-1 §4.5: category %q must have ≥ 4 fixtures; got %d", cat, counts[cat])
	}
}
