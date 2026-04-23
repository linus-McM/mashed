//go:build !windows

package claudecli

import (
	"os"
	"syscall"
)

var termSignal os.Signal = syscall.SIGTERM
