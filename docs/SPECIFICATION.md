# Mashed — Desktop Application Specification

> Notification-first IDE for multi-agent development.
> A Wails v2 desktop app (Go backend + Svelte frontend) for monitoring, orchestrating, and interacting with Claude Code sessions across `~/Development` repos, with live terminal embedding, workflow automation, and integrated code editing.

---

## Changelog

### BMAD Terminal Bridge (bridge-01 through bridge-04)

- BMAD tmux sessions are now named `bmad-{repo}-{branch}-{label}-{hash}` (e.g. `bmad-surfseer-main-create-story-a1b2c3d4`) instead of the cryptic `bmad-{nodeID}-{unix}` format.
- **Migration note:** Sessions created by older builds cannot be re-attached via View Terminal — they must be restarted. Orphaned `bmad-*` sessions from prior runs are automatically cleaned up on next app startup.
- View Terminal now works for running BMAD nodes, streaming live Claude CLI output through a new `TmuxAdapter` bridge (FIFO-based with polling fallback).
- The terminal modal title shows the parsed friendly form (`Terminal — repo · branch · label`) instead of the raw tmux target.

---

## 1. Product Overview

### What It Does

Mashed is a desktop application that:

1. **Discovers** all git repos under a configurable development directory and detects running Claude Code CLI sessions via process inspection
2. **Displays** agents in a notification feed showing status, token usage, model, and activity sparklines
3. **Orchestrates** multi-step workflows via the BMAD engine — a visual DAG-based automation system with process nodes, conditions, loops, transforms, and merge nodes
4. **Streams** live session data by tailing JSONL session files in real-time via fsnotify
5. **Embeds** live terminal sessions via a WebSocket bridge to xterm.js — agents run in tmux, rendered in the desktop app
6. **Edits** code with an integrated Monaco editor, diff viewer, and file tree browser
7. **Manages** git operations — branches, commits, pushes, PRs, merges, worktrees — all from the UI

### Why Wails + Go

- **Single binary** — Wails compiles Go + Svelte into one native app, no runtime dependencies
- **Native tmux** — Go spawns and manages tmux sessions directly for agent isolation
- **WebSocket terminal** — `creack/pty` + `gorilla/websocket` bridge real terminal sessions to xterm.js
- **Goroutines** — concurrent process scanning, file watching, and session parsing map naturally to goroutines
- **Svelte frontend** — reactive UI with Monaco editor, xyflow canvas, xterm.js terminal, all in a frameless native window

### Design Philosophy

Industrial/Utilitarian aesthetic. Linear meets Bloomberg Terminal. Neon green (`#00e57a`) accent, dark-only (`#07080a` base), Geist typography, compact density. See `DESIGN.md` for the full design system.

---

## 2. Architecture

### Project Structure

