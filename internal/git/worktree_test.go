package git

import (
	"testing"
)

func TestSplitWorktreeBlocks_Empty(t *testing.T) {
	blocks := splitWorktreeBlocks("")
	if len(blocks) != 0 {
		t.Errorf("splitWorktreeBlocks(\"\") returned %d blocks, want 0", len(blocks))
	}
}

func TestSplitWorktreeBlocks_WhitespaceOnly(t *testing.T) {
	blocks := splitWorktreeBlocks("\n\n\n")
	if len(blocks) != 0 {
		t.Errorf("splitWorktreeBlocks(whitespace) returned %d blocks, want 0", len(blocks))
	}
}

func TestParseWorktreeBlock_Porcelain(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantCount  int
		wantPaths  []string
		wantBranch []string
	}{
		{
			name: "single worktree with branch",
			raw: `worktree /Users/dev/project
HEAD abc123def456
branch refs/heads/main
`,
			wantCount:  1,
			wantPaths:  []string{"/Users/dev/project"},
			wantBranch: []string{"main"},
		},
		{
			name: "two worktrees",
			raw: `worktree /Users/dev/project
HEAD abc123def456
branch refs/heads/main

worktree /Users/dev/project/.claude/worktrees/agent-abc
HEAD 789def012345
branch refs/heads/feature-branch
`,
			wantCount:  2,
			wantPaths:  []string{"/Users/dev/project", "/Users/dev/project/.claude/worktrees/agent-abc"},
			wantBranch: []string{"main", "feature-branch"},
		},
		{
			name: "detached HEAD (no branch line)",
			raw: `worktree /Users/dev/project
HEAD abc123def456
detached
`,
			wantCount:  1,
			wantPaths:  []string{"/Users/dev/project"},
			wantBranch: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := splitWorktreeBlocks(tt.raw)
			if len(blocks) != tt.wantCount {
				t.Fatalf("got %d blocks, want %d", len(blocks), tt.wantCount)
			}

			for i, block := range blocks {
				wt, ok := parseWorktreeBlock(block)
				if !ok {
					t.Fatalf("block %d: parseWorktreeBlock returned ok=false", i)
				}
				if wt.Path != tt.wantPaths[i] {
					t.Errorf("block %d: Path = %q, want %q", i, wt.Path, tt.wantPaths[i])
				}
				if wt.Branch != tt.wantBranch[i] {
					t.Errorf("block %d: Branch = %q, want %q", i, wt.Branch, tt.wantBranch[i])
				}
			}
		})
	}
}

func TestParseWorktreeBlock_NoWorktreeLine(t *testing.T) {
	// A block without "worktree " prefix should return ok=false.
	block := []string{"HEAD abc123", "branch refs/heads/main"}
	_, ok := parseWorktreeBlock(block)
	if ok {
		t.Error("expected ok=false for block without 'worktree' line")
	}
}
