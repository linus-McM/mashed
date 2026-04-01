package git

import (
	"testing"
)

func TestParseDiffStatLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantOK      bool
		wantPath    string
		wantAdded   int
		wantRemoved int
		wantBinary  bool
	}{
		{
			name:        "simple additions",
			line:        " internal/scanner/processes.go | 10 ++++++++++",
			wantOK:      true,
			wantPath:    "internal/scanner/processes.go",
			wantAdded:   10,
			wantRemoved: 0,
		},
		{
			name:        "mixed additions and removals",
			line:        " internal/agent/engine.go | 8 +++-----",
			wantOK:      true,
			wantPath:    "internal/agent/engine.go",
			wantAdded:   3,
			wantRemoved: 5,
		},
		{
			name:        "only removals",
			line:        " old_file.go | 5 -----",
			wantOK:      true,
			wantPath:    "old_file.go",
			wantAdded:   0,
			wantRemoved: 5,
		},
		{
			name:       "binary file",
			line:       " assets/logo.png | Bin 0 -> 12345 bytes",
			wantOK:     true,
			wantPath:   "assets/logo.png",
			wantBinary: true,
		},
		{
			name:   "no pipe separator",
			line:   "this line has no pipe",
			wantOK: false,
		},
		{
			name:   "summary line skipped",
			line:   " 3 files changed, 10 insertions(+), 5 deletions(-)",
			wantOK: false, // parseDiffStatLine should not match summaries (no " | ")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, ok := parseDiffStatLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("parseDiffStatLine(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
			}
			if !ok {
				return
			}

			if fs.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", fs.Path, tt.wantPath)
			}
			if fs.IsBinary != tt.wantBinary {
				t.Errorf("IsBinary = %v, want %v", fs.IsBinary, tt.wantBinary)
			}
			if !tt.wantBinary {
				if fs.Added != tt.wantAdded {
					t.Errorf("Added = %d, want %d", fs.Added, tt.wantAdded)
				}
				if fs.Removed != tt.wantRemoved {
					t.Errorf("Removed = %d, want %d", fs.Removed, tt.wantRemoved)
				}
			}
		})
	}
}

func TestParseDiffStatLine_EmptyPath(t *testing.T) {
	// A line with empty path part should return ok=false.
	_, ok := parseDiffStatLine(" | 5 +++++")
	if ok {
		t.Error("expected ok=false for empty path")
	}
}