```
main.go                         ← Wails entry point, window config, asset embedding
app.go                          ← App struct, startup/shutdown, config, theme management
app_scan.go                     ← Process scanning loop, agent discovery, notification engine
app_sessions.go                 ← JSONL session watching, parsing, status inference
app_tmux.go                     ← Agent/terminal spawning, kill, log retrieval
app_git.go                      ← Git operations (branches, commits, push, PR, merge, diff, worktrees)
app_bmad.go                     ← BMAD workflow CRUD, execution control, agent/module management
app_explain.go                  ← AI-powered diff explanation via Claude CLI
font_scanner.go                 ← Local font discovery + Nerd Fonts catalog
theme_scanner.go                ← VSCodium theme import + conversion
internal/
  agent/
    engine.go                   ← NotificationEngine — event classification and emission
  bmad/
    types.go                    ← WorkflowDef, WorkflowNode, WorkflowEdge, NodeType, etc.
    storage.go                  ← Workflow/agent persistence to ~/.mashed/
    executor.go                 ← DAG-based workflow execution (dynamic ready-set algorithm)
    registry.go                 ← Process catalog (built-in BMAD processes by phase)
    modules.go                  ← Module catalog (reusable BMAD modules)
    templates.go                ← Built-in workflow templates
    condition.go                ← Condition evaluation engine (comparisons, regex, contains)
    sprint.go                   ← Sprint status YAML parsing/updating
    artifacts.go                ← Artifact path resolution and verification
  domain/
    types.go                    ← All domain structs (Agent, Repo, Workflow, SessionData, etc.)
  explain/
    explain.go                  ← Diff explanation via Claude CLI subprocess
  git/
    diff.go                     ← Git diff parsing, scoped diffs
    worktree.go                 ← Git worktree management
    errors.go                   ← Git-specific error types
  scanner/
    processes.go                ← Discover Claude CLI sessions via ps + lsof
    repos.go                    ← Scan dev directory for git repos + metadata
    sessions.go                 ← Parse JSONL session files for token/log data
    watcher.go                  ← fsnotify-based live JSONL tailing
    claude.go                   ← ClaudeCodeProvider (AgentProvider implementation)
    errors.go                   ← Scanner-specific error types
  terminal/
    bridge.go                   ← WebSocket bridge (Go pty ↔ xterm.js in browser)
    panes.go                    ← tmux pane discovery, PID-to-pane mapping
frontend/
  src/
    App.svelte                  ← Root component, view routing
    views/
      Setup.svelte              ← First-run dev directory picker
      NotificationFeed.svelte   ← Main dashboard — agent notification stream
      AgentDetail.svelte        ← Agent drill-down with terminal + logs + diff
      WorkflowBuilder.svelte    ← BMAD visual workflow canvas
      Settings.svelte           ← Theme, font, path configuration
      SpawnAgent.svelte         ← New agent session launcher
      NewSessionModal.svelte    ← Session creation modal
      BranchModal.svelte        ← Branch creation
      SwitchBranchModal.svelte  ← Branch switching
      MergeModal.svelte         ← Branch merge
      ForcePushModal.svelte     ← Force push confirmation
    components/
      TitleBar.svelte           ← Frameless window title bar with controls
      Terminal.svelte           ← xterm.js terminal (WebSocket to Go bridge)
      MonacoEditor.svelte       ← Monaco editor integration
      CodeEditor.svelte         ← Lightweight code editor
      DiffView.svelte           ← Side-by-side diff viewer
      FileTree.svelte           ← File browser tree
      SparkLine.svelte          ← Token burn sparkline (block characters)
      StatusBadge.svelte        ← Agent status pill badge
      NewRepoModal.svelte       ← Repository creation dialog
      bmad/
        CanvasPane.svelte       ← xyflow DAG canvas for workflow builder
        ProcessNode.svelte      ← Standard process node
        ConditionNode.svelte    ← If/else branch node
        LoopNode.svelte         ← Loop N times node
        LoopUntilNode.svelte    ← Loop until condition node
        TransformNode.svelte    ← Data extraction/transform node
        MergeNode.svelte        ← Branch merge node
        DeletableEdge.svelte    ← Edge with delete button
        ProcessSidebar.svelte   ← Process catalog sidebar
        NodeConfigPanel.svelte  ← Node configuration editor
        ExecutionBar.svelte     ← Run/pause/stop controls
        TemplatePicker.svelte   ← Workflow template selection
        AgentConfigModal.svelte ← Custom agent configuration
        RepoContextBar.svelte   ← Repository context display
        SprintPanel.svelte      ← Sprint status panel
        OutputViewerModal.svelte← Node output inspection
    lib/
      stores/theme.js           ← Theme reactive store
      stores/font.js            ← Font reactive store
      themes.js                 ← Built-in theme definitions
      themeConverter.js         ← VSCodium → Mashed theme conversion
      themeInit.js              ← Theme initialization on startup
      monacoTheme.js            ← Monaco editor theme adapter
      fileTree.js               ← File tree data structures
      sprintColors.js           ← BMAD sprint status colors
```

### Key Dependencies

**Go (go.mod)**

| Package | Purpose |
|---------|---------|
| `github.com/wailsapp/wails/v2` | Desktop app framework (Go ↔ Svelte bridge) |
| `github.com/creack/pty` | Pseudo-terminal allocation for embedded terminals |
| `github.com/gorilla/websocket` | WebSocket server for terminal bridge |
| `github.com/fsnotify/fsnotify` | File system watching for live JSONL tailing |
| `github.com/stretchr/testify` | Test assertions and mocking |
| `gopkg.in/yaml.v3` | YAML parsing (sprint status, configs) |

**Frontend (package.json)**

| Package | Purpose |
|---------|---------|
| `svelte` ^4.2 | Reactive UI framework |
| `vite` ^5.0 | Build tool and dev server |
| `@xyflow/svelte` ^0.1 | Node-based workflow canvas (DAG visualization) |
| `monaco-editor` ^0.55 | Code editor (syntax highlighting, diff view) |
| `@xterm/xterm` 5.5 | Terminal emulator in the browser |
| `@xterm/addon-fit` | Auto-resize terminal to container |
| `@xterm/addon-canvas` | Canvas-based terminal renderer |
| `lucide-svelte` | Icon library |

