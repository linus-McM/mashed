# Mashed — Desktop Application Specification

> Notification-first IDE for multi-agent development.
> A Wails v2 desktop app (Go backend + Svelte frontend) for monitoring, orchestrating, and interacting with Claude Code sessions across `~/Development` repos, with live terminal embedding, workflow automation, and integrated code editing.

---

## Changelog

### BMAD Terminal Bridge

- BMAD tmux sessions are named `bmad-{repo}-{branch}-{label}-{shortHash}` (e.g. `bmad-surfseer-main-create-story-a1b2c3d4`). Built by `internal/bmad.BuildSessionName(repoPath, branch, nodeLabel, nodeID, nowNanos)` and parsed back by `ParseSessionName`. Total length capped to prevent tmux rejection (≤ 88 bytes); long components are slugified and truncated while the short hash is preserved for uniqueness.
- **Migration note:** Sessions created by older builds cannot be re-attached via View Terminal — they must be restarted. Orphaned `bmad-*` sessions from prior runs whose names do not belong to a tracked execution are cleaned up at executor startup (`internal/bmad/cleanup.go`).
- View Terminal works for running BMAD nodes, streaming live Claude CLI output through the `internal/terminal.TmuxAdapter` (FIFO-based with polling fallback).
- The terminal modal title shows the parsed friendly form (`Terminal — repo · branch · label`) instead of the raw tmux target; parsing is mirrored on the frontend in `frontend/src/lib/bmadSessionName.ts`.

### Out-of-Process PTY Helper

- Terminal I/O is no longer performed in-process. `main.go` resolves and launches a separate `mashed-pty-helper` binary (source: `cmd/pty-helper/main.go`, client in `internal/terminal/helper/client.go`, server in `internal/terminal/helper/server.go`) and communicates over a Unix domain socket using the protocol defined in `internal/terminal/helper/protocol.go`. The Wails process talks to the helper via `helper.Client`; terminal sessions are registered per-window through `app_terminal_registry.go`.
- The helper is discovered in `build/bin/` first (dev) and falls back to the `.app` bundle in production. Startup waits for the socket to appear before binding Wails methods. Shutdown signals the helper, closes the client, and waits for the child process to exit cleanly.

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
main.go                             ← Wails entry point, menu bar, PTY helper lifecycle, asset embedding
app.go                              ← App struct, startup/shutdown, config, theme management, active context
app_scan.go                         ← Process scanning loop, agent discovery, notification engine wiring
app_sessions.go                     ← JSONL session watching, parsing, status inference
app_spawn.go                        ← SpawnAgent / SpawnAgentWithCommand / SpawnTerminal / KillAgent
app_terminal_registry.go            ← Per-window terminal session registry backed by helper.Client
app_git.go                          ← Git operations (branches, commits, push, PR, merge, diff, worktrees)
app_bmad.go                         ← BMAD workflow CRUD, execution control, agent/module management
app_bmad_question_test.go           ← Question-response flow tests
app_claude.go                       ← Claude-CLI-backed integrations (ListAllAgents, etc.)
app_models.go                       ← Model catalog (ListModels)
app_review.go                       ← Code review streaming (StreamCodeReviewSummary, SpawnPRReview)
app_review_scoped.go                ← Scoped review / refactor-plan streaming (StreamScopedAdvice, SpawnRefactorPlan)
app_explain.go                      ← AI-powered diff explanation via Claude CLI
font_scanner.go                     ← Local font discovery + Nerd Fonts catalog
theme_scanner.go                    ← VSCodium theme import + conversion
cmd/
  pty-helper/
    main.go                         ← Out-of-process PTY helper binary (mashed-pty-helper)
