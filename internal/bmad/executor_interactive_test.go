// Package bmad contains tests for the executor's interactive routing path.
// Story bmad-interactive-02: Executor routing for interactive nodes.
//
// RED Phase: These tests define expected behaviour.
// They FAIL until the go-engineer implements the feature (GREEN phase).
//
// Harness conventions (inherited from executor_command_test.go):
//   - All tests in package bmad (white-box).
//   - newHarness() / newSessionState() from executor_test.go / executor_session_test.go.
//   - successRunner() / idleCycleTrackingRunner() / idleMockRunner() from mock_helpers_test.go.
//   - t.TempDir() for file fixtures.
//   - waitForNodeStatus() for async node-completion assertions.
//
// Routing assertion strategy (AC-1 / AC-2):
//   Two package-level hook variables are declared in
//   executor_interactive_testhooks_test.go.  The go-engineer wires them inside
//   executeProcessNode (or whatever dispatch point gains the Mode switch) so
//   each hook is called with the nodeID when the respective code path is taken.
//   Tests register closures before starting the workflow and deregister in
//   t.Cleanup to avoid cross-test pollution.

package bmad

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Fixtures ──────────────────────────────────────────────────────────────────

// testInteractProcessID is the ID used to register a test-only ProcessDef with
// an InteractGuided mode in the in-memory registry. The go-engineer must expose
// a test-only registration path (e.g. registerTestProcess / setTestProcesses) or
// the tests will fail to resolve the ProcessDef via ProcessByID.
// Alternative: the engineer can inline the ProcessDef into execState.Nodes and
// add an override map on Executor for test injection — either pattern is fine as
// long as ProcessByID("test-interact-guided") returns the def below.
const testInteractProcessID = "test-interact-guided"

// interactGuidedProcessDef is the canonical test-only ProcessDef used across
// all interactive-routing tests. Mode = InteractGuided, no user-sourced
// InputSpecs, single memory OutputSpec — the simplest valid interactive process.
var interactGuidedProcessDef = ProcessDef{
	ID:        testInteractProcessID,
	Name:      "Test Interactive Guided",
	Mode:      InteractGuided,
	SkillName: "test-skill",
	InputSpecs: []InputSpec{}, // no inputs → resolveInputs returns empty resolved, nil missing, nil err
	OutputSpecs: []OutputSpec{
		{ID: "out1", Target: OutputToMemory},
	},
}

// saveInteractiveWorkflow persists a single-node WorkflowDef whose process ID
// is processID and whose NodeType is NodeTypeProcess.
func saveInteractiveWorkflow(t *testing.T, s *Storage, processID string) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          fmt.Sprintf("wf-interactive-%s", processID),
		Name:        "interactive-routing-fixture",
		Description: "single-node interactive workflow for S2 tests",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: processID,
				Label:     "Interactive Node",
				Status:    NodePending,
				NodeType:  NodeTypeProcess,
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// saveDownstreamWorkflow persists a two-node workflow: n1 (interactive process)
// → n2 (transform stub). Used for AC-3 downstream-start assertion.
func saveDownstreamWorkflow(t *testing.T, s *Storage) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          "wf-interactive-downstream",
		Name:        "interactive-downstream-fixture",
		Description: "n1→n2 DAG for downstream-node assertion",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: testInteractProcessID,
				Label:     "Interactive",
				Status:    NodePending,
				NodeType:  NodeTypeProcess,
			},
			{
				ID:        "n2",
				ProcessID: "bmad-brainstorming", // any real registry entry
				Label:     "Downstream",
				Status:    NodePending,
				NodeType:  NodeTypeProcess,
			},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// ── AC-1: Autonomous nodes route through executeNode (executeProcessNode) ─────

