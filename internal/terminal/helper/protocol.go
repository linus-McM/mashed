package helper

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

// MessageType identifies the kind of IPC message.
type MessageType string

const (
	MsgSpawn MessageType = "spawn"
	MsgKill  MessageType = "kill"
)

// Sentinel errors for the helper protocol.
var (
	ErrHelperNotRunning  = errors.New("helper: not running")
	ErrSpawnFailed       = errors.New("helper: spawn failed")
	ErrConnectionClosed  = errors.New("helper: connection closed")
)

// Envelope is the top-level wire format for IPC messages.
type Envelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// SpawnRequest asks the helper to create a new PTY session.
type SpawnRequest struct {
	ID    string   `json:"id"`
	Shell string   `json:"shell"`
	Args  []string `json:"args"`
	Env   []string `json:"env"`
	Cwd   string   `json:"cwd"`
	Cols  uint16   `json:"cols"`
	Rows  uint16   `json:"rows"`
}

// SpawnResponse reports the result of a spawn request.
type SpawnResponse struct {
	ID    string `json:"id"`
	PID   int    `json:"pid"`
	Error string `json:"error,omitempty"`
}

// KillRequest asks the helper to signal a session.
type KillRequest struct {
	ID     string `json:"id"`
	Signal int    `json:"signal"`
}

// WriteMessage encodes data as a length-prefixed JSON Envelope and writes it to conn.
// The wire format is: [4-byte big-endian uint32 length][JSON payload].
func WriteMessage(conn io.Writer, msgType MessageType, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("protocol: marshal data: %w", err)
	}

	env := Envelope{
		Type: msgType,
		Data: json.RawMessage(raw),
	}

	payload, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("protocol: marshal envelope: %w", err)
	}

	// Write 4-byte big-endian length prefix.
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	if _, err := conn.Write(lenBuf[:]); err != nil {
		return fmt.Errorf("protocol: write length: %w", err)
	}

	if _, err := conn.Write(payload); err != nil {
		return fmt.Errorf("protocol: write payload: %w", err)
	}

	return nil
}

// ReadMessage reads a length-prefixed JSON Envelope from r.
func ReadMessage(r io.Reader) (Envelope, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return Envelope{}, fmt.Errorf("protocol: read length: %w", err)
	}

	length := binary.BigEndian.Uint32(lenBuf[:])
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Envelope{}, fmt.Errorf("protocol: read payload: %w", err)
	}

	var env Envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return Envelope{}, fmt.Errorf("protocol: unmarshal envelope: %w", err)
	}

	return env, nil
}

// SendFd sends a file descriptor over a Unix domain socket using SCM_RIGHTS.
// macOS requires at least 1 byte of regular data alongside OOB data.
func SendFd(conn *net.UnixConn, fd int) error {
	rights := syscall.UnixRights(fd)
	// Send 1 byte of regular data (required on macOS).
	_, _, err := conn.WriteMsgUnix([]byte{0}, rights, nil)
	if err != nil {
		return fmt.Errorf("protocol: send fd: %w", err)
	}
	return nil
}

// RecvFd receives a file descriptor from a Unix domain socket using SCM_RIGHTS.
func RecvFd(conn *net.UnixConn) (int, error) {
	buf := make([]byte, 1)
	oob := make([]byte, syscall.CmsgSpace(4))

	_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return -1, fmt.Errorf("protocol: recv fd: %w", err)
	}

	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, fmt.Errorf("protocol: parse control message: %w", err)
	}

	for _, msg := range msgs {
		fds, err := syscall.ParseUnixRights(&msg)
		if err != nil {
			continue
		}
		if len(fds) > 0 {
			return fds[0], nil
		}
	}

	return -1, fmt.Errorf("protocol: recv fd: no file descriptor in message")
}
