// Package uiadapter — race-mode probe (race detector OFF build).
//
// The alloc-budget tests (Story 3 AC-8 and Story 4 AC-4.8) call
// testing.AllocsPerRun which is incompatible with the race detector: the
// detector inserts read/write tracking that surfaces as additional Go
// allocations on hot paths exercising regexp / atomic operations. Story 3's
// hot paths (cache map lookups under sync.RWMutex) survive the noise; Story
// 4's fastpath rule loop with five regexp.FindStringSubmatch calls does
// not. Rather than relax the tolerance, the alloc test is skipped under
// race mode — the value of the AC is to catch instrumentation-induced
// allocations and that signal is preserved in the non-race CI lane.
//
//go:build !race

package uiadapter

// raceDetectorEnabled reports whether the binary was built with `-race`.
// A separate `_test.go` file behind the `race` build tag returns true when
// the detector is active.
const raceDetectorEnabled = false
