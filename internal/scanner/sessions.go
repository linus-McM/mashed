package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"conductor/internal/domain"
)

// sessionParserState tracks incremental parsing state for a single JSONL file.
type sessionParserState struct {
	offset    int64
	inode     uint64
	lineCount int
}

// toolKindMap maps tool names to LogKind for display styling.
var toolKindMap = map[string]domain.LogKind{
	"Bash":  domain.LogInfo,
	"Read":  domain.LogDim,
	"Write": domain.LogOK,
	"Edit":  domain.LogOK,
	"Glob":  domain.LogDim,
	"Grep":  domain.LogDim,
	"Agent": domain.LogSystem,
	"Skill": domain.LogSystem,
}

// ParseSession parses a complete JSONL session file and returns structured data.
func (p *ClaudeCodeProvider) ParseSession(path string) (*domain.SessionData, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ParseError{Path: path, Err: fmt.Errorf("opening file: %w", ErrSessionNotFound)}
		}
		return nil, &ParseError{Path: path, Err: fmt.Errorf("opening file: %w", err)}
	}
	defer f.Close()

	data := &domain.SessionData{
		SessionID: sessionIDFromPath(path),
	}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line

	lineNum := 0
	subAgentMap := make(map[string]*domain.SubAgentInfo)

	for sc.Scan() {
		lineNum++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		// Skip incomplete lines (no closing brace)
		if line[len(line)-1] != '}' {
			continue
		}
		if err := parseJSONLLine(line, data, subAgentMap); err != nil {
			continue // skip malformed lines
		}
	}

	if err := sc.Err(); err != nil {
		return data, &ParseError{Path: path, Line: lineNum, Err: fmt.Errorf("scanning: %w", err)}
	}

	for _, sub := range subAgentMap {
		data.SubAgents = append(data.SubAgents, *sub)
	}

	// Cap log lines at last 50
	if len(data.LogLines) > 50 {
		data.LogLines = data.LogLines[len(data.LogLines)-50:]
	}

	return data, nil
}

// ParseSessionIncremental reads only new lines from the last known offset.
// Detects inode changes for file rotation and resets state accordingly.
func (p *ClaudeCodeProvider) ParseSessionIncremental(path string) (*domain.SessionData, error) {
	p.parserMu.Lock()
	state, ok := p.parsers[path]
	if !ok {
		state = &sessionParserState{}
		p.parsers[path] = state
	}
	p.parserMu.Unlock()

	fi, err := os.Stat(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: fmt.Errorf("stat: %w", err)}
	}

	currentInode := fileInode(fi)
	if state.inode != 0 && currentInode != state.inode {
		// File rotated — reset to beginning
		state.offset = 0
		state.lineCount = 0
		state.inode = currentInode
	}
	if state.inode == 0 {
		state.inode = currentInode
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: fmt.Errorf("opening: %w", err)}
	}
	defer f.Close()

	if state.offset > 0 {
		if _, err := f.Seek(state.offset, 0); err != nil {
			return nil, &ParseError{Path: path, Offset: state.offset, Err: fmt.Errorf("seeking: %w", err)}
		}
	}

	data := &domain.SessionData{
		SessionID: sessionIDFromPath(path),
	}
	subAgentMap := make(map[string]*domain.SubAgentInfo)

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var bytesRead int64
	for sc.Scan() {
		line := sc.Bytes()
		bytesRead += int64(len(line)) + 1 // +1 for newline

		if len(line) == 0 {
			continue
		}
		// Skip incomplete trailing lines
		if line[len(line)-1] != '}' {
			bytesRead -= int64(len(line)) + 1 // don't advance past incomplete line
			break
		}

		state.lineCount++
		if err := parseJSONLLine(line, data, subAgentMap); err != nil {
			continue
		}
	}

	// Only advance offset for complete lines
	if sc.Err() == nil {
		state.offset += bytesRead
	}

	for _, sub := range subAgentMap {
		data.SubAgents = append(data.SubAgents, *sub)
	}

	return data, sc.Err()
}

// fileInode extracts the inode number from os.FileInfo via syscall.
func fileInode(fi os.FileInfo) uint64 {
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return stat.Ino
}

// --- JSONL message types ---

type jsonlMessage struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

type assistantMessage struct {
	Usage   tokenUsage        `json:"usage"`
	Content []json.RawMessage `json:"content"`
}

type tokenUsage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheReadTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_input_tokens"`
}