internal/
  advice/
    loader.go                       ← Markdown advice mode loader (user + bundled defaults)
    types.go                        ← Advice mode struct
    defaults/                       ← Bundled modes: clean-code, domain-driven-design,
                                        extreme-programming, performance, pragmatic,
                                        security-first, solid, strategic
  agent/
    engine.go                       ← NotificationEngine — event classification and emission
    tokensamples.go                 ← Token burn sample ring buffer helpers
  bmad/
    types.go                        ← WorkflowDef, WorkflowNode, WorkflowEdge, NodeType, BmadPhase, etc.
    storage.go                      ← Workflow/agent persistence to ~/.mashed/
    executor.go                     ← DAG-based workflow execution (dynamic ready-set algorithm)
    registry.go                     ← Process catalog (built-in BMAD processes by phase)
    modules.go                      ← Module catalog (reusable BMAD modules)
    templates.go                    ← Built-in workflow templates
    condition.go                    ← Condition evaluation engine (comparisons, regex, contains)
    sprint.go                       ← Sprint status YAML parsing/updating
    artifacts.go                    ← Artifact path resolution and verification
    question.go                     ← AskUserQuestion response routing for BMAD nodes
    session_naming.go               ← BuildSessionName / ParseSessionName for tmux session labels
    cleanup.go                      ← Orphaned tmux session cleanup at executor startup
    assets.go                       ← "Mashed-ready" skill/agent/command asset loader
    skillgen.go                     ← Skill generation helper
  domain/
    types.go                        ← All domain structs (Agent, Repo, Workflow, SessionData, etc.)
    models.go                       ← Model metadata catalog (name, token limit)
  explain/
    explain.go                      ← Diff explanation via Claude CLI subprocess
  git/
    diff.go                         ← Git diff parsing, scoped diffs
    worktree.go                     ← Git worktree management
    errors.go                       ← Git-specific error types
  scanner/
    processes.go                    ← Discover Claude CLI sessions via ps + lsof
    repos.go                        ← Scan dev directory for git repos + metadata
    sessions.go                     ← Parse JSONL session files for token/log data
    watcher.go                      ← fsnotify-based live JSONL tailing
    claude.go                       ← ClaudeCodeProvider (AgentProvider implementation)
    errors.go                       ← Scanner-specific error types
  terminal/
    bridge.go                       ← WebSocket bridge (pty ↔ xterm.js in browser)
    manager.go                      ← Terminal session manager (lifecycle, registry)
    session.go                      ← Terminal session abstraction
    panes.go                        ← tmux pane discovery, PID-to-pane mapping
    tmux_adapter.go                 ← FIFO-based tmux pane capture adapter
    tmux_escape.go                  ← tmux control-mode / escape helpers
    stub.go                         ← Build-tag stubs for non-tmux environments
    helper/
      client.go                     ← Unix-socket client used by the Wails process
      server.go                     ← Server running inside mashed-pty-helper
      protocol.go                   ← Wire protocol (request/response framing)
scripts/
  check-coverage.sh                 ← CI test coverage gate
  tmux-bmad-viewer.sh               ← Dev helper for attaching to running BMAD sessions
