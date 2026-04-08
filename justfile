# Repo name used as tmux session prefix (auto-detected from directory name)
repo := `basename $(git rev-parse --show-toplevel 2>/dev/null || basename $PWD)`
# Random 4-digit suffix to allow multiple sessions of the same type
rand := `printf '%04d' $((RANDOM % 10000))`

# Start a Gemini session in YOLO approval mode
g_session:
    @gemini --approval-mode yolo

# Start a Claude session (default model) in tmux
c_session:
    @claude --dangerously-skip-permissions

# Start Claude with Sonnet model in tmux
sonnet:
    @tmux new-session -d -s {{repo}}-sonnet-{{rand}} 'claude --dangerously-skip-permissions --model "sonnet"' && tmux attach -t {{repo}}-sonnet-{{rand}}

# Start Claude with Opus model in tmux
opus:
    @tmux new-session -d -s {{repo}}-opus-{{rand}} 'claude --dangerously-skip-permissions' && tmux attach -t {{repo}}-opus-{{rand}}

# Start Claude with Haiku model in tmux
haiku:
    @tmux new-session -d -s {{repo}}-haiku-{{rand}} 'claude --dangerously-skip-permissions --model "haiku"' && tmux attach -t {{repo}}-haiku-{{rand}}
dev:
    PATH="$HOME/go/bin:$PATH" wails dev

# Build and launch Mashed
run: build
    open build/bin/mashed.app

# Build Mashed production binary
build:
    cd frontend && npm install && cd ..
    PATH="$HOME/go/bin:$PATH" wails build
    cp -r fonts build/bin/mashed.app/Contents/Resources/fonts

# Run Go tests
test:
    go test ./internal/... -count=1


repomixer:
    repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml
    repomix --remote https://github.com/xyflow/xyflow  --compress -o ./docs/repomixer/xyflow/xyflow.xml --style xml
    repomix --remote https://github.com/bmad-code-org/BMAD-METHOD  --compress -o ./docs/repomixer/bmad-method/bmad-method.xml --style xml


