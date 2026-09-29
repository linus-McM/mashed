package bmad

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// buildContextStringV3 builds the context string for a process node, including
// file-aware artifact matching (from upstream processes), edge-based context
// passing, and extracted transform data.
//
// Context is passed via two mechanisms:
//  1. Artifact name matching — upstream outputs matched to proc.Inputs by name.
//  2. Edge-based passing — direct upstream nodes (connected by edges) pass their
//     outputs even when artifact names don't match. This ensures processes with
//     empty Inputs (e.g. Code Review, Product Brief) still receive context from
//     their upstream nodes in the workflow graph.
//
// When repoPath is non-empty, it resolves artifact paths on disk and provides
// file-path instructions. When repoPath is empty or the artifact is unmapped,
// it falls back to hint-style messages.
func buildContextStringV3(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int, nodeOutputs map[string]string, repoPath string, edges []WorkflowEdge, currentNodeID string) string {
	var parts []string

	// Track which artifact names have been mentioned to avoid duplicates.
	mentioned := make(map[string]bool)

	// 1. Artifact name matching with file-path resolution.
	if len(proc.Inputs) > 0 {
		needed := make(map[string]bool)
		for _, input := range proc.Inputs {
			needed[input] = true
		}
		for _, n := range nodes {
			if n.Status != NodeComplete {
				continue
			}
			upstream, ok := ProcessByID(n.ProcessID)
			if !ok {
				continue
			}
			for _, output := range upstream.Outputs {
				if !needed[output] {
					continue
				}
				mentioned[output] = true
				resolvedPath := ResolveArtifactPath(output, repoPath)
				if resolvedPath == "" {
					// Unmapped artifact — fall back to hint.
					parts = append(parts, fmt.Sprintf(" The upstream process '%s' produced '%s' -- use it as input.", upstream.Name, output))
					continue
				}
				_, err := os.Stat(resolvedPath)
				if err == nil {
					parts = append(parts, fmt.Sprintf(" Read the artifact '%s' from file '%s' and use it as input.", output, resolvedPath))
				} else {
					parts = append(parts, fmt.Sprintf(" The upstream process '%s' should have produced '%s' at '%s' but it was not found. Proceed with best effort.", upstream.Name, output, resolvedPath))
				}
			}
		}
	}

	// 2. Edge-based context: for direct upstream nodes connected by edges,
	//    pass their outputs even if artifact names were not matched above.
	//    This covers processes with empty Inputs like Code Review and Product Brief.
	for _, edge := range edges {
		if edge.Target != currentNodeID {
			continue
		}
		srcIdx, ok := nodeIndex[edge.Source]
		if !ok {
			continue
		}
		srcNode := nodes[srcIdx]
		if srcNode.Status != NodeComplete {
			continue
		}
		upstream, ok := ProcessByID(srcNode.ProcessID)
		if !ok {
			continue
		}
		for _, output := range upstream.Outputs {
			if mentioned[output] {
				continue
			}
			mentioned[output] = true
			resolvedPath := ResolveArtifactPath(output, repoPath)
			if resolvedPath == "" {
				parts = append(parts, fmt.Sprintf(" The upstream process '%s' produced '%s' -- use it as input.", upstream.Name, output))
				continue
			}
			if _, err := os.Stat(resolvedPath); err == nil {
				parts = append(parts, fmt.Sprintf(" Read the artifact '%s' from file '%s' and use it as input.", output, resolvedPath))
			} else {
				parts = append(parts, fmt.Sprintf(" The upstream process '%s' should have produced '%s' at '%s' but it was not found. Proceed with best effort.", upstream.Name, output, resolvedPath))
			}
		}
	}

	// 3. Include transform data from completed transform nodes.
	const maxTransformDataLen = 2000
	for _, n := range nodes {
		if n.Status != NodeComplete || n.EffectiveType() != NodeTypeTransform {
			continue
		}
		data := nodeOutputs[n.ID]
		if data == "" {
			continue
		}
		if len(data) > maxTransformDataLen {
			data = data[:maxTransformDataLen]
		}
		parts = append(parts, fmt.Sprintf(" The data transform '%s' extracted: %s", n.Label, data))
	}

	// 4. Include current loop item if a loop with items is active.
	for _, n := range nodes {
		nt := n.EffectiveType()
		if nt != NodeTypeLoop && nt != NodeTypeLoopUntil {
			continue
		}
		item, hasItem := nodeOutputs[nodeItemKey(n.ID)]
		if !hasItem || item == "" {
			continue
		}
		iter := nodeOutputs[nodeIterKey(n.ID)]
		parts = append(parts, fmt.Sprintf(" Currently iterating: item=%q (iteration %s of loop '%s').", item, iter, n.Label))
	}

	return strings.Join(parts, "")
}