frontend/
  src/
    App.svelte                      ← Root component, view routing
    main.js                         ← Svelte entry point
    views/
      Setup.svelte                  ← First-run dev directory picker
      NotificationFeed.svelte       ← Main dashboard — agent notification stream
      AgentDetail.svelte            ← Agent drill-down with terminal + logs + diff + editor
      WorkflowBuilder.svelte        ← BMAD visual workflow canvas
      Settings.svelte               ← Theme, font, path, editor, advice configuration
      SpawnAgent.svelte             ← New agent session launcher
      NewSessionModal.svelte        ← Session creation modal
      SummarisationModal.svelte     ← Streaming advice / review summarisation modal
      BranchModal.svelte            ← Branch creation
      SwitchBranchModal.svelte      ← Branch switching
      MergeModal.svelte             ← Branch merge
      ForcePushModal.svelte         ← Force push confirmation
    components/
      TitleBar.svelte               ← Frameless window title bar with controls
      Terminal.svelte               ← xterm.js terminal (WebSocket bridge to helper)
      MonacoEditor.svelte           ← Monaco editor integration
      CodeEditor.svelte             ← Lightweight code editor
      EditorRouter.svelte           ← Dispatches to Monaco / Markdown / Image editor by file type
      MarkdownEditor.svelte         ← Crepe / Milkdown markdown editor
      ImageViewer.svelte            ← Image preview component
      DiffView.svelte               ← Side-by-side diff viewer
      FileTree.svelte               ← File browser tree
      SparkLine.svelte              ← Token burn sparkline (block characters)
      StatusBadge.svelte            ← Agent status pill badge
      NewRepoModal.svelte           ← Repository creation dialog
      AboutModal.svelte             ← About / version dialog
      bmad/
        CanvasPane.svelte           ← xyflow DAG canvas for workflow builder
        ProcessNode.svelte          ← Standard process node
        ConditionNode.svelte        ← If/else branch node
        LoopNode.svelte             ← Loop N times node
        LoopUntilNode.svelte        ← Loop until condition node
        TransformNode.svelte        ← Data extraction/transform node
        MergeNode.svelte            ← Branch merge node
        DeletableEdge.svelte        ← Edge with delete button
        ProcessSidebar.svelte       ← Process catalog sidebar
        NodeConfigPanel.svelte      ← Node configuration editor
        ExecutionBar.svelte         ← Run/pause/stop controls
        TemplatePicker.svelte       ← Workflow template selection
        AgentConfigModal.svelte     ← Custom agent configuration
        ArrayEditorModal.svelte     ← Array field editor (for node config lists)
        NameWorkflowModal.svelte    ← New workflow naming dialog
        QuestionResponseModal.svelte ← AskUserQuestion response modal
        QuestionSnackbarStack.svelte ← Stacked question-toast notifications
        OutputViewerModal.svelte    ← Node output inspection
        GitPanel.svelte             ← Embedded git panel on the workflow canvas
        RepoContextBar.svelte       ← Repository context display
        SprintPanel.svelte          ← Sprint status panel
    config/
      claude-cli.json               ← Claude CLI spawn defaults (model, args, env)
    lib/
      stores/theme.js               ← Theme reactive store
      stores/font.js                ← Font reactive store
      stores/editorSettings.js      ← Editor settings (font size, tab width, etc.)
      stores/sessions.js            ← Active session store
      themes.js                     ← Built-in theme definitions
      themeConverter.js             ← VSCodium → Mashed theme conversion
      themeInit.js                  ← Theme initialization on startup
      monacoTheme.js                ← Monaco editor theme adapter
      fileTree.js                   ← File tree data structures
      sprintColors.js               ← BMAD sprint status colors
      bmadSessionName.ts            ← Parses BMAD tmux session names (mirrors Go ParseSessionName)
    scripts/
      check-orphan-tokens.mjs       ← Design-token guardrail script
  wailsjs/                          ← Auto-generated Wails bindings (do not edit)
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

### Two-process architecture

Terminal I/O is split across two processes for crash isolation:

1. **Wails process** (`main.go` → `App`) — owns the UI, Wails bindings, and a `helper.Client` connected via a Unix domain socket.
2. **`mashed-pty-helper`** (`cmd/pty-helper/main.go`) — owns `creack/pty` file descriptors and tmux interaction. Runs `helper.Server` (`internal/terminal/helper/server.go`) behind the socket.

At startup `main.go`:

1. Calls `resolveHelperPath()` — prefers `build/bin/mashed-pty-helper` (dev) over the `.app` bundle copy.
2. Launches the helper subprocess inheriting stdin/out/err.
3. Calls `waitForSocket(path, timeout)` until the helper's Unix socket file appears.
4. Opens `helper.Client` over the socket.
5. Hands the client to the Wails `App` before `wails.Run(...)`.

On shutdown: closes the client, signals the helper subprocess, and waits for exit to avoid orphaned helpers.

### Helper Protocol (`internal/terminal/helper/protocol.go`)

Framed request/response messages over the Unix socket. Operations include: spawn a pty attached to a shell or tmux pane, read/write bytes, resize, kill, and list active sessions. `client.go` and `server.go` implement the two ends; both have integration tests (`integration_test.go`).

### WebSocket Bridge (`internal/terminal/bridge.go`)

Inside the Wails process, `bridge.go` runs a WebSocket server on a dynamic port. The port is exposed to the frontend via `GetTerminalPort()`. `Terminal.svelte` connects xterm.js to `ws://localhost:{port}` and sends a tmux target string on connect. The bridge proxies bytes in both directions between the frontend socket and the helper session.

### Session Registry (`app_terminal_registry.go`)

Per-window terminal sessions are registered so that `KillTerminalSession(sessionName)` and reconnects can locate the correct helper session. `internal/terminal/manager.go` and `session.go` define the server-side lifecycle; `panes.go` handles tmux pane discovery; `tmux_adapter.go` and `tmux_escape.go` provide the FIFO-based tmux capture used by the BMAD "View Terminal" feature.

