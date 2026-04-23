//go:build windows

package claudecli

import "os"

// Windows has no SIGTERM; use Kill equivalent. CommandContext's default
// cancel path will kill the process.
var termSignal os.Signal = os.Kill
