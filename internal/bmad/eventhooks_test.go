// Package bmad — test-only note for the emit event hook.
// Story bmad-interactive-03: Suspension primitive + RespondToInput Wails binding.
//
// The production emit path (e.emit wrapper in executor.go) calls
// testEventHook when non-nil. The var declaration itself lives in
// executor.go so the production code can reference it; tests set and clear
// the hook via hookEvents() in executor_suspend_test.go.
package bmad
