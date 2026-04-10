# Story bridge-01: Descriptive BMAD Session Names

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** ready

## Description

Replace the cryptic `bmad-{nodeID}-{unixTimestamp}` tmux session naming scheme with a self-describing format: `bmad-{repo}-{branch}-{nodeLabel}-{shortHash}` (e.g. `bmad-surfseer-main-create-story-a1b2c3d4`). This makes `tmux ls`, log messages, and the View Terminal modal title immediately identifiable by repo, branch, and workflow step — no more guessing which session belongs to which workflow. The BMAD executor's existing tmux target format (`{sessionName}:0.0`) and all downstream behavior (polling, `RespondToQuestion`, capture) stay intact because only the `sessionName` formula changes.

## Developer Notes

### Architecture

- **New file:** `internal/bmad/session_naming.go`
  - `func BuildSessionName(repoPath, branch, nodeLabel, nodeID string, nowNanos int64) string`
  - `func ParseSessionName(name string) (repo, branch, label, shortHash string, ok bool)` — best-effort, returns `ok=false` for unrecognised formats (used by frontend-facing helpers and logs)
  - `func slugifyComponent(s string, maxLen int) string` — internal, lowercases, strips to ASCII alnum/`_`/`-`, collapses runs of `-`, trims leading/trailing `-`, truncates to `maxLen`
  - Exported constant `SessionNamePrefix = "bmad-"`
  - Exported constant `MaxSessionNameBytes = 100`
- **New file:** `internal/bmad/session_naming_test.go` — table-driven, 20+ cases covering slug rules, truncation, unicode, collision avoidance, empty inputs
- **Modify file:** `internal/bmad/executor.go`
  - Replace line 850 `sessionName := fmt.Sprintf("bmad-%s-%d", nodeID, time.Now().Unix())` with a call to `BuildSessionName(state.exec.RepoPath, branch, node.Label, nodeID, time.Now().UnixNano())`
  - Derive `branch` from the execution or repo. If not already tracked, read it via `e.runCmd(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")`. On error fall back to the literal string `"detached"`. Do **not** fail the node on a branch lookup error.
  - Derive `repo` from `filepath.Base(repoPath)`.
  - `nodeLabel` comes from `state.exec.Nodes[idx].Label` (the user-facing label). If empty, fall back to `state.exec.Nodes[idx].ProcessID`.
- **Modify file:** `internal/bmad/executor_test.go`
  - Any test asserting `sessionName` starts with `bmad-node-` or matches `bmad-{nodeID}-{unix}` must switch to asserting it starts with `SessionNamePrefix` and passes `ParseSessionName`.

### Key Method Signatures

```go
// session_naming.go
const (
    SessionNamePrefix   = "bmad-"
    MaxSessionNameBytes = 100
)

func BuildSessionName(repoPath, branch, nodeLabel, nodeID string, nowNanos int64) string
func ParseSessionName(name string) (repo, branch, label, shortHash string, ok bool)
```

### Technical Considerations

- **Slug rules:** lowercase, replace any rune that is not `[a-z0-9_-]` with `-`, collapse runs of `-`, trim leading/trailing `-`. Per-component max length is 24 bytes so the full name stays under 100 bytes even with all four components populated.
- **Short hash:** `fmt.Sprintf("%x", sha256.Sum256([]byte(nodeID + "|" + strconv.FormatInt(nowNanos, 10))))[:8]` — uses nanos (not unix seconds) to minimise collision when two nodes launch in the same second.
- **Fallbacks:**
  - Empty `repoPath` → `"unknown"`
  - Empty `branch` → `"unknown"`
  - Empty `nodeLabel` AND empty `nodeID` → `"unlabeled"`
  - All empty → `bmad-unknown-unknown-unlabeled-{hash}`
- **Length cap:** If the assembled name exceeds `MaxSessionNameBytes`, truncate each component proportionally but always preserve the `bmad-` prefix and the 8-char hash suffix.
- **`ParseSessionName`:** splits on `-` from the right (hash is always the last 8 chars). Returns `ok=true` only if name starts with `SessionNamePrefix` and has a valid 8-hex-char tail.
- **No breaking change to `TmuxTarget` format:** Still `{sessionName}:0.0`. `RespondToQuestion`, polling, and `captureOutput` all operate on targets, not session names.

### Risks & Edge Cases