### tmux Pane Discovery (`internal/terminal/panes.go`)

```bash
tmux list-panes -a -F '#{pane_pid}:#{pane_id}:#{session_name}:#{window_index}.#{pane_index}'
```

Agent PIDs are rarely direct children of tmux panes (shell → node → claude). Discovery walks the PPID chain up to 8 levels to match an agent PID to a pane. The pane list is cached (short TTL) to avoid repeated `tmux list-panes` calls; the cache is invalidated when new sessions are spawned.

### Agent Spawning (`app_spawn.go`)

```go
func (a *App) SpawnAgent(repoPath string, model string) (string, error)
func (a *App) SpawnAgentWithCommand(repoPath, command string) (string, error)
func (a *App) SpawnTerminal(repoPath string) (string, error)
func (a *App) KillAgent(agentID string, pid int, tmuxTarget string) error
func (a *App) KillTerminalSession(sessionName string) error
```

BMAD sessions are named via `internal/bmad.BuildSessionName(repoPath, branch, nodeLabel, nodeID, nowNanos)` producing `bmad-{repo}-{branch}-{label}-{shortHash}`. Ad-hoc agents / terminals spawned from the UI use session names minted in `app_spawn.go` and tracked via the helper protocol rather than being derived by format string — see the code for the exact format when interoperating with external tooling.

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

### Process Registry (`internal/bmad/registry.go`, phases in `types.go`)

Processes are organized by BMAD lifecycle phase. Phase constants (`internal/bmad/types.go`):

```go
type BmadPhase string

const (
    PhaseAnalysis       BmadPhase = "analysis"
    PhasePlanning       BmadPhase = "planning"
    PhaseSolutioning    BmadPhase = "solutioning"
    PhaseImplementation BmadPhase = "implementation"
    PhaseSupport        BmadPhase = "support"
    PhaseUtilities      BmadPhase = "utilities"
)
```

Two additional phase labels — `"strategic"` and `"review"` — are used at the Global process layer and inside story status (`StoryReview`). `GetBmadProcessesByPhase(phase)` filters the catalog for the sidebar's collapsible phase groups.

### Storage (`internal/bmad/storage.go`)

- Workflows: `~/.mashed/workflows/{id}.json`
- BMAD agents: `~/.mashed/bmad-agents/{id}.json`
- Per-repo workflows via the `RepoPath` field on `WorkflowDef`

### Artifact System (`internal/bmad/artifacts.go`)

- Artifacts resolve to `{repoPath}/_bmad-output/{category}/{artifact}` via a canonical path map
- `ResolveArtifactPath(name, repoPath)` → full path
- `VerifyArtifacts(repoPath, outputNames)` → checks existence
- Unmapped artifacts (`"code"`, `"tests"`, `"any-doc"`) return `""` — handled by callers
- Exposed to the frontend via `GetArtifactStatus(...)` for per-node status decoration

### Question Flow (`internal/bmad/question.go`)

AskUserQuestion tool calls emitted by running BMAD nodes are routed through the question store and surfaced to the frontend as `QuestionResponseModal` / `QuestionSnackbarStack`. The frontend answers via `RespondToQuestion(...)`, which posts the response back into the originating node's session.

### Asset Loader (`internal/bmad/assets.go`)

Walks user and bundled skill/agent/command directories and keeps the "mashed-ready" asset set. An asset is considered mashed-ready when it carries a `mashedRole` frontmatter value. Missing frontmatter or missing role silently skips the asset without erroring. Exposed via `ListAllMashedAssets`.

### Session Cleanup (`internal/bmad/cleanup.go`)

On executor startup, orphaned tmux sessions whose names start with the BMAD prefix but do not correspond to any tracked execution are killed. `liveSessionNames()` is the source of truth for "in use"; `bareSessionName(target)` strips optional `:window.pane` suffixes before comparison.

### Templates (`internal/bmad/templates.go`)

