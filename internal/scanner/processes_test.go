package scanner

import (
	"testing"
	"time"
)

func TestParsePSLine_HappyPath(t *testing.T) {
	// Three claude processes with varying etime formats.
	lines := []struct {
		name    string
		line    string
		wantPID int
		wantPPID int
		wantModel string
		wantSID   string
	}{
		{
			name:      "basic claude process",
			line:      "  1234   999  05:30 /usr/local/bin/claude --model opus",
			wantPID:   1234,
			wantPPID:  999,
			wantModel: "opus",
			wantSID:   "",
		},
		{
			name:      "claude with session-id",
			line:      "  5678  1000  01:02:03 /usr/local/bin/claude --model sonnet --session-id abc-123",
			wantPID:   5678,
			wantPPID:  1000,
			wantModel: "sonnet",
			wantSID:   "abc-123",
		},
		{
			name:      "claude with day etime",
			line:      "  9012  2000  01-02:03:04 /usr/local/bin/claude --session-id xyz-789",
			wantPID:   9012,
			wantPPID:  2000,
			wantModel: "claude",
			wantSID:   "xyz-789",
		},
	}

	for _, tt := range lines {
		t.Run(tt.name, func(t *testing.T) {
			rp, err := parsePSLine(tt.line)
			if err != nil {
				t.Fatalf("parsePSLine(%q) error: %v", tt.line, err)
			}
			if rp.pid != tt.wantPID {
				t.Errorf("pid = %d, want %d", rp.pid, tt.wantPID)
			}
			if rp.ppid != tt.wantPPID {
				t.Errorf("ppid = %d, want %d", rp.ppid, tt.wantPPID)
			}
			if rp.model != tt.wantModel {
				t.Errorf("model = %q, want %q", rp.model, tt.wantModel)
			}
			if rp.sessionID != tt.wantSID {
				t.Errorf("sessionID = %q, want %q", rp.sessionID, tt.wantSID)
			}
		})
	}
}

func TestIsExcludedProcess(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		excluded bool
	}{
		{
			name:     "Electron Claude Helper excluded",
			line:     "  111  222  01:00 /Applications/Claude.app/Contents/Frameworks/Claude Helper (Renderer)",
			excluded: true,
		},
		{
			name:     "grep excluded",
			line:     "  333  444  00:01 grep claude",
			excluded: true,
		},
		{
			name:     "context-mode excluded",
			line:     "  555  666  02:00 /usr/local/bin/claude --context-mode",
			excluded: true,
		},
		{
			name:     "zsh excluded",
			line:     "  777  888  03:00 /bin/zsh -l claude",
			excluded: true,
		},
		{
			name:     "Claude.app MacOS excluded",
			line:     "  900  100  04:00 /Applications/Claude.app/Contents/MacOS/Claude",
			excluded: true,
		},
		{
			name:     "legitimate claude CLI not excluded",
			line:     "  1234  999  05:30 /usr/local/bin/claude --model opus",
			excluded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExcludedProcess(tt.line)
			if got != tt.excluded {
				t.Errorf("isExcludedProcess(%q) = %v, want %v", tt.line, got, tt.excluded)
			}
		})
	}
}

func TestParseEtime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:  "days-hours:minutes:seconds",
			input: "01-02:03:04",
			want:  1*24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second,
		},
		{
			name:  "hours:minutes:seconds",
			input: "05:06:07",
			want:  5*time.Hour + 6*time.Minute + 7*time.Second,
		},
		{
			name:  "minutes:seconds",
			input: "08:09",
			want:  8*time.Minute + 9*time.Second,
		},
		{
			name:  "seconds only",
			input: "10",
			want:  10 * time.Second,
		},
		{
			name:  "multi-day with leading zeros",
			input: "07-00:00:00",
			want:  7 * 24 * time.Hour,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEtime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseEtime(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEtime(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseEtime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractModel(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "model flag present",
			args: "/usr/local/bin/claude --model opus --verbose",
			want: "opus",
		},
		{
			name: "model flag with spaces before value",
			args: "/usr/local/bin/claude   --model   claude-opus-4-6",
			want: "claude-opus-4-6",
		},
		{
			name: "no model flag defaults to claude",
			args: "/usr/local/bin/claude --session-id abc",
			want: "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModel(tt.args)
			if got != tt.want {
				t.Errorf("extractModel(%q) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "session-id present",
			args: "/usr/local/bin/claude --session-id abc-123-def",
			want: "abc-123-def",
		},
		{
			name: "session-id with other flags",
			args: "/usr/local/bin/claude --model opus --session-id sess-999 --verbose",
			want: "sess-999",
		},
		{
			name: "no session-id returns empty",
			args: "/usr/local/bin/claude --model opus",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionID(tt.args)
			if got != tt.want {
				t.Errorf("extractSessionID(%q) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
