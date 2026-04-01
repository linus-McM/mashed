// Package scanner discovers Claude Code processes, parses session files, and watches for changes.
package scanner

import (
	"errors"
	"fmt"
)

// Sentinel errors for the scanner package.
var (
	ErrProcessNotFound = errors.New("scanner: process not found")
	ErrNoWorkingDir    = errors.New("scanner: cannot determine working directory")
	ErrSessionNotFound = errors.New("scanner: session file not found")
	ErrInvalidJSONL    = errors.New("scanner: invalid JSONL line")
	ErrWatcherClosed   = errors.New("scanner: watcher closed")
	ErrNotGitRepo      = errors.New("scanner: not a git repository")
	ErrInodeChanged    = errors.New("scanner: file inode changed (rotation detected)")
)

// ScanError represents an error during process or repo scanning.
type ScanError struct {
	PID int
	Op  string
	Err error
}

func (e *ScanError) Error() string {
	if e.PID > 0 {
		return fmt.Sprintf("scan %s (pid %d): %v", e.Op, e.PID, e.Err)
	}
	return fmt.Sprintf("scan %s: %v", e.Op, e.Err)
}

func (e *ScanError) Unwrap() error { return e.Err }

// ParseError represents an error during JSONL session parsing.
type ParseError struct {
	Path   string
	Line   int
	Offset int64
	Err    error
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse %s (line %d, offset %d): %v", e.Path, e.Line, e.Offset, e.Err)
	}
	return fmt.Sprintf("parse %s: %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// WatchError represents an error during filesystem watching.
type WatchError struct {
	Path string
	Op   string
	Err  error
}

func (e *WatchError) Error() string {
	return fmt.Sprintf("watch %s (%s): %v", e.Op, e.Path, e.Err)
}

func (e *WatchError) Unwrap() error { return e.Err }
