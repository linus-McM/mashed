// Package bmad — test-only event hook for suspension/respond assertions.
// Story bmad-interactive-03: Suspension primitive + RespondToInput Wails binding.
//
// The production emit path (e.emitEvent) calls testEventHook when non-nil.
// GREEN-phase wiring (go-engineer):
//
//	In every e.emitEvent call site, or in a wrapper func (e *Executor) emit(name string, data interface{}):
//
//	  if testEventHook != nil {
//	      testEventHook(name, data)
//	  }
//	  e.emitEvent(name, data)
//
// Tests set and clear this hook in t.Cleanup to prevent cross-test pollution.
package bmad

// testEventHook, when non-nil, is called synchronously by the executor's emit
// wrapper each time an event fires. Declared here (not executor.go) to keep
// test scaffolding out of the production binary.
var testEventHook func(event string, payload interface{})
