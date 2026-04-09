package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Client communicates with the PTY helper process over a Unix domain socket.
type Client struct {
	conn   *net.UnixConn
	mu     sync.Mutex // serializes Spawn calls to prevent SCM_RIGHTS ordering issues
	closed atomic.Bool
}

// Dial connects to the helper's Unix domain socket and returns a Client.
func Dial(sockPath string) (*Client, error) {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: sockPath, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("client: dial %s: %w", sockPath, err)
	}
	return &Client{conn: conn}, nil
}

// Spawn requests the helper to create a new PTY session and returns the PTY
// master file descriptor, the child PID, and any error.
// Calls are serialized via mutex to prevent SCM_RIGHTS fd ordering issues.
func (c *Client) Spawn(ctx context.Context, req SpawnRequest) (*os.File, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return nil, 0, ErrConnectionClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	// Watch for context cancellation to unblock I/O by setting a past deadline.
	done := make(chan struct{})
	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		select {
		case <-ctx.Done():
			c.conn.SetDeadline(time.Now())
		case <-done:
		}
	}()
	defer func() {
		close(done)
		<-waitDone
		c.conn.SetDeadline(time.Time{})
	}()

	if err := WriteMessage(c.conn, MsgSpawn, req); err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, fmt.Errorf("client: write spawn: %w", err)
	}

	env, err := ReadMessage(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, fmt.Errorf("client: read spawn response: %w", err)
	}

	var resp SpawnResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		return nil, 0, fmt.Errorf("client: unmarshal spawn response: %w", err)
	}

	if resp.Error != "" {
		return nil, 0, fmt.Errorf("%w: %s", ErrSpawnFailed, resp.Error)
	}

	fd, err := RecvFd(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		return nil, 0, fmt.Errorf("client: recv fd: %w", err)
	}

	return os.NewFile(uintptr(fd), "ptmx"), resp.PID, nil
}

// Kill sends a signal to a helper-managed session. Fire-and-forget — no response expected.
func (c *Client) Kill(id string, sig syscall.Signal) error {
	if c.closed.Load() {
		return ErrConnectionClosed
	}
	return WriteMessage(c.conn, MsgKill, KillRequest{ID: id, Signal: int(sig)})
}

// Close marks the client as closed and shuts down the connection.
func (c *Client) Close() error {
	c.closed.Store(true)
	return c.conn.Close()
}
