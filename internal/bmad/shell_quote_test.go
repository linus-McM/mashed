package bmad

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapInBashExec(t *testing.T) {
	tests := []struct {
		name  string
		inner string
		want  string
	}{
		{
			name:  "plain command",
			inner: `claude --model opus`,
			want:  `bash -c 'claude --model opus; exec bash'`,
		},
		{
			name:  "single quote inside",
			inner: `echo it's ready`,
			want:  `bash -c 'echo it'\''s ready; exec bash'`,
		},
		{
			name:  "backtick inside",
			inner: "echo `date`",
			want:  "bash -c 'echo `date`; exec bash'",
		},
		{
			name:  "dollar sign inside",
			inner: `echo $HOME`,
			want:  `bash -c 'echo $HOME; exec bash'`,
		},
		{
			name:  "mixed special characters",
			inner: "echo it's `date` $HOME",
			want:  "bash -c 'echo it'\\''s `date` $HOME; exec bash'",
		},
		{
			name:  "multiple single quotes",
			inner: `it's a 'test' here`,
			want:  `bash -c 'it'\''s a '\''test'\'' here; exec bash'`,
		},
		{
			name:  "realistic claude command",
			inner: `claude --dangerously-skip-permissions --model sonnet "use bmad-brainstorming context: file at '/tmp/repo/plan.md'"`,
			want:  `bash -c 'claude --dangerously-skip-permissions --model sonnet "use bmad-brainstorming context: file at '\''/tmp/repo/plan.md'\''"; exec bash'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInBashExec(tt.inner)
			assert.Equal(t, tt.want, got)
		})
	}
}
