// Package eval is the offline Gemma UI AST eval harness (Story ui-ast-U9).
// It loads a hand-curated corpus of raw Claude-turn captures + companion
// expected-widget labels and scores an Adapter against the §4.5 thresholds.
// The harness itself lives in internal/uiadapter under the ollama_eval build
// tag — this package holds the data types, loader, and scoring code so unit
// tests can exercise them without a live Ollama instance.
package eval

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Fixture pairs a raw Claude-turn capture with its expected widget-shape
// label. ID is the fixture filename stem so scorecard diagnostics can cite
// the exact file a regression came from.
type Fixture struct {
	ID       string
	Raw      string
	Expected Expected
}

// Expected is the companion label schema committed alongside every raw
// fixture. It declares the widget-shape mix the adapter is expected to emit,
// plus optional preservation assertions (URLs / code blocks) per §7.2.
type Expected struct {
	DecisionGroupCount    int      `json:"decision_group_count"`
	WidgetTypes           []string `json:"widget_types"`
	MustContainURLs       []string `json:"must_contain_urls,omitempty"`
	MustContainCodeBlocks []string `json:"must_contain_code_blocks,omitempty"`
	FallbackAnswerShape   string   `json:"fallback_answer_shape,omitempty"`
	// Category gates AC-1's per-category count check. See CORPUS.md for the
	// closed set: brainstorming|elicitation|product-brief|party|freeform|adversarial.
	Category string `json:"category"`
}

//go:embed testdata/*-raw.txt testdata/*-expected.json testdata/CORPUS.md
var corpusFS embed.FS

const (
	rawSuffix      = "-raw.txt"
	expectedSuffix = "-expected.json"
	testdataDir    = "testdata"
)

// LoadCorpus returns every embedded fixture sorted by ID so scorecard output
// is deterministic across runs. Panics on a malformed pair (missing
// expected.json, unparseable JSON) — fixtures are checked-in test data, so a
// corruption is a build-time error the developer must fix before shipping,
// not something callers should defensively handle.
func LoadCorpus() []Fixture {
	entries, err := fs.ReadDir(corpusFS, testdataDir)
	if err != nil {
		panic(fmt.Sprintf("eval: read testdata dir: %v", err))
	}

	ids := collectFixtureIDs(entries)
	corpus := make([]Fixture, 0, len(ids))
	for _, id := range ids {
		corpus = append(corpus, loadFixture(id))
	}
	return corpus
}

func collectFixtureIDs(entries []fs.DirEntry) []string {
	seen := map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, rawSuffix) {
			continue
		}
		seen[strings.TrimSuffix(name, rawSuffix)] = struct{}{}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func loadFixture(id string) Fixture {
	rawPath := testdataDir + "/" + id + rawSuffix
	expPath := testdataDir + "/" + id + expectedSuffix

	rawBytes, err := corpusFS.ReadFile(rawPath)
	if err != nil {
		panic(fmt.Sprintf("eval: read %s: %v", rawPath, err))
	}
	expBytes, err := corpusFS.ReadFile(expPath)
	if err != nil {
		panic(fmt.Sprintf("eval: read companion %s: %v", expPath, err))
	}
	var exp Expected
	if err := json.Unmarshal(expBytes, &exp); err != nil {
		panic(fmt.Sprintf("eval: decode %s: %v", expPath, err))
	}
	return Fixture{ID: id, Raw: string(rawBytes), Expected: exp}
}
