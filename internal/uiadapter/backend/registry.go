package backend

import (
	"fmt"
	"sort"
	"sync"

	"mashed/internal/uiadapter"
)

// Constructor builds a concrete LLMBackend from the adapter Config.
// Registration happens from each subpackage's init().
type Constructor func(cfg uiadapter.Config) (LLMBackend, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]Constructor{}
)

// Register installs a constructor under name. Panics on duplicate names —
// that only happens via an init() collision, which is a programmer error
// (Plan §6.5 "Never panic in library code. Panic is for programmer bugs
// only.").
func Register(name string, ctor Constructor) {
	if name == "" {
		panic("backend.Register: empty name")
	}
	if ctor == nil {
		panic("backend.Register: nil constructor for " + name)
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		panic("backend.Register: duplicate name " + name)
	}
	registry[name] = ctor
}

// From resolves a backend by name. Unknown names produce an error wrapping
// ErrUnknownBackend with the sorted list of available options — AC-C.2.
func From(name string, cfg uiadapter.Config) (LLMBackend, error) {
	registryMu.RLock()
	ctor, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w %q; available: %v", ErrUnknownBackend, name, Available())
	}
	return ctor(cfg)
}

// Available returns the sorted list of registered backend names. Used for
// error messages and for the title-bar selector (Story v3-18).
func Available() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// reset — test-only; callable from _test.go files in this package to purge
// the registry between tests so stub init doesn't collide across runs.
// Lowercase keeps it package-private.
func reset() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = map[string]Constructor{}
}
