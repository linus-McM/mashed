package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteReadMessage_SpawnRequest(t *testing.T) {
	tests := []struct {
		name string
		req  SpawnRequest
	}{
		{
			name: "basic spawn request",
			req: SpawnRequest{
				ID:    "sess-1",
				Shell: "/bin/zsh",
				Args:  []string{"-l"},
				Env:   []string{"TERM=xterm-256color", "HOME=/Users/test"},
				Cwd:   "/tmp",
				Cols:  80,
				Rows:  24,
			},
		},
		{
			name: "empty fields",
			req: SpawnRequest{
				ID:    "sess-2",
				Shell: "/bin/sh",
			},
		},
		{
			name: "unicode in env",
			req: SpawnRequest{
				ID:    "sess-3",
				Shell: "/bin/bash",
				Env:   []string{"LANG=ja_JP.UTF-8", "GREETING=こんにちは"},
				Cols:  120,
				Rows:  40,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := WriteMessage(&buf, MsgSpawn, tt.req)
			require.NoError(t, err)

			env, err := ReadMessage(&buf)
			require.NoError(t, err)
			assert.Equal(t, MsgSpawn, env.Type)

			var got SpawnRequest
			err = json.Unmarshal(env.Data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.req, got)
		})
	}
}

func TestWriteReadMessage_KillRequest(t *testing.T) {
	tests := []struct {
		name string
		req  KillRequest
	}{
		{
			name: "SIGTERM",
			req:  KillRequest{ID: "sess-1", Signal: 15},
		},
		{
			name: "SIGKILL",
			req:  KillRequest{ID: "sess-2", Signal: 9},
		},
		{
			name: "zero signal (check existence)",
			req:  KillRequest{ID: "sess-3", Signal: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := WriteMessage(&buf, MsgKill, tt.req)
			require.NoError(t, err)

			env, err := ReadMessage(&buf)
			require.NoError(t, err)
			assert.Equal(t, MsgKill, env.Type)

			var got KillRequest
			err = json.Unmarshal(env.Data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.req, got)
		})
	}
}

func TestWriteReadMessage_SpawnResponse(t *testing.T) {
	tests := []struct {
		name string
		resp SpawnResponse
	}{
		{
			name: "success",
			resp: SpawnResponse{ID: "sess-1", PID: 12345},
		},
		{
			name: "error response",
			resp: SpawnResponse{ID: "sess-2", Error: "exec: not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := WriteMessage(&buf, MsgSpawn, tt.resp)
			require.NoError(t, err)

			env, err := ReadMessage(&buf)
			require.NoError(t, err)

			var got SpawnResponse
			err = json.Unmarshal(env.Data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.resp, got)
		})
	}
}

func TestWriteReadMessage_LargePayload(t *testing.T) {
	// 200 env vars of 1KB each — round-trip without truncation.
	envVars := make([]string, 200)
	for i := range envVars {
		envVars[i] = "VAR_" + strings.Repeat("x", 1024)
	}

	req := SpawnRequest{
		ID:    "large-env",
		Shell: "/bin/sh",
		Env:   envVars,
		Cols:  80,
		Rows:  24,
	}

	var buf bytes.Buffer

	err := WriteMessage(&buf, MsgSpawn, req)
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 200*1024, "payload should be > 200KB")

	env, err := ReadMessage(&buf)
	require.NoError(t, err)
	assert.Equal(t, MsgSpawn, env.Type)

	var got SpawnRequest
	err = json.Unmarshal(env.Data, &got)
	require.NoError(t, err)
	assert.Equal(t, req, got)
}