Built-in workflow templates deep-copied into user workflows via `CreateFromTemplate(templateID, repoPath)`.

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
- **Canvas** (`CanvasPane.svelte`): drag-and-drop node placement, edge connections, `DeletableEdge` for edge removal
- **Sidebar** (`ProcessSidebar.svelte`): process catalog with collapsible BMAD phase groups (analysis, planning, solutioning, implementation, support, utilities + strategic/review globals)
- **Config panel** (`NodeConfigPanel.svelte`) + **Array field editor** (`ArrayEditorModal.svelte`): edit node parameters and list-valued fields
- **Execution bar** (`ExecutionBar.svelte`): start/pause/resume/stop workflow execution
- **Template picker** (`TemplatePicker.svelte`) + **Name dialog** (`NameWorkflowModal.svelte`): create workflows from templates, then name them
- **Agent config** (`AgentConfigModal.svelte`): edit custom BMAD agents
- **Node types**: `ProcessNode`, `ConditionNode`, `LoopNode`, `LoopUntilNode`, `TransformNode`, `MergeNode`
- **Question flow** (`QuestionResponseModal.svelte`, `QuestionSnackbarStack.svelte`): answer AskUserQuestion prompts from running nodes
- **Output inspection** (`OutputViewerModal.svelte`): view captured terminal output per node
- **Sprint panel** (`SprintPanel.svelte`): sprint status parsed from `sprint-status.yaml`
- **Git panel** (`GitPanel.svelte`): inline git operations (commit, push, branch switch) without leaving the canvas
- **Repo context bar** (`RepoContextBar.svelte`): which repo / branch / worktree the workflow is running against

### Settings (`Settings.svelte`)
- Theme selection (built-in, bundled, and VSCodium import)
- Font selection (system fonts + Nerd Fonts)
- Font size, editor settings (tab width, etc.)
- Dev directory path
- VSCodium extension path
- Advice mode selection (list from `ListAdviceModes()`)

### Summarisation (`SummarisationModal.svelte`)
Streaming modal used for `StreamAdvice`, `StreamCodeReviewSummary`, and `StreamScopedAdvice`. Token-by-token render of the model output with a cancel button.

### About (`AboutModal.svelte`)
Version + build info dialog, reachable from the TitleBar menu.

---

## 10. Wails Bindings (Go → Svelte API)

All public methods on the `App` struct are exposed to the Svelte frontend via Wails bindings.

> The exhaustive list is the public (exported) method set of `*App` across the root `app*.go` files. Sections below group them by purpose.

### Configuration & Context
| Method | Purpose |
|--------|---------|
| `PickDirectory()` | Native OS directory picker dialog |
| `PickFile()` | Native OS file picker dialog |
| `SetDevDir(dir)` | Save dev directory and start scanning |
| `GetDevDir()` | Current dev directory |
| `GetConfig()` | Full persisted config |
| `SetTheme(id)` | Persist theme selection |
| `SetMonoFont(family)` | Persist font selection |
| `SetFontSize(size)` | Persist font size |
| `SetVSCodiumExtPath(path)` | Persist VSCodium path |
| `SetSidebarWidth(px)` | Persist sidebar layout width |
| `SetActiveContext(...)` | Update the focused repo / agent context |
| `GetEditorSettings()` | Current editor settings (font size, tab width, etc.) |
| `SetEditorSettings(s)` | Persist editor settings |
| `DefaultEditorSettings()` | Factory defaults |

### Theme Management
| Method | Purpose |
|--------|---------|
| `GetSavedThemes()` | All imported themes as JSON |
| `SaveTheme(id, json)` | Persist converted theme |
| `RemoveTheme(id)` | Delete saved theme |
| `ListVSCodiumThemes()` | Discover VSCodium color themes |
| `ReadThemeFile(path)` | Read raw theme JSON |
| `SetImportedTheme(path)` | Import a VSCodium theme |
| `ListBundledThemes()` | Themes shipped inside the app bundle |
| `ReadBundledThemeFile(id)` | Read a bundled theme JSON |

### Font Discovery
| Method | Purpose |
|--------|---------|
| `ListLocalFonts()` | System monospace fonts |
| `ListNerdFonts()` | Nerd Fonts catalog |
| `GetFontsDir()` | Font installation directory |
| `OpenFontsDir()` | Open font dir in Finder |