// TestRoutingDispatchesAutonomousNodesToExecuteNode (AC-1)
// Given a process node whose ProcessDef.Mode == "" (legacy/autonomous).
// When runDynamic picks it from the ready-set.
// Then the executeNode / executeProcessNode path is taken.
// And executeInteractiveNode is NOT entered for that node.
//
// Instrumented via testHookExecuteNode / testHookExecuteInteractiveNode
// (declared in executor_interactive_testhooks_test.go; wired by go-engineer).
func TestRoutingDispatchesAutonomousNodesToExecuteNode(t *testing.T) {
	// NOT parallel: hook globals (testHookExecuteNode, testHookExecuteInteractiveNode)
	// are shared across tests; parallel execution causes spurious cross-counts.
	var executeNodeCalls int32
	var executeInteractiveCalls int32

	// Wire routing hooks — deregister in cleanup to avoid cross-test pollution.
	testHookExecuteNode = func(nodeID string) {
		atomic.AddInt32(&executeNodeCalls, 1)
	}
	testHookExecuteInteractiveNode = func(nodeID string) {
		atomic.AddInt32(&executeInteractiveCalls, 1)
	}
	t.Cleanup(func() {
		testHookExecuteNode = nil
		testHookExecuteInteractiveNode = nil
	})

	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	// "bmad-brainstorming" has Mode == "" (autonomous legacy process).
	wfID := saveSingleNodeWorkflow(t, h.storage, NodeTypeProcess, map[string]string{})
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeComplete)

	assert.GreaterOrEqual(t, atomic.LoadInt32(&executeNodeCalls), int32(1),
		"executeNode/executeProcessNode must be called for Mode==\"\" nodes")
	assert.Equal(t, int32(0), atomic.LoadInt32(&executeInteractiveCalls),
		"executeInteractiveNode must NOT be called for autonomous nodes")
}

// ── AC-2: Interactive modes route to executeInteractiveNode ──────────────────

// TestRoutingDispatchesInteractiveModesToExecuteInteractiveNode (AC-2)
// Table: Mode = InteractGuided, InteractIterative, InteractParty.
// For each: build state with one process node pointing to a test-only ProcessDef.
// Assert executeInteractiveNode reached; executeNode NOT reached.
func TestRoutingDispatchesInteractiveModesToExecuteInteractiveNode(t *testing.T) {
	tests := []struct {
		name      string
		mode      InteractionMode
		processID string
	}{
		{
			name:      "InteractGuided routes to executeInteractiveNode",
			mode:      InteractGuided,
			processID: "test-interact-guided",
		},
		{
			name:      "InteractIterative routes to executeInteractiveNode",
			mode:      InteractIterative,
			processID: "test-interact-iterative",
		},
		{
			name:      "InteractParty routes to executeInteractiveNode",
			mode:      InteractParty,
			processID: "test-interact-party",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// NOT parallel: routing hook globals are shared across tests.
			var executeNodeCalls int32
			var executeInteractiveCalls int32

			testHookExecuteNode = func(nodeID string) {
				atomic.AddInt32(&executeNodeCalls, 1)
			}
			testHookExecuteInteractiveNode = func(nodeID string) {
				atomic.AddInt32(&executeInteractiveCalls, 1)
			}
			t.Cleanup(func() {
				testHookExecuteNode = nil
				testHookExecuteInteractiveNode = nil
			})

			h := newHarness(t)
			// Interactive node: use idleCycleTrackingRunner so the session completes
			// cleanly if executeInteractiveNode falls back to the session path.
			runner, _ := idleCycleTrackingRunner()
			h.executor.SetCommandRunner(runner)

			// The test-only ProcessDef for this mode must be registered by the
			// go-engineer via a test registration hook.  The test references the ID;
			// registration is a GREEN-phase concern.
			wfID := saveInteractiveWorkflow(t, h.storage, tt.processID)
			exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
			require.NoError(t, err)
			t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

			// Node must reach a terminal state (complete or failed — either proves
			// executeInteractiveNode was entered, not executeNode).
			require.Eventuallyf(t, func() bool {
				got, err := h.executor.GetExecution(exec.ID)
				if err != nil {
					return false
				}
				for _, n := range got.Nodes {
					if n.ID == "n1" {
						return n.Status == NodeComplete || n.Status == NodeFailed
					}
				}
				return false
			}, 3*time.Second, 50*time.Millisecond,
				"node n1 must reach a terminal state")

			assert.GreaterOrEqual(t, atomic.LoadInt32(&executeInteractiveCalls), int32(1),
				"executeInteractiveNode must be called for Mode=%q nodes", tt.mode)
			assert.Equal(t, int32(0), atomic.LoadInt32(&executeNodeCalls),
				"executeNode/executeProcessNode must NOT be called for Mode=%q nodes", tt.mode)
		})
	}
}

