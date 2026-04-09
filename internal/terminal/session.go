package terminal

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// ErrSessionDead is returned when an operation is attempted on a dead session.
var ErrSessionDead = errors.New("terminal: session is not alive")

const (
	// scrollBufferDefaultCap is the default scrollback buffer capacity (1MB).
	scrollBufferDefaultCap = 1 << 20
	// ptyReadSize is the read buffer size for PTY output.
	ptyReadSize = 4096
)

// ---------------------------------------------------------------------------
// scrollBuffer — circular byte buffer
// ---------------------------------------------------------------------------

// wsWriteTimeout is the deadline for writing a single WebSocket message.
const wsWriteTimeout = 500 * time.Millisecond

// scrollBuffer is a thread-safe circular byte buffer that stores scrollback data.
type scrollBuffer struct {
	mu       sync.Mutex
	buf      []byte
	capacity int
	pos      int  // next write position
	full     bool // true once the buffer has wrapped at least once
}

func newScrollBuffer(capacity int) *scrollBuffer {
	return &scrollBuffer{
		buf:      make([]byte, capacity),
		capacity: capacity,
	}
}

// Write appends data to the ring buffer, discarding oldest bytes when over capacity.
func (sb *scrollBuffer) Write(p []byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(p) >= sb.capacity {
		copy(sb.buf, p[len(p)-sb.capacity:])
		sb.pos = 0
		sb.full = true
		return
	}

	n := copy(sb.buf[sb.pos:], p)
	if n < len(p) {
		copy(sb.buf, p[n:])
		sb.pos = len(p) - n
		sb.full = true
	} else {
		sb.pos += n
		if sb.pos == sb.capacity {
			sb.pos = 0
			sb.full = true
		}
	}
}

// Snapshot returns an independent copy of the buffered data.
func (sb *scrollBuffer) Snapshot() []byte {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if !sb.full && sb.pos == 0 {
		return []byte{}
	}

	if !sb.full {
		out := make([]byte, sb.pos)
		copy(out, sb.buf[:sb.pos])
		return out
	}

	out := make([]byte, sb.capacity)
	n := copy(out, sb.buf[sb.pos:])
	copy(out[n:], sb.buf[:sb.pos])
	return out
}

// ---------------------------------------------------------------------------
// ManagedSession
// ---------------------------------------------------------------------------

// ManagedSession owns a PTY, an exec.Cmd, a scrollback buffer,
// and a set of connected WebSocket clients.
type ManagedSession struct {
	name     string
	cmd      *exec.Cmd
	ptmx     *os.File
	scroll   *scrollBuffer
	clients  map[*websocket.Conn]*sync.Mutex
	mu       sync.Mutex
	done     chan struct{}
	pid      int
	killOnce sync.Once
}

// newManagedSession creates a session for an already-started command.
// Precondition: cmd must have been started with SysProcAttr.Setpgid = true
// so that Kill() can send SIGTERM to the entire process group.
func newManagedSession(name string, cmd *exec.Cmd, ptmx *os.File) *ManagedSession {
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setpgid {
		log.Printf("terminal: newManagedSession(%s): cmd lacks Setpgid — process group kill may fail", name)
	}
	ms := &ManagedSession{
		name:    name,
		cmd:     cmd,
		ptmx:    ptmx,
		scroll:  newScrollBuffer(scrollBufferDefaultCap),
		clients: make(map[*websocket.Conn]*sync.Mutex),
		done:    make(chan struct{}),
		pid:     cmd.Process.Pid,
	}
	go ms.readLoop()
	return ms
}

// newRemoteManagedSession creates a session for a PTY fd received from the
// helper process. There is no local exec.Cmd — the child is managed by the helper.
func newRemoteManagedSession(name string, pid int, ptmx *os.File) *ManagedSession {
	ms := &ManagedSession{
		name:    name,
		ptmx:    ptmx,
		scroll:  newScrollBuffer(scrollBufferDefaultCap),
		clients: make(map[*websocket.Conn]*sync.Mutex),
		done:    make(chan struct{}),
		pid:     pid,
	}
	go ms.readLoop()
	return ms
}