1. **Branch lookup latency:** `git rev-parse` is fast but runs synchronously before tmux launch. Acceptable — executeNode already runs multiple subprocesses.
2. **Branch with slashes** (`feature/foo/bar`): slugify collapses them to `feature-foo-bar`.
3. **Non-ASCII repo/branch names:** slugifier strips non-ASCII runes to `-`.
4. **Very long repo paths / workflow labels:** per-component truncation kicks in before total-length check.
5. **Session name collision** (two workflows, same repo/branch/label, same nanosecond): the 8-char hash differs because nanos differ. Documented as accepted risk — probability is effectively zero in single-user desktop use. Retry with regenerated hash is a Story 3 concern (not here).
6. **Existing old-format sessions:** Not reconnectable — documented as a known gap; users must restart workflows. bridge-04 handles cleanup.

### Reference Files

- `internal/bmad/executor.go` lines 820-865 — `executeNode` tmux launch site (modification target)
- `internal/bmad/executor.go` lines 22-80 — `CommandRunner`, `Executor` struct, `runCmd` wiring (for test harness pattern)
- `internal/bmad/executor.go` lines 94-129 — `types.go` `WorkflowExecution` struct (`RepoPath` already present; no new field needed)
- `internal/bmad/question.go` lines 128-145 — `escapeTmuxLiteral` (same package style for helper functions)
- `internal/bmad/executor_test.go` — `newHarness` pattern with mock `CommandRunner`

Reference skills: `/golang-testing`, `/golang-error-handling`, `/simplify`.

## Acceptance Criteria

**AC-1: BuildSessionName produces self-describing names**
- Given `repoPath="/Users/linus/Development/surfseer"`, `branch="main"`, `nodeLabel="Create Story"`, `nodeID="node-1775795467345"`, and `nowNanos=1712600000000000000`
- When `BuildSessionName(...)` is called
- Then the returned string starts with `"bmad-surfseer-main-create-story-"`
- And the final 8 characters are lowercase hex

**AC-2: All components are slugified to lowercase ASCII alnum/dash/underscore**
- Given any input containing spaces, slashes, or non-ASCII runes
- When `BuildSessionName` is called
- Then the output contains only characters from `[a-z0-9_-]`
- And no component starts or ends with `-`

**AC-3: Empty inputs fall back to sentinel components**
- Given `repoPath=""`, `branch=""`, `nodeLabel=""`, `nodeID=""`
- When `BuildSessionName("", "", "", "", 123)` is called
- Then the result matches `bmad-unknown-unknown-unlabeled-{8 hex chars}`

**AC-4: Total name length never exceeds MaxSessionNameBytes**
- Given inputs with very long values (e.g. 500-char repo, 500-char branch, 500-char label, 500-char nodeID)
- When `BuildSessionName` is called
- Then `len(result) <= MaxSessionNameBytes`
- And the result still begins with `SessionNamePrefix`
- And the result still ends with an 8-char hex hash

**AC-5: ParseSessionName round-trips valid names**
- Given a name produced by `BuildSessionName("/x/myrepo", "main", "Create Story", "n1", 100)`
- When `ParseSessionName(name)` is called
- Then it returns `repo="myrepo"`, `branch="main"`, `label="create-story"`, an 8-char `shortHash`, and `ok=true`

**AC-6: ParseSessionName rejects unrecognised names**
- Given a name `"not-a-bmad-session"` or `"bmad-short"`
- When `ParseSessionName(name)` is called
- Then it returns `ok=false`
- And returned string fields are empty

**AC-7: Executor uses BuildSessionName for new tmux sessions**
- Given a BMAD workflow started on repo `/tmp/testrepo` on branch `main` with node labelled `"Draft PRD"`
- When the node begins executing
- Then `tmux new-session -s {name}` is invoked where `{name}` passes `ParseSessionName` and decodes to `repo="testrepo"`, `branch="main"`, `label="draft-prd"`
- And `state.exec.Nodes[idx].TmuxTarget` equals `{name}:0.0`

**AC-8: Executor degrades gracefully on branch lookup failure**
- Given the mock `CommandRunner` returns an error for `git rev-parse --abbrev-ref HEAD`
- When a node launches
- Then the session name still builds successfully
- And the branch component is `"detached"`
- And the node is NOT failed due to the git error

## BDD Test Scenarios

### Scenario 1: Basic naming

```gherkin
Feature: Descriptive BMAD session names

  Scenario: Typical case with all fields populated
    Given repoPath "/Users/linus/Development/surfseer"
    And branch "main"
    And nodeLabel "Create Story"
    And nodeID "node-1775795467345"
    And nowNanos 1712600000000000000
    When BuildSessionName is called
    Then the result starts with "bmad-surfseer-main-create-story-"
    And the result has exactly one 8-char hex suffix after the last "-"

  Scenario: Feature branch with slashes
    Given branch "feature/foo/bar"
    When BuildSessionName is called
    Then the branch component is "feature-foo-bar"
```

### Scenario 2: Slugification rules

