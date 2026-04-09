package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

// osExit is a variable so tests can replace it.
var osExit = os.Exit

// helperSession tracks a single PTY session managed by the helper.
type helperSession struct {
	cmd  *exec.Cmd
	ptmx *os.File
}

// Server manages PTY sessions spawned by the helper binary.
type Server struct {
	sessions  sync.Map // map[string]*helperSession
	parentPID int
	spawnWg   sync.WaitGroup // tracks in-flight spawn operations only
}

// NewServer creates a new helper Server that monitors the given parent PID.
func NewServer(parentPID int) *Server {
	return &Server{
		parentPID: parentPID,
	}
}

// Serve accepts connections on ln and dispatches messages.
// Each connection is handled in its own goroutine.
func (s *Server) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handleConn(conn.(*net.UnixConn))
	}
}

// handleConn reads messages from a single connection in a loop.
func (s *Server) handleConn(conn *net.UnixConn) {
	defer conn.Close()

	for {
		env, err := ReadMessage(conn)
		if err != nil {
			return // connection closed or read error
		}

		switch env.Type {
		case MsgSpawn:
			var req SpawnRequest
			if err := json.Unmarshal(env.Data, &req); err != nil {
				log.Printf("helper: unmarshal spawn request: %v", err)
				continue
			}
			s.handleSpawn(conn, req)

		case MsgKill:
			var req KillRequest
			if err := json.Unmarshal(env.Data, &req); err != nil {
				log.Printf("helper: unmarshal kill request: %v", err)
				continue
			}
			s.handleKill(conn, req)

		default:
			log.Printf("helper: unknown message type: %s", env.Type)
		}
	}
}

// handleSpawn creates a new PTY session and sends back the response + fd.
func (s *Server) handleSpawn(conn *net.UnixConn, req SpawnRequest) {
	s.spawnWg.Add(1)
	defer s.spawnWg.Done()
	shell := req.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/zsh"
		}
	}

	args := req.Args
	cmd := exec.Command(shell, args...)
	cmd.Dir = req.Cwd
	if len(req.Env) > 0 {
		cmd.Env = req.Env
	}
	// NOTE: Setpgid is intentionally omitted. macOS Sequoia blocks setpgid()
	// combined with PTY fork/exec (EPERM). The shell's own process group is
	// sufficient — Kill sends SIGHUP to the PTY which propagates to the child.

	cols := req.Cols
	if cols == 0 {
		cols = 80
	}
	rows := req.Rows
	if rows == 0 {
		rows = 24
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		// Send error response — no fd follows.
		resp := SpawnResponse{
			ID:    req.ID,
			PID:   0,
			Error: fmt.Sprintf("spawn: %v", err),
		}
		if wErr := WriteMessage(conn, MsgSpawn, resp); wErr != nil {
			log.Printf("helper: write spawn error response: %v", wErr)
		}
		return
	}

	// Store session.
	s.sessions.Store(req.ID, &helperSession{
		cmd:  cmd,
		ptmx: ptmx,
	})

	// Send SpawnResponse FIRST (client expects response before fd).
	resp := SpawnResponse{
		ID:  req.ID,
		PID: cmd.Process.Pid,
	}
	if err := WriteMessage(conn, MsgSpawn, resp); err != nil {
		log.Printf("helper: write spawn response: %v", err)
		ptmx.Close()
		s.sessions.Delete(req.ID)
		return
	}

	// Send fd via SCM_RIGHTS (do NOT close ptmx — helper keeps it open).
	if err := SendFd(conn, int(ptmx.Fd())); err != nil {
		log.Printf("helper: send fd for %s: %v", req.ID, err)
	}

	// Wait for the child process in the background to avoid zombies.
	go func() {
		cmd.Wait()
	}()
}

// handleKill signals a session's process group and cleans up.
func (s *Server) handleKill(conn *net.UnixConn, req KillRequest) {
	val, ok := s.sessions.Load(req.ID)
	if !ok {
		resp := struct {
			ID    string `json:"id"`
			Error string `json:"error,omitempty"`
		}{
			ID:    req.ID,
			Error: fmt.Sprintf("session %q not found", req.ID),
		}
		WriteMessage(conn, MsgKill, resp)
		return
	}

	session := val.(*helperSession)

	// Signal the process directly (not process group — Setpgid is not used
	// on macOS Sequoia due to EPERM). Closing ptmx sends SIGHUP to the child.
	if session.cmd.Process != nil {
		sig := syscall.Signal(req.Signal)
		if sig == 0 {
			sig = syscall.SIGTERM
		}
		session.cmd.Process.Signal(sig)
	}

	// Close the pty master — sends SIGHUP to child and all its descendants.
	session.ptmx.Close()

	// Remove from map.
	s.sessions.Delete(req.ID)

	resp := struct {
		ID string `json:"id"`
	}{
		ID: req.ID,
	}
	WriteMessage(conn, MsgKill, resp)
}

// Shutdown waits for in-flight spawn operations to finish, then kills all sessions.
func (s *Server) Shutdown() {
	s.spawnWg.Wait() // let in-flight handleSpawn finish before closing ptmx
	s.sessions.Range(func(key, val any) bool {
		session := val.(*helperSession)
		if session.cmd.Process != nil {
			session.cmd.Process.Signal(syscall.SIGTERM)
		}
		session.ptmx.Close()
		s.sessions.Delete(key)
		return true
	})
}

// MonitorParent polls whether the parent process is still alive.
// If the parent is gone (ESRCH), calls osExit(0).
func (s *Server) MonitorParent(parentPID int) {
	for {
		err := syscall.Kill(parentPID, 0)
		if err == syscall.ESRCH {
			log.Printf("helper: parent PID %d gone, exiting", parentPID)
			osExit(0)
			return
		}
		time.Sleep(2 * time.Second)
	}
}