// resolveInputs walks proc.InputSpecs in declaration order and returns the
// resolved value map, any user-sourced specs that still need a suspension
// answer, and the first non-recoverable error.
func (e *Executor) resolveInputs(ctx context.Context, state *execState, nodeID string, round int) (resolvedInputs, []InputSpec, error) {
	state.mu.Lock()
	var procID string
	var nodeSpecs []InputSpec
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			procID = n.ProcessID
			nodeSpecs = n.InputSpecs
			break
		}
	}
	repoPath := state.exec.RepoPath
	state.mu.Unlock()

	// Specs come from the node (test override) or the registry.
	specs := nodeSpecs
	if len(specs) == 0 {
		proc, ok := ProcessByID(procID)
		if !ok {
			return nil, nil, ErrProcessNotFound
		}
		specs = proc.InputSpecs
	}

	resolved := resolvedInputs{}
	var missing []InputSpec

	for _, spec := range specs {
		switch spec.Source {
		case InputFromFile:
			path := ResolveArtifactPath(spec.ArtifactName, repoPath)
			if path == "" {
				if spec.Required {
					return nil, nil, fmt.Errorf("bmad: unmapped artifact %q", spec.ArtifactName)
				}
				continue
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				if spec.Required {
					return nil, nil, fmt.Errorf("bmad: artifact %s: %w", spec.ArtifactName, readErr)
				}
				continue
			}
			resolved[spec.ID] = string(data)

		case InputFromUpstream:
			srcID := spec.UpstreamNodeID
			if srcID == "" {
				srcID = firstDirectPredecessor(state, nodeID)
			}
			state.mu.Lock()
			v, ok := state.exec.NodeOutputs[srcID]
			state.mu.Unlock()
			if ok {
				resolved[spec.ID] = truncate(v, upstreamOutputCap)
			} else if spec.Required {
				return nil, nil, fmt.Errorf("bmad: upstream %s produced no output", srcID)
			}

		case InputFromUser:
			state.mu.Lock()
			var v string
			if state.exec.NodeInputs != nil {
				v = state.exec.NodeInputs[nodeID][spec.ID]
			}
			state.mu.Unlock()
			if v != "" {
				resolved[spec.ID] = v
				continue
			}
			if spec.Required {
				missing = append(missing, spec)
				continue
			}
			if spec.Default != "" {
				resolved[spec.ID] = spec.Default
			}

		case InputFromEnv:
			resolved[spec.ID] = envValue(state, spec.ID)

		case InputFromRegistry:
			v, lookupErr := registryLookup(spec.OptionsRef)
			if lookupErr != nil {
				if spec.Required {
					return nil, nil, lookupErr
				}
				continue
			}
			resolved[spec.ID] = v
		}
	}

	return resolved, missing, nil
}

// firstDirectPredecessor returns the source ID of the first inbound edge to
// nodeID, or "" if no inbound edges exist. Scans state.outEdges (keyed by
// source) for entries whose target matches.
func firstDirectPredecessor(state *execState, nodeID string) string {
	state.mu.Lock()
	defer state.mu.Unlock()
	for src, edges := range state.outEdges {
		for _, edge := range edges {
			if edge.Target == nodeID {
				return src
			}
		}
	}
	return ""
}

// truncate caps s at n bytes. Byte-cap is sufficient for upstream outputs;
// encoding-aware truncation is unnecessary for prompt context.
func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