```gherkin
Feature: Slugification

  Scenario: Non-ASCII runes are stripped
    Given nodeLabel "Créate Störy 日本語"
    When BuildSessionName is called
    Then the label component contains only [a-z0-9_-]

  Scenario: Runs of dashes collapse
    Given nodeLabel "foo   bar"
    When BuildSessionName is called
    Then the label component is "foo-bar"

  Scenario: Leading and trailing dashes trimmed
    Given nodeLabel "  create  "
    When slugifyComponent is called
    Then the result is "create"
```

### Scenario 3: Fallbacks

```gherkin
Feature: Empty input fallbacks

  Scenario: All empty
    Given repoPath "" and branch "" and nodeLabel "" and nodeID ""
    When BuildSessionName is called with nowNanos 123
    Then the result matches "bmad-unknown-unknown-unlabeled-[0-9a-f]{8}"

  Scenario: Empty label only
    Given nodeLabel "" and nodeID "fallback-id"
    When BuildSessionName is called
    Then the label component is "fallback-id"
```

### Scenario 4: Length cap

```gherkin
Feature: Max session name length

  Scenario: Very long inputs get truncated
    Given a repoPath of 500 "a" chars and branch of 500 "b" chars
    When BuildSessionName is called
    Then len(result) <= 100
    And the result starts with "bmad-"
    And the last 8 characters are hex
```

### Scenario 5: Round-trip parsing

```gherkin
Feature: ParseSessionName

  Scenario: Valid name parses
    Given a name produced by BuildSessionName("/x/myrepo", "main", "Create Story", "n1", 100)
    When ParseSessionName is called
    Then repo is "myrepo"
    And branch is "main"
    And label is "create-story"
    And ok is true

  Scenario: Invalid prefix
    Given a name "foo-bar-baz-12345678"
    When ParseSessionName is called
    Then ok is false
```

### Scenario 6: Executor integration

```gherkin
Feature: Executor uses BuildSessionName

  Scenario: Node launches with descriptive name
    Given a mock CommandRunner capturing all tmux invocations
    And a workflow on repo "/tmp/testrepo" branch "main"
    And a node labelled "Draft PRD"
    When the executor launches the node
    Then the "tmux new-session -s ..." call uses a name starting with "bmad-testrepo-main-draft-prd-"
    And TmuxTarget is set to that name suffixed with ":0.0"

  Scenario: Git branch lookup fails
    Given the mock CommandRunner returns an error for "git rev-parse --abbrev-ref HEAD"
    When a node launches
    Then the session name's branch component is "detached"
    And the node continues running (not failed)
```

## Tasks / Subtasks

- [ ] Task 1: Create `session_naming.go` with BuildSessionName, ParseSessionName, slugifyComponent (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Subtask 1a: Implement `slugifyComponent` with ASCII filter, dash collapse, trim, per-component truncation
  - [ ] Subtask 1b: Implement `BuildSessionName` assembling slugified components + sha256 short hash
  - [ ] Subtask 1c: Implement total-length cap with proportional truncation preserving prefix and hash suffix
  - [ ] Subtask 1d: Implement `ParseSessionName` reverse operation with `ok` bool

- [ ] Task 2: Write table-driven tests in `session_naming_test.go` (AC: AC-1 through AC-6)
  - [ ] Subtask 2a: 20+ slugify cases (spaces, unicode, dashes, mixed case, empty)
  - [ ] Subtask 2b: BuildSessionName happy path + fallback matrix
  - [ ] Subtask 2c: Length cap test with 500-byte inputs
  - [ ] Subtask 2d: ParseSessionName round-trip + invalid input cases

- [ ] Task 3: Wire BuildSessionName into `executor.go` (AC: AC-7, AC-8)
  - [ ] Subtask 3a: Derive `repo` from `filepath.Base(state.exec.RepoPath)` and `nodeLabel` from the node struct
  - [ ] Subtask 3b: Look up branch via `e.runCmd(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")`, fall back to `"detached"` on error
  - [ ] Subtask 3c: Replace line 850 sessionName formula with `BuildSessionName(repo, branch, label, nodeID, time.Now().UnixNano())`
  - [ ] Subtask 3d: Keep `target := fmt.Sprintf("%s:0.0", sessionName)` unchanged

- [ ] Task 4: Update `executor_test.go` assertions (AC: AC-7, AC-8)
  - [ ] Subtask 4a: Replace any `strings.HasPrefix(name, "bmad-node-")` assertions with `ParseSessionName` checks
  - [ ] Subtask 4b: Add a test for the git-rev-parse error path asserting branch="detached" and node continues

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/bmad/session_naming.go` and modified lines in `internal/bmad/executor.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Story status updated to `done`
