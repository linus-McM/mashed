// Package uiadapter — race-mode probe (race detector ON build).
//
// See `race_off_test.go` for rationale.
//
//go:build race

package uiadapter

// raceDetectorEnabled reports whether the binary was built with `-race`.
const raceDetectorEnabled = true