type contentBlock struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Text      string          `json:"text,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

// parseJSONLLine dispatches a single JSONL line to the appropriate handler.
func parseJSONLLine(line []byte, data *domain.SessionData, subAgents map[string]*domain.SubAgentInfo) error {
	var msg jsonlMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", ErrInvalidJSONL)
	}

	switch msg.Type {
	case "assistant":
		data.LastMessageType = "assistant"
		return parseAssistantMessage(msg.Message, data, subAgents)
	case "user":
		data.LastMessageType = "user"
		data.HasPendingToolUse = false // user responded, tool use resolved
		return parseUserMessage(msg.Message, subAgents)
	}
	return nil
}

func parseAssistantMessage(raw json.RawMessage, data *domain.SessionData, subAgents map[string]*domain.SubAgentInfo) error {
	var am assistantMessage
	if err := json.Unmarshal(raw, &am); err != nil {
		return fmt.Errorf("unmarshal assistant: %w", ErrInvalidJSONL)
	}

	// Accumulate tokens
	data.InputTokens += am.Usage.InputTokens
	data.OutputTokens += am.Usage.OutputTokens
	data.CacheReadTokens += am.Usage.CacheReadTokens
	data.CacheCreateTokens += am.Usage.CacheCreationTokens
	data.TotalTokens += am.Usage.InputTokens + am.Usage.OutputTokens +
		am.Usage.CacheReadTokens + am.Usage.CacheCreationTokens

	data.HasPendingToolUse = false
	data.LastToolName = ""

	for _, rawBlock := range am.Content {
		var block contentBlock
		if err := json.Unmarshal(rawBlock, &block); err != nil {
			continue
		}

		switch block.Type {
		case "tool_use":
			data.LogLines = append(data.LogLines, toolUseToLogLine(block))
			data.HasPendingToolUse = true
			data.LastToolName = block.Name

			if block.Name == "Agent" {
				info := parseAgentToolInput(block)
				subAgents[block.ID] = &info
			}

		case "text":
			text := strings.TrimSpace(block.Text)
			text = collapseNewlines(text)
			if len(text) >= 6 {
				data.LogLines = append(data.LogLines, domain.LogLine{
					Kind: domain.LogInfo,
					Text: truncate(text, 200),
					Ts:   time.Now(),
				})
			}
		}
	}

	return nil
}

func parseUserMessage(raw json.RawMessage, subAgents map[string]*domain.SubAgentInfo) error {
	var um struct {
		Content []json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &um); err != nil {
		return nil // not all user messages have structured content
	}

	for _, rawBlock := range um.Content {
		var block contentBlock
		if err := json.Unmarshal(rawBlock, &block); err != nil {
			continue
		}

		if block.Type != "tool_result" || block.ToolUseID == "" {
			continue
		}

		sub, ok := subAgents[block.ToolUseID]
		if !ok {
			continue
		}

		sub.Status = "done"

		// Extract result text from content
		var resultStr string
		if err := json.Unmarshal(block.Content, &resultStr); err == nil {
			sub.Result = truncate(resultStr, 500)
		}
	}

	return nil
}

// toolUseToLogLine converts a tool_use content block to a LogLine.
func toolUseToLogLine(block contentBlock) domain.LogLine {
	displayName := block.Name

	// Handle MCP tool names: mcp__server__ns__tool -> last segment
	if strings.HasPrefix(displayName, "mcp__") {
		parts := strings.Split(displayName, "__")
		displayName = parts[len(parts)-1]
	}

	kind, ok := toolKindMap[block.Name]
	if !ok {
		kind = domain.LogInfo
	}

	detail := extractToolDetail(block)

	text := displayName
	if detail != "" {
		text = displayName + ": " + detail
	}

	return domain.LogLine{
		Kind: kind,
		Text: truncate(text, 200),
		Ts:   time.Now(),
	}
}

// extractToolDetail pulls the most relevant field from a tool's input.
func extractToolDetail(block contentBlock) string {
	if len(block.Input) == 0 {
		return ""
	}

	var input map[string]interface{}
	if err := json.Unmarshal(block.Input, &input); err != nil {
		return ""
	}

	switch block.Name {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return truncate(cmd, 100)
		}
	case "Read":
		if fp, ok := input["file_path"].(string); ok {
			return filepath.Base(fp)
		}
	case "Write":
		if fp, ok := input["file_path"].(string); ok {
			return filepath.Base(fp)
		}
	case "Edit":
		if fp, ok := input["file_path"].(string); ok {
			return filepath.Base(fp)
		}
	case "Glob":
		if p, ok := input["pattern"].(string); ok {
			return p
		}
	case "Grep":
		if p, ok := input["pattern"].(string); ok {
			return p
		}
	case "Agent":
		if d, ok := input["description"].(string); ok {
			return d
		}
	case "Skill":
		if s, ok := input["skill"].(string); ok {
			return s
		}
	}

	return ""
}

// parseAgentToolInput extracts sub-agent metadata from an Agent tool_use block.
func parseAgentToolInput(block contentBlock) domain.SubAgentInfo {
	var input map[string]interface{}
	if err := json.Unmarshal(block.Input, &input); err != nil {
		return domain.SubAgentInfo{
			Name:      "unknown",
			ToolUseID: block.ID,
			Status:    "running",
		}
	}

	name, _ := input["name"].(string)
	if name == "" {
		name, _ = input["description"].(string)
	}
	if name == "" {
		name = "sub-agent"
	}

	desc, _ := input["description"].(string)

	return domain.SubAgentInfo{
		Name:        name,
		Description: desc,
		ToolUseID:   block.ID,
		Status:      "running",
	}
}

func sessionIDFromPath(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".jsonl")
}

func collapseNewlines(s string) string {
	for strings.Contains(s, "\n\n") {
		s = strings.ReplaceAll(s, "\n\n", "\n")
	}
	return strings.ReplaceAll(s, "\n", " ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
