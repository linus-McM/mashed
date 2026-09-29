package bmad

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// executeTransformNode reads the output of a source node, applies an extraction
// (regex or line range), and stores the result. Transform nodes are synchronous.
func (e *Executor) executeTransformNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID string) {
	idx := nodeIndex[nodeID]

	// Mark running.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	node := state.exec.Nodes[idx]
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Read config.
	sourceNodeID := node.Config["sourceNode"]
	extractType := node.Config["extractType"]
	extractPattern := node.Config["extractPattern"]

	// Get source output.
	state.mu.Lock()
	sourceOutput := state.exec.NodeOutputs[sourceNodeID]
	state.mu.Unlock()

	// Apply extraction.
	var result string
	switch extractType {
	case "regex":
		result = extractRegex(sourceOutput, extractPattern)
	case "lines":
		result = extractLines(sourceOutput, extractPattern)
	default:
		result = sourceOutput // passthrough if unknown type
	}

	// Cap at 100KB.
	if len(result) > maxCaptureBytes {
		result = result[len(result)-maxCaptureBytes:]
	}

	// Store result.
	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = result
	state.mu.Unlock()

	e.completeNode(state, idx, nodeID)
}

// executeMultiFileLoader resolves the configured {label, path} entries of a
// MultiFileLoader node and surfaces them via OutputPaths + a
// NodeArtifactEvent. See Story breadcrumbs-07.
//
// Behaviour:
//   - Parse Config["entries"] as a JSON array of MultiFileEntry.
//   - Reject >64 entries (ErrMultiFileTooMany).
//   - Reject duplicate non-empty labels (ErrDuplicateMultiFileLabel).
//   - For each entry, resolve relative Path against repoPath, stat the result,
//     and record it under entry.Label (or "file[N]" for empty labels).
//   - Missing files are tolerated: the key is appended to Missing rather than
//     failing the node.
func (e *Executor) executeMultiFileLoader(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath string) {
	state.mu.Lock()
	idx := nodeIndex[nodeID]
	cfg := make(map[string]string, len(state.exec.Nodes[idx].Config))
	for k, v := range state.exec.Nodes[idx].Config {
		cfg[k] = v
	}
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.mu.Unlock()

	// failValidation records the error text, marks the node failed, and
	// terminates the execution with ExecFailed. MultiFileLoader validation
	// errors (malformed config) are NOT retriable — unlike process-node
	// failures they should not leave the execution in ExecPaused awaiting
	// user intervention.
	failValidation := func(msg string) {
		state.mu.Lock()
		state.exec.NodeOutputs[nodeID] = msg
		state.mu.Unlock()
		e.failNode(state, idx, nodeID)
		state.mu.Lock()
		state.exec.Status = ExecFailed
		state.paused = false
		state.cancel()
		state.mu.Unlock()
		e.emit("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecFailed})
	}

	// Parse entries JSON.
	var entries []MultiFileEntry
	if raw := cfg["entries"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &entries); err != nil {
			failValidation(fmt.Errorf("bmad: parse entries: %w", err).Error())
			return
		}
	}

	// Validate: cap.
	if len(entries) > 64 {
		failValidation(ErrMultiFileTooMany.Error())
		return
	}

	// Validate: duplicate non-empty labels.
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.Label == "" {
			continue
		}
		if _, dup := seen[entry.Label]; dup {
			failValidation(fmt.Errorf("%w: %s", ErrDuplicateMultiFileLabel, entry.Label).Error())
			return
		}
		seen[entry.Label] = struct{}{}
	}

	// Emit running status.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Resolve each entry.
	outputPaths := make(map[string]string, len(entries))
	missing := make([]string, 0)
	found := make([]string, 0, len(entries))
	for i, entry := range entries {
		key := entry.Label
		if key == "" {
			key = fmt.Sprintf("file[%d]", i)
		}
		path := entry.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoPath, path)
		}
		// Containment: a relative entry must not escape repoPath via "..".
		// Absolute entries are allowed (user explicitly chose an out-of-repo
		// file); only the join-with-repoPath path requires the check.
		if !filepath.IsAbs(entry.Path) {
			rel, err := filepath.Rel(repoPath, path)
			if err != nil || strings.HasPrefix(rel, "..") {
				missing = append(missing, key)
				continue
			}
		}
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, key)
			continue
		}
		outputPaths[key] = path
		found = append(found, key)
	}

	// Write OutputPaths + mark complete.
	state.mu.Lock()
	if state.exec.Nodes[idx].OutputPaths == nil {
		state.exec.Nodes[idx].OutputPaths = make(map[string]string, len(outputPaths))
	}
	for k, v := range outputPaths {
		state.exec.Nodes[idx].OutputPaths[k] = v
	}
	state.exec.Nodes[idx].Status = NodeComplete
	state.mu.Unlock()

	e.emit("bmad:node:artifacts", NodeArtifactEvent{
		ExecID:  state.exec.ID,
		NodeID:  nodeID,
		Found:   found,
		Missing: missing,
		Paths:   outputPaths,
	})
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})
}