### Data Flow

```
Process Scanner (goroutine, 5s tick)
  └─ ps -eo pid,ppid,etime,args → parse → AgentSession[]
  └─ lsof -p {pid} → working directory
  └─ git rev-parse --show-toplevel → repo root
  └─ git branch/status/log → RepoInfo

JSONL Watcher (goroutine, fsnotify)
  └─ watches ~/.claude/projects/{repo-key}/*.jsonl
  └─ on file change → triggers doScan()
  └─ parse session → token counts + log entries + sub-agents + status inference

NotificationEngine (goroutine, event channel)
  └─ receives Agent updates from scanner
  └─ classifies events (needs_response, error, completed, running, started)
  └─ emits prioritized NotificationEvent to frontend via Wails events

Terminal Bridge (goroutine, WebSocket server)
  └─ frontend connects xterm.js via WebSocket
  └─ Go side: creack/pty attaches to tmux pane
  └─ bidirectional: keystrokes → pty, output → xterm.js

tmux Pane Discovery (on-demand, cached)
  └─ tmux list-panes -a → PID-to-pane mapping
  └─ walk PPID chain (up to 8 levels) to match agent PIDs

All → Wails runtime.EventsEmit() → Svelte frontend reactive updates
```

---

## 3. Domain Types

```go
type AgentStatus string

const (
    StatusRunning  AgentStatus = "running"  // actively processing (tool calls, generating)
    StatusOpen     AgentStatus = "open"     // idle — Claude spoke last, not waiting for user
    StatusFinished AgentStatus = "finished" // task complete, awaiting next instruction
    StatusWaiting  AgentStatus = "waiting"  // actively waiting for user response (AskUserQuestion)
    StatusBlocked  AgentStatus = "blocked"  // legacy — maps to waiting
    StatusError    AgentStatus = "error"
    StatusQueued   AgentStatus = "queued"
    StatusDone     AgentStatus = "done"
)

type LogKind string

const (
    LogOK     LogKind = "ok"     // green  ✓  — write/edit success
    LogInfo   LogKind = "info"   // blue   ℹ  — bash commands
    LogWarn   LogKind = "warn"   // amber  ⚠  — warnings
    LogErr    LogKind = "err"    // red    ✗  — errors
    LogDim    LogKind = "dim"    // dim    ·  — reads, searches
    LogSystem LogKind = "system" // purple ⬡  — agent/skill calls
)

type LogLine struct {
    Kind LogKind   `json:"kind"`
    Text string    `json:"text"`
    Ts   time.Time `json:"ts"`
}

type Agent struct {
    ID           string        `json:"id"`          // "pid-{pid}" or "{parentID}-sub-{name}-{toolUseID}"
    Name         string        `json:"name"`        // model name or sub-agent name
    Status       AgentStatus   `json:"status"`
    TokensUsed   int64         `json:"tokensUsed"`
    TokensMax    int64         `json:"tokensMax"`   // 1M for opus, 200K for others
    Model        string        `json:"model"`       // "claude-opus-4-6", "sonnet", etc.
    LogLines     []LogLine     `json:"logLines"`    // last 50 log entries
    Elapsed      time.Duration `json:"elapsed"`
    PID          int           `json:"pid"`         // 0 for in-process sub-agents
    HasTmuxPane  bool          `json:"hasTmuxPane"`
    TmuxTarget   string        `json:"tmuxTarget"`  // tmux pane target string
    RepoPath     string        `json:"repoPath"`    // working directory path
    SubAgentInfo *SubAgentInfo `json:"-"`           // set when this agent represents a sub-agent
}

type DagEdge struct {
    From string `json:"from"` // parent agent ID
    To   string `json:"to"`   // child agent ID
}

type Workflow struct {
    ID        string      `json:"id"`
    Branch    string      `json:"branch"`
    Status    AgentStatus `json:"status"`
    Agents    []Agent     `json:"agents"`
    Edges     []DagEdge   `json:"edges"`
    StartedAt time.Time   `json:"startedAt"`
}

type Repo struct {
    ID               string    `json:"id"`
    Name             string    `json:"name"`
    Path             string    `json:"path"`
    Branch           string    `json:"branch"`
    LastCommit       string    `json:"lastCommit"`
    LastCommitDate   time.Time `json:"lastCommitDate"`
    Dirty            bool      `json:"dirty"`
    Workflow         *Workflow `json:"workflow,omitempty"`
    TokenBurnHistory [60]int64 `json:"-"` // ring buffer, last 60 samples
    BurnIndex        int       `json:"-"` // current write position
}

type SessionData struct {
    SessionID         string         `json:"sessionId"`
    TotalTokens       int64          `json:"totalTokens"`
    InputTokens       int64          `json:"inputTokens"`
    OutputTokens      int64          `json:"outputTokens"`
    CacheReadTokens   int64          `json:"cacheReadTokens"`
    CacheCreateTokens int64          `json:"cacheCreateTokens"`
    LogLines          []LogLine      `json:"logLines"`
    SubAgents         []SubAgentInfo `json:"subAgents"`
    LastMessageType   string         `json:"lastMessageType"`   // "assistant" or "user"
    LastToolName      string         `json:"lastToolName"`      // last tool_use name
    HasPendingToolUse bool           `json:"hasPendingToolUse"` // tool_use awaiting result
}

type SubAgentInfo struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    ToolUseID   string    `json:"toolUseId"`
    Status      string    `json:"status"`      // "running" | "done"
    Result      string    `json:"result"`      // summary (max 500 chars)
    OutputFile  string    `json:"outputFile"`  // path to sub-agent JSONL
    LogLines    []LogLine `json:"logLines"`
}

type AgentSession struct {
    PID       int       `json:"pid"`
    PPID      int       `json:"ppid"`
    Model     string    `json:"model"`
    StartedAt time.Time `json:"startedAt"`
    SessionID string    `json:"sessionId"` // empty if no --session-id flag
}

type RepoInfo struct {
    Name           string         `json:"name"`
    Path           string         `json:"path"`
    Branch         string         `json:"branch"`
    LastCommit     string         `json:"lastCommit"`
    LastCommitDate time.Time      `json:"lastCommitDate"`
    Dirty          bool           `json:"dirty"`
    Agents         []AgentSession `json:"agents"`
}

// NotificationEvent is a single event surfaced to the developer.
type NotificationEvent struct {
    ID         string    `json:"id"`
    AgentID    string    `json:"agentId"`
    AgentName  string    `json:"agentName"`
    Model      string    `json:"model"`
    RepoName   string    `json:"repoName"`
    RepoPath   string    `json:"repoPath"`
    RepoBranch string    `json:"repoBranch"`
    EventType  EventType `json:"eventType"`
    Summary    string    `json:"summary"`
    Timestamp  time.Time `json:"timestamp"`
    Read       bool      `json:"read"`
    Priority   int       `json:"priority"`   // 0=needs-response .. 3=running
    TokensUsed int64     `json:"tokensUsed"`
    TokensMax  int64     `json:"tokensMax"`
    TmuxTarget string    `json:"tmuxTarget"`
    PID        int       `json:"pid"`

    // Sub-agent fields
    IsSubAgent       bool      `json:"isSubAgent"`
    ParentAgentID    string    `json:"parentAgentId,omitempty"`
    SubAgentName     string    `json:"subAgentName,omitempty"`
    SubAgentDesc     string    `json:"subAgentDesc,omitempty"`
    SubAgentStatus   string    `json:"subAgentStatus,omitempty"`
    SubAgentResult   string    `json:"subAgentResult,omitempty"`
    SubAgentLogLines []LogLine `json:"subAgentLogLines,omitempty"`
}

// AgentProvider abstracts the agent platform (Claude Code, future: Aider, Codex, Gemini).
type AgentProvider interface {
    ScanProcesses() ([]AgentSession, error)
    GetWorkingDir(pid int) (string, error)
    ParseSession(path string) (*SessionData, error)
    WatchSessions(ctx context.Context) (<-chan SessionEvent, error)
    SessionDir(repoPath string) string
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

6. **Resolve to repo root**
   ```bash
   git -C {dir} rev-parse --show-toplevel
   ```

7. **Filter**: Only keep agents whose repo root is under the configured dev directory

### Repo Scanning (`scanner/repos.go`)

For each directory in `$devDir` (configured at first launch):
- Check `git rev-parse --is-inside-work-tree`
- Get branch: `git branch --show-current` (fallback: `git rev-parse --short HEAD`)
- Get last commit: `git log --oneline -1 --format=%s`
- Get commit date: `git log -1 --format=%ct` (epoch seconds)
- Get dirty: `git status --porcelain` (non-empty = dirty)

### Session-to-Agent Pairing

Sessions are sorted newest-first. For each agent:
1. If agent has `--session-id`: exact match by session ID in the session directory
2. Fallback: claim the most recently modified unclaimed `.jsonl` file in the session directory
3. Claimed sessions are tracked per-scan to prevent two agents from sharing one file

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
- Track `LastMessageType`, `LastToolName`, `HasPendingToolUse` for status inference

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
- Clears `HasPendingToolUse` flag

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

### Status Inference (`app_sessions.go:inferStatus`)

Based on parsed session state:

| Condition | Status |
|-----------|--------|
| No log lines | `running` |
| Last log line is error | `error` |
| Last tool was `AskUserQuestion` + pending | `waiting` |
| Has pending tool_use | `running` |
| Last message from assistant, info log kind | `finished` |
| Last message from assistant, other | `open` |
| Default | `running` |

### Sub-Agent Detection

When `Agent` tool_use is found:
1. Store `SubAgentInfo{Name: input.name || input.description, Status: "running"}`
2. Key by `tool_use.id`
3. When `tool_result` arrives with matching ID → status = "done"
4. When `<task-notification>` arrives → parse `<result>`, `<output-file>`

### Live Tailing (`scanner/watcher.go`)

Uses `fsnotify` to watch `~/.claude/projects/{repo-key}/`:
- On file change: triggers a full scan cycle (`doScan()`)
- Parse incrementally via the session parser
- Updated data flows through NotificationEngine → Wails events → Svelte

---

## 6. Notification Engine

### Overview (`internal/agent/engine.go`)

The NotificationEngine sits between the scanner and the frontend. It:
- Receives `Agent` updates from `doScan()`
- Classifies each update into an `EventType` with priority
- Emits `NotificationEvent` to the Wails frontend via `runtime.EventsEmit()`
- Maintains a channel-based event stream consumed by `consumeEngineEvents()`

### Event Types and Priority

| Priority | EventType | Trigger |
|----------|-----------|---------|
| 0 | `needs_response` | AskUserQuestion, permission prompt |
| 1 | `error` | Tool error, test failure |
| 2 | `completed` | Session ended cleanly |
| 3 | `running` | Actively producing output |
| 4 | `started` | New agent detected |

### Agent Pruning

Each scan cycle tracks which agent PIDs are still alive. Agents whose processes have exited are pruned from the notification list. Sub-agents are pruned when their parent is gone.

---

## 7. Terminal Bridge

### WebSocket Architecture (`internal/terminal/bridge.go`)

The terminal bridge enables the Svelte frontend to display live terminal sessions:

1. **Go side**: Starts a WebSocket server on a dynamic port at startup
2. **Frontend**: `Terminal.svelte` connects xterm.js to the WebSocket URL
3. **Connection flow**:
   - Frontend sends tmux target string on connect
   - Go attaches to the tmux pane via `creack/pty`
   - Bidirectional streaming: keystrokes → pty stdin, pty stdout → xterm.js

### tmux Pane Discovery (`internal/terminal/panes.go`)

```bash
tmux list-panes -a -F '#{pane_pid}:#{pane_id}:#{session_name}:#{window_index}.#{pane_index}'
```

Agent PIDs are not always direct children of tmux panes (shell → node → claude). The discovery walks the PPID chain up to 8 levels to find a matching pane.

Pane list is cached (5s TTL) to avoid repeated `tmux list-panes` calls. Cache is invalidated when new sessions are spawned.

### Agent Spawning (`app_tmux.go`)

```go
// SpawnAgent creates a new Claude session in a tmux pane
func (a *App) SpawnAgent(repoPath, model string) (string, error)

