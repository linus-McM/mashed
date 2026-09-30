// Package bmad — test-only hook variables for interactive routing assertions.
// Story bmad-interactive-02: Executor routing for interactive nodes.
//
// These hooks are declared here (not in executor.go) to avoid shipping
// test scaffolding in the production binary.
//
// GREEN-phase wiring (go-engineer task):
//
//   Inside executeProcessNode (or wherever the Mode switch is inserted), add:
//
//     if testHookExecuteNode != nil {
//         testHookExecuteNode(nodeID)
//     }
//
//   Inside executeInteractiveNode, at the very top after idx resolution, add:
//
//     if testHookExecuteInteractiveNode != nil {
//         testHookExecuteInteractiveNode(nodeID)
//     }
//
//   Both hooks are nil in production (hook vars only exist in test binaries).
//
// Hook concurrency:
//   Hooks are written from a single test goroutine before StartWorkflow and
//   cleared in t.Cleanup after the test settles.  The hooks are read inside
//   goroutines spawned by runDynamic.  This is safe under the Go memory model
//   as long as the write happens-before StartWorkflow and the cleanup runs
//   after all node goroutines have returned (guaranteed by waitForNodeStatus +
//   StopWorkflow cleanup).

package bmad

// Hooks and the upstreamOutputCap constant are now declared in executor.go
// so production code can reference them. Tests import them as package-level
// identifiers. Keeping this file for doc + the package directive.