// fileLoaderMaxBytes caps how much of the configured file flows into
// NodeOutputs. 256 KiB is generous for text artifacts (the spec docs cap
// at ~8 KiB for context injection; the full contents live on disk via
// OutputPaths for callers that need them).
const fileLoaderMaxBytes = 256 * 1024

// executeFileLoader reads the file configured on a util-file-loader node
// into NodeOutputs (contents, capped) and OutputPaths["file-path"]
// (absolute path). This is the synchronous replacement for the previous
// behaviour of spawning an empty-skill claude session that did nothing
// useful. Downstream nodes — both autonomous (buildContextStringV3) and
// interactive (buildInteractivePrompt) — pick up the upstream output
// through the normal context path.
func (e *Executor) executeFileLoader(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath string) {
	state.mu.Lock()
	idx := nodeIndex[nodeID]
	cfg := make(map[string]string, len(state.exec.Nodes[idx].Config))
	for k, v := range state.exec.Nodes[idx].Config {
		cfg[k] = v
	}
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.mu.Unlock()

	// Mark running so the UI shows the transition before the read.
	e.setStatus(state, idx, nodeID, NodeRunning)

	raw := strings.TrimSpace(cfg["filePath"])
	if raw == "" {
		e.recordNodeError(state, idx, nodeID, "bmad: file loader: filePath not configured")
		e.failNode(state, idx, nodeID)
		return
	}

	abs := raw
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(repoPath, raw)
	}
	abs = filepath.Clean(abs)

	// Containment check for relative paths mirrors MultiFileLoader. Absolute
	// paths are allowed (user explicitly picked a file outside the repo).
	if !filepath.IsAbs(raw) && repoPath != "" {
		rel, err := filepath.Rel(filepath.Clean(repoPath), abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: %s is outside repo root", raw))
			e.failNode(state, idx, nodeID)
			return
		}
	}

	info, err := os.Stat(abs)
	if err != nil {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: stat %s: %v", abs, err))
		e.failNode(state, idx, nodeID)
		return
	}
	if info.IsDir() {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: %s is a directory", abs))
		e.failNode(state, idx, nodeID)
		return
	}

	contents, err := os.ReadFile(abs)
	if err != nil {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: read %s: %v", abs, err))
		e.failNode(state, idx, nodeID)
		return
	}
	truncated := false
	if len(contents) > fileLoaderMaxBytes {
		contents = contents[:fileLoaderMaxBytes]
		truncated = true
	}

	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = string(contents)
	if state.exec.Nodes[idx].OutputPaths == nil {
		state.exec.Nodes[idx].OutputPaths = map[string]string{}
	}
	state.exec.Nodes[idx].OutputPaths["file-path"] = abs
	state.mu.Unlock()

	e.emit("bmad:node:artifacts", NodeArtifactEvent{
		ExecID: state.exec.ID,
		NodeID: nodeID,
		Found:  []string{"file-path"},
		Paths:  map[string]string{"file-path": abs},
	})
	if truncated {
		log.Printf("bmad: file loader %s truncated %s at %d bytes", nodeID, abs, fileLoaderMaxBytes)
	}

	e.completeNode(state, idx, nodeID)
}

// extractRegex applies a regex to input and returns the first capture group
// (or the full match if no groups). Returns "" on no match or invalid pattern.
func extractRegex(input, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(input)
	if len(matches) == 0 {
		return ""
	}
	if len(matches) > 1 {
		return matches[1]
	}
	return matches[0]
}

// extractLines extracts lines from input by pattern: "2-4" (range), "-3" (last N), "5" (single).
// Line numbers are 1-indexed.
func extractLines(input, pattern string) string {
	lines := strings.Split(input, "\n")

	// Last N lines: "-3".
	if strings.HasPrefix(pattern, "-") {
		n, err := strconv.Atoi(pattern[1:])
		if err != nil || n <= 0 {
			return ""
		}
		if n > len(lines) {
			n = len(lines)
		}
		return strings.Join(lines[len(lines)-n:], "\n")
	}

	// Range or single: "2-4" or "5".
	parts := strings.SplitN(pattern, "-", 2)
	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 1 {
		return ""
	}
	start-- // convert to 0-indexed

	end := start + 1
	if len(parts) == 2 {
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return ""
		}
	}

	// Clamp.
	if start >= len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}
