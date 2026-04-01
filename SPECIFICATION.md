# Claude Conductor — Go TUI Specification

> Multi-repository Claude agent orchestration dashboard.
> A terminal-native tool for monitoring and interacting with Claude Code sessions across `~/Development` repos, with live tmux pane embedding and sub-agent drill-down.

---

## 1. Product Overview

### What It Does

Claude Conductor is a TUI dashboard that:

1. **Discovers** all git repos under `~/Development` and detects running Claude Code CLI sessions via process inspection
2. **Displays** repos in a navigable grid showing agent status, token usage, and activity sparklines
3. **Organizes** agents into hierarchical workflows — orchestrator sessions spawn sub-agents, forming a DAG
4. **Streams** live session data by tailing JSONL session files in real-time
5. **Embeds** live tmux panes directly in the terminal when agents run inside tmux — no browser, no WebSocket bridge

### Why Go

- **Single binary** — `go build` produces one artifact, no runtime dependencies
- **Native tmux** — the app runs *inside* tmux and can split panes, attach sessions, and capture output directly
- **Charm ecosystem** — bubbletea (Elm-architecture TUI), lipgloss (styling), bubbles (components) are battle-tested
- **Goroutines** — concurrent file watching, process scanning, and pane capture map naturally to goroutines
- **Remote access** — Charm's `wish` library enables `ssh dashboard@host` for team-wide visibility

### Navigation

```
┌─────────────────────────────────────────────────────────┐
│  REPOS view                                             │
│  Grid of repo cards with agent panels + sparklines      │
│  j/k or ↑↓ to select, Enter to drill in                │
└──────────────────────┬──────────────────────────────────┘
                       │ Enter
┌──────────────────────▼──────────────────────────────────┐
│  WORKFLOW view                                          │
│  Left: agent tree (orchestrator + sub-agents)           │
│  Right: log viewer for selected agent                   │
│  j/k to select agent, Enter to drill in, Esc to back   │
└──────────────────────┬──────────────────────────────────┘
                       │ Enter
┌──────────────────────▼──────────────────────────────────┐
│  AGENT view                                             │
│  Full-screen: metrics strip + live terminal/log         │
│  If tmux pane exists: split-pane to show live session   │
│  Esc to back                                            │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Architecture

```
cmd/conductor/main.go           ← entry point, CLI flags, tea.NewProgram
internal/
  tui/
    model.go                    ← root bubbletea Model (state machine)
    repos.go                    ← repos grid view
    workflow.go                 ← split-pane: agent tree + log viewer
    agent.go                    ← full-screen agent detail + tmux embed
    chrome.go                   ← header bar, breadcrumb, status bar
    theme.go                    ← lipgloss styles, color palette
  scanner/
    processes.go                ← discover Claude CLI sessions via ps + lsof
    repos.go                    ← scan ~/Development for git repos + metadata
    sessions.go                 ← parse JSONL session files for token/log data
    watcher.go                  ← fsnotify-based live JSONL tailing
  tmux/
    panes.go                    ← map agent PIDs to tmux panes
    control.go                  ← tmux control mode (-C) attach/capture
    embed.go                    ← split pane to show live agent session
  domain/
    types.go                    ← all domain structs (Agent, Repo, Workflow, etc.)
```

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | Terminal styling (colors, borders, layout) |
| `github.com/charmbracelet/bubbles` | Components (list, viewport, textinput, spinner) |
| `github.com/fsnotify/fsnotify` | File system watching for JSONL tailing |

### Data Flow

```
Process Scanner (goroutine, 5s tick)
  └─ ps -eo pid,ppid,etime,args → parse → AgentSession[]
  └─ lsof -p {pid} → working directory
  └─ git branch/status/log → RepoInfo

