# UI AST Eval Corpus

This directory holds the hand-curated fixture corpus the offline Gemma eval
harness (Story ui-ast-U9) scores the adapter against. It is the test data
driving `internal/uiadapter/eval.LoadCorpus()` and the tagged
`TestEval_FullCorpus_MeetsThresholds` harness.

## Layout

Every fixture is a pair of files keyed by a shared ID stem:

- `<id>-raw.txt` — the raw Claude-turn capture fed into `Adapter.Translate`.
- `<id>-expected.json` — the companion label file declaring the widget-shape
  mix the adapter should emit. Schema:

  ```json
  {
    "decision_group_count": 1,
    "widget_types": ["choice"],
    "must_contain_urls": ["https://example.com"],
    "must_contain_code_blocks": ["```python\nprint('hi')\n```"],
    "fallback_answer_shape": "choice",
    "category": "brainstorming"
  }
  ```

  Fields:
  - `decision_group_count` — number of `decision_group` nodes the adapter is
    expected to emit.
  - `widget_types` — closed set from §3.2 (`choice`, `multi`, `approval`,
    `free`, `number`, `file`, `json`).
  - `must_contain_urls` / `must_contain_code_blocks` — verbatim strings the
    rendered AST must preserve per §7.2.
  - `fallback_answer_shape` — expected top-level
    `UIAST.FallbackAnswerShape` value when the adapter degrades to a
    single-markdown fallback.
  - `category` — one of the six AC-1 buckets (see below).

## Categories

AC-1 requires ≥ 30 fixtures with ≥ 4 per category:

| Category | Description |
|---|---|
| `brainstorming` | Method picker / technique selection / idea organisation prompts |
| `elicitation` | 5-method elicitation + r/a/x options |
| `product-brief` | Stage-by-stage brief questions |
| `party` | Freeform + code-block cases (§7.2 preservation) |
| `freeform` | Single open question |
| `adversarial` | Contains URL + code block; validator must preserve |

## Capture process

Fixtures are **synthetic** — hand-crafted representative shapes of real Claude
turns. They exist to exercise the eval harness's scoring code (valid-JSON
rate, validator-pass rate, per-widget P/R) and its threshold enforcement, not
to prove a specific model meets a specific accuracy bar in a controlled
benchmark. That's the job of the tagged harness running against live Ollama.

When replacing a fixture with a real capture:

1. **Scrub PII.** No user names, email addresses, internal URLs, absolute
   paths containing `/Users/<name>`, API keys, or repo-specific project
   identifiers.
2. **Collapse whitespace** only when the original used tabs interchangeably
   with spaces — otherwise leave whitespace verbatim so validator edge cases
   (markdown fence detection) are preserved.
3. **Commit the pair atomically.** A `<id>-raw.txt` without its
   `<id>-expected.json` fails `LoadCorpus` at panic-with-clear-error and
   blocks the test suite, which is the intended safety net.
4. **Update this file** with any new category or sub-shape rationale.

## Maintenance

- **Adding a fixture** — drop `<id>-raw.txt` + `<id>-expected.json` into this
  directory. `LoadCorpus()` picks them up automatically via `//go:embed`.
- **Removing a fixture** — delete both files. Ensure the per-category count
  stays ≥ 4 or `TestEval_Corpus_MinimumCount` will fail.
- **Renaming** — remember fixtures are keyed by filename stem; renaming the
  raw file without the expected file causes an atomic-pair panic at load.