// SpawnTerminal creates a plain shell session
func (a *App) SpawnTerminal(repoPath string) (string, error)
```

tmux sessions are named `mashed-{repoName}-{timestamp}` (agents) or `term-{repoName}-{timestamp}` (terminals).

---

## 8. BMAD Workflow Engine

### Overview

BMAD (Breakthrough Method of Agile AI-driven Development) is the visual workflow automation system. Users build DAGs of process nodes on an xyflow canvas, then execute them to orchestrate multi-step Claude agent workflows.

### Workflow Definition (`internal/bmad/types.go`)

```go
type WorkflowDef struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Nodes       []WorkflowNode `json:"nodes"`
    Edges       []WorkflowEdge `json:"edges"`
    IsTemplate  bool           `json:"isTemplate"`
    TemplateID  string         `json:"templateId,omitempty"`
    RepoPath    string         `json:"repoPath,omitempty"`
    CreatedAt   string         `json:"createdAt"`
    UpdatedAt   string         `json:"updatedAt"`
}
```

### Node Types

| Type | Purpose | Execution |
|------|---------|-----------|
| `process` | Standard BMAD process (analysis, architecture, etc.) | Spawns Claude in tmux |
| `condition` | If/else branch based on output | Evaluates condition, activates one branch |
| `loop` | Repeat N times | Mini ready-set loop over body nodes |
| `loopUntil` | Repeat until condition met | Loop with condition check each iteration |
| `transform` | Extract/transform data | Synchronous — regex or line extraction |
| `merge` | Join branches | Waits for all incoming edges |

### Execution Engine (`internal/bmad/executor.go`)

Uses a **dynamic ready-set algorithm** (`runDynamic()`):

1. Build in-degree map from edges
2. Initialize ready set: nodes with in-degree 0
3. While ready set is non-empty:
   - Execute all ready nodes concurrently
   - On completion: decrement in-degree of downstream nodes
   - Move newly-ready nodes (in-degree 0) to the ready set
4. Control flow nodes have special handling:
   - **Conditions**: `activeOutEdges()` filters by `SourceHandle` to route branches
   - **Loops**: `executeLoopNode()` runs a mini ready-set loop for body nodes
   - **Transforms**: synchronous, no tmux session
   - **Merges**: standard in-degree wait behavior

`topoSort()` is kept only for cycle detection.

### Process Registry (`internal/bmad/registry.go`)

Processes are organized by BMAD lifecycle phase:
- `analysis` — requirements, stakeholder mapping
- `architecture` — system design, data modeling
- `planning` — sprint planning, story creation
- `implementation` — code generation, testing
- `review` — code review, QA

### Storage (`internal/bmad/storage.go`)

- Workflows: `~/.mashed/workflows/{id}.json`
- Agents: `~/.mashed/bmad-agents/{id}.json`
- Per-repo workflows via `RepoPath` field

### Artifact System (`internal/bmad/artifacts.go`)

- Artifacts resolve to `{repoPath}/_bmad-output/{category}/{artifact}` via a canonical path map
- `ResolveArtifactPath(name, repoPath)` → full path
- `VerifyArtifacts(repoPath, outputNames)` → checks existence
- Unmapped artifacts (`"code"`, `"tests"`) return `""` — handled by callers

### Templates (`internal/bmad/templates.go`)

6 built-in workflow templates that can be deep-copied into user workflows via `CreateFromTemplate()`.

---

## 9. UI Views

### Setup View (`Setup.svelte`)
First-run experience. Native directory picker dialog to select the development root. Persists to `~/.mashed/config.json`.

### Notification Feed (`NotificationFeed.svelte`)
Main dashboard. Priority-sorted list of agent notification events:
- Left accent stripe colored by status
- Repo name, agent model, summary text, timestamp
- Click to drill into agent detail
- Token usage display, sparklines

### Agent Detail (`AgentDetail.svelte`)
Full agent view with:
- Live terminal via xterm.js (WebSocket to Go bridge)
- Parsed log lines from JSONL session
- Token usage metrics
- File tree and diff viewer
- Monaco editor for code inspection
- Git operations (commit, push, PR, branch management)

### Workflow Builder (`WorkflowBuilder.svelte`)
Visual DAG editor powered by `@xyflow/svelte`:
- **Canvas** (`CanvasPane.svelte`): drag-and-drop node placement, edge connections
- **Sidebar** (`ProcessSidebar.svelte`): process catalog organized by BMAD phase
- **Config panel** (`NodeConfigPanel.svelte`): edit node parameters
- **Execution bar** (`ExecutionBar.svelte`): start/pause/stop workflow execution
- **Template picker** (`TemplatePicker.svelte`): start from built-in templates
- **Node types**: ProcessNode, ConditionNode, LoopNode, LoopUntilNode, TransformNode, MergeNode
- **Sprint panel** (`SprintPanel.svelte`): sprint status from YAML

### Settings (`Settings.svelte`)
- Theme selection (built-in + VSCodium import)
- Font selection (system fonts + Nerd Fonts)
- Font size
- Dev directory path
- VSCodium extension path

---

## 10. Wails Bindings (Go → Svelte API)

All public methods on the `App` struct are exposed to the Svelte frontend via Wails bindings.

### Configuration
| Method | Purpose |
|--------|---------|
| `PickDirectory()` | Native OS directory picker dialog |
| `SetDevDir(dir)` | Save dev directory and start scanning |
| `GetDevDir()` | Current dev directory |
| `GetConfig()` | Full persisted config |
| `SetTheme(id)` | Persist theme selection |
| `SetMonoFont(family)` | Persist font selection |
| `SetFontSize(size)` | Persist font size |
| `SetVSCodiumExtPath(path)` | Persist VSCodium path |

### Theme Management
| Method | Purpose |
|--------|---------|
| `GetSavedThemes()` | All imported themes as JSON |
| `SaveTheme(id, json)` | Persist converted theme |
| `RemoveTheme(id)` | Delete saved theme |
| `ListVSCodiumThemes()` | Discover VSCodium color themes |
| `ReadThemeFile(path)` | Read raw theme JSON |
| `SetImportedTheme(path)` | Import a VSCodium theme |

### Font Discovery
| Method | Purpose |
|--------|---------|
| `ListLocalFonts()` | System monospace fonts |
| `ListNerdFonts()` | Nerd Fonts catalog |
| `GetFontsDir()` | Font installation directory |
| `OpenFontsDir()` | Open font dir in Finder |

### Agent Management
| Method | Purpose |
|--------|---------|
| `GetNotifications()` | Priority-sorted notification list |
| `SpawnAgent(repoPath, model)` | Start Claude in new tmux session |
| `SpawnAgentWithCommand(repoPath, cmd)` | Start with custom CLI command |
| `SpawnTerminal(repoPath)` | Start plain shell session |
| `KillAgent(agentID, pid, tmuxTarget)` | Terminate agent + tmux session |
| `GetAgentLog(repoPath)` | Parsed log lines for latest session |
| `MarkRead(agentID)` | Mark notification as read |
| `GetTerminalPort()` | WebSocket bridge port |

### Git Operations
| Method | Purpose |
|--------|---------|
| `ListRepoChoices()` | Available repos for selection |
| `CreateRepo(name, isPublic, installBmad)` | Create new git repo |
| `RepoMtimes(repoPath)` | Cheap change detection (mtime polling) |
| `RepoStatus(repoPath)` | Dirty, ahead/behind, protected status |
| `GitListBranches(repoPath)` | Local branch list |
| `GitSwitchBranch(repoPath, branch, autoCommit)` | Switch branch |
| `GitCreateBranch(repoPath, prefix, name, autoCommit)` | Create + checkout |
| `GitCommit(repoPath)` | AI-generated commit |
| `GitCommitStreaming(repoPath)` | Streaming commit with progress |
| `GitCommitAndPush(repoPath)` | Commit + push |
| `GitPush(repoPath)` | Push current branch |
| `GitForcePush(repoPath)` | Force push (with confirmation) |
| `GitPull(repoPath)` | Pull from remote |
| `GitMergeInto(repoPath, target, autoCommit)` | Merge into target branch |
| `GitCommitPushAndPR(repoPath)` | Full ship: commit + push + create PR |
| `GetScopedDiff(dir)` | Files changed in directory |
| `GetWorktrees(repoPath)` | Git worktrees for repo |
| `ListRepoFiles(repoPath)` | File listing for tree view |
| `WriteFile(path, content)` | Write file contents |
| `ReadFile(path)` | Read file contents |
| `ReadFileDiff(repoPath, filePath)` | Unified diff for file |
| `ReadFileAtHead(repoPath, filePath)` | File contents at HEAD |
| `SpawnPRReview(repoPath)` | Launch PR review agent |

### Diff Explanation
| Method | Purpose |
|--------|---------|
| `ExplainDiffHunk(repoPath, filePath, hunk)` | AI explanation of a diff |
| `IsExplainAvailable()` | Check if claude CLI is on PATH |

### BMAD Workflows
| Method | Purpose |
|--------|---------|
| `ListBmadWorkflows()` | All saved workflows |
| `GetBmadWorkflow(id)` | Load single workflow |
| `SaveBmadWorkflow(wf)` | Persist workflow |
| `DeleteBmadWorkflow(id)` | Remove workflow |
| `ListBmadWorkflowsByRepo(repoPath)` | Workflows for a repo |
| `ListBmadTemplates()` | Built-in templates |
| `CreateFromTemplate(templateID, repoPath)` | Copy template → workflow |
| `GetBmadProcesses()` | Full process catalog |
| `GetBmadProcessesByPhase(phase)` | Processes by lifecycle phase |
| `GetBmadModules()` | Available BMAD modules |
| `GetControlFlowNodes()` | Control flow node type list |

### BMAD Execution
| Method | Purpose |
|--------|---------|
| `StartBmadWorkflow(workflowID, repoPath, model)` | Begin execution |
| `PauseBmadWorkflow(execID)` | Pause running execution |
| `ResumeBmadWorkflow(execID)` | Resume paused execution |
| `StopBmadWorkflow(execID)` | Cancel execution |
| `GetBmadExecution(execID)` | Current execution state |
| `GetNodeOutput(execID, nodeID)` | Captured terminal output for node |

### BMAD Agents
| Method | Purpose |
|--------|---------|
| `ListBmadAgents()` | Custom agent configs |
| `SaveBmadAgent(agent)` | Persist agent config |
| `DeleteBmadAgent(id)` | Remove agent config |

### Sprint Management
| Method | Purpose |
|--------|---------|
| `GetSprintStatus(repoPath)` | Parse sprint-status.yaml |
| `UpdateStoryStatus(repoPath, storyID, status)` | Update story status |

---

## 11. Wails Events (Go → Svelte push)

| Event | Payload | Trigger |
|-------|---------|---------|
| `needs-setup` | `bool` | DevDir initialization complete |
| `repos` | `[]RepoInfo` | Every scan cycle (5s) |
| `agent:removed` | `string` (agentID) | Agent killed |
| `notification` | `NotificationEvent` | Via NotificationEngine |
| `bmad:execution:*` | execution state | Workflow execution updates |
| `git:commit:progress` | progress data | Streaming commit updates |

---

## 12. Configuration & Persistence

### Config File (`~/.mashed/config.json`)

```json
{
  "devDir": "/Users/linus/Development",
  "theme": "dark-default",
  "vscodiumExtPath": "/path/to/extensions",
  "importedTheme": "One Dark Pro",
  "monoFont": "JetBrains Mono",
  "fontSize": 13
}
```

### Storage Paths

| Path | Purpose |
|------|---------|
| `~/.mashed/config.json` | User preferences |
| `~/.mashed/themes.json` | Imported VSCodium themes |
| `~/.mashed/workflows/{id}.json` | BMAD workflow definitions |
| `~/.mashed/bmad-agents/{id}.json` | Custom agent configurations |
| `~/.claude/projects/{repo-key}/` | Claude Code session files (read-only) |

### Repo Key Derivation

```
/Users/linus/Development/mashed → -Users-linus-Development-mashed
```
(Replace all `/` with `-`)

---

## 13. Concurrency Model

```
main goroutine
  └─ wails.Run(app)
       ├─ app.startup()
       │    ├─ NotificationEngine (goroutine, channel consumer)
       │    ├─ Terminal Bridge WebSocket server (goroutine)
       │    └─ initScanning()
       │         ├─ scanLoop (goroutine, 5s tick)
       │         │    └─ doScan(): ps, lsof, git, session parse, engine updates
       │         ├─ watchSessions (goroutine, fsnotify)
       │         │    └─ on file change → doScan()
       │         └─ consumeEngineEvents (goroutine, channel reader)
       │              └─ NotificationEvent → Wails EventsEmit → Svelte
       └─ BMAD Executor (on-demand goroutines per workflow execution)
            └─ runDynamic() → concurrent node execution
```

All Wails-bound methods are called from the Svelte frontend on the main thread. Background goroutines communicate via channels and `runtime.EventsEmit()`. The `App.mu` mutex protects the notification list.

---

## 14. Build & Distribution

```bash
# Development (hot-reload frontend)
wails dev

# Production build
wails build

# The output binary embeds all frontend assets via //go:embed
```

### Window Configuration

| Setting | Value |
|---------|-------|
| Title | "Mashed" |
| Default size | 1280 x 800 |
| Min size | 800 x 600 |
| Frameless | true (custom TitleBar.svelte) |
| Background | `#07080a` (design system deepest) |
| Mac | Hidden inset title bar, transparent webview |

---

*End of specification.*
