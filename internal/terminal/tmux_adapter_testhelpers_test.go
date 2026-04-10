// Package terminal test-only helpers for TmuxAdapter tests.
//
// These helpers are compiled only in the test binary (file name ends in
// _test.go) and exist solely to grant tests access to unexported fields
// that must be mockable — specifically the injectable mkfifo function
// and the fallbackPolling flag on *TmuxAttachment.
//
// The go-engineer implementing tmux_adapter.go (Task #3) must define:
//   - `mkfifo func(path string, mode uint32) error` as a field on
//     *TmuxAdapter, defaulting to syscall.Mkfifo.
//   - `fallbackPolling bool` as a field on *TmuxAttachment, set to true
//     when the mkfifo field returns an error during Attach.
//
// If either field is renamed, update this file too (both helpers fail to
// compile loudly, which is the desired behavior).

package terminal

// overrideMkfifo replaces the adapter's mkfifo function with a test
// stand-in. Tests use this to force the polling-fallback code path.
func overrideMkfifo(a *TmuxAdapter, fn func(path string, mode uint32) error) {
	a.mkfifo = fn
}

// attachmentIsPolling reports whether the attachment is running in the
// capture-pane polling fallback instead of the FIFO + pipe-pane path. It
// accepts the TmuxSession interface (the post-bridge-03 return type of
// *TmuxAdapter.Attach) and type-asserts to the concrete *TmuxAttachment so
// callsites do not need to assert at every use.
func attachmentIsPolling(a TmuxSession) bool {
	att, ok := a.(*TmuxAttachment)
	if !ok {
		return false
	}
	return att.fallbackPolling
}
