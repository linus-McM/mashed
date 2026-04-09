package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"mashed/internal/terminal/helper"
)

func main() {
	sockPath := os.Getenv("MASHED_PTY_SOCK")
	if sockPath == "" {
		log.Fatal("MASHED_PTY_SOCK not set")
	}

	ppidStr := os.Getenv("MASHED_PARENT_PID")
	ppid, _ := strconv.Atoi(ppidStr)

	// Clean stale socket.
	os.Remove(sockPath)

	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: sockPath, Net: "unix"})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	defer os.Remove(sockPath)

	fmt.Println("READY")

	srv := helper.NewServer(ppid)
	if ppid > 0 {
		go srv.MonitorParent(ppid)
	}

	// Handle signals for clean shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		srv.Shutdown()
		os.Exit(0)
	}()

	srv.Serve(ln)
}
