package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startMockServer creates a Unix listener and runs handler on the first accepted connection.
// Uses /tmp for short socket paths (macOS sun_path limit is 104 bytes).
func startMockServer(t *testing.T, handler func(conn *net.UnixConn)) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "hs")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	sockPath := filepath.Join(dir, "s.sock")
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: sockPath, Net: "unix"})
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.AcceptUnix()
		if err != nil {
			return
		}
		defer conn.Close()
		handler(conn)
	}()

	return sockPath
}

func TestClientSpawn_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "ptmx-*")
	require.NoError(t, err)
	testData := "pty-data"
	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Sync())
	tmpFd := int(tmpFile.Fd()) // capture before handler goroutine to avoid race with Close
	defer tmpFile.Close()

	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		env, err := ReadMessage(conn)
		if err != nil {
			return
		}
		var req SpawnRequest
		json.Unmarshal(env.Data, &req)

		resp := SpawnResponse{ID: req.ID, PID: 42}
		WriteMessage(conn, MsgSpawn, resp)
		SendFd(conn, tmpFd)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "sess-1",
		Shell: "/bin/zsh",
		Cols:  80,
		Rows:  24,
	})
	require.NoError(t, err)
	assert.Equal(t, 42, pid)
	require.NotNil(t, ptmx)
	defer ptmx.Close()

	_, err = ptmx.Seek(0, io.SeekStart)
	require.NoError(t, err)
	data, err := io.ReadAll(ptmx)
	require.NoError(t, err)
	assert.Equal(t, testData, string(data))
}

func TestClientSpawn_Error(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		ReadMessage(conn)
		resp := SpawnResponse{ID: "sess-1", Error: "exec: not found"}
		WriteMessage(conn, MsgSpawn, resp)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	ptmx, pid, err := client.Spawn(context.Background(), SpawnRequest{
		ID:    "sess-1",
		Shell: "/nonexistent",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSpawnFailed)
	assert.Contains(t, err.Error(), "exec: not found")
	assert.Nil(t, ptmx)
	assert.Equal(t, 0, pid)
}

func TestClientSpawn_Serialization(t *testing.T) {
	const spawnCount = 3
	const delay = 50 * time.Millisecond

	tmpFile, err := os.CreateTemp(t.TempDir(), "ptmx-*")
	require.NoError(t, err)
	tmpFd := int(tmpFile.Fd()) // capture before handler goroutine to avoid race with Close
	defer tmpFile.Close()

	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		for i := 0; i < spawnCount; i++ {
			env, err := ReadMessage(conn)
			if err != nil {
				return
			}
			var req SpawnRequest
			json.Unmarshal(env.Data, &req)

			time.Sleep(delay)

			resp := SpawnResponse{ID: req.ID, PID: 100 + i}
			WriteMessage(conn, MsgSpawn, resp)
			SendFd(conn, tmpFd)
		}
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	var wg sync.WaitGroup
	errs := make([]error, spawnCount)

	start := time.Now()
	for i := 0; i < spawnCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ptmx, _, err := client.Spawn(context.Background(), SpawnRequest{
				ID:    fmt.Sprintf("sess-%d", i),
				Shell: "/bin/sh",
			})
			errs[i] = err
			if ptmx != nil {
				ptmx.Close()
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)

	for i, err := range errs {
		assert.NoError(t, err, "spawn %d should succeed", i)
	}
	// With mutex serialization: >= 150ms (3 * 50ms).
	// Without serialization: ~50ms.
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(100),
		"spawn calls should be serialized (expected >= 150ms, got %v)", elapsed)
}

func TestClientSpawn_ContextCancelled(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		// Read request but never respond — client will time out.
		ReadMessage(conn)
		time.Sleep(5 * time.Second)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _, err = client.Spawn(ctx, SpawnRequest{ID: "sess-1", Shell: "/bin/sh"})
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestClientSpawn_AlreadyCancelledContext(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		time.Sleep(time.Second)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = client.Spawn(ctx, SpawnRequest{ID: "sess-1", Shell: "/bin/sh"})
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestClientKill(t *testing.T) {
	received := make(chan KillRequest, 1)

	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		env, err := ReadMessage(conn)
		if err != nil {
			return
		}
		var req KillRequest
		json.Unmarshal(env.Data, &req)
		received <- req
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	err = client.Kill("sess-1", syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case req := <-received:
		assert.Equal(t, "sess-1", req.ID)
		assert.Equal(t, int(syscall.SIGTERM), req.Signal)
	case <-time.After(time.Second):
		t.Fatal("mock server did not receive kill request")
	}
}

func TestClientClose(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		time.Sleep(time.Second)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)
	assert.True(t, client.closed.Load())

	// Subsequent Spawn should return ErrConnectionClosed.
	_, _, err = client.Spawn(context.Background(), SpawnRequest{ID: "sess-1"})
	assert.ErrorIs(t, err, ErrConnectionClosed)

	// Subsequent Kill should return ErrConnectionClosed.
	err = client.Kill("sess-1", syscall.SIGTERM)
	assert.ErrorIs(t, err, ErrConnectionClosed)
}

func TestClientSpawn_WriteError(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		// Close immediately — client's WriteMessage will fail.
		conn.Close()
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	// Give mock server time to close its end.
	time.Sleep(20 * time.Millisecond)

	_, _, err = client.Spawn(context.Background(), SpawnRequest{ID: "sess-1", Shell: "/bin/sh"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client: write spawn")
}

func TestClientSpawn_BadResponse(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		ReadMessage(conn)
		// Send raw garbage as a length-prefixed message — valid framing, invalid JSON payload.
		payload := []byte("not-json{{{")
		var lenBuf [4]byte
		lenBuf[3] = byte(len(payload))
		conn.Write(lenBuf[:])
		conn.Write(payload)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	defer client.Close()

	_, _, err = client.Spawn(context.Background(), SpawnRequest{ID: "sess-1", Shell: "/bin/sh"})
	require.Error(t, err)
	// Could be unmarshal envelope error (from ReadMessage) or unmarshal spawn response error.
	assert.Error(t, err)
}

func TestDial_Failure(t *testing.T) {
	_, err := Dial(filepath.Join(t.TempDir(), "nonexistent.sock"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client: dial")
}

func TestDial_Success(t *testing.T) {
	sockPath := startMockServer(t, func(conn *net.UnixConn) {
		time.Sleep(time.Second)
	})

	client, err := Dial(sockPath)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.conn)
	assert.False(t, client.closed.Load())
	client.Close()
}
