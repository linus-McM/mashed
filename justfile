# Repo name used as tmux session prefix (auto-detected from directory name)
repo := `basename $(git rev-parse --show-toplevel 2>/dev/null || basename $PWD)`

# Start a Gemini session in YOLO approval mode
g_session:
    @gemini --approval-mode yolo

# Start a Claude session (default model) in tmux
c_session:
    @tmux new-session -d -s {{repo}}-opus-c_session 'claude --dangerously-skip-permissions' && tmux attach -t {{repo}}-opus-c_session

# Start Claude with Sonnet model in tmux
sonnet:
    @tmux new-session -d -s {{repo}}-sonnet 'claude --dangerously-skip-permissions --model "sonnet"' && tmux attach -t {{repo}}-sonnet

# Start Claude with Opus model in tmux
opus:
    @tmux new-session -d -s {{repo}}-opus 'claude --dangerously-skip-permissions' && tmux attach -t {{repo}}-opus

# Start Claude with Haiku model in tmux
haiku:
    @tmux new-session -d -s {{repo}}-haiku 'claude --dangerously-skip-permissions --model "haiku"' && tmux attach -t {{repo}}-haiku

create-story:
	@tmux new-session -d -s {{repo}}-create-story 'claude --dangerously-skip-permissions --model opus "/bmad-agent-sm  CS"' && tmux attach -t {{repo}}-create-story

validate-create-story: # (internal command to validate story preparation before dev)
	@tmux new-session -d -s {{repo}}-validate-story 'claude --dangerously-skip-permissions --model opus "/bmad-agent-sm  VCS"' && tmux attach -t {{repo}}-validate-story

dev-story:
	@tmux new-session -d -s {{repo}}-dev-story 'claude --dangerously-skip-permissions --model opus "/bmad-agent-dev DS"' && tmux attach -t {{repo}}-dev-story

code-review:
	@tmux new-session -d -s {{repo}}-code-review 'claude --dangerously-skip-permissions --model opus "/bmad-agent-dev CR"' && tmux attach -t {{repo}}-code-review

# Launch Mashed in dev mode (hot reload)
dev:
    PATH="$HOME/go/bin:$PATH" wails dev

# Build and launch Mashed
run: build
    open build/bin/conductor.app

# Build Mashed production binary
build:
    cd frontend && npm install && cd ..
    PATH="$HOME/go/bin:$PATH" wails build

# Run Go tests
test:
    go test ./internal/... -count=1


repomixer:
    repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml 