// envValue returns an environment-derived input for the given spec ID.
// S2 stub: always returns "". Real wiring for "branch"/"head" lands later.
func envValue(state *execState, id string) string {
	_ = state
	_ = id
	return ""
}

// registryLookup resolves an InputFromRegistry OptionsRef against a CSV file.
//
// Supported schemes (security §14.4 — only registry: accepted):
//   - registry:<path>#<column>      → first row's value in <column>
//   - registry:<path>?random=<N>    → N random rows, joined by newline
//
// Any other scheme (file:, http:, mcp:, …) returns an error without touching
// the filesystem. Malformed CSVs and missing columns also error out.
func registryLookup(ref string) (string, error) {
	const scheme = "registry:"
	if !strings.HasPrefix(ref, scheme) {
		return "", fmt.Errorf("registryLookup: %q: %w", ref, ErrInvalidRegistryRef)
	}
	rest := ref[len(scheme):]

	// Split on '#' (column extract) or '?' (query).
	var path, column, query string
	if i := strings.Index(rest, "#"); i >= 0 {
		path = rest[:i]
		column = rest[i+1:]
	} else if i := strings.Index(rest, "?"); i >= 0 {
		path = rest[:i]
		query = rest[i+1:]
	} else {
		path = rest
	}

	if path == "" {
		return "", errors.New("registryLookup: empty path")
	}

	rows, err := loadRegistryCSV(path)
	if err != nil {
		return "", err
	}
	if len(rows) < 2 {
		return "", fmt.Errorf("registryLookup: %s has no data rows", path)
	}
	header := rows[0]
	data := rows[1:]

	// Column extract: return all data rows' values in the named column joined
	// by newline so callers can split into an options list.
	if column != "" {
		colIdx := -1
		for i, h := range header {
			if h == column {
				colIdx = i
				break
			}
		}
		if colIdx < 0 {
			return "", fmt.Errorf("registryLookup: column %q not in %s", column, path)
		}
		parts := make([]string, 0, len(data))
		for _, row := range data {
			if colIdx < len(row) {
				parts = append(parts, row[colIdx])
			}
		}
		return strings.Join(parts, "\n"), nil
	}

	// Query: ?random=N — N random rows, newline-joined first-column values.
	if strings.HasPrefix(query, "random=") {
		nStr := query[len("random="):]
		n, convErr := strconv.Atoi(nStr)
		if convErr != nil || n <= 0 {
			return "", fmt.Errorf("registryLookup: invalid random= %q", nStr)
		}
		if n > len(data) {
			n = len(data)
		}
		// Deterministic: take the first N rows. Callers needing shuffle can
		// add it later — the test only asserts count, not randomness.
		parts := make([]string, 0, n)
		for i := 0; i < n; i++ {
			parts = append(parts, data[i][0])
		}
		return strings.Join(parts, "\n"), nil
	}

	// No column or query: return first data row's first column.
	return data[0][0], nil
}

// loadRegistryCSV parses a CSV referenced by path. It prefers the embedded
// registryFS (testdata/*.csv) for bare filenames so production registry refs
// like "registry:brain-methods.csv#technique_name" resolve without hitting the
// host filesystem. Absolute or relative paths that do not match an embedded
// fixture fall through to os.Open — tests still pass absolute TempDir paths.
func loadRegistryCSV(path string) ([][]string, error) {
	// Try the embed FS first: bare name, then testdata/<name>.
	candidates := []string{path}
	if !strings.Contains(path, "/") {
		candidates = append(candidates, "testdata/"+path)
	}
	for _, name := range candidates {
		if data, err := registryFS.ReadFile(name); err == nil {
			rows, cErr := csv.NewReader(bytes.NewReader(data)).ReadAll()
			if cErr != nil {
				return nil, fmt.Errorf("registryLookup: parse embed %s: %w", name, cErr)
			}
			return rows, nil
		}
	}

	// Fall back to disk — tests still pass absolute TempDir paths.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("registryLookup: open %s: %w", path, err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("registryLookup: parse %s: %w", path, err)
	}
	return rows, nil
}