JSONL Watcher (goroutine per active session)
  └─ fsnotify on ~/.claude/projects/{key}/*.jsonl
  └─ tail new lines → parse → token counts + log entries + sub-agents

tmux Scanner (goroutine, 5s tick)
  └─ tmux list-panes -a → PID-to-pane mapping
  └─ walk PPID chain to match agent PIDs

All → tea.Msg → Model.Update() → Model.View() → terminal render
```

---

## 3. Domain Types

```go
type AgentStatus string

const (
    StatusRunning AgentStatus = "running"
    StatusBlocked AgentStatus = "blocked"
    StatusError   AgentStatus = "error"
    StatusQueued  AgentStatus = "queued"
    StatusDone    AgentStatus = "done"
)

type LogKind string

const (
    LogOK     LogKind = "ok"      // green  ✓  — write/edit success
    LogInfo   LogKind = "info"    // blue   ℹ  — bash commands
    LogWarn   LogKind = "warn"    // amber  ⚠  — warnings
    LogErr    LogKind = "err"     // red    ✗  — errors
    LogDim    LogKind = "dim"     // dim    ·  — reads, searches
    LogSystem LogKind = "system"  // purple ⬡  — agent/skill calls
)

type LogLine struct {
    Kind LogKind
    Text string
    Ts   time.Time
}

type Agent struct {
    ID           string       // "pid-{pid}" or "{parentID}-sub-{name}"
    Name         string       // model name or sub-agent name
    Status       AgentStatus
    TokensUsed   int64
    TokensMax    int64        // 1M for opus, 200K for others
    Model        string       // "claude-opus-4-6", "sonnet", etc.
    LogLines     []LogLine    // last 50 log entries
    Elapsed      time.Duration
    PID          int          // 0 for in-process sub-agents
    HasTmuxPane  bool
}

type DagEdge struct {
    From string // parent agent ID
    To   string // child agent ID
}

type Workflow struct {
    ID        string
    Branch    string
    Status    AgentStatus
    Agents    []Agent
    Edges     []DagEdge
    StartedAt time.Time
}

type Repo struct {
    ID               string
    Name             string
    Branch           string
    LastCommit       string
    LastCommitDate   time.Time
    Dirty            bool
    Workflow         *Workflow   // nil if no agents running
    TokenBurnHistory []int64    // last 60 samples (~60s)
    Accent           lipgloss.Color
}

// Parsed from a single JSONL session file
type SessionData struct {
    SessionID         string
    TotalTokens       int64
    InputTokens       int64
    OutputTokens      int64
    CacheReadTokens   int64
    CacheCreateTokens int64
    LogLines          []LogLine
    SubAgents         []SubAgentInfo
}

type SubAgentInfo struct {
    Name        string
    Description string
    Status      string // "running" | "done"
    Result      string // summary text (max 500 chars)
    OutputFile  string // path to full output JSONL
    LogLines    []LogLine
}

// Live process metadata
type AgentSession struct {
    PID       int
    PPID      int
    Model     string
    StartedAt time.Time
    SessionID string // empty if no --session-id flag
}

type RepoInfo struct {
    Name           string
    Branch         string
    LastCommit     string
    LastCommitDate time.Time
    Dirty          bool
    Agents         []AgentSession
}
```

---

## 4. Process Scanning

### Algorithm (`scanner/processes.go`)

**Goal**: Find all running `claude` CLI processes and map each to its working directory.

1. **Execute ps**
   ```bash
   ps -eo pid,ppid,etime,args
   ```
   Filter lines matching `\bclaude\b`, excluding:
   - `grep` itself
   - `Claude Helper`, `Claude.app/Contents/MacOS` (Electron)
   - `context-mode` (MCP daemon)
   - `/bin/zsh` (shell wrappers)

2. **Parse each line**
   ```
   Regex: ^(\d+)\s+(\d+)\s+([\d:.-]+)\s+(.+)$
   Fields: PID, PPID, ETIME, ARGS
   ```

3. **Parse etime** → `time.Duration`
   ```
   Formats: dd-hh:mm:ss | hh:mm:ss | mm:ss | ss
   startedAt = time.Now().Add(-elapsed)
   ```

4. **Extract from args**
   ```
   --model (\S+)       → model (default: "claude")
   --session-id (\S+)  → sessionID (default: "")
   ```

5. **Get working directory**
   ```bash
   lsof -p {pid} 2>/dev/null | grep cwd | awk '{print $NF}'
   ```

6. **Group by directory** → `map[string][]AgentSession`

### Repo Scanning (`scanner/repos.go`)

For each directory in `$CONDUCTOR_DEV_DIR` (default `~/Development`):
- Check `git rev-parse --is-inside-work-tree`
- Get branch: `git branch --show-current` (fallback: `git rev-parse --short HEAD`)
- Get last commit: `git log --oneline -1 --format=%s`
- Get commit date: `git log -1 --format=%ct` (epoch seconds)
- Get dirty: `git status --porcelain` (non-empty = dirty)
- Merge with agent sessions by matching repo directory

---

## 5. JSONL Session Parsing

### File Location

```
~/.claude/projects/{repo-key}/{session-id}.jsonl

repo-key = strings.ReplaceAll(repoAbsPath, "/", "-")
Example:  /Users/linus/Development/mashed
        → -Users-linus-Development-mashed
```

### Line Format

Each line is a JSON object with `"type"` field:

#### `type: "assistant"` — Claude's response

```json
{
  "type": "assistant",
  "message": {
    "usage": {
      "input_tokens": 150,
      "output_tokens": 50,
      "cache_read_input_tokens": 0,
      "cache_creation_input_tokens": 100
    },
    "content": [
      {"type": "tool_use", "id": "toolu_01...", "name": "Bash", "input": {"command": "git status"}},
      {"type": "text", "text": "The repo is clean."}
    ]
  }
}
```

**Processing**:
- Accumulate all token fields
- For `tool_use` blocks: extract tool name + detail → LogLine
- For `text` blocks: collapse newlines, skip if < 6 chars → LogLine
- For `Agent` tool_use: track as sub-agent (keyed by `block.id`)

#### `type: "user"` — Tool results / task notifications

```json
{
  "type": "user",
  "message": {
    "content": [
      {"type": "tool_result", "tool_use_id": "toolu_01...", "content": "..."}
    ]
  }
}
```

**Processing**:
- `tool_result` matching a sub-agent tool_use_id → mark sub-agent as done
- String content containing `<task-notification>` → parse sub-agent result + output file path

### Tool → LogLine Mapping

| Tool Name | LogKind | Detail Source |
|-----------|---------|---------------|
| Bash | `info` | `input.command` |
| Read | `dim` | `input.file_path` (basename) |
| Write | `ok` | `input.file_path` (basename) |
| Edit | `ok` | `input.file_path` (basename) |
| Glob | `dim` | `input.pattern` |
| Grep | `dim` | `input.pattern` |
| Agent | `system` | `input.description` |
| Skill | `system` | `input.skill` |
| *(other)* | `info` | tool name only |

MCP tool names like `mcp__server__ns__tool` → extract last segment after `__`.

### Sub-Agent Detection

When `Agent` tool_use is found:
1. Store `SubAgentInfo{Name: input.name || input.description, Status: "running"}`
2. Key by `tool_use.id`
3. When `tool_result` arrives with matching ID → status = "done"
4. When `<task-notification>` arrives → parse `<result>`, `<output-file>`
5. If output file exists, parse its JSONL for sub-agent log lines (last 50)

### Live Tailing (`scanner/watcher.go`)

Use `fsnotify` to watch `~/.claude/projects/{repo-key}/`:
- On file modify: read new lines from last known offset
- Parse incrementally (same logic as full parse)
- Send `tea.Msg` with updated session data
- Only watch repos that have active agents (avoid watching all ~50 repos)

---

## 6. DAG Construction

### Edge Detection

Two sources of parent→child relationships:

1. **Separate processes** (PPID chain):
   ```go
   pidSet := set of all agent PIDs
   for _, session := range sessions {
       if pidSet.Contains(session.PPID) {
           edges = append(edges, DagEdge{
               From: fmt.Sprintf("pid-%d", session.PPID),
               To:   fmt.Sprintf("pid-%d", session.PID),
           })
       }
   }
   ```

2. **In-process sub-agents** (Agent tool calls in JSONL):
   ```go
   // Only for orchestrator sessions (model == "claude", no explicit --model flag)
   for _, sub := range sessionData.SubAgents {
       subID := parentID + "-sub-" + sub.Name
       agents = append(agents, Agent{ID: subID, Name: sub.Name, ...})
       edges = append(edges, DagEdge{From: parentID, To: subID})
   }
   ```

### Session-to-Data Pairing

1. If session has `--session-id`: exact match by session ID
2. Fallback: sort sessions by startedAt, sort data by totalTokens desc, pair positionally

---

## 7. UI Views

### Color Palette

```go
var Theme = struct {
    BG, BG1, BG2, BG3         lipgloss.Color
    Border, Border2            lipgloss.Color
    Green, GreenDim            lipgloss.Color
    Amber, Red, Blue, Purple   lipgloss.Color
    Teal                       lipgloss.Color
    Text, TextDim, TextMuted   lipgloss.Color
}{
    BG: "#07080a", BG1: "#0d0f12", BG2: "#12151a", BG3: "#181c23",
    Border: "#1e2530", Border2: "#2a3340",
    Green: "#00e57a", GreenDim: "#006636",
    Amber: "#f0a500", Red: "#e84545", Blue: "#3d9eff", Purple: "#9d6fff",
    Teal: "#00c4b3",
    Text: "#c8d4e0", TextDim: "#4a5a6a", TextMuted: "#2e3d4d",
}

// Accent colors rotate per repo
var Accents = []lipgloss.Color{"#00e57a", "#3d9eff", "#9d6fff", "#f0a500", "#00c4b3", "#e84545"}
```

### Repos View (`tui/repos.go`)

2-column grid of repo cards. Each card contains:

```
┌─────────────────────────────────────────┐
│ ━━━━━━━━━━━━━━━━━━━━━━━ (accent bar)   │
│ repo-name                    [RUNNING]  │
│ ⎇ main                                 │
│                                         │
│ TOKEN BURN / 60s                        │
│ ▁▂▃▅▆▇█▇▅▃▂▁▂▃▅▆ (sparkline)          │
│                                         │
│ ┌─ ● claude-opus-4-6    RUNNING ──────┐ │
│ │   claude-opus-4-6         109.1M    │ │
│ └─────────────────────────────────────┘ │
│ ┌─ ● claude              RUNNING ────┐  │
│ │   claude                    8.0M   │  │
│ └────────────────────────────────────┘  │
│   ┌─ · arch-reviewer          DONE ─┐  │
│   └──────────────────────────────────┘  │
│   ┌─ · quality-reviewer       DONE ─┐  │
│   └──────────────────────────────────┘  │
│                                         │
│ 3          116.2M         696m 32s      │
│ agents     tokens         elapsed       │
└─────────────────────────────────────────┘
```

- **Agent panels**: thin border per agent, left accent stripe colored by status
- **Sub-agents**: indented, smaller, dimmer border
- **Sparkline**: last 60 token-total samples rendered as block characters (▁▂▃▄▅▆▇█)
- **Status badge**: colored pill (RUNNING/BLOCKED/ERROR/DONE/QUEUED)

### Workflow View (`tui/workflow.go`)

Split layout using lipgloss `JoinHorizontal`:

```
┌── AGENTS ──────────────┬── TERMINAL ─────────────────────────────┐
│                        │                                         │
│ ▾ ● claude-opus-4-6   │  claude-opus-4-6          RUNNING       │
│     RUNNING (5)        │  claude-opus-4-6    109.1M / 1M         │
│   ┃                    │  ─────────────────────────────────────   │
│   ├─ · arch-reviewer   │                                         │
│   │    DONE            │  08:15:23  ℹ  Bash: git status          │
│   ├─ · quality-rev     │  08:15:24  ✓  Edit: main.go             │
│   │    DONE            │  08:15:25  ·  Read: config.yaml          │
│   ├─ · security-rev    │  08:15:26  ⬡  Agent: researcher         │
│   │    DONE            │  08:15:27  ✓  Write: handler.go          │
│   └─ · ux-reviewer     │  08:15:28  ✗  Bash: go test ./...       │
│        DONE            │  08:15:30  ✓  Edit: handler_test.go      │
│                        │  ▊                                       │
│ ● claude               │                                         │
│   RUNNING              │                                         │
│                        │                                         │
└────────────────────────┴─────────────────────────────────────────┘
```

- **Left pane** (fixed 30-col width): collapsible agent tree using bubbles `list`
  - Root agents: expand/collapse chevron + status dot + name + sub-count
  - Sub-agents: tree connector lines (┃├─└─) + status dot + name
  - Click/Enter to select, shown highlighted
- **Right pane**: log viewport for selected agent using bubbles `viewport`
  - Timestamp (dim) + prefix glyph (colored) + text
  - Auto-scroll to bottom, scroll up to see history
  - Blinking cursor block when agent is running

### Agent View (`tui/agent.go`)

Full-screen detail for a single agent:

```
┌── STATUS ──── MODEL ──────── TOKENS ──────── ELAPSED ────────────┐
│  RUNNING      opus           109.1M / 1M     1328m 25s           │
│  ████████████████████████████░░░░░░░░░░░ (progress bar)          │
├──────────────────────────────────────────────────────────────────┤
│  [LIVE] TERMINAL                    click to focus / Esc to exit │
│                                                                   │
│  (live tmux pane content OR scrollable log lines)                │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

**Two modes**:
1. **Live tmux** (when agent has a tmux pane): use `tmux capture-pane -t {target} -p -e` polled at 200ms to show real terminal content. On Enter/focus: `tmux select-pane -t {target}` to switch to the live pane.
2. **Log fallback**: scrollable viewport of parsed JSONL log lines (same as workflow right pane)

### Chrome (`tui/chrome.go`)

- **Header**: `⬡ CLAUDE CONDUCTOR` centered, clock right-aligned
- **Breadcrumb**: `repos › repo-name › agent-name` with navigation
- **Status bar**: mode badge (REPOS/WORKFLOW/AGENT), keyboard hints, `N agents running` with pulse

---

## 8. tmux Integration

### Overview

The Go app runs inside tmux. This is the key advantage over the web-based approach — tmux pane management is native.

### Pane Discovery (`tmux/panes.go`)

```bash
tmux list-panes -a -F '#{pane_pid}:#{pane_id}:#{session_name}:#{window_index}.#{pane_index}'
```

Parse into:
```go
type PaneInfo struct {
    PanePID int
    PaneID  string // e.g. "%3"
    Target  string // e.g. "work:0.1"
}
```

### PID-to-Pane Matching (`tmux/panes.go`)

Agent PIDs are not always direct children of tmux panes (shell → node → claude). Walk the PPID chain:

```go
func FindPane(pid int, panes []PaneInfo) *PaneInfo {
    // Direct match
    for _, p := range panes {
        if p.PanePID == pid { return &p }
    }
    // Walk PPID chain (up to 8 levels)
    current := pid
    for i := 0; i < 8; i++ {
        ppid := getParentPID(current) // ps -o ppid= -p {current}
        if ppid <= 1 { break }
        for _, p := range panes {
            if p.PanePID == ppid { return &p }
        }
        current = ppid
    }
    return nil
}
```

Cache pane list for 5 seconds to avoid repeated `tmux list-panes` calls.

### Pane Content Capture (`tmux/control.go`)

For read-only terminal display in the agent view:

```bash
tmux capture-pane -t {target} -p -e    # -e preserves ANSI escape sequences
```

Poll every 200ms when the agent view is active. Parse ANSI escapes for colored rendering in lipgloss.

### Live Pane Switching (`tmux/embed.go`)

When the user wants to interact with an agent's terminal:

```bash
tmux select-pane -t {target}      # switch focus to the agent's pane
# OR
tmux split-window -t {target}     # show agent pane alongside dashboard
```

The dashboard can also create new panes for agents:

```bash
tmux split-window -h "claude --model opus --dangerously-skip-permissions"
```

---

## 9. Keyboard Shortcuts

| Key | Repos View | Workflow View | Agent View |
|-----|-----------|---------------|------------|
| `j` / `↓` | Next repo | Next agent | Scroll down |
| `k` / `↑` | Prev repo | Prev agent | Scroll up |
| `Enter` | → Workflow | → Agent | Focus tmux pane |
| `Esc` | — | → Repos | → Workflow |
| `q` | Quit | → Repos | → Workflow |
| `Tab` | Switch column | Toggle expand | — |
| `/` | Filter repos | Filter agents | — |
| `?` | Help overlay | Help overlay | Help overlay |

---

## 10. Configuration

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONDUCTOR_DEV_DIR` | `~/Development` | Root directory to scan for git repos |
| `HOME` | (system) | Used to locate `~/.claude/projects/` |

### Paths

| Path | Purpose |
|------|---------|
| `$CONDUCTOR_DEV_DIR/` | Git repos to scan |
| `~/.claude/projects/` | Claude Code session storage |
| `~/.claude/projects/{repo-key}/` | Per-repo session directory |
| `~/.claude/projects/{repo-key}/{uuid}.jsonl` | Single session log |

**Repo key derivation**:
```
/Users/linus/Development/mashed → -Users-linus-Development-mashed
```
(Replace all `/` with `-`)

---

## 11. Concurrency Model

```
main goroutine
  └─ tea.NewProgram(model)
       ├─ Process Scanner (tea.Tick, 5s)
       │    └─ exec ps, lsof, git commands
       │    └─ sends ScanResultMsg
       ├─ JSONL Watcher (long-running goroutine)
       │    └─ fsnotify watches active session files
       │    └─ sends SessionUpdateMsg per new line
       ├─ tmux Pane Scanner (tea.Tick, 5s)
       │    └─ exec tmux list-panes
       │    └─ sends PaneUpdateMsg
       ├─ Pane Capture (tea.Tick, 200ms, only when agent view active)
       │    └─ exec tmux capture-pane
       │    └─ sends PaneCaptureMsg
       └─ Token Burn Sampler (tea.Tick, 1s)
            └─ samples total tokens per repo
            └─ sends BurnSampleMsg
```

All external commands run via `tea.Cmd` returning `tea.Msg` — no shared mutable state, no locks.

---

## 12. Build & Distribution

```bash
# Development
go run ./cmd/conductor

# Build
go build -o conductor ./cmd/conductor

# Install
go install ./cmd/conductor

# Run (must be inside tmux for pane features)
conductor
conductor --dev-dir /path/to/repos
```

### CLI Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--dev-dir` | `~/Development` | Override repo scan directory |
| `--poll-interval` | `5s` | Process/repo scan interval |
| `--no-tmux` | `false` | Disable tmux features (pure log mode) |

---

## 13. Implementation Phases

### Phase 1: Foundation
- [ ] Project scaffolding (`cmd/conductor/main.go`, `go.mod`)
- [ ] Domain types (`domain/types.go`)
- [ ] Process scanner (`scanner/processes.go`)
- [ ] Repo scanner (`scanner/repos.go`)
- [ ] Root bubbletea model with view switching

### Phase 2: Repos View
- [ ] 2-column grid layout with lipgloss
- [ ] Repo card rendering (name, branch, status badge)
- [ ] Agent panels with thin borders inside each card
- [ ] Sparkline component (block characters)
- [ ] Stats footer (agents, tokens, elapsed)
- [ ] Keyboard navigation (j/k, Enter to drill in)

### Phase 3: Session Parsing
- [ ] JSONL parser (`scanner/sessions.go`)
- [ ] Token accumulation
- [ ] Tool call → log line extraction
- [ ] Sub-agent detection (Agent tool_use + task-notification)
- [ ] DAG edge construction (PPID + in-process)

### Phase 4: Workflow View
- [ ] Split-pane layout
- [ ] Agent tree with collapsible groups
- [ ] Log viewport with colored output
- [ ] Agent selection + keyboard navigation

### Phase 5: Live Tailing
- [ ] fsnotify watcher (`scanner/watcher.go`)
- [ ] Incremental JSONL parsing
- [ ] Real-time token count updates
- [ ] Auto-scroll in log viewport

### Phase 6: tmux Integration
- [ ] Pane discovery + PID mapping (`tmux/panes.go`)
- [ ] capture-pane polling (`tmux/control.go`)
- [ ] ANSI rendering in agent view
- [ ] Pane switching on Enter (`tmux/embed.go`)

### Phase 7: Polish
- [ ] Chrome (header, breadcrumb, status bar)
- [ ] Help overlay (`?`)
- [ ] Filter/search (`/`)
- [ ] Graceful degradation when not in tmux
- [ ] Error handling for missing repos, dead sessions

---

*End of specification.*
