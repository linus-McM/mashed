package terminal

// NewStubSession returns a ManagedSession with no real PTY, suitable for unit tests.
func NewStubSession(name string) *ManagedSession {
	return &ManagedSession{
		name: name,
		done: make(chan struct{}),
	}
}
