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


# List all tmux sessions related to this repo, grouped by parent/child
sessions:
    #!/usr/bin/env bash
    repo=$(basename "$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")")
    echo "tmux sessions for: $repo"
    echo "════════════════════════════════════════════════"
    # Parent sessions: match repo name prefix
    parents=$(tmux list-sessions -F '#{session_name}|#{session_created}|#{session_attached}|#{session_windows}' 2>/dev/null | grep "^${repo}" || true)
    # Child sessions: common agent prefixes (surfseer, bmad, clawteam, term)
    children=$(tmux list-sessions -F '#{session_name}|#{session_created}|#{session_attached}|#{session_windows}' 2>/dev/null | grep -E "^(surfseer|bmad|clawteam|term)-" || true)
    if [ -z "$parents" ] && [ -z "$children" ]; then
        echo "  No sessions found."
        exit 0
    fi
    if [ -n "$parents" ]; then
        echo ""
        echo "  Parent sessions (${repo}-*):"
        echo "$parents" | while IFS='|' read -r name created attached windows; do
            ts=$(date -r "$created" '+%H:%M' 2>/dev/null || echo "$created")
            status=""
            [ "$attached" = "1" ] && status=" ◀ attached"
            printf "    %-40s %s  (%s win)%s\n" "$name" "$ts" "$windows" "$status"
        done
    fi
    if [ -n "$children" ]; then
        echo ""
        echo "  Agent sessions (surfseer-*, bmad-*, clawteam-*, term-*):"
        echo "$children" | while IFS='|' read -r name created attached windows; do
            ts=$(date -r "$created" '+%H:%M' 2>/dev/null || echo "$created")
            status=""
            [ "$attached" = "1" ] && status=" ◀ attached"
            printf "    %-40s %s  (%s win)%s\n" "$name" "$ts" "$windows" "$status"
        done
    fi
    echo ""
    p_count=0; [ -n "$parents" ] && p_count=$(echo "$parents" | wc -l | tr -d ' ')
    c_count=0; [ -n "$children" ] && c_count=$(echo "$children" | wc -l | tr -d ' ')
    total=$((p_count + c_count))
    echo "  Total: $total sessions"

# Attach to all repo-related tmux sessions in a single tmux window (switch with prefix+s)
attach-all:
    #!/usr/bin/env bash
    repo=$(basename "$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")")
    sessions=$(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep -E "^(${repo}|surfseer|bmad|clawteam|term)-" || true)
    if [ -z "$sessions" ]; then
        echo "No repo sessions found."
        exit 0
    fi
    count=$(echo "$sessions" | wc -l | tr -d ' ')
    first=$(echo "$sessions" | head -1)
    echo "Found $count sessions. Attaching to: $first"
    echo "Use prefix+s to switch between sessions, or prefix+( / prefix+) to cycle."
    tmux attach -t "$first"

# Check for Claude Code updates and install if available
update:
    @claude --version && claude update

# Alias for update
upgrade: update

repomixer:
    repomix --remote https://github.com/peteromallet/desloppify --compress -o ./docs/repomixer/desloppify/desloppify.xml --style xml
    repomix --remote https://github.com/wailsapp/wails  --compress -o ./docs/repomixer/wails/wails.xml --style xml
    repomix --remote https://github.com/xyflow/xyflow  --compress -o ./docs/repomixer/xyflow/xyflow.xml --style xml
    repomix --remote https://github.com/bmad-code-org/BMAD-METHOD  --compress -o ./docs/repomixer/bmad-method/bmad-method.xml --style xml