// ── AC-3: executeInteractiveNode happy path ───────────────────────────────────

// TestExecuteInteractiveNodeHappyPath (AC-3)
// ProcessDef: Mode=InteractGuided, no user-sourced InputSpecs, one OutputSpec
// with Target=OutputToMemory.
// When executeInteractiveNode runs end-to-end:
//   - node transitions pending → running → complete
//   - NodeOutputs[nodeID] is populated from captured round output
//   - downstream node (n2) begins execution (i.e. its in-degree decrements to 0)
func TestExecuteInteractiveNodeHappyPath(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	// idleCycleTrackingRunner simulates: pane alive, captures settle to idle prompt,
	// completing the waitForIdleCompletion call inside executeInteractiveNode.
	runner, _ := idleCycleTrackingRunner()
	h.executor.SetCommandRunner(runner)

	wfID := saveDownstreamWorkflow(t, h.storage)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	// n1 (interactive) must complete.
	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeComplete)

	got, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	// NodeOutputs["n1"] must be non-empty (captureRoundOutput populated it).
	assert.NotEmpty(t, got.NodeOutputs["n1"],
		"NodeOutputs[n1] must be populated after executeInteractiveNode completes")

	// n2 must have started (or completed) — proves activeOutEdges fired
	// and decremented n2's in-degree after n1 completed.
	require.Eventuallyf(t, func() bool {
		snapshot, err := h.executor.GetExecution(exec.ID)
		if err != nil {
			return false
		}
		for _, n := range snapshot.Nodes {
			if n.ID == "n2" {
				return n.Status == NodeRunning || n.Status == NodeComplete || n.Status == NodeFailed
			}
		}
		return false
	}, 3*time.Second, 50*time.Millisecond,
		"n2 must start executing after n1 completes (downstream in-degree decremented)")
}

// ── AC-4: resolveInputs table ─────────────────────────────────────────────────