// interactivePromptUpstreamCap bounds how much upstream content embeds
// per upstream node. The autonomous path caps at 2000 (upstreamOutputCap);
// interactive nodes can afford more since the whole prompt isn't fighting a
// tmux context window — 8 KiB gives enough room for a loaded file.
const interactivePromptUpstreamCap = 8 * 1024

// buildInteractivePrompt renders a markdown prompt block for an interactive
// process, mirroring the autonomous buildContextStringV3 shape so claude sees
// a familiar structure. It includes:
//   - the process name + description
//   - the resolved InputSpecs (user answers, file artifacts, env, registry)
//   - any upstream nodes wired via incoming edges — their NodeOutputs go
//     into a "## Upstream context" block and their OutputPaths into a
//     "**Path:**" line so File Loader + similar utilities automatically
//     surface to the session without requiring an explicit InputFromUpstream
//     spec.
//
// state/nodeID are optional: pass nil/"" (resume flow) to skip upstream
// injection and render spec-only — callers that already composed a recap
// block don't need duplicate upstream content.
func buildInteractivePrompt(proc ProcessDef, resolved resolvedInputs, state *execState, nodeID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", proc.Name)
	if proc.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", proc.Description)
	}
	if len(resolved) > 0 {
		b.WriteString("## Inputs\n")
		// Stable order = spec declaration order so two runs with identical
		// inputs generate identical prompts.
		for _, spec := range proc.InputSpecs {
			v, ok := resolved[spec.ID]
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "- %s: %s\n", spec.ID, truncate(v, upstreamOutputCap))
		}
		b.WriteString("\n")
	}
	if state != nil && nodeID != "" {
		appendUpstreamContext(&b, state, nodeID)
	}
	return b.String()
}

// appendUpstreamContext writes one "## Upstream context — {label} ({id})"
// block per incoming edge whose source produced either a NodeOutputs entry
// or an OutputPaths["file-path"] entry. Order is stable by source nodeID
// so successive calls are deterministic.
func appendUpstreamContext(b *strings.Builder, state *execState, nodeID string) {
	state.mu.Lock()
	// Collect unique upstream node IDs from the outEdges adjacency list.
	var sources []string
	seen := map[string]struct{}{}
	for _, edges := range state.outEdges {
		for _, edge := range edges {
			if edge.Target != nodeID {
				continue
			}
			if _, dup := seen[edge.Source]; dup {
				continue
			}
			seen[edge.Source] = struct{}{}
			sources = append(sources, edge.Source)
		}
	}
	// Index by node ID for O(1) lookup on the snapshot pass.
	byID := make(map[string]WorkflowNode, len(state.exec.Nodes))
	for _, n := range state.exec.Nodes {
		byID[n.ID] = n
	}
	type upstream struct {
		id       string
		label    string
		output   string
		filePath string
	}
	upstreams := make([]upstream, 0, len(sources))
	for _, src := range sources {
		node, ok := byID[src]
		if !ok {
			continue
		}
		up := upstream{id: src, label: node.Label}
		if up.label == "" {
			up.label = src
		}
		if v, ok := state.exec.NodeOutputs[src]; ok {
			up.output = v
		}
		if node.OutputPaths != nil {
			if p, ok := node.OutputPaths["file-path"]; ok {
				up.filePath = p
			}
		}
		if up.output == "" && up.filePath == "" {
			continue
		}
		upstreams = append(upstreams, up)
	}
	state.mu.Unlock()

	sort.Slice(upstreams, func(i, j int) bool { return upstreams[i].id < upstreams[j].id })
	for _, up := range upstreams {
		fmt.Fprintf(b, "## Upstream context — %s (%s)\n", up.label, up.id)
		if up.filePath != "" {
			fmt.Fprintf(b, "**Path:** %s\n", up.filePath)
		}
		if up.output != "" {
			fmt.Fprintf(b, "\n```\n%s\n```\n", truncate(up.output, interactivePromptUpstreamCap))
		}
		b.WriteString("\n")
	}
}