func TestReadMessage_TruncatedLength(t *testing.T) {
	// Only 2 bytes instead of 4.
	r := bytes.NewReader([]byte{0x00, 0x01})
	_, err := ReadMessage(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read length")
}

func TestReadMessage_TruncatedPayload(t *testing.T) {
	// Length says 100 bytes but only 5 are present.
	var buf bytes.Buffer
	buf.Write([]byte{0x00, 0x00, 0x00, 0x64}) // length = 100
	buf.Write([]byte("hello"))                  // only 5 bytes

	_, err := ReadMessage(&buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read payload")
}

func TestReadMessage_InvalidJSON(t *testing.T) {
	payload := []byte("not valid json{{{")
	var buf bytes.Buffer
	lenBuf := make([]byte, 4)
	lenBuf[0] = 0
	lenBuf[1] = 0
	lenBuf[2] = 0
	lenBuf[3] = byte(len(payload))
	buf.Write(lenBuf)
	buf.Write(payload)

	_, err := ReadMessage(&buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal envelope")
}

func TestReadMessage_EmptyReader(t *testing.T) {
	r := bytes.NewReader([]byte{})
	_, err := ReadMessage(r)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

func TestWriteMessage_MultipleMessages(t *testing.T) {
	var buf bytes.Buffer

	// Write two messages back to back.
	err := WriteMessage(&buf, MsgSpawn, SpawnRequest{ID: "a", Shell: "/bin/sh"})
	require.NoError(t, err)
	err = WriteMessage(&buf, MsgKill, KillRequest{ID: "a", Signal: 9})
	require.NoError(t, err)

	// Read them back in order.
	env1, err := ReadMessage(&buf)
	require.NoError(t, err)
	assert.Equal(t, MsgSpawn, env1.Type)

	env2, err := ReadMessage(&buf)
	require.NoError(t, err)
	assert.Equal(t, MsgKill, env2.Type)
}

func TestSendRecvFd(t *testing.T) {
	// Create a Unix socketpair.
	fds, err := createSocketPair()
	require.NoError(t, err)

	senderFile := os.NewFile(uintptr(fds[0]), "sender")
	receiverFile := os.NewFile(uintptr(fds[1]), "receiver")
	defer senderFile.Close()
	defer receiverFile.Close()

	senderConn, err := net.FileConn(senderFile)
	require.NoError(t, err)
	defer senderConn.Close()

	receiverConn, err := net.FileConn(receiverFile)
	require.NoError(t, err)
	defer receiverConn.Close()

	sender := senderConn.(*net.UnixConn)
	receiver := receiverConn.(*net.UnixConn)

	// Create a temp file with known content.
	tmpFile, err := os.CreateTemp(t.TempDir(), "fdtest-*")
	require.NoError(t, err)
	testData := "hello from fd passing"
	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Sync())

	// Send the fd.
	err = SendFd(sender, int(tmpFile.Fd()))
	require.NoError(t, err)
	tmpFile.Close()

	// Receive the fd.
	recvFd, err := RecvFd(receiver)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, recvFd, 0)

	// Read through the received fd and verify content.
	recvFile := os.NewFile(uintptr(recvFd), "received")
	defer recvFile.Close()

	_, err = recvFile.Seek(0, io.SeekStart)
	require.NoError(t, err)

	data, err := io.ReadAll(recvFile)
	require.NoError(t, err)
	assert.Equal(t, testData, string(data))
}

func TestRecvFd_ClosedConn(t *testing.T) {
	fds, err := createSocketPair()
	require.NoError(t, err)

	senderFile := os.NewFile(uintptr(fds[0]), "sender")
	receiverFile := os.NewFile(uintptr(fds[1]), "receiver")

	receiverConn, err := net.FileConn(receiverFile)
	require.NoError(t, err)
	receiver := receiverConn.(*net.UnixConn)

	// Close both ends of the sender so RecvFd gets an error.
	senderFile.Close()
	receiverFile.Close()

	// Close the sender side — receiver should get an error.
	_, err = RecvFd(receiver)
	require.Error(t, err)
	receiver.Close()
}

func TestSentinelErrors(t *testing.T) {
	assert.EqualError(t, ErrHelperNotRunning, "helper: not running")
	assert.EqualError(t, ErrSpawnFailed, "helper: spawn failed")
	assert.EqualError(t, ErrConnectionClosed, "helper: connection closed")
}

func TestMessageTypeConstants(t *testing.T) {
	assert.Equal(t, MessageType("spawn"), MsgSpawn)
	assert.Equal(t, MessageType("kill"), MsgKill)
}

func TestEnvelopeJSONTags(t *testing.T) {
	env := Envelope{
		Type: MsgSpawn,
		Data: json.RawMessage(`{"id":"test"}`),
	}
	b, err := json.Marshal(env)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"type":"spawn"`)
	assert.Contains(t, string(b), `"data":`)
}

func TestSpawnResponseOmitEmpty(t *testing.T) {
	resp := SpawnResponse{ID: "sess-1", PID: 100}
	b, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.NotContains(t, string(b), `"error"`)

	resp2 := SpawnResponse{ID: "sess-2", Error: "failed"}
	b2, err := json.Marshal(resp2)
	require.NoError(t, err)
	assert.Contains(t, string(b2), `"error":"failed"`)
}

// createSocketPair creates a Unix SOCK_STREAM socketpair and returns both fds.
func createSocketPair() ([2]int, error) {
	var fds [2]int
	rawFds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fds, err
	}
	fds[0] = rawFds[0]
	fds[1] = rawFds[1]
	return fds, nil
}