// readLoop reads PTY output in chunks, writes to scrollBuffer, and fans out to clients.
func (ms *ManagedSession) readLoop() {
	defer ms.cleanupClients()

	buf := make([]byte, ptyReadSize)
	for {
		n, err := ms.ptmx.Read(buf)
		if n > 0 {
			data := buf[:n]

			// Hold ms.mu only for scroll write + client map snapshot.
			ms.mu.Lock()
			ms.scroll.Write(data)
			clients := make(map[*websocket.Conn]*sync.Mutex, len(ms.clients))
			for ws, wsMu := range ms.clients {
				clients[ws] = wsMu
			}
			ms.mu.Unlock()

			// Fan out without holding session lock.
			var dead []*websocket.Conn
			for ws, wsMu := range clients {
				wsMu.Lock()
				_ = ws.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
				writeErr := ws.WriteMessage(websocket.BinaryMessage, data)
				wsMu.Unlock()
				if writeErr != nil {
					dead = append(dead, ws)
				}
			}
			if len(dead) > 0 {
				ms.mu.Lock()
				for _, ws := range dead {
					delete(ms.clients, ws)
					ws.Close()
				}
				ms.mu.Unlock()
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("terminal: readLoop(%s): %v", ms.name, err)
			}
			ms.killOnce.Do(func() {
				ms.ptmx.Close()
				close(ms.done)
				if ms.cmd != nil {
					go ms.cmd.Wait()
				}
			})
			return
		}
	}
}

// cleanupClients closes all remaining client connections on session teardown.
func (ms *ManagedSession) cleanupClients() {
	ms.mu.Lock()
	for ws := range ms.clients {
		ws.Close()
	}
	ms.clients = nil
	ms.mu.Unlock()
}

// AddClient sends the scrollback snapshot then registers ws for live fan-out.
// The snapshot send and registration are atomic under session mu.
func (ms *ManagedSession) AddClient(ws *websocket.Conn) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.clients == nil {
		ws.Close()
		return
	}

	snap := ms.scroll.Snapshot()
	if len(snap) > 0 {
		_ = ws.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
		if err := ws.WriteMessage(websocket.BinaryMessage, snap); err != nil {
			ws.Close()
			return
		}
	}

	ms.clients[ws] = &sync.Mutex{}
	go ms.clientReader(ws)
}

// RemoveClient disconnects and deregisters a WebSocket client.
func (ms *ManagedSession) RemoveClient(ws *websocket.Conn) {
	ms.mu.Lock()
	delete(ms.clients, ws)
	ms.mu.Unlock()
	ws.Close()
}

// wsReadLimit is the maximum size of a single WebSocket message from a client.
const wsReadLimit = 64 * 1024

// clientReader reads messages from the client WebSocket and writes to the PTY.
func (ms *ManagedSession) clientReader(ws *websocket.Conn) {
	ws.SetReadLimit(wsReadLimit)
	for {
		if !ms.IsAlive() {
			return
		}
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			ms.RemoveClient(ws)
			return
		}
		switch msgType {
		case websocket.BinaryMessage:
			_, _ = ms.ptmx.Write(data)
		case websocket.TextMessage:
			var rm resizeMsg
			if err := json.Unmarshal(data, &rm); err == nil && rm.Type == "resize" {
				_ = pty.Setsize(ms.ptmx, &pty.Winsize{
					Cols: rm.Cols,
					Rows: rm.Rows,
				})
			}
		}
	}
}

// Kill terminates the session process, closes the PTY, and signals done.
// Closing ptmx sends SIGHUP to the child and its descendants via the TTY.
func (ms *ManagedSession) Kill() {
	ms.killOnce.Do(func() {
		if ms.pid > 0 {
			if err := syscall.Kill(ms.pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
				log.Printf("terminal: kill(%s, pid=%d): %v", ms.name, ms.pid, err)
			}
		}
		ms.ptmx.Close()
		close(ms.done)
		if ms.cmd != nil {
			go ms.cmd.Wait()
		}
	})
}

// IsAlive returns true if the session has not been terminated.
func (ms *ManagedSession) IsAlive() bool {
	select {
	case <-ms.done:
		return false
	default:
		return true
	}
}

// Wait blocks until the session is done.
func (ms *ManagedSession) Wait() {
	<-ms.done
}

// Name returns the session name.
func (ms *ManagedSession) Name() string {
	return ms.name
}