// TestResolveInputsTable (AC-4)
// Covers all four non-user source types plus user source (required/optional).
// resolveInputs is called directly — it does not exist yet (RED phase).
func TestResolveInputsTable(t *testing.T) {
	// Create a temp directory tree for file-based fixtures.
	tmpDir := t.TempDir()
	bmadOut := filepath.Join(tmpDir, "_bmad-output", "analysis-artifacts")
	require.NoError(t, os.MkdirAll(bmadOut, 0o755))

	// Write product-brief artifact file.
	briefContent := "# Product Brief\nSome content."
	briefPath := filepath.Join(bmadOut, "product-brief.md")
	require.NoError(t, os.WriteFile(briefPath, []byte(briefContent), 0o644))

	// Write a registry CSV for the registry lookup test.
	registryDir := filepath.Join(tmpDir, "registry")
	require.NoError(t, os.MkdirAll(registryDir, 0o755))
	csvPath := filepath.Join(registryDir, "methods.csv")
	f, err := os.Create(csvPath)
	require.NoError(t, err)
	w := csv.NewWriter(f)
	require.NoError(t, w.WriteAll([][]string{
		{"method_name", "description"},
		{"SCAMPER", "Creative ideation technique"},
		{"TRIZ",    "Inventive problem solving"},
	}))
	w.Flush()
	require.NoError(t, f.Close())

	tests := []struct {
		name           string
		specs          []InputSpec
		nodeOutputs    map[string]string // pre-seeded NodeOutputs (upstream values)
		nodeInputs     map[string]map[string]string // pre-seeded NodeInputs (user values)
		repoPath       string
		wantResolved   map[string]string // expected keys in resolved map
		wantMissingIDs []string          // expected spec IDs in missing slice
		wantErr        bool
	}{
		{
			name: "file present resolves to content",
			specs: []InputSpec{
				{ID: "brief", Source: InputFromFile, ArtifactName: "product-brief", Required: true},
			},
			repoPath:     tmpDir,
			wantResolved: map[string]string{"brief": briefContent},
		},
		{
			name: "file missing required returns error",
			specs: []InputSpec{
				{ID: "missing-file", Source: InputFromFile, ArtifactName: "domain-research", Required: true},
			},
			repoPath: tmpDir, // domain-research.md not written
			wantErr:  true,
		},
		{
			name: "file missing optional is skipped without error",
			specs: []InputSpec{
				{ID: "opt-file", Source: InputFromFile, ArtifactName: "market-research", Required: false},
			},
			repoPath:     tmpDir, // market-research.md not written
			wantResolved: map[string]string{}, // no key added for optional missing
		},
		{
			name: "upstream present resolves via NodeOutputs",
			specs: []InputSpec{
				{ID: "prior", Source: InputFromUpstream, UpstreamNodeID: "upstream-node", Required: true},
			},
			nodeOutputs:  map[string]string{"upstream-node": "upstream output text"},
			wantResolved: map[string]string{"prior": "upstream output text"},
		},
		{
			name: "upstream missing required returns error",
			specs: []InputSpec{
				{ID: "prior-missing", Source: InputFromUpstream, UpstreamNodeID: "absent-node", Required: true},
			},
			nodeOutputs: map[string]string{},
			wantErr:     true,
		},
		{
			name: "user required with no answer goes to missing slice",
			specs: []InputSpec{
				{ID: "topic", Source: InputFromUser, Required: true, Shape: ShapeFree},
			},
			wantResolved:   map[string]string{},
			wantMissingIDs: []string{"topic"},
		},
		{
			name: "user optional with Default resolves to default",
			specs: []InputSpec{
				{ID: "method", Source: InputFromUser, Required: false, Default: "SCAMPER"},
			},
			wantResolved: map[string]string{"method": "SCAMPER"},
		},
		{
			name: "user provided value resolves to provided value",
			specs: []InputSpec{
				{ID: "topic", Source: InputFromUser, Required: true, Shape: ShapeFree},
			},
			nodeInputs:   map[string]map[string]string{"n1": {"topic": "my topic"}},
			wantResolved: map[string]string{"topic": "my topic"},
		},
		{
			name: "env source returns stub string (no error)",
			specs: []InputSpec{
				{ID: "branch", Source: InputFromEnv},
			},
			wantResolved: map[string]string{"branch": ""}, // envValue stub returns ""
		},
		{
			name: "registry lookup column extract",
			specs: []InputSpec{
				{
					ID:         "technique",
					Source:     InputFromRegistry,
					OptionsRef: fmt.Sprintf("registry:%s#method_name", csvPath),
					Required:   true,
				},
			},
			wantResolved: map[string]string{"technique": "SCAMPER"}, // first row after header
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			proc := ProcessDef{
				ID:         testInteractProcessID,
				Name:       "test",
				InputSpecs: tt.specs,
			}

			nodes := []WorkflowNode{
				{
					ID:         "n1",
					ProcessID:  proc.ID,
					Label:      "Test Node",
					Status:     NodeRunning,
					NodeType:   NodeTypeProcess,
					InputSpecs: tt.specs,
				},
			}
			state, _ := newSessionState(nodes, nil)
			// Inject repo path via state.exec.RepoPath.
			if tt.repoPath != "" {
				state.exec.RepoPath = tt.repoPath
			}
			// Seed NodeOutputs for upstream tests.
			if tt.nodeOutputs != nil {
				state.exec.NodeOutputs = tt.nodeOutputs
			}
			// Seed NodeInputs for user-provided tests.
			if tt.nodeInputs != nil {
				state.exec.NodeInputs = tt.nodeInputs
			}

			e := NewExecutor(nil, func(string, interface{}) {})

			// resolveInputs does not exist yet — this call fails to compile (RED).
			resolved, missing, err := e.resolveInputs(context.Background(), state, "n1", 1)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify expected resolved entries.
			for k, want := range tt.wantResolved {
				assert.Equal(t, want, resolved[k], "resolved[%q]", k)
			}

			// Verify missing IDs.
			var missingIDs []string
			for _, spec := range missing {
				missingIDs = append(missingIDs, spec.ID)
			}
			assert.Equal(t, tt.wantMissingIDs, missingIDs,
				"missing spec IDs must match expected")
		})
	}
}

// ── AC-5: verifyOutputs failure on missing required file ─────────────────────