### Agents, Sessions & Terminals
| Method | Purpose |
|--------|---------|
| `GetNotifications()` | Priority-sorted notification list |
| `SpawnAgent(repoPath, model)` | Start Claude in new tmux session |
| `SpawnAgentWithCommand(repoPath, cmd)` | Start with a custom CLI command |
| `SpawnTerminal(repoPath)` | Start plain shell session |
| `KillAgent(agentID, pid, tmuxTarget)` | Terminate agent + tmux session |
| `KillTerminalSession(name)` | Terminate registered terminal session |
| `GetAgentLog(repoPath)` | Parsed log lines for latest session |
| `MarkRead(agentID)` | Mark notification as read |
| `GetTerminalPort()` | WebSocket bridge port |
| `ListRepoSessions(repoPath)` | All known sessions for a repo |
| `ListAllAgents()` | All agents visible across repos |
| `ListModels()` | Available model names (opus, sonnet, haiku, ...) |
| `WriteConsoleLog(line)` | Append a line to the app console log |
| `TakeScreenshot()` | Save a full-window screenshot to disk |

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

### Files & Editor I/O
| Method | Purpose |
|--------|---------|
| `ListRepoFiles(repoPath)` | File listing for tree view |
| `WriteFile(path, content)` | Write file contents |
| `ReadFile(path)` | Read file contents |
| `ReadFileBase(path)` | Base64 read (for images / binaries) |
| `ReadFileDiff(repoPath, filePath)` | Unified diff for file |
| `ReadFileAtHead(repoPath, filePath)` | File contents at HEAD |

### Review, Explain & Advice (streaming)
| Method | Purpose |
|--------|---------|
| `ExplainDiffHunk(repoPath, filePath, hunk)` | AI explanation of a diff |
| `IsExplainAvailable()` | Check if claude CLI is on PATH |
| `StreamCodeReviewSummary(repoPath, mode)` | Streaming code review summary |
| `StreamAdvice(repoPath, mode)` | Streaming advice for a whole repo / branch |
| `StreamScopedAdvice(repoPath, paths, mode)` | Streaming advice scoped to selected files |
| `SpawnPRReview(repoPath)` | Launch PR review agent |
| `SpawnRefactorPlan(repoPath, advice, paths)` | Launch refactor-plan agent with pre-selected context |
| `ListAdviceModes()` | All available advice modes (user + bundled) |
| `ListAllMashedAssets()` | Mashed-ready skills / agents / commands |

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
| `GetArtifactStatus(repoPath, outputs)` | Existence check for workflow artifacts |

### BMAD Execution & Questions
| Method | Purpose |
|--------|---------|
| `StartBmadWorkflow(workflowID, repoPath, model)` | Begin execution |
| `PauseBmadWorkflow(execID)` | Pause running execution |
| `ResumeBmadWorkflow(execID)` | Resume paused execution |
| `StopBmadWorkflow(execID)` | Cancel execution |
| `GetBmadExecution(execID)` | Current execution state |
| `GetBmadCurrentExecution()` | Active execution for the current context |
| `GetNodeOutput(execID, nodeID)` | Captured terminal output for node |
| `RespondToQuestion(...)` | Answer an AskUserQuestion prompt from a running node |

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

Events are emitted via `runtime.EventsEmit(ctx, name, payload...)` and consumed in Svelte via `EventsOn(name, cb)`. The complete set of events the frontend listens for:

| Event | Payload | Trigger |
|-------|---------|---------|
| `repos` | `[]RepoInfo` | Every scan cycle |
| `agent:notification` | `NotificationEvent` | NotificationEngine classifies an agent update |
| `agent:removed` | `string` (agentID) | Agent process exited or was killed |
| `terminal:session:added` | session info | New terminal session registered via helper |
| `terminal:session:removed` | `string` (sessionName) | Terminal session closed |
| `git:commit:progress` | progress payload | Streaming commit updates |
| `repo:create:progress` | progress payload | `CreateRepo` progress updates |
| `review:summary:progress` | partial text | Streaming code review tokens |
| `review:summary:done` | final summary | Code review stream complete |
| `review:advice:progress` | partial text | Streaming advice tokens |
| `screenshot:inject` | bytes / path | Programmatic screenshot capture |
| `screenshot:taken` | `string` (path) | Screenshot saved to disk |
| `menu:navigate` | `string` (route) | Native menu navigation request |
| `menu:about` | — | Native menu "About" item clicked |
| `bmad:execution:status` | execution state | BMAD workflow execution status change |
| `bmad:node:status` | node state | Per-node status update |
| `bmad:node:artifacts` | artifact set | Artifact existence / verification update |
| `bmad:node:question` | question event | Running node requested a user answer |
| `bmad:node:question:dismissed` | question event | Question dismissed / answered |
| `bmad:node:idle` | idle event | Node entered idle state |
| `bmad:node:idle:dismissed` | idle event | Idle state dismissed |
| `bmad:sprint:updated` | sprint state | sprint-status.yaml changed |

