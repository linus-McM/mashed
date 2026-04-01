// Package domain defines all domain types for Claude Conductor.
package domain

import (
	"context"
	"time"
)

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	StatusRunning AgentStatus = "running"
	StatusBlocked AgentStatus = "blocked"
	StatusError   AgentStatus = "error"
	StatusQueued  AgentStatus = "queued"
	StatusDone    AgentStatus = "done"
)

// LogKind categorizes log line types for display styling.
type LogKind string

const (
	LogOK     LogKind = "ok"     // green  ✓  — write/edit success
	LogInfo   LogKind = "info"   // blue   ℹ  — bash commands
	LogWarn   LogKind = "warn"   // amber  ⚠  — warnings
	LogErr    LogKind = "err"    // red    ✗  — errors
	LogDim    LogKind = "dim"    // dim    ·  — reads, searches
	LogSystem LogKind = "system" // purple ⬡  — agent/skill calls
)

// LogLine is a single parsed event from a JSONL session file.
type LogLine struct {
	Kind LogKind   `json:"kind"`
	Text string    `json:"text"`
	Ts   time.Time `json:"ts"`
}

// Agent represents a running or completed Claude Code agent.
type Agent struct {
	ID          string        `json:"id"`          // "pid-{pid}" or "{parentID}-sub-{name}-{toolUseID}"
	Name        string        `json:"name"`        // model name or sub-agent name
	Status      AgentStatus   `json:"status"`
	TokensUsed  int64         `json:"tokensUsed"`
	TokensMax   int64         `json:"tokensMax"`   // 1M for opus, 200K for others
	Model       string        `json:"model"`       // "claude-opus-4-6", "sonnet", etc.
	LogLines    []LogLine     `json:"logLines"`    // last 50 log entries
	Elapsed     time.Duration `json:"elapsed"`
	PID         int           `json:"pid"`         // 0 for in-process sub-agents
	HasTmuxPane bool          `json:"hasTmuxPane"`
	TmuxTarget  string        `json:"tmuxTarget"`  // tmux pane target string
	RepoPath    string        `json:"repoPath"`    // working directory path
}

// DagEdge represents a parent→child relationship between agents.
type DagEdge struct {
	From string `json:"from"` // parent agent ID
	To   string `json:"to"`   // child agent ID
}

// Workflow groups agents working in the same repo into a DAG.
type Workflow struct {
	ID        string      `json:"id"`
	Branch    string      `json:"branch"`
	Status    AgentStatus `json:"status"`
	Agents    []Agent     `json:"agents"`
	Edges     []DagEdge   `json:"edges"`
	StartedAt time.Time   `json:"startedAt"`
}

// Repo represents a git repository with optional active workflow.
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
	BurnIndex        int       `json:"-"` // current write position in ring buffer
}

// SessionData holds parsed data from a single JSONL session file.
type SessionData struct {
	SessionID         string         `json:"sessionId"`
	TotalTokens       int64          `json:"totalTokens"`
	InputTokens       int64          `json:"inputTokens"`
	OutputTokens      int64          `json:"outputTokens"`
	CacheReadTokens   int64          `json:"cacheReadTokens"`
	CacheCreateTokens int64          `json:"cacheCreateTokens"`
	LogLines          []LogLine      `json:"logLines"`
	SubAgents         []SubAgentInfo `json:"subAgents"`
}

// SubAgentInfo tracks a sub-agent spawned via the Agent tool.
type SubAgentInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ToolUseID   string    `json:"toolUseId"`   // tool_use.id for pairing with tool_result
	Status      string    `json:"status"`      // "running" | "done"
	Result      string    `json:"result"`      // summary text (max 500 chars)
	OutputFile  string    `json:"outputFile"`  // path to full output JSONL
	LogLines    []LogLine `json:"logLines"`
}

// AgentSession holds live process metadata for a running Claude CLI instance.
type AgentSession struct {
	PID       int       `json:"pid"`
	PPID      int       `json:"ppid"`
	Model     string    `json:"model"`
	StartedAt time.Time `json:"startedAt"`
	SessionID string    `json:"sessionId"` // empty if no --session-id flag
}

// RepoInfo combines git metadata with running agent sessions.
type RepoInfo struct {
	Name           string         `json:"name"`
	Path           string         `json:"path"`
	Branch         string         `json:"branch"`
	LastCommit     string         `json:"lastCommit"`
	LastCommitDate time.Time      `json:"lastCommitDate"`
	Dirty          bool           `json:"dirty"`
	Agents         []AgentSession `json:"agents"`
}

// EventType classifies notification events by urgency.
type EventType string

const (
	EventNeedsResponse EventType = "needs_response" // AskUserQuestion, permission prompt
	EventError         EventType = "error"           // tool error, test failure
	EventCompleted     EventType = "completed"       // session ended cleanly
	EventRunning       EventType = "running"         // actively producing output
	EventStarted       EventType = "started"         // new agent detected
)

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
	Priority   int       `json:"priority"`   // 0=needs-response, 1=error, 2=completed, 3=running
	TokensUsed int64     `json:"tokensUsed"`
	TokensMax  int64     `json:"tokensMax"`  // context window size
	TmuxTarget string    `json:"tmuxTarget"` // tmux pane target for drill-down (empty if no pane)
	PID        int       `json:"pid"`
}

// WorktreeInfo describes a git worktree associated with an agent branch.
type WorktreeInfo struct {
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	IsOrphaned bool   `json:"isOrphaned"` // true if branch was deleted
}

// ScopedDiff shows files changed by an agent.
type ScopedDiff struct {
	Files []DiffFileStat `json:"files"`
}

// DiffFileStat is a single file's change summary.
type DiffFileStat struct {
	Path      string `json:"path"`
	Added     int    `json:"added"`
	Removed   int    `json:"removed"`
	IsBinary  bool   `json:"isBinary"`
	IsNew     bool   `json:"isNew"`
}

// AgentProvider abstracts the agent platform.
// ClaudeCodeProvider implements this for Claude Code.
// Future providers (Aider, Codex CLI, Gemini CLI) implement the same interface.
type AgentProvider interface {
	// ScanProcesses discovers running agent sessions.
	ScanProcesses() ([]AgentSession, error)

	// GetWorkingDir returns the working directory for a process.
	// CLAUDE_SPECIFIC: uses lsof on macOS, /proc on Linux.
	GetWorkingDir(pid int) (string, error)

	// ParseSession parses a session log file and returns structured data.
	// CLAUDE_SPECIFIC: reads JSONL from ~/.claude/projects/{repo-key}/.
	ParseSession(path string) (*SessionData, error)

	// WatchSessions watches for new session activity and sends events.
	// CLAUDE_SPECIFIC: watches ~/.claude/projects/ for new .jsonl files.
	WatchSessions(ctx context.Context) (<-chan SessionEvent, error)

	// SessionDir returns the directory where session files are stored for a repo.
	// CLAUDE_SPECIFIC: ~/.claude/projects/{repo-key}/
	SessionDir(repoPath string) string
}

// SessionEvent is emitted by WatchSessions when session files change.
type SessionEvent struct {
	SessionPath string    // path to the .jsonl file
	RepoPath    string    // path to the repo
	IsNew       bool      // true if this is a new session file
	Timestamp   time.Time
}