// TestVerifyOutputsRequiredFileMissing (AC-5, companion rows)
// verifyOutputs does not exist yet — calls fail to compile (RED).
func TestVerifyOutputsRequiredFileMissing(t *testing.T) {
	tests := []struct {
		name         string
		spec         OutputSpec
		createFile   bool // whether to create the artifact on disk before calling
		wantErr      bool
	}{
		{
			name: "required file absent → error",
			spec: OutputSpec{
				ID:           "brief-out",
				Target:       OutputToFile,
				ArtifactName: "product-brief",
				Optional:     false,
			},
			createFile: false,
			wantErr:    true,
		},
		{
			name: "required file present → nil error",
			spec: OutputSpec{
				ID:           "brief-out",
				Target:       OutputToFile,
				ArtifactName: "product-brief",
				Optional:     false,
			},
			createFile: true,
			wantErr:    false,
		},
		{
			name: "optional file absent → nil error",
			spec: OutputSpec{
				ID:           "brief-out",
				Target:       OutputToFile,
				ArtifactName: "product-brief",
				Optional:     true,
			},
			createFile: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			if tt.createFile {
				// Mirror the artifactPaths layout: _bmad-output/analysis-artifacts/product-brief.md
				artifactDir := filepath.Join(tmpDir, "_bmad-output", "analysis-artifacts")
				require.NoError(t, os.MkdirAll(artifactDir, 0o755))
				require.NoError(t, os.WriteFile(
					filepath.Join(artifactDir, "product-brief.md"),
					[]byte("content"),
					0o644,
				))
			}

			state, _ := newSessionState([]WorkflowNode{
				{ID: "n1", ProcessID: testInteractProcessID, Status: NodeRunning, NodeType: NodeTypeProcess},
			}, nil)

			e := NewExecutor(nil, func(string, interface{}) {})

			// verifyOutputs does not exist yet — RED phase.
			err := e.verifyOutputs(state, "n1", []OutputSpec{tt.spec}, tmpDir)

			if tt.wantErr {
				require.Error(t, err,
					"verifyOutputs must return error when required file is absent")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestVerifyOutputsMemoryTargetAlwaysOK (AC-5 companion)
// OutputToMemory outputs are always satisfied; no filesystem check needed.
func TestVerifyOutputsMemoryTargetAlwaysOK(t *testing.T) {
	t.Parallel()

	spec := OutputSpec{
		ID:       "out1",
		Target:   OutputToMemory,
		Optional: false,
	}

	state, _ := newSessionState([]WorkflowNode{
		{ID: "n1", ProcessID: testInteractProcessID, Status: NodeRunning, NodeType: NodeTypeProcess},
	}, nil)

	e := NewExecutor(nil, func(string, interface{}) {})

	// verifyOutputs does not exist yet — RED phase.
	err := e.verifyOutputs(state, "n1", []OutputSpec{spec}, t.TempDir())
	assert.NoError(t, err,
		"OutputToMemory must always pass verifyOutputs regardless of filesystem")
}

// ── AC-6: activeOutEdges defensive log ───────────────────────────────────────

// TestActiveOutEdgesLogsOnNonCompleteNode (AC-6)
// Given a node in NodeRunning status.
// When activeOutEdges is called on it.
// Then the log output contains "activeOutEdges called for node".
// And the return value is unchanged vs the same call with NodeComplete.
//
// Uses log.SetOutput to a buffer — not parallel (global log state).
func TestActiveOutEdgesLogsOnNonCompleteNode(t *testing.T) {
	// Not parallel: captures global log output.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	// Build minimal state: one running node, one pending downstream node, one edge.
	nodes := []WorkflowNode{
		{ID: "n1", Label: "Running", Status: NodeRunning, NodeType: NodeTypeProcess,
			ProcessID: "bmad-brainstorming"},
		{ID: "n2", Label: "Downstream", Status: NodePending, NodeType: NodeTypeProcess,
			ProcessID: "bmad-brainstorming"},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "n1", Target: "n2"},
	}
	state, _ := newSessionState(nodes, edges)

	e := NewExecutor(nil, func(string, interface{}) {})

	// Call activeOutEdges with NodeRunning — should log a diagnostic.
	runningEdges := e.activeOutEdges(state, "n1", "", NodeTypeProcess)

	assert.Contains(t, logBuf.String(), "activeOutEdges called for node",
		"log must contain diagnostic when node status is not NodeComplete")

	// Reset log buffer and flip n1 to NodeComplete.
	logBuf.Reset()
	state.mu.Lock()
	state.exec.Nodes[0].Status = NodeComplete
	state.mu.Unlock()

	// Call again with NodeComplete — must NOT log the diagnostic.
	completeEdges := e.activeOutEdges(state, "n1", "", NodeTypeProcess)

	assert.NotContains(t, logBuf.String(), "activeOutEdges called for node",
		"log must NOT contain diagnostic when node status is NodeComplete")

	// Both calls must return the same edges (semantic unchanged).
	assert.Equal(t, completeEdges, runningEdges,
		"return value must be identical regardless of status — no semantic change")
}

// ── registryLookup security + correctness ────────────────────────────────────

// TestRegistryLookupRejectsNonRegistryScheme
// registryLookup does not exist yet — RED phase.
func TestRegistryLookupRejectsNonRegistryScheme(t *testing.T) {
	t.Parallel()

	// Build a real CSV fixture for the valid-ref row.
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "methods.csv")
	f, err := os.Create(csvPath)
	require.NoError(t, err)
	w := csv.NewWriter(f)
	require.NoError(t, w.WriteAll([][]string{
		{"method_name", "description"},
		{"SCAMPER", "Creative technique"},
		{"TRIZ", "Inventive problems"},
	}))
	w.Flush()
	require.NoError(t, f.Close())

	tests := []struct {
		name       string
		ref        string
		wantErr    bool
		wantNonEmpty bool
		wantCount  int // >0 means expect a multi-value response
	}{
		{
			name:    "file:// scheme rejected for security",
			ref:     "file:../secrets",
			wantErr: true,
		},
		{
			name:    "http:// scheme rejected for security",
			ref:     "http://evil.example.com",
			wantErr: true,
		},
		{
			name:    "mcp: scheme rejected (reserved, not yet supported)",
			ref:     "mcp:x#y",
			wantErr: true,
		},
		{
			name:         "registry: column extract returns first row value",
			ref:          fmt.Sprintf("registry:%s#method_name", csvPath),
			wantNonEmpty: true,
		},
		{
			name:       "registry: random=2 returns multiple CSV rows",
			ref:        fmt.Sprintf("registry:%s?random=2", csvPath),
			wantNonEmpty: true,
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// registryLookup does not exist yet — RED phase.
			result, err := registryLookup(tt.ref)

			if tt.wantErr {
				require.Error(t, err,
					"registryLookup must reject non-registry scheme refs")
				return
			}

			require.NoError(t, err)

			if tt.wantNonEmpty {
				assert.NotEmpty(t, result,
					"registryLookup must return non-empty result for valid ref")
			}

			if tt.wantCount > 0 {
				// For random=N queries, result is expected to be a
				// newline-separated or comma-separated list of N values.
				// Accept either delimiter; the go-engineer decides the format.
				parts := strings.FieldsFunc(result, func(r rune) bool {
					return r == '\n' || r == ','
				})
				assert.GreaterOrEqual(t, len(parts), tt.wantCount,
					"registryLookup random=2 must return at least 2 values")
			}
		})
	}
}

// ── truncate cap ─────────────────────────────────────────────────────────────

// TestTruncateCap
// upstreamOutputCap = 2000 per §4 "Bounded memory".
// truncate does not exist yet — RED phase.
func TestTruncateCap(t *testing.T) {
	t.Parallel()

	const cap = upstreamOutputCap // must equal 2000 per §4

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantSame bool // true → output must equal input byte-for-byte
	}{
		{
			name:     "empty string passes through",
			input:    "",
			wantSame: true,
		},
		{
			name:     "string shorter than cap passes through",
			input:    strings.Repeat("x", cap-1),
			wantSame: true,
		},
		{
			name:    "string exactly at cap passes through",
			input:   strings.Repeat("x", cap),
			wantLen: cap,
		},
		{
			name:    "string longer than cap is cut at cap",
			input:   strings.Repeat("x", cap+500),
			wantLen: cap,
		},
		{
			name:    "very long string is cut at cap",
			input:   strings.Repeat("a", cap*3),
			wantLen: cap,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// truncate does not exist yet — RED phase.
			got := truncate(tt.input, cap)

			if tt.wantSame {
				assert.Equal(t, tt.input, got,
					"strings shorter than cap must pass through unchanged")
				return
			}

			assert.Equal(t, tt.wantLen, len(got),
				"truncated string must have length == cap")
		})
	}
}