---

## 12. Configuration & Persistence

### Config Struct (`app.go`)

```go
type mashedConfig struct {
    DevDir          string          `json:"devDir"`
    Theme           string          `json:"theme,omitempty"`
    VSCodiumExtPath string          `json:"vscodiumExtPath,omitempty"`
    ImportedTheme   string          `json:"importedTheme,omitempty"`
    MonoFont        string          `json:"monoFont,omitempty"`
    FontSize        int             `json:"fontSize,omitempty"`
    SidebarWidth    int             `json:"sidebarWidth,omitempty"`
    EditorSettings  *EditorSettings `json:"editorSettings,omitempty"`
}
```

Persisted to `~/.mashed/config.json` via `loadConfig()` / `saveConfig()`. `EditorSettings` holds Monaco-style options (font size, tab width, smooth scrolling, etc.).

### Storage Paths

| Path | Purpose |
|------|---------|
| `~/.mashed/config.json` | User preferences (see struct above) |
| `~/.mashed/themes.json` | Imported VSCodium themes |
| `~/.mashed/workflows/{id}.json` | BMAD workflow definitions |
| `~/.mashed/bmad-agents/{id}.json` | Custom BMAD agent configurations |
| `~/.mashed/advice/` | User-authored advice mode markdown files |
| `~/.claude/projects/{repo-key}/` | Claude Code session JSONL files (read-only) |

### Repo Key Derivation

```
/Users/linus/Development/mashed → -Users-linus-Development-mashed
```
(Replace all `/` with `-`)

---

## 13. Concurrency Model

```
mashed-pty-helper (subprocess)                 Wails process
───────────────────────────                    ─────────────
helper.Server                       Unix       main goroutine
  ├─ pty sessions (creack/pty)  <── socket ──▶   └─ main() resolves + launches helper,
  ├─ tmux pane attachments                           waits for socket, opens helper.Client
  └─ protocol.* request handlers                    └─ wails.Run(app)
                                                          ├─ app.startup()
                                                          │    ├─ NotificationEngine (goroutine, channel consumer)
                                                          │    ├─ Terminal Bridge WebSocket server (goroutine)
                                                          │    ├─ Terminal registry (uses helper.Client)
                                                          │    └─ initScanning()
                                                          │         ├─ scanLoop (goroutine, tick)
                                                          │         │    └─ doScan(): ps, lsof, git, session parse, engine updates
                                                          │         ├─ watchSessions (goroutine, fsnotify)
                                                          │         │    └─ on file change → doScan()
                                                          │         └─ consumeEngineEvents (goroutine, channel reader)
                                                          │              └─ NotificationEvent → Wails EventsEmit → Svelte
                                                          ├─ BMAD Executor (on-demand goroutines per workflow execution)
                                                          │    └─ runDynamic() → concurrent node execution
                                                          └─ Shutdown: close helper.Client, signal helper, wait for exit
```

All Wails-bound methods are invoked from the Svelte frontend on the main thread. Background goroutines communicate via channels and `runtime.EventsEmit()`. The `App.mu` mutex protects shared state including the notification list and terminal registry. Crashes inside the PTY helper do not take down the Wails process — the helper is restarted on the next operation.

---

## 14. Build & Distribution

```bash
# Development (hot-reload frontend) — see justfile for the full dev task
wails dev

# Production build
wails build

# The output binary embeds all frontend assets via //go:embed (main.go).
# The PTY helper is built separately and placed at build/bin/mashed-pty-helper
# (dev) or inside the .app bundle (release). main.go's resolveHelperPath()
# prefers build/bin/ so freshly-signed helper binaries take precedence over
# a stale bundled copy during development.
```

Top-level build glue lives in `justfile` (tasks for Go build, helper build, frontend test, coverage gate via `scripts/check-coverage.sh`, etc.) and `lefthook.yml` for git hooks.

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
